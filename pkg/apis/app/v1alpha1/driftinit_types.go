package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type ComponentType string

const (
	ZOOKEEPER ComponentType = "ZooKeeper"
)

type PvcInfo struct {
	StorageClass string                       `json:"storageClass"`
	VolumeMode   *corev1.PersistentVolumeMode `json:"volumeMode"`
	Storage      string                       `json:"storage"`
}

type DriftInitSpec struct {
	NameSpace   string                    `json:"namespace"`     //集群所在的命名空间
	Components  []ComponentType           `json:"components"`    //需要安装哪些组件
	CurrentPath string                    `json:"currentPath"`   //当前初始化过程进行到的路径
	Active      int32                     `json:"active"`        //当前初始化过程处在第几步
	Complete    bool                      `json:"complete"`      //初始化是否完成
	Pvc         map[ComponentType]PvcInfo `json:"pvc,omitempty"` //各组件用到的 pvc 信息
}

type DriftInitStatus struct {
}

type DriftInit struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DriftInitSpec   `json:"spec,omitempty"`
	Status DriftInitStatus `json:"status,omitempty"`
}

type DriftInitList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DriftInit `json:"items"`
}

func (in *DriftInitList) DeepCopyObject() runtime.Object {
	out := DriftInitList{}
	in.DeepCopyInto(&out)

	return &out
}

func (in *DriftInit) DeepCopyObject() runtime.Object {
	out := DriftInit{}
	in.DeepCopyInto(&out)

	return &out
}

func init() {
	SchemeBuilder.Register(&DriftInit{}, &DriftInitList{})
}
