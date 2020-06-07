package zookeeper

import (
	"context"
	corev1 "k8s.io/api/core/v1"
	policy "k8s.io/api/policy/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func DeleteAllZooKeeperResource(c client.Client, request reconcile.Request) {
	deleteClientService(c, request)
	deleteHeadlessService(c, request)
	deletePdb(c, request)
}

func deleteClientService(c client.Client, request reconcile.Request) {
	var clientService = &corev1.Service{}
	if err := c.Get(context.TODO(), client.ObjectKey{
		Namespace: request.Namespace,
		Name:      request.Name + "-client-service",
	}, clientService); err != nil {
		log.Info("查找 client service 失败")
		return
	}
	err := c.Delete(context.TODO(), clientService)
	if err != nil {
		log.Info("删除 client service 失败")
		return
	}
	log.Info("删除 client service 成功")
}

func deleteHeadlessService(c client.Client, request reconcile.Request) {
	var headlessService = &corev1.Service{}
	if err := c.Get(context.TODO(), client.ObjectKey{
		Namespace: request.Namespace,
		Name:      request.Name + "-headless-service",
	}, headlessService); err != nil {
		log.Info("查找 headless service 失败")
		return
	}
	err := c.Delete(context.TODO(), headlessService)
	if err != nil {
		log.Info("删除 headless service 失败")
		return
	}
	log.Info("删除 headless service 成功")
}

func deletePdb(c client.Client, request reconcile.Request) {
	var pdb = &policy.PodDisruptionBudget{}
	if err := c.Get(context.TODO(), client.ObjectKey{
		Namespace: request.Namespace,
		Name:      request.Name + "-pdb",
	}, pdb); err != nil {
		log.Info("查找 pdb 失败")
		return
	}
	err := c.Delete(context.TODO(), pdb)
	if err != nil {
		log.Info("删除 pdb 失败")
		return
	}
	log.Info("删除 pdb 成功")
}
