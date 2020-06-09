package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type KafkaSpec struct {
	Namespace        string `json:"namespace"`
	Size             int32  `json:"size"`
	MinSize          int32  `json:"minSize"`
	ClientPort       int32  `json:"clientPort"`
	DataDir          string `json:"dataDir"`
	ZooKeeperAddress string `json:"zookeeperAddress"`
	StorageClass     string `json:"storageClass,omitempty"`
	Storage          string `json:"storage,omitempty"`
}

type KafkaStatus struct{}

type Kafka struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KafkaSpec   `json:"spec,omitempty"`
	Status KafkaStatus `json:"status,omitempty"`
}

type KafkaList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Kafka `json:"items"`
}

func (in *KafkaList) DeepCopyObject() runtime.Object {
	return in.DeepCopy()
}

func (in *Kafka) DeepCopyObject() runtime.Object {
	return in.DeepCopy()
}

func init() {
	SchemeBuilder.Register(&Kafka{}, &KafkaList{})
}
