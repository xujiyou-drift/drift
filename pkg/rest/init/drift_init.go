package init

import (
	"context"
	"github.com/gin-gonic/gin"
	appv1alpha1 "github.com/xujiyou-drift/drift/pkg/apis/app/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"log"
)

const DriftName = "drift-instance"

type PvcRecord struct {
	Pvc         map[appv1alpha1.ComponentType]appv1alpha1.PvcInfo `json:"pvc"`
	CurrentPath string                                            `json:"currentPath"`
	Active      int32                                             `json:"active"`
}

func FindDriftInitCr(c *gin.Context) {
	driftInitInstance := &appv1alpha1.DriftInit{}
	err := Mgr.GetClient().Get(context.TODO(), types.NamespacedName{Name: DriftName, Namespace: ""}, driftInitInstance)
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

	namespaceErr := Mgr.GetClient().Create(context.TODO(), &namespace)
	if namespaceErr != nil && !errors.IsAlreadyExists(namespaceErr) {
		log.Println("创建 Namespace 失败：", namespaceErr)
		c.JSON(200, gin.H{"code": 1, "errMsg": "创建 Namespace 失败"})
		return
	}

	initErr := Mgr.GetClient().Create(context.TODO(), &driftInit)
	if initErr != nil && errors.IsAlreadyExists(initErr) {
		_ = Mgr.GetClient().Get(context.TODO(), types.NamespacedName{Name: driftInit.Name}, &driftInit)
		driftInit.Spec = driftInitSpec
		initErr = Mgr.GetClient().Update(context.TODO(), &driftInit)
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
	err = Mgr.GetClient().Get(context.TODO(), types.NamespacedName{Name: DriftName}, &driftInit)
	if err != nil {
		c.JSON(200, gin.H{"code": 2, "errMsg": "数据查找失败"})
		return
	}

	driftInit.Spec.CurrentPath = pvcRecord.CurrentPath
	driftInit.Spec.Active = pvcRecord.Active

	if pvcRecord.Pvc != nil {
		driftInit.Spec.Pvc = pvcRecord.Pvc
	}

	err = Mgr.GetClient().Update(context.TODO(), &driftInit)
	if err != nil {
		c.JSON(200, gin.H{"code": 3, "errMsg": "数据更新失败"})
		return
	}
	c.JSON(200, gin.H{"code": 0, "errMsg": "更新成功"})
}

func Complete(c *gin.Context) {
	var driftInit appv1alpha1.DriftInit
	err := Mgr.GetClient().Get(context.TODO(), types.NamespacedName{Name: DriftName}, &driftInit)
	if err != nil {
		c.JSON(200, gin.H{"code": 2, "errMsg": "数据查找失败"})
		return
	}

	driftInit.Spec.Complete = true
	err = Mgr.GetClient().Update(context.TODO(), &driftInit)
	if err != nil {
		c.JSON(200, gin.H{"code": 3, "errMsg": "数据更新失败"})
		return
	}
	c.JSON(200, gin.H{"code": 0, "errMsg": "更新成功"})
}
