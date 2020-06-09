package kafka

import (
	appv1alpha1 "github.com/xujiyou-drift/drift/pkg/apis/app/v1alpha1"
	policy "k8s.io/api/policy/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func NewPodDisruptionBudget(kafka *appv1alpha1.Kafka) *policy.PodDisruptionBudget {
	labels := map[string]string{
		"app": kafka.Name,
	}
	return &policy.PodDisruptionBudget{
		ObjectMeta: metav1.ObjectMeta{
			Name:      kafka.Name + "-pdb",
			Namespace: kafka.Namespace,
			Labels:    labels,
		},
		Spec: policy.PodDisruptionBudgetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			MinAvailable: &intstr.IntOrString{
				Type:   intstr.Int,
				IntVal: kafka.Spec.MinSize,
			},
		},
	}
}
