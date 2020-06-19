package kafka

import (
	"context"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/types"
	"reflect"
	"time"

	appv1alpha1 "github.com/xujiyou-drift/drift/pkg/apis/app/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_kafka")

func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileKafka{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

func add(mgr manager.Manager, r reconcile.Reconciler) error {
	c, err := controller.New("kafka-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &appv1alpha1.Kafka{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &appv1alpha1.Kafka{},
	})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileKafka{}

type ReconcileKafka struct {
	client client.Client
	scheme *runtime.Scheme
}

func (r *ReconcileKafka) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling Kafka")

	kafkaInstance := &appv1alpha1.Kafka{}
	err := r.client.Get(context.TODO(), request.NamespacedName, kafkaInstance)
	if err != nil {
		if errors.IsNotFound(err) {
			DeleteAllKafkaResource(r.client, request)
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}

	clientService := NewClientService(kafkaInstance)
	headlessService := NewHeadlessService(kafkaInstance)
	podDisruptionBudget := NewPodDisruptionBudget(kafkaInstance)
	statefulSet := NewStatefulSet(kafkaInstance)

	if err := controllerutil.SetControllerReference(kafkaInstance, statefulSet, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	result, err := CreateOrUpdateClientService(context.TODO(), r.client, clientService)
	if err != nil {
		reqLogger.Error(err, "Failed to create or update client service")
		return reconcile.Result{}, err
	} else {
		reqLogger.Info("Create or update kafka client service result", "result", string(result))
	}

	result, err = CreateOrUpdateHeadlessService(context.TODO(), r.client, headlessService)
	if err != nil {
		reqLogger.Error(err, "Failed to create or update headless service")
		return reconcile.Result{}, err
	} else {
		reqLogger.Info("Create or update kafka headless service result", "result", string(result))
	}

	result, err = CreateOrUpdatePodDisruptionBudget(context.TODO(), r.client, podDisruptionBudget)
	if err != nil {
		reqLogger.Error(err, "Failed to create or update pod disruption budget")
		return reconcile.Result{}, err
	} else {
		reqLogger.Info("Create or update kafka pod disruption budget result", "result", string(result))
	}

	found := &appsv1.StatefulSet{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: kafkaInstance.Name, Namespace: kafkaInstance.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new StatefulSet", "StatefulSet.Namespace", kafkaInstance.Namespace, "StatefulSet.Name", kafkaInstance.Name)
		err = r.client.Create(context.TODO(), statefulSet)
		if err != nil {
			reqLogger.Error(err, "Failed to create new StatefulSet", "StatefulSet.Namespace", kafkaInstance.Namespace, "StatefulSet.Name", kafkaInstance.Name)
			return reconcile.Result{}, err
		}
		return reconcile.Result{Requeue: true, RequeueAfter: time.Second * 20}, nil
	} else if err != nil {
		reqLogger.Error(err, "Failed to get StatefulSet")
		return reconcile.Result{}, err
	}

	result, err = UpdateStatefulSet(context.TODO(), r.client, statefulSet)
	if err != nil {
		reqLogger.Error(err, "Failed to create or update stateful set")
		return reconcile.Result{}, err
	} else {
		reqLogger.Info("Update a Kafka StatefulSet", "StatefulSet.Namespace", kafkaInstance.Namespace, "StatefulSet.Name", kafkaInstance.Name)
	}

	podList := &corev1.PodList{}
	listOpts := []client.ListOption{
		client.InNamespace(kafkaInstance.Namespace),
		client.MatchingLabels(
			map[string]string{
				"app": kafkaInstance.Name,
			},
		),
	}
	if err = r.client.List(context.TODO(), podList, listOpts...); err != nil {
		reqLogger.Error(err, "Failed to list pods", "Kafka.Namespace", kafkaInstance.Namespace, "Kafka.Name", kafkaInstance.Name)
		return reconcile.Result{}, err
	}

	if podList.Items == nil || len(podList.Items) == 0 {
		return reconcile.Result{Requeue: true, RequeueAfter: time.Second * 20}, nil
	}

	podNames := getPodNames(podList.Items)

	// Update status.Nodes if needed
	if !reflect.DeepEqual(podNames, kafkaInstance.Status.Nodes) {
		kafkaInstance.Status.Nodes = podNames
		err := r.client.Status().Update(context.TODO(), kafkaInstance)
		if err != nil {
			reqLogger.Error(err, "Failed to update Kafka status")
			return reconcile.Result{}, err
		}

		for _, pod := range podList.Items {
			pod.Labels["pod-name"] = pod.Name
			_ = r.client.Update(context.TODO(), &pod)
			var podService = NewPodService(kafkaInstance, pod.Name)
			_ = r.client.Create(context.TODO(), podService)
		}
	}

	return reconcile.Result{}, nil
}

func getPodNames(pods []corev1.Pod) []string {
	var podNames []string
	for _, pod := range pods {
		podNames = append(podNames, pod.Name)

	}
	return podNames
}
