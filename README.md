# rabbitmq-operator

Kubernetes operator for RabbitMQ

Sample CRD:
```
---
apiVersion: rabbitmq.improvado.io/v1
kind: Rabbitmq
metadata:
  name: imp20rabbit
spec:
  replicas: 2
  image:
    name: rabbitmq
    tag: 3-alpine
  #secret_credentials: rabbit-users
  #secret_service_account: rabbit-service
  memory_high_watermark: 256M
  k8s_host: "kubernetes.default.svc.cluster.imp"
  k8s_addrtype: hostname
  cluster_node_cleanup_interval: 10
  volume_size: 1Gi
#  policies:
#    - name: ha-three
#      vhost: "rabbit"
#      pattern: ".*"
#      definition:
#        ha-mode: "exactly"
#        ha-params: 3
#        ha-sync-mode: "automatic"
#      priority: 0
#      apply-to: all
```
