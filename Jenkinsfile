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

    
    stage("Test") {
        sh "kubectl get pod -n default"
    }
}
