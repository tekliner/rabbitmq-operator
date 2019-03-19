operator-sdk generate k8s
eval $(aws ecr get-login --no-include-email --region us-east-1 --profile staging | sed 's|https://||')
operator-sdk build 716309063777.dkr.ecr.us-east-1.amazonaws.com/rabbitmq-operator
docker push 716309063777.dkr.ecr.us-east-1.amazonaws.com/rabbitmq-operator
kubectl delete po -l app=rabbitmq-operator
