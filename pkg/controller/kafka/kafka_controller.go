package kafka

import (
	"context"

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
	reqLogger.Info("Create or update client service result", "result", string(result))
	if err != nil {
		reqLogger.Error(err, "Failed to create or update client service")
		return reconcile.Result{}, err
	}

	result, err = CreateOrUpdateHeadlessService(context.TODO(), r.client, headlessService)
	reqLogger.Info("Create or update headless service result", "result", string(result))
	if err != nil {
		reqLogger.Error(err, "Failed to create or update headless service")
		return reconcile.Result{}, err
	}

	result, err = CreateOrUpdatePodDisruptionBudget(context.TODO(), r.client, podDisruptionBudget)
	reqLogger.Info("Create or update pod disruption budget result", "result", string(result))
	if err != nil {
		reqLogger.Error(err, "Failed to create or update pod disruption budget")
		return reconcile.Result{}, err
	}

	result, err = CreateOrUpdateStatefulSet(context.TODO(), r.client, statefulSet)
	reqLogger.Info("Create or update stateful set result", "result", string(result))
	if err != nil {
		reqLogger.Error(err, "Failed to create or update stateful set")
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}
