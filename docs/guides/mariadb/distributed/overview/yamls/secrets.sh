# The script returns a kubeconfig for the service account given
# you need to have kubectl on PATH with the context set to the cluster you want to create the config for

# Cosmetics for the created config
firstWorkerSecretName=$1

# cluster name what you given in clusters registration
clusterName=$2

# the Namespace and ServiceAccount name that is used for the config
namespace=$3

# Need to give correct network interface value like ens160, eth0 etc
networkInterface=$4

# kubectl cluster-info of respective worker-cluster
worker_endpoint=$5


######################
# actual script starts
set -o errexit

### Fetch Worker cluster Secrets ###
PROJECT_NAMESPACE=$(kubectl get secrets $firstWorkerSecretName -n $namespace  -o jsonpath={.data.namespace})
CONTROLLER_ENDPOINT=$(kubectl get secrets $firstWorkerSecretName -n $namespace  -o jsonpath={.data.controllerEndpoint})
CA_CRT=$(kubectl get secrets $firstWorkerSecretName -n $namespace  -o jsonpath='{.data.ca\.crt}')
TOKEN=$(kubectl get secrets $firstWorkerSecretName -n $namespace  -o jsonpath={.data.token})

echo "
---
## Base64 encoded secret values from controller cluster
controllerSecret:
  namespace: ${PROJECT_NAMESPACE}
  endpoint: ${CONTROLLER_ENDPOINT}
  ca.crt: ${CA_CRT}
  token: ${TOKEN}
cluster:
  name: ${clusterName}
  endpoint: ${worker_endpoint}
netop:
  networkInterface: ${networkInterface}
"