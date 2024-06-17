# Kubernetes general information

We run our infrastructure on AWS, more specifically on AWS EKS, it is a managed Kubernetes cluster platform where you can deploy your infrastructure and use AWS

#### Get list of pods running
    kubectl get pods

#### Find a bit more about the pod
    kubectl describe pod <pod name>

#### Get the log output
    kubectl logs <pod name> --all-containers
    kubectl logs <pod name> --all-containers --tail 50
    kubectl logs <pod name> --all-containers --tail 50 --follow

#### Get persistent volumes
    kubectl get pvc

#### Deploy a new version of a service
    kubectl apply -f <folder or file.yaml or https://url/yaml.file>

#### Bash into a running instance/pod
    kubectl exec -it <pod name> -- bash

#### Get ingress information
    kubectl get ingress

These will be the basic commands which should keep you going, if you would like to have a more in-depth understanding of kubectl and extending your knowledge please visit https://kubernetes.io/docs/reference/kubectl/cheatsheet/