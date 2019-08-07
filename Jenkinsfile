node {
    checkout scm
    def branch = env.BRANCH_NAME.toLowerCase()
    def registry = "716309063777.dkr.ecr.us-east-1.amazonaws.com"
    def build = env.BUILD_NUMBER
    def image

    stage("Build image") {
        sh 'docker build -t rabbitmq-operator .'
        image = docker.image("rabbitmq-operator")
    }

    stage("Push image") {
        docker.withRegistry("https://"+registry+"/", 'ecr:us-east-1:3c5c323b-afed-4bf0-ae1a-3b19d1c904fe') {
            image.push("${branch}-${build}")
        }
    }

    // check branch and select cluster to deploy
    if (branch == 'master') {
        stage ('Wait for confirmation of build promotion') {
            input message: 'Is this build ready for production?', submitter: 'tekliner'
        }
        stage ('Deploy to production') {
            writeFile file: 'operator.yaml', text: """
---
apiVersion: apps/v1beta1
kind: Deployment
metadata:
  name: rabbitmq-operator
spec:
  replicas: 1
  selector:
    matchLabels:
      app: rabbitmq-operator
  template:
    metadata:
      labels:
        app: rabbitmq-operator
    spec:
      serviceAccountName: rabbitmq-operator
      containers:
        - name: rabbitmq-operator
          image: 716309063777.dkr.ecr.us-east-1.amazonaws.com/rabbitmq-operator:${branch}-${build}
          command:
          - rabbitmq-operator
          imagePullPolicy: Always
          env:
            - name: POD_NAME
              valueFrom:
                fieldRef:
                  fieldPath: metadata.name
            - name: OPERATOR_NAME
              value: "rabbitmq-operator"
            - name: WATCH_NAMESPACE
              value: "messaging"
"""
            archiveArtifacts: 'operator.yaml'
            sh "kubectl apply -f operator.yaml -n messaging"
        }
    } else {
      // deploy PR to sandbox
        stage ('Deploy to sandbox') {
            writeFile file: 'operator.yaml', text: """
---
apiVersion: apps/v1beta1
kind: Deployment
metadata:
  name: rabbitmq-operator
spec:
  replicas: 1
  selector:
    matchLabels:
      app: rabbitmq-operator
  template:
    metadata:
      labels:
        app: rabbitmq-operator
    spec:
      serviceAccountName: rabbitmq-operator
      containers:
        - name: rabbitmq-operator
          image: 716309063777.dkr.ecr.us-east-1.amazonaws.com/rabbitmq-operator:${branch}-${build}
          command:
          - rabbitmq-operator
          imagePullPolicy: Always
          env:
            - name: POD_NAME
              valueFrom:
                fieldRef:
                  fieldPath: metadata.name
            - name: OPERATOR_NAME
              value: "rabbitmq-operator"
            - name: WATCH_NAMESPACE
              value: "rabbitmq-operator-${branch}"
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: rabbitmq-operator
---
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: rabbitmq-operator
subjects:
- apiGroup: rbac.authorization.k8s.io
  kind: User
  name: system:serviceaccount:rabbitmq-operator-${branch}:rabbitmq-operator
roleRef:
  kind: ClusterRole
  name: rabbitmq-operator
  apiGroup: rbac.authorization.k8s.io
---
kind: RoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: rabbitmq-operator
subjects:
- kind: ServiceAccount
  name: rabbitmq-operator
roleRef:
  kind: ClusterRole
  name: rabbitmq-operator
  apiGroup: rbac.authorization.k8s.io
"""
            archiveArtifacts: 'operator.yaml'
            // create separated namespace and deploy operator into it
            sh "HOME=/root;KUBECONFIG=/root/.kube/sandbox.config kubectl create ns rabbitmq-operator-${branch} || true"
            sh "HOME=/root;KUBECONFIG=/root/.kube/sandbox.config kubectl apply -f operator.yaml -n rabbitmq-operator-${branch} || true"
        }
    }
}