package kafka

import (
	appv1alpha1 "github.com/xujiyou-drift/drift/pkg/apis/app/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewStatefulSet(kafka *appv1alpha1.Kafka) *appsv1.StatefulSet {
	labels := map[string]string{
		"app": kafka.Name,
	}
	var statefulSet = &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      kafka.Name,
			Namespace: kafka.Namespace,
			Labels:    labels,
		},
		Spec: appsv1.StatefulSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			ServiceName: kafka.Name + "-headless-service",
			Replicas:    &kafka.Spec.Size,
			UpdateStrategy: appsv1.StatefulSetUpdateStrategy{
				Type: appsv1.RollingUpdateStatefulSetStrategyType,
			},
			PodManagementPolicy: appsv1.ParallelPodManagement, //Kafka 不要求顺序性
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "drift-kafka",
							Image: "registry.prod.bbdops.com/common/drift-kafka:v0.0.22",
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: kafka.Spec.ClientPort,
									Name:          "client",
								},
							},
							Env: []corev1.EnvVar{
								{
									Name:  "EXTERNAL_ADDRESS",
									Value: kafka.Spec.ExternalAddress,
								},
								{
									Name:  "DATA_DIR",
									Value: kafka.Spec.DataDir,
								}, {
									Name:  "ZOOKEEPER_ADDRESS",
									Value: kafka.Spec.ZooKeeperAddress,
								},
							},
						},
					},
				},
			},
		},
	}

	if kafka.Spec.StorageClass != "" && kafka.Spec.Storage != "" {
		quantity, _ := resource.ParseQuantity(kafka.Spec.Storage)
		statefulSet.Spec.VolumeClaimTemplates = []corev1.PersistentVolumeClaim{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "data-dir",
				},
				Spec: corev1.PersistentVolumeClaimSpec{
					AccessModes: []corev1.PersistentVolumeAccessMode{
						corev1.ReadWriteOnce,
					},
					StorageClassName: &kafka.Spec.StorageClass,
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
				MountPath: kafka.Spec.DataDir,
				SubPath:   "data",
			},
		}
	}

	return statefulSet
}
