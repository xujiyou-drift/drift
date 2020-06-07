package first

import (
	"context"
	"github.com/gin-gonic/gin"
	appv1alpha1 "github.com/xujiyou-drift/drift/pkg/apis/app/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"log"
)

const ZooKeeperName = "zookeeper-cluster"

func CreateZooKeeper(c *gin.Context) {
	var zookeeperSpec appv1alpha1.ZooKeeperSpec
	err := c.BindJSON(&zookeeperSpec)
	log.Println(zookeeperSpec)

	if err != nil {
		c.JSON(200, gin.H{"code": 1, "errMsg": "数据解析失败"})
		return
	}

	var zookeeperInstance = appv1alpha1.ZooKeeper{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ZooKeeperName,
			Namespace: zookeeperSpec.Namespace,
		},
		Spec: zookeeperSpec,
		Status: appv1alpha1.ZooKeeperStatus{
			Nodes: []string{
				"",
			},
		},
	}

	err = Mgr.GetClient().Create(context.TODO(), &zookeeperInstance)
	if err != nil && errors.IsAlreadyExists(err) {
		_ = Mgr.GetClient().Get(context.TODO(), types.NamespacedName{Name: zookeeperInstance.Name}, &zookeeperInstance)
		zookeeperInstance.Spec = zookeeperSpec
		err = Mgr.GetClient().Update(context.TODO(), &zookeeperInstance)
		if err != nil {
			log.Println("更新 ZooKeeper 失败：", err)
			c.JSON(200, gin.H{"code": 1, "errMsg": "更新 ZooKeeper 失败"})
			return
		}
	} else if err != nil {
		log.Println("创建 ZooKeeper 失败：", err)
		c.JSON(200, gin.H{"code": 1, "errMsg": "创建 ZooKeeper 失败"})
		return
	}

	var driftInit appv1alpha1.DriftInit
	err = Mgr.GetClient().Get(context.TODO(), types.NamespacedName{Name: DriftName}, &driftInit)
	if err != nil {
		c.JSON(200, gin.H{"code": 2, "errMsg": "数据查找失败"})
		return
	}

	driftInit.Spec.CurrentPath = "/first/complete"
	driftInit.Spec.Active = 4

	err = Mgr.GetClient().Update(context.TODO(), &driftInit)
	if err != nil {
		c.JSON(200, gin.H{"code": 3, "errMsg": "数据更新失败"})
		return
	}

	c.JSON(200, gin.H{"code": 0, "errMsg": "创建成功"})

}
