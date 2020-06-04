package zookeeper

import (
	"context"
	appv1alpha1 "github.com/xujiyou-drift/drift/pkg/apis/app/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_zookeeper")

func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileZooKeeper{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

func add(mgr manager.Manager, r reconcile.Reconciler) error {
	c, err := controller.New("zookeeper-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &appv1alpha1.ZooKeeper{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &appv1alpha1.ZooKeeper{},
	})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileZooKeeper{}

type ReconcileZooKeeper struct {
	client client.Client
	scheme *runtime.Scheme
}

func (r *ReconcileZooKeeper) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling ZooKeeper")

	zooKeeperInstance := &appv1alpha1.ZooKeeper{}
	err := r.client.Get(context.TODO(), request.NamespacedName, zooKeeperInstance)
	if err != nil {
		if errors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}

	clientService := NewClientService(zooKeeperInstance)
	headlessService := NewHeadlessService(zooKeeperInstance)
	podDisruptionBudget := NewPodDisruptionBudget(zooKeeperInstance)
	statefulSet := NewStatefulSet(zooKeeperInstance)

	if err := controllerutil.SetControllerReference(zooKeeperInstance, statefulSet, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	found := &appsv1.StatefulSet{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: statefulSet.Name, Namespace: statefulSet.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new ZooKeeper Cluster", "StatefulSet.Namespace", statefulSet.Namespace, "StatefulSet.Name", statefulSet.Name)
		err = r.client.Create(context.TODO(), clientService)
		err = r.client.Create(context.TODO(), headlessService)
		err = r.client.Create(context.TODO(), podDisruptionBudget)
		err = r.client.Create(context.TODO(), statefulSet)
		if err != nil {
			return reconcile.Result{}, err
		}

		return reconcile.Result{}, nil
	} else if err != nil {
		reqLogger.Error(err, "Failed to get StatefulSet")
		return reconcile.Result{}, err
	}

	err = r.client.Update(context.TODO(), clientService)
	err = r.client.Update(context.TODO(), headlessService)
	err = r.client.Update(context.TODO(), podDisruptionBudget)
	err = r.client.Update(context.TODO(), statefulSet)
	if err != nil {
		reqLogger.Error(err, "Failed Update ZooKeeper Cluster", "StatefulSet.Namespace", statefulSet.Namespace, "statefulSet.Name", statefulSet.Name)
		return reconcile.Result{}, err
	}

	podList := &corev1.PodList{}
	listOpts := []client.ListOption{
		client.InNamespace(zooKeeperInstance.Namespace),
		client.MatchingLabels(map[string]string{
			"app": zooKeeperInstance.Name,
		}),
	}
	if err = r.client.List(context.TODO(), podList, listOpts...); err != nil {
		reqLogger.Error(err, "Failed to list pods", "ZooKeeper.Namespace", zooKeeperInstance.Namespace, "ZooKeeper.Name", zooKeeperInstance.Name)
		return reconcile.Result{}, err
	}
	podNames := getPodNames(podList.Items)

	if !reflect.DeepEqual(podNames, zooKeeperInstance.Status.Nodes) {
		zooKeeperInstance.Status.Nodes = podNames
		err := r.client.Status().Update(context.TODO(), zooKeeperInstance)
		if err != nil {
			reqLogger.Error(err, "Failed to update ZooKeeper status")
			return reconcile.Result{}, err
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
