package zookeeper

import (
	appv1alpha1 "github.com/xujiyou-drift/drift/pkg/apis/app/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strconv"
)

func NewStatefulSet(zookeeper *appv1alpha1.ZooKeeper) *appsv1.StatefulSet {
	labels := map[string]string{
		"app": zookeeper.Name,
	}
	var statefulSet = &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      zookeeper.Name,
			Namespace: zookeeper.Namespace,
			Labels:    labels,
		},
		Spec: appsv1.StatefulSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			ServiceName: zookeeper.Name + "-headless-service",
			Replicas:    &zookeeper.Spec.Size,
			UpdateStrategy: appsv1.StatefulSetUpdateStrategy{
				Type: appsv1.RollingUpdateStatefulSetStrategyType,
			},
			PodManagementPolicy: appsv1.ParallelPodManagement, //ZooKeeper 不要求顺序性
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "drift-zookeeper",
							Image: "registry.prod.bbdops.com/common/drift-zookeeper:v0.0.14",
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: zookeeper.Spec.ClientPort,
									Name:          "client",
								}, {
									ContainerPort: zookeeper.Spec.MetricsPort,
									Name:          "metrics",
								}, {
									ContainerPort: zookeeper.Spec.ServerPort,
									Name:          "server",
								}, {
									ContainerPort: zookeeper.Spec.LeaderPort,
									Name:          "leader-election",
								},
							},
							Env: []corev1.EnvVar{
								{
									Name:  "SERVER_NUMBER",
									Value: strconv.Itoa(int(zookeeper.Spec.Size)),
								}, {
									Name:  "STATEFUL_SET_NAME",
									Value: zookeeper.Name,
								}, {
									Name:  "SERVER_PORT",
									Value: strconv.Itoa(int(zookeeper.Spec.ServerPort)),
								}, {
									Name:  "ELECTION_PORT",
									Value: strconv.Itoa(int(zookeeper.Spec.LeaderPort)),
								}, {
									Name:  "DATA_DIR",
									Value: zookeeper.Spec.DataDir,
								},
							},
						},
					},
				},
			},
		},
	}

	if zookeeper.Spec.StorageClass != "" && zookeeper.Spec.Storage != "" {
		quantity, _ := resource.ParseQuantity(zookeeper.Spec.Storage)
		statefulSet.Spec.VolumeClaimTemplates = []corev1.PersistentVolumeClaim{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "data-dir",
				},
				Spec: corev1.PersistentVolumeClaimSpec{
					AccessModes: []corev1.PersistentVolumeAccessMode{
						corev1.ReadWriteOnce,
					},
					StorageClassName: &zookeeper.Spec.StorageClass,
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceStorage: quantity,
						},
					},
				},
			},
		}
		statefulSet.Spec.Template.Spec.Containers[0].VolumeMounts = []corev1.VolumeMount{
			{
				Name:      "data-dir",
				MountPath: zookeeper.Spec.DataDir,
				SubPath:   "data",
			},
		}
	}

	return statefulSet
}
