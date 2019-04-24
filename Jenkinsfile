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

    stage ('Run deployments to staging') {
        if (branch == 'master') {
        stage ('Wait for confirmation of build promotion') {
            input message: 'Is this build ready for production?', submitter: 'tekliner'
        }
        stage ('Deploy to production') {
            writeFile file: 'operator.yaml', text: """
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: rabbitmq-operator
spec:
  replicas: 1
  selector:
    matchLabels:
      name: rabbitmq-operator
  template:
    metadata:
      labels:
        name: rabbitmq-operator
    spec:
      serviceAccountName: rabbitmq-operator
      containers:
        - name: rabbitmq-operator
          image: 716309063777.dkr.ecr.us-east-1.amazonaws.com/rabbitmq-operator:${branch}-${build}
          command:
          - rabbitmq-operator
          imagePullPolicy: Always
          env:
            - name: WATCH_NAMESPACE
              value: ""
            - name: POD_NAME
              valueFrom:
                fieldRef:
                  fieldPath: metadata.name
            - name: OPERATOR_NAME
              value: "rabbitmq-operator"
"""
            archiveArtifacts: 'operator.yaml'
            sh "kubectl apply -f operator.yaml -n messaging"
            }
        }
    }
}