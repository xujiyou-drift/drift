package zookeeper

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

type FindInfo struct {
	Namespace string `json:"namespace"`
}

func FindStatus(c *gin.Context) {
	var findInfo FindInfo
	_ = c.BindJSON(&findInfo)

	c.Header("X-Content-Type-Options", "nosniff")
	var labels = map[string]string{
		"app": "zookeeper-cluster",
	}
	var podList = &corev1.PodList{}
	ticker := time.NewTicker(5 * time.Second)
	for _ = range ticker.C {
		err := Mgr.GetClient().List(context.TODO(), podList, client.InNamespace(findInfo.Namespace), client.MatchingLabels(labels))
		if err != nil {
			log.Println(err)
			return
		}
		jsons, _ := json.Marshal(podList)
		_, _ = fmt.Fprintf(c.Writer, "%s", jsons)
		c.Writer.Flush()

		var running = true
		for index := range podList.Items {
			if podList.Items[index].Status.Phase != "Running" {
				running = false
			}
		}
		if running {
			return
		}
	}

}
