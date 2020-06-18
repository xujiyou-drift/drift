package kafka

import (
	appv1alpha1 "github.com/xujiyou-drift/drift/pkg/apis/app/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strconv"
	"strings"
)

func NewPodService(kafka *appv1alpha1.Kafka, podName string) *corev1.Service {
	ss := strings.Split(podName, "-")
	last := ss[len(ss)-1]
	var nodePortStr = "3109" + last
	nodePort, _ := strconv.Atoi(nodePortStr)
	labels := map[string]string{
		"app":                                kafka.Name,
		"statefulset.kubernetes.io/pod-name": podName,
	}
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      podName + "-pod-service",
			Namespace: kafka.Namespace,
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			Selector: labels,
			Ports: []corev1.ServicePort{
				{
					Name:     "client",
					Port:     9094,
					NodePort: int32(nodePort),
				},
			},
			Type: corev1.ServiceTypeNodePort,
		},
	}
}

func NewClientService(kafka *appv1alpha1.Kafka) *corev1.Service {
	labels := map[string]string{
		"app": kafka.Name,
	}
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      kafka.Name + "-client-service",
			Namespace: kafka.Namespace,
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			Selector: labels,
			Ports: []corev1.ServicePort{
				{
					Name: "client",
					Port: kafka.Spec.ClientPort,
				},
			},
		},
	}
}

func NewHeadlessService(kafka *appv1alpha1.Kafka) *corev1.Service {
	labels := map[string]string{
		"app": kafka.Name,
	}
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      kafka.Name + "-headless-service",
			Namespace: kafka.Namespace,
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			Selector: labels,
			Ports: []corev1.ServicePort{
				{
					Name: "server",
					Port: kafka.Spec.ClientPort,
				},
			},
			ClusterIP: "None",
		},
	}
}
