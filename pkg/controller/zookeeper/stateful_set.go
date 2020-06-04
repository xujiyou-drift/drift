package zookeeper

import (
	appv1alpha1 "github.com/xujiyou-drift/drift/pkg/apis/app/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strconv"
)

func NewStatefulSet(zookeeper *appv1alpha1.ZooKeeper) *appsv1.StatefulSet {
	labels := map[string]string{
		"app": zookeeper.Name,
	}
	return &appsv1.StatefulSet{
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
							Image: "registry.cn-chengdu.aliyuncs.com/bbd-image/drift-zookeeper:v0.0.10",
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: zookeeper.Spec.ClientPort,
									Name:          "client",
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
									Value: "3",
								}, {
									Name:  "STATEFUL_SET_NAME",
									Value: zookeeper.Name,
								}, {
									Name:  "SERVER_PORT",
									Value: strconv.Itoa(int(zookeeper.Spec.ServerPort)),
								}, {
									Name:  "ELECTION_PORT",
									Value: strconv.Itoa(int(zookeeper.Spec.LeaderPort)),
								},
							},
						},
					},
					ImagePullSecrets: []corev1.LocalObjectReference{
						{
							Name: "docker-secret",
						},
					},
				},
			},
		},
	}
}
