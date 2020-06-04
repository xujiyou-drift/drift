package rest

import (
	"context"
	"github.com/gin-gonic/gin"
	appv1alpha1 "github.com/xujiyou-drift/drift/pkg/apis/app/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"log"
	"strings"
)

const DriftName = "drift-instance"

type PvcRecord struct {
	Pvc         map[appv1alpha1.ComponentType]appv1alpha1.PvcInfo `json:"pvc"`
	CurrentPath string                                            `json:"currentPath"`
	Active      int32                                             `json:"active"`
}

func FindDriftInitCr(c *gin.Context) {
	driftInitInstance := &appv1alpha1.DriftInit{}
	err := mgr.GetClient().Get(context.TODO(), types.NamespacedName{
		Name:      DriftName,
		Namespace: "",
	}, driftInitInstance)
	if err != nil {
		log.Println(err)
		if errors.IsNotFound(err) {
			c.JSON(200, gin.H{"code": 1, "errMsg": "not found"})
			return
		} else {
			c.JSON(200, gin.H{"code": 2, "errMsg": "查找 DriftInit 错误"})
		}
	} else {
		c.JSON(200, gin.H{"code": 0, "errMsg": "success", "data": driftInitInstance})
	}
}

func CreateDriftInit(c *gin.Context) {
	var driftInitSpec appv1alpha1.DriftInitSpec
	err := c.BindJSON(&driftInitSpec)
	if err != nil {
		c.JSON(200, gin.H{"code": 1, "errMsg": "数据解析失败"})
		return
	}
	var driftInit = appv1alpha1.DriftInit{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "app.drift.com/v1alpha1",
			Kind:       "DriftInit",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: DriftName,
		},
		Spec: driftInitSpec,
	}

	var namespace = corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: driftInitSpec.NameSpace,
		},
	}

	namespaceErr := mgr.GetClient().Create(context.TODO(), &namespace)
	if namespaceErr != nil && !errors.IsAlreadyExists(namespaceErr) {
		log.Println("创建 Namespace 失败：", namespaceErr)
		c.JSON(200, gin.H{"code": 1, "errMsg": "创建 Namespace 失败"})
		return
	}

	initErr := mgr.GetClient().Create(context.TODO(), &driftInit)
	if initErr != nil && errors.IsAlreadyExists(initErr) {
		_ = mgr.GetClient().Get(context.TODO(), types.NamespacedName{Name: driftInit.Name}, &driftInit)
		driftInit.Spec = driftInitSpec
		initErr = mgr.GetClient().Update(context.TODO(), &driftInit)
		if initErr != nil {
			log.Println("更新 DriftInit 失败：", initErr)
			c.JSON(200, gin.H{"code": 1, "errMsg": "更新 DriftInit 失败"})
			return
		}
	} else if initErr != nil {
		log.Println("创建 DriftInit 失败：", initErr)
		c.JSON(200, gin.H{"code": 1, "errMsg": "创建 DriftInit 失败"})
		return
	}

	c.JSON(200, gin.H{"code": 0, "errMsg": "创建成功"})
}

func RecordPvc(c *gin.Context) {
	var pvcRecord PvcRecord
	err := c.BindJSON(&pvcRecord)
	if err != nil {
		c.JSON(200, gin.H{"code": 1, "errMsg": "数据解析失败"})
		return
	}

	var driftInit appv1alpha1.DriftInit
	err = mgr.GetClient().Get(context.TODO(), types.NamespacedName{Name: DriftName}, &driftInit)
	if err != nil {
		c.JSON(200, gin.H{"code": 2, "errMsg": "数据查找失败"})
		return
	}

	driftInit.Spec.CurrentPath = pvcRecord.CurrentPath
	driftInit.Spec.Active = pvcRecord.Active

	if pvcRecord.Pvc != nil {
		driftInit.Spec.Pvc = pvcRecord.Pvc
		for componentType, PvcInfo := range driftInit.Spec.Pvc {
			log.Println("创建PVC...")
			err = createPvc(componentType, PvcInfo, driftInit.Spec.NameSpace)
			if err != nil {
				c.JSON(200, gin.H{"code": 4, "errMsg": "创建PVC失败"})
				log.Println("创建PVC失败", err)
				return
			}
		}
	}

	err = mgr.GetClient().Update(context.TODO(), &driftInit)
	if err != nil {
		c.JSON(200, gin.H{"code": 3, "errMsg": "数据更新失败"})
		return
	}
	c.JSON(200, gin.H{"code": 0, "errMsg": "更新成功"})
}

func createPvc(componentType appv1alpha1.ComponentType, pvcInfo appv1alpha1.PvcInfo, namespace string) error {
	quantity, err := resource.ParseQuantity(pvcInfo.Storage)
	if err != nil {
		log.Println("解析数据失败:", err)
		return err
	}
	var pvcName = strings.ToLower(string(componentType)) + "-pvc"
	var pvc = corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      pvcName,
			Namespace: namespace,
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{
				corev1.ReadWriteOnce,
			},
			VolumeMode: pvcInfo.VolumeMode,
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: quantity,
				},
			},
			StorageClassName: &pvcInfo.StorageClass,
		},
	}

	err = mgr.GetClient().Create(context.TODO(), &pvc)
	if err != nil && errors.IsAlreadyExists(err) {
		return nil
	} else if err != nil {
		return err
	}

	return nil
}
