package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type ZooKeeperSpec struct {
	Namespace    string `json:"namespace"`
	Size         int32  `json:"size"`
	MinSize      int32  `json:"minSize"`
	ClientPort   int32  `json:"clientPort"`
	ServerPort   int32  `json:"serverPort"`
	LeaderPort   int32  `json:"leaderPort"`
	DataDir      string `json:"dataDir"`
	StorageClass string `json:"storageClass,omitempty"`
	Storage      string `json:"storage,omitempty"`
}

type ZooKeeperStatus struct {
	Nodes []string `json:"nodes,omitempty"`
}

type ZooKeeper struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ZooKeeperSpec   `json:"spec,omitempty"`
	Status ZooKeeperStatus `json:"status,omitempty"`
}

type ZooKeeperList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ZooKeeper `json:"items"`
}

func (in *ZooKeeperList) DeepCopyObject() runtime.Object {
	//out := ZooKeeperList{}
	//in.DeepCopyInto(&out)
	//
	//return &out
	return in.DeepCopy()
}

func (in *ZooKeeper) DeepCopyObject() runtime.Object {
	return in.DeepCopy()
}

func init() {
	SchemeBuilder.Register(&ZooKeeper{}, &ZooKeeperList{})
}
