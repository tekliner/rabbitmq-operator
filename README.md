# rabbitmq-operator

Kubernetes operator for RabbitMQ. Code is highly fresh, be patient.

Sample CRD (whole specs see in rabbitmq_types.go):
```
---
apiVersion: rabbitmq.improvado.io/v1
kind: Rabbitmq
metadata:
  name: imp20rabbit
spec:
  replicas: 2
  
  # set affinity and anti-affinity
  affinity:
    podAffinity:
      requiredDuringSchedulingIgnoredDuringExecution:
      - labelSelector:
          matchExpressions:
          - key: security
            operator: In
            values:
            - S1
        topologyKey: failure-domain.beta.kubernetes.io/zonei

  # set rabbitmq docker image, use hub.docker.com or your own
  image:
    name: rabbitmq
    tag: 3-alpine
  
  # use custom names for secrets instead of default based on CRD name
  # default_user, default_password and cookie is generated once at first start
  #secret_credentials: rabbit-users
  #secret_service_account: rabbit-service
  
  # set vm_memory_high_watermark.absolute value
  memory_high_watermark: 256M
  
  # clusterize rabbit
  k8s_host: "kubernetes.default.svc.cluster.imp"
  k8s_addrtype: hostname
  cluster_node_cleanup_interval: 10
  cluster_formation.node_cleanup.only_log_warning: true
  cluster_partition_handling: autoheal

  loopback_users.guest = false

  hipe_compile: false

  # PVC
  volume_size: 1Gi

  # Set custom ENV
  env:
    - name: VARIABLENAME
      value: test

  policies:
    - name: ha-three
      vhost: "rabbit"
      pattern: ".*"
      definition:
        ha-mode: "exactly"
        ha-params: 3
        ha-sync-mode: "automatic"
      priority: 0
      apply-to: all

  plugins:
    - rabbitmq_management_agent
```

Default plugins:

* rabbitmq_consistent_hash_exchange,
* rabbitmq_federation,
* rabbitmq_federation_management,
* rabbitmq_management,
* rabbitmq_peer_discovery_k8s,
* rabbitmq_shovel,
* rabbitmq_shovel_management

In future:
* SSL
* Additional users
* Custom k8s labels
* RabbitMQ limits based on pods limits
