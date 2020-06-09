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

const KafkaName = "kafka-cluster"

func CreateKafka(c *gin.Context) {
	var kafkaSpec appv1alpha1.KafkaSpec
	err := c.BindJSON(&kafkaSpec)
	log.Println(kafkaSpec)
	if err != nil {
		c.JSON(200, gin.H{"code": 1, "errMsg": "数据解析失败"})
		return
	}

	var kafkaInstance = appv1alpha1.Kafka{
		ObjectMeta: metav1.ObjectMeta{
			Name:      KafkaName,
			Namespace: kafkaSpec.Namespace,
		},
		Spec: kafkaSpec,
	}

	err = Mgr.GetClient().Create(context.TODO(), &kafkaInstance)
	if err != nil && errors.IsAlreadyExists(err) {
		_ = Mgr.GetClient().Get(context.TODO(), types.NamespacedName{Name: kafkaInstance.Name, Namespace: kafkaSpec.Namespace}, &kafkaInstance)
		kafkaInstance.Spec = kafkaSpec
		err = Mgr.GetClient().Update(context.TODO(), &kafkaInstance)
		if err != nil {
			log.Println("更新 Kafka 失败：", err)
			c.JSON(200, gin.H{"code": 1, "errMsg": "更新 Kafka 失败"})
			return
		}
	} else if err != nil {
		log.Println("创建 Kafka 失败：", err)
		c.JSON(200, gin.H{"code": 1, "errMsg": "创建 Kafka 失败"})
		return
	}
	c.JSON(200, gin.H{"code": 0, "errMsg": "创建 Kafka 成功"})
}
