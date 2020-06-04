package zookeeper

import (
	appv1alpha1 "github.com/xujiyou-drift/drift/pkg/apis/app/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewClientService(zookeeper *appv1alpha1.ZooKeeper) *corev1.Service {
	labels := map[string]string{
		"app": zookeeper.Name,
	}
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      zookeeper.Name + "-client-service",
			Namespace: zookeeper.Namespace,
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			Selector: labels,
			Ports: []corev1.ServicePort{
				{
					Name: "client",
					Port: zookeeper.Spec.ClientPort,
				},
			},
		},
	}
}

func NewHeadlessService(zookeeper *appv1alpha1.ZooKeeper) *corev1.Service {
	labels := map[string]string{
		"app": zookeeper.Name,
	}
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      zookeeper.Name + "-headless-service",
			Namespace: zookeeper.Namespace,
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			Selector: labels,
			Ports: []corev1.ServicePort{
				{
					Name: "server",
					Port: zookeeper.Spec.ServerPort,
				}, {
					Name: "leader-election",
					Port: zookeeper.Spec.LeaderPort,
				},
			},
			ClusterIP: "None",
		},
	}
}
