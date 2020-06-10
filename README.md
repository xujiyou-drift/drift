# Drift

> 目标是做一个 Kubernetes 之上的 "CDH"
>
> 实际上是一个 Kubernetes Operator
>
> 可以通过界面部署和管理大数据组件
>
> 支持持久化储存

## 支持的组件

- [x] ZooKeeper 3.6.1
- [x] Kafka 2.5.0
- [ ] Yarn
- [ ] HDFS
- [ ] HBase
- [ ] Hive
- [ ] Spark
- [ ] Flink

## Getting Start

前提是有一个 Kubernetes 集群，并且集群中安装有 Ingress Controller。

第一步，创建命名空间：
```bash
kubectl create ns drift
```

第二步，使用 helm 部署 Drift：
```bash
git clone https://github.com/xujiyou-drift/drift-helm-chart.git
cd drift-helm-chart
helm install my-drift ./drift \
 --namespace drift \
 --set ingress.host=drift.test.bbdops.com
```

等待 Pod 创建完成：
```bash
kubectl get pods -n drift --watch
```

第三步，配置 hosts 或 dns，比如我这里在本机配置 hosts：
```
10.28.109.30 drift.test.bbdops.com
```

第四步，浏览器打开界面：http://drift.test.bbdops.com ，用户名及密码为 admin/admin

![login](./images/login.png)

第五步，输入各组件所在的命名空间，和要选择的组件，目前仅支持 ZooKeeper 和 Kafka，ZooKeeper 是必须的：
![select](./images/select.png)

第六步，数据储存类和卷大小，如果集群没有配置储存类，也可以选择跳过：
![pvc](./images/pvc.png)

第七步，配置组件，默认即可：
![config](./images/config.png)

第八步，点击完成：
![complete](./images/complete.png)

等待 Pod 创建完成：
![home](./images/home.png)

或者使用命令观察 Pod 创建过程：
```bash
kubectl get pods -n bigdata --watch
```

创建完成如下图所示：
![pods](./images/pods.png)

## 使用的镜像

使用到的镜像都是定制的镜像，不可以使用其他人做的镜像。
构建镜像的代码在：https://github.com/xujiyou-drift/drift-images

## 前端

前端使用 Vue.JS 构建，代码在：https://github.com/xujiyou-drift/drift-vue

## 卸载

```bash
helm uninstall my-drift ./drift --namespace drift
```