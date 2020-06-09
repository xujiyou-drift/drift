package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	corev1 "k8s.io/api/core/v1"
	"log"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"
)

func FindStatus(c *gin.Context) {
	c.Header("X-Content-Type-Options", "nosniff")
	var labels = map[string]string{
		"app": "kafka-cluster",
	}
	var podList = &corev1.PodList{}
	ticker := time.NewTicker(5 * time.Second)
	for _ = range ticker.C {
		fmt.Println(time.Now())
		err := Mgr.GetClient().List(context.TODO(), podList, client.InNamespace("drift-test"), client.MatchingLabels(labels))
		if err != nil {
			log.Println(err)
			return
		}
		jsons, _ := json.Marshal(podList)
		_, _ = fmt.Fprintf(c.Writer, "%s", jsons)
		c.Writer.Flush()
	}

}
