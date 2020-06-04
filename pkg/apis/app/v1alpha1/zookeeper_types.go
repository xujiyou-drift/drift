package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type ZooKeeperSpec struct {
	Size       int32 `json:"size"`
	MinSize    int32 `json:"minSize"`
	ClientPort int32 `json:"clientPort"`
	ServerPort int32 `json:"serverPort"`
	LeaderPort int32 `json:"leaderPort"`
}

type ZooKeeperStatus struct {
	Nodes []string `json:"nodes"`
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
	out := ZooKeeperList{}
	in.DeepCopyInto(&out)

	return &out
}

func (in *ZooKeeper) DeepCopyObject() runtime.Object {
	out := ZooKeeper{}
	in.DeepCopyInto(&out)

	return &out
}

func init() {
	SchemeBuilder.Register(&ZooKeeper{}, &ZooKeeperList{})
}
