# rabbitmq-operator

Kubernetes operator for RabbitMQ

Sample CRD (whole specs see in rabbitmq_types.go):
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
  
  # do not use default secret names
  #secret_credentials: rabbit-users
  #secret_service_account: rabbit-service
  
  memory_high_watermark: 256M
  
  # clusterize rabbit
  k8s_host: "kubernetes.default.svc.cluster.imp"
  k8s_addrtype: hostname
  cluster_node_cleanup_interval: 10
  
  # PVC
  volume_size: 1Gi
  
  # you know 
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
