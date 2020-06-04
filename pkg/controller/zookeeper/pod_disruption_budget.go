package zookeeper

import (
	appv1alpha1 "github.com/xujiyou-drift/drift/pkg/apis/app/v1alpha1"
	policy "k8s.io/api/policy/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func NewPodDisruptionBudget(zookeeper *appv1alpha1.ZooKeeper) *policy.PodDisruptionBudget {
	labels := map[string]string{
		"app": zookeeper.Name,
	}
	return &policy.PodDisruptionBudget{
		ObjectMeta: metav1.ObjectMeta{
			Name:      zookeeper.Name + "-pdb",
			Namespace: zookeeper.Namespace,
			Labels:    labels,
		},
		Spec: policy.PodDisruptionBudgetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			MinAvailable: &intstr.IntOrString{
				Type:   intstr.Int,
				IntVal: zookeeper.Spec.MinSize,
			},
		},
	}
}
