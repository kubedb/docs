---
title: CLI | KubeDB
menu:
  docs_{{ .version }}:
    identifier: kf-cli-cli
    name: Quickstart
    parent: kf-cli-kafka
    weight: 100
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Manage KubeDB managed Kafka objects using CLIs

## KubeDB CLI

KubeDB comes with its own cli. It is called `kubedb` cli. `kubedb` can be used to manage any KubeDB object. `kubedb` cli also performs various validations to improve ux. To install KubeDB cli on your workstation, follow the steps [here](/docs/setup/README.md).

### How to Create objects

`kubectl create` creates a database CRD object in `default` namespace by default. Following command will create a Kafka object as specified in `kafka.yaml`.

```bash
$ kubectl create -f druid-quickstart.yaml
kafka.kubedb.com/kafka created
```

You can provide namespace as a flag `--namespace`. Provided namespace should match with namespace specified in input file.

```bash
$ kubectl create -f druid-quickstart.yaml --namespace=kube-system
kafka.kubedb.com/kafka created
```

`kubectl create` command also considers `stdin` as input.

```bash
cat druid-quickstart.yaml | kubectl create -f -
```

### How to List Objects

`kubectl get` command allows users to list or find any KubeDB object. To list all Kafka objects in `default` namespace, run the following command:

```bash
$ kubectl get kafka
NAME    TYPE                  VERSION   STATUS   AGE
kafka   kubedb.com/v1alpha2   3.6.1     Ready    36m
```

You can also use short-form (`kf`) for kafka CR.

```bash
$ kubectl get kf
NAME    TYPE                  VERSION   STATUS   AGE
kafka   kubedb.com/v1alpha2   3.6.1     Ready    36m
```

To get YAML of an object, use `--output=yaml` or `-oyaml` flag. Use `-n` flag for referring namespace.

```yaml
$ kubectl get kf kafka -n demo -oyaml
apiVersion: kubedb.com/v1alpha2
kind: Kafka
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: |
      {"apiVersion":"kubedb.com/v1alpha2","kind":"Kafka","metadata":{"annotations":{},"name":"kafka","namespace":"demo"},"spec":{"authSecret":{"name":"kafka-admin-cred"},"enableSSL":true,"healthChecker":{"failureThreshold":3,"periodSeconds":20,"timeoutSeconds":10},"keystoreCredSecret":{"name":"kafka-keystore-cred"},"storageType":"Durable","deletionPolicy":"DoNotTerminate","tls":{"certificates":[{"alias":"server","secretName":"kafka-server-cert"},{"alias":"client","secretName":"kafka-client-cert"}],"issuerRef":{"apiGroup":"cert-manager.io","kind":"Issuer","name":"kafka-ca-issuer"}},"topology":{"broker":{"replicas":3,"resources":{"limits":{"memory":"1Gi"},"requests":{"cpu":"500m","memory":"1Gi"}},"storage":{"accessModes":["ReadWriteOnce"],"resources":{"requests":{"storage":"1Gi"}},"storageClassName":"standard"},"suffix":"broker"},"controller":{"replicas":3,"resources":{"limits":{"memory":"1Gi"},"requests":{"cpu":"500m","memory":"1Gi"}},"storage":{"accessModes":["ReadWriteOnce"],"resources":{"requests":{"storage":"1Gi"}},"storageClassName":"standard"},"suffix":"controller"}},"version":"3.6.1"}}
  creationTimestamp: "2023-03-29T07:01:29Z"
  finalizers:
    - kubedb.com
  generation: 1
  name: kafka
  namespace: demo
  resourceVersion: "570445"
  uid: ed5f6197-0238-4aba-a7d9-7dc771b2564c
spec:
  authSecret:
    name: kafka-admin-cred
  enableSSL: true
  healthChecker:
    failureThreshold: 3
    periodSeconds: 20
    timeoutSeconds: 10
  keystoreCredSecret:
    name: kafka-keystore-cred
  podTemplate:
    controller: {}
    metadata: {}
    spec:
      resources: {}
  storageType: Durable
  deletionPolicy: DoNotTerminate
  tls:
    certificates:
      - alias: server
        secretName: kafka-server-cert
      - alias: client
        secretName: kafka-client-cert
    issuerRef:
      apiGroup: cert-manager.io
      kind: Issuer
      name: kafka-ca-issuer
  topology:
    broker:
      replicas: 3
      resources:
        limits:
          memory: 1Gi
        requests:
          cpu: 500m
          memory: 1Gi
      storage:
        accessModes:
          - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
        storageClassName: standard
      suffix: broker
    controller:
      replicas: 3
      resources:
        limits:
          memory: 1Gi
        requests:
          cpu: 500m
          memory: 1Gi
      storage:
        accessModes:
          - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
        storageClassName: standard
      suffix: controller
  version: 3.6.1
status:
  conditions:
    - lastTransitionTime: "2023-03-29T07:01:29Z"
      message: 'The KubeDB operator has started the provisioning of Kafka: demo/kafka'
      observedGeneration: 1
      reason: DatabaseProvisioningStartedSuccessfully
      status: "True"
      type: ProvisioningStarted
    - lastTransitionTime: "2023-03-29T07:02:46Z"
      message: All desired replicas are ready.
      observedGeneration: 1
      reason: AllReplicasReady
      status: "True"
      type: ReplicaReady
    - lastTransitionTime: "2023-03-29T07:02:37Z"
      message: 'The Kafka: demo/kafka is accepting client requests'
      observedGeneration: 1
      reason: DatabaseAcceptingConnectionRequest
      status: "True"
      type: AcceptingConnection
    - lastTransitionTime: "2023-03-29T07:03:37Z"
      message: 'The Kafka: demo/kafka is ready.'
      observedGeneration: 1
      reason: ReadinessCheckSucceeded
      status: "True"
      type: Ready
    - lastTransitionTime: "2023-03-29T07:03:41Z"
      message: 'The Kafka: demo/kafka is successfully provisioned.'
      observedGeneration: 1
      reason: DatabaseSuccessfullyProvisioned
      status: "True"
      type: Provisioned
  phase: Ready
```

To get JSON of an object, use `--output=json` or `-ojson` flag.

```bash
$ kubectl get kf kafka -n demo -ojson
{
    "apiVersion": "kubedb.com/v1alpha2",
    "kind": "Kafka",
    "metadata": {
        "annotations": {
            "kubectl.kubernetes.io/last-applied-configuration": "{\"apiVersion\":\"kubedb.com/v1alpha2\",\"kind\":\"Kafka\",\"metadata\":{\"annotations\":{},\"name\":\"kafka\",\"namespace\":\"demo\"},\"spec\":{\"authSecret\":{\"name\":\"kafka-admin-cred\"},\"enableSSL\":true,\"healthChecker\":{\"failureThreshold\":3,\"periodSeconds\":20,\"timeoutSeconds\":10},\"keystoreCredSecret\":{\"name\":\"kafka-keystore-cred\"},\"storageType\":\"Durable\",\"deletionPolicy\":\"DoNotTerminate\",\"tls\":{\"certificates\":[{\"alias\":\"server\",\"secretName\":\"kafka-server-cert\"},{\"alias\":\"client\",\"secretName\":\"kafka-client-cert\"}],\"issuerRef\":{\"apiGroup\":\"cert-manager.io\",\"kind\":\"Issuer\",\"name\":\"kafka-ca-issuer\"}},\"topology\":{\"broker\":{\"replicas\":3,\"resources\":{\"limits\":{\"memory\":\"1Gi\"},\"requests\":{\"cpu\":\"500m\",\"memory\":\"1Gi\"}},\"storage\":{\"accessModes\":[\"ReadWriteOnce\"],\"resources\":{\"requests\":{\"storage\":\"1Gi\"}},\"storageClassName\":\"standard\"},\"suffix\":\"broker\"},\"controller\":{\"replicas\":3,\"resources\":{\"limits\":{\"memory\":\"1Gi\"},\"requests\":{\"cpu\":\"500m\",\"memory\":\"1Gi\"}},\"storage\":{\"accessModes\":[\"ReadWriteOnce\"],\"resources\":{\"requests\":{\"storage\":\"1Gi\"}},\"storageClassName\":\"standard\"},\"suffix\":\"controller\"}},\"version\":\"3.6.1\"}}\n"
        },
        "creationTimestamp": "2023-03-29T07:01:29Z",
        "finalizers": [
            "kubedb.com"
        ],
        "generation": 1,
        "name": "kafka",
        "namespace": "demo",
        "resourceVersion": "570445",
        "uid": "ed5f6197-0238-4aba-a7d9-7dc771b2564c"
    },
    "spec": {
        "authSecret": {
            "name": "kafka-admin-cred"
        },
        "enableSSL": true,
        "healthChecker": {
            "failureThreshold": 3,
            "periodSeconds": 20,
            "timeoutSeconds": 10
        },
        "keystoreCredSecret": {
            "name": "kafka-keystore-cred"
        },
        "podTemplate": {
            "controller": {},
            "metadata": {},
            "spec": {
                "resources": {}
            }
        },
        "storageType": "Durable",
        "deletionPolicy": "DoNotTerminate",
        "tls": {
            "certificates": [
                {
                    "alias": "server",
                    "secretName": "kafka-server-cert"
                },
                {
                    "alias": "client",
                    "secretName": "kafka-client-cert"
                }
            ],
            "issuerRef": {
                "apiGroup": "cert-manager.io",
                "kind": "Issuer",
                "name": "kafka-ca-issuer"
            }
        },
        "topology": {
            "broker": {
                "replicas": 3,
                "resources": {
                    "limits": {
                        "memory": "1Gi"
                    },
                    "requests": {
                        "cpu": "500m",
                        "memory": "1Gi"
                    }
                },
                "storage": {
                    "accessModes": [
                        "ReadWriteOnce"
                    ],
                    "resources": {
                        "requests": {
                            "storage": "1Gi"
                        }
                    },
                    "storageClassName": "standard"
                },
                "suffix": "broker"
            },
            "controller": {
                "replicas": 3,
                "resources": {
                    "limits": {
                        "memory": "1Gi"
                    },
                    "requests": {
                        "cpu": "500m",
                        "memory": "1Gi"
                    }
                },
                "storage": {
                    "accessModes": [
                        "ReadWriteOnce"
                    ],
                    "resources": {
                        "requests": {
                            "storage": "1Gi"
                        }
                    },
                    "storageClassName": "standard"
                },
                "suffix": "controller"
            }
        },
        "version": "3.6.1"
    },
    "status": {
        "conditions": [
            {
                "lastTransitionTime": "2023-03-29T07:01:29Z",
                "message": "The KubeDB operator has started the provisioning of Kafka: demo/kafka",
                "observedGeneration": 1,
                "reason": "DatabaseProvisioningStartedSuccessfully",
                "status": "True",
                "type": "ProvisioningStarted"
            },
            {
                "lastTransitionTime": "2023-03-29T07:02:46Z",
                "message": "All desired replicas are ready.",
                "observedGeneration": 1,
                "reason": "AllReplicasReady",
                "status": "True",
                "type": "ReplicaReady"
            },
            {
                "lastTransitionTime": "2023-03-29T07:02:37Z",
                "message": "The Kafka: demo/kafka is accepting client requests",
                "observedGeneration": 1,
                "reason": "DatabaseAcceptingConnectionRequest",
                "status": "True",
                "type": "AcceptingConnection"
            },
            {
                "lastTransitionTime": "2023-03-29T07:03:37Z",
                "message": "The Kafka: demo/kafka is ready.",
                "observedGeneration": 1,
                "reason": "ReadinessCheckSucceeded",
                "status": "True",
                "type": "Ready"
            },
            {
                "lastTransitionTime": "2023-03-29T07:03:41Z",
                "message": "The Kafka: demo/kafka is successfully provisioned.",
                "observedGeneration": 1,
                "reason": "DatabaseSuccessfullyProvisioned",
                "status": "True",
                "type": "Provisioned"
            }
        ],
        "phase": "Ready"
    }
}
```

To list all KubeDB objects managed by KubeDB including secrets, use following command:

```bash
$ kubectl get all,secret -A -l app.kubernetes.io/managed-by=kubedb.com -owide
NAMESPACE   NAME                     READY   STATUS    RESTARTS      AGE   IP            NODE                 NOMINATED NODE   READINESS GATES
demo        pod/kafka-broker-0       1/1     Running   0             45m   10.244.0.49   kind-control-plane   <none>           <none>
demo        pod/kafka-broker-1       1/1     Running   0             45m   10.244.0.53   kind-control-plane   <none>           <none>
demo        pod/kafka-broker-2       1/1     Running   0             45m   10.244.0.57   kind-control-plane   <none>           <none>
demo        pod/kafka-controller-0   1/1     Running   0             45m   10.244.0.51   kind-control-plane   <none>           <none>
demo        pod/kafka-controller-1   1/1     Running   0             45m   10.244.0.55   kind-control-plane   <none>           <none>
demo        pod/kafka-controller-2   1/1     Running   0             45m   10.244.0.58   kind-control-plane   <none>           <none>

NAMESPACE   NAME                       TYPE        CLUSTER-IP   EXTERNAL-IP   PORT(S)              AGE   SELECTOR
demo        service/kafka-broker       ClusterIP   None         <none>        9092/TCP,29092/TCP   46m   app.kubernetes.io/instance=kafka,app.kubernetes.io/managed-by=kubedb.com,app.kubernetes.io/name=kafkas.kubedb.com,kubedb.com/role=broker
demo        service/kafka-controller   ClusterIP   None         <none>        9093/TCP             46m   app.kubernetes.io/instance=kafka,app.kubernetes.io/managed-by=kubedb.com,app.kubernetes.io/name=kafkas.kubedb.com,kubedb.com/role=controller

NAMESPACE   NAME                                READY   AGE   CONTAINERS   IMAGES
demo        statefulset.apps/kafka-broker       3/3     45m   kafka        ghcr.io/appscode-images/kafka-kraft:3.6.1@sha256:e251d3c0ceee0db8400b689e42587985034852a8a6c81b5973c2844e902e6d11
demo        statefulset.apps/kafka-controller   3/3     45m   kafka        ghcr.io/appscode-images/kafka-kraft:3.6.1@sha256:e251d3c0ceee0db8400b689e42587985034852a8a6c81b5973c2844e902e6d11

NAMESPACE   NAME                                       TYPE               VERSION   AGE
demo        appbinding.appcatalog.appscode.com/kafka   kubedb.com/kafka   3.4.0     45m

NAMESPACE   NAME                             TYPE                       DATA   AGE
demo        secret/kafka-admin-cred          kubernetes.io/basic-auth   2      46m
demo        secret/kafka-broker-config       Opaque                     3      46m
demo        secret/kafka-client-cert         kubernetes.io/tls          3      46m
demo        secret/kafka-controller-config   Opaque                     3      45m
demo        secret/kafka-keystore-cred       Opaque                     3      46m
demo        secret/kafka-server-cert         kubernetes.io/tls          5      46m
```

Flag `--output=wide` or `-owide` is used to print additional information. List command supports short names for each object types. You can use it like `kubectl get <short-name>`.

You can print labels with objects. The following command will list all Snapshots with their corresponding labels.

```bash
$ kubectl get pods -n demo --show-labels
NAME                 READY   STATUS    RESTARTS      AGE   LABELS
kafka-broker-0       1/1     Running   0             47m   app.kubernetes.io/component=database,app.kubernetes.io/instance=kafka,app.kubernetes.io/managed-by=kubedb.com,app.kubernetes.io/name=kafkas.kubedb.com,controller-revision-hash=kafka-broker-5f568d57c9,kubedb.com/role=broker,statefulset.kubernetes.io/pod-name=kafka-broker-0
kafka-broker-1       1/1     Running   0             47m   app.kubernetes.io/component=database,app.kubernetes.io/instance=kafka,app.kubernetes.io/managed-by=kubedb.com,app.kubernetes.io/name=kafkas.kubedb.com,controller-revision-hash=kafka-broker-5f568d57c9,kubedb.com/role=broker,statefulset.kubernetes.io/pod-name=kafka-broker-1
kafka-broker-2       1/1     Running   0             47m   app.kubernetes.io/component=database,app.kubernetes.io/instance=kafka,app.kubernetes.io/managed-by=kubedb.com,app.kubernetes.io/name=kafkas.kubedb.com,controller-revision-hash=kafka-broker-5f568d57c9,kubedb.com/role=broker,statefulset.kubernetes.io/pod-name=kafka-broker-2
kafka-controller-0   1/1     Running   0             47m   app.kubernetes.io/component=database,app.kubernetes.io/instance=kafka,app.kubernetes.io/managed-by=kubedb.com,app.kubernetes.io/name=kafkas.kubedb.com,controller-revision-hash=kafka-controller-96ddd885f,kubedb.com/role=controller,statefulset.kubernetes.io/pod-name=kafka-controller-0
kafka-controller-1   1/1     Running   0             47m   app.kubernetes.io/component=database,app.kubernetes.io/instance=kafka,app.kubernetes.io/managed-by=kubedb.com,app.kubernetes.io/name=kafkas.kubedb.com,controller-revision-hash=kafka-controller-96ddd885f,kubedb.com/role=controller,statefulset.kubernetes.io/pod-name=kafka-controller-1
kafka-controller-2   1/1     Running   3 (47m ago)   47m   app.kubernetes.io/component=database,app.kubernetes.io/instance=kafka,app.kubernetes.io/managed-by=kubedb.com,app.kubernetes.io/name=kafkas.kubedb.com,controller-revision-hash=kafka-controller-96ddd885f,kubedb.com/role=controller,statefulset.kubernetes.io/pod-name=kafka-controller-2
```

You can also filter list using `--selector` flag.

```bash
$ kubectl get services -n demo --selector='app.kubernetes.io/name=kafkas.kubedb.com' --show-labels
NAME               TYPE        CLUSTER-IP   EXTERNAL-IP   PORT(S)              AGE   LABELS
kafka-broker       ClusterIP   None         <none>        9092/TCP,29092/TCP   49m   app.kubernetes.io/component=database,app.kubernetes.io/instance=kafka,app.kubernetes.io/managed-by=kubedb.com,app.kubernetes.io/name=kafkas.kubedb.com
kafka-controller   ClusterIP   None         <none>        9093/TCP             49m   app.kubernetes.io/component=database,app.kubernetes.io/instance=kafka,app.kubernetes.io/managed-by=kubedb.com,app.kubernetes.io/name=kafkas.kubedb.com
```

To print only object name, run the following command:

```bash
$ kubectl get all -o name -n demo
pod/kafka-broker-0
pod/kafka-broker-1
pod/kafka-broker-2
pod/kafka-controller-0
pod/kafka-controller-1
pod/kafka-controller-2
service/kafka-broker
service/kafka-controller
statefulset.apps/kafka-broker
statefulset.apps/kafka-controller
appbinding.appcatalog.appscode.com/kafka
```

### How to Describe Objects

`kubectl describe` command allows users to describe any KubeDB object. The following command will describe Kafka instance `kafka` with relevant information.

```bash
$ kubectl describe -n demo kf kafka
Name:         kafka
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  kubedb.com/v1alpha2
Kind:         Kafka
Metadata:
  Creation Timestamp:  2023-03-29T07:01:29Z
  Finalizers:
    kubedb.com
  Generation:  1
  Managed Fields:
    API Version:  kubedb.com/v1alpha2
    Fields Type:  FieldsV1
    fieldsV1:
      f:metadata:
        f:finalizers:
          .:
          v:"kubedb.com":
    Manager:      kafka-operator
    Operation:    Update
    Time:         2023-03-29T07:01:29Z
    API Version:  kubedb.com/v1alpha2
    Fields Type:  FieldsV1
    fieldsV1:
      f:metadata:
        f:annotations:
          .:
          f:kubectl.kubernetes.io/last-applied-configuration:
      f:spec:
        .:
        f:authSecret:
        f:enableSSL:
        f:healthChecker:
          .:
          f:failureThreshold:
          f:periodSeconds:
          f:timeoutSeconds:
        f:keystoreCredSecret:
        f:storageType:
        f:deletionPolicy:
        f:tls:
          .:
          f:certificates:
          f:issuerRef:
        f:topology:
          .:
          f:broker:
            .:
            f:replicas:
            f:resources:
              .:
              f:limits:
                .:
                f:memory:
              f:requests:
                .:
                f:cpu:
                f:memory:
            f:storage:
              .:
              f:accessModes:
              f:resources:
                .:
                f:requests:
                  .:
                  f:storage:
              f:storageClassName:
            f:suffix:
          f:controller:
            .:
            f:replicas:
            f:resources:
              .:
              f:limits:
                .:
                f:memory:
              f:requests:
                .:
                f:cpu:
                f:memory:
            f:storage:
              .:
              f:accessModes:
              f:resources:
                .:
                f:requests:
                  .:
                  f:storage:
              f:storageClassName:
            f:suffix:
        f:version:
    Manager:      kubectl-client-side-apply
    Operation:    Update
    Time:         2023-03-29T07:01:29Z
    API Version:  kubedb.com/v1alpha2
    Fields Type:  FieldsV1
    fieldsV1:
      f:status:
        .:
        f:conditions:
        f:phase:
    Manager:         kafka-operator
    Operation:       Update
    Subresource:     status
    Time:            2023-03-29T07:01:34Z
  Resource Version:  570445
  UID:               ed5f6197-0238-4aba-a7d9-7dc771b2564c
Spec:
  Auth Secret:
    Name:      kafka-admin-cred
  Enable SSL:  true
  Health Checker:
    Failure Threshold:  3
    Period Seconds:     20
    Timeout Seconds:    10
  Keystore Cred Secret:
    Name:  kafka-keystore-cred
  Pod Template:
    Controller:
    Metadata:
    Spec:
      Resources:
  Storage Type:        Durable
  Deletion Policy:     DoNotTerminate
  Tls:
    Certificates:
      Alias:        server
      Secret Name:  kafka-server-cert
      Alias:        client
      Secret Name:  kafka-client-cert
    Issuer Ref:
      API Group:  cert-manager.io
      Kind:       Issuer
      Name:       kafka-ca-issuer
  Topology:
    Broker:
      Replicas:  3
      Resources:
        Limits:
          Memory:  1Gi
        Requests:
          Cpu:     500m
          Memory:  1Gi
      Storage:
        Access Modes:
          ReadWriteOnce
        Resources:
          Requests:
            Storage:         1Gi
        Storage Class Name:  standard
      Suffix:                broker
    Controller:
      Replicas:  3
      Resources:
        Limits:
          Memory:  1Gi
        Requests:
          Cpu:     500m
          Memory:  1Gi
      Storage:
        Access Modes:
          ReadWriteOnce
        Resources:
          Requests:
            Storage:         1Gi
        Storage Class Name:  standard
      Suffix:                controller
  Version:                   3.4.0
Status:
  Conditions:
    Last Transition Time:  2023-03-29T07:01:29Z
    Message:               The KubeDB operator has started the provisioning of Kafka: demo/kafka
    Observed Generation:   1
    Reason:                DatabaseProvisioningStartedSuccessfully
    Status:                True
    Type:                  ProvisioningStarted
    Last Transition Time:  2023-03-29T07:02:46Z
    Message:               All desired replicas are ready.
    Observed Generation:   1
    Reason:                AllReplicasReady
    Status:                True
    Type:                  ReplicaReady
    Last Transition Time:  2023-03-29T07:02:37Z
    Message:               The Kafka: demo/kafka is accepting client requests
    Observed Generation:   1
    Reason:                DatabaseAcceptingConnectionRequest
    Status:                True
    Type:                  AcceptingConnection
    Last Transition Time:  2023-03-29T07:03:37Z
    Message:               The Kafka: demo/kafka is ready.
    Observed Generation:   1
    Reason:                ReadinessCheckSucceeded
    Status:                True
    Type:                  Ready
    Last Transition Time:  2023-03-29T07:03:41Z
    Message:               The Kafka: demo/kafka is successfully provisioned.
    Observed Generation:   1
    Reason:                DatabaseSuccessfullyProvisioned
    Status:                True
    Type:                  Provisioned
  Phase:                   Ready
Events:
  Type     Reason      Age   From                         Message
  ----     ------      ----  ----                         -------
  Warning  Failed      50m   KubeDB Ops-manager Operator  Fail to be ready database: "kafka". Reason: services "kafka-broker" not found
  Warning  Failed      50m   KubeDB Ops-manager Operator  Fail to be ready database: "kafka". Reason: services "kafka-broker" not found
  Warning  Failed      50m   KubeDB Ops-manager Operator  Fail to be ready database: "kafka". Reason: services "kafka-broker" not found
  Warning  Failed      50m   KubeDB Ops-manager Operator  Fail to be ready database: "kafka". Reason: services "kafka-broker" not found
  Warning  Failed      50m   KubeDB Ops-manager Operator  Fail to be ready database: "kafka". Reason: services "kafka-broker" not found
  Normal   Successful  50m   KubeDB Ops-manager Operator  Successfully created Kafka server certificates
  Normal   Successful  50m   KubeDB Ops-manager Operator  Successfully created Kafka client-certificates
```

`kubectl describe` command provides following basic information about a database.

- PetSet
- Storage (Persistent Volume)
- Service
- Secret (If available)
- Topology (If available)
- Monitoring system (If available)

To hide details about PetSet & Service, use flag `--show-workload=false`
To hide details about Secret, use flag `--show-secret=false`
To hide events on KubeDB object, use flag `--show-events=false`

To describe all Kafka objects in `default` namespace, use following command

```bash
$ kubectl describe kf
```

To describe all Kafka objects from every namespace, provide `--all-namespaces` flag.

```bash
$ kubectl describe kf --all-namespaces
```

You can also describe KubeDb objects with matching labels. The following command will describe all Kafka objects with specified labels from every namespace.

```bash
$ kubectl describe kf --all-namespaces --selector='app.kubernetes.io/component=database'
```

To learn about various options of `describe` command, please visit [here](/docs/reference/cli/kubectl-dba_describe.md).


#### Edit restrictions

Various fields of a KubeDb object can't be edited using `edit` command. The following fields are restricted from updates for all KubeDB objects:

- apiVersion
- kind
- metadata.name
- metadata.namespace
- status

If PetSets or Deployments exists for a database, following fields can't be modified as well.

Kafka:

- spec.init
- spec.storageType
- spec.storage
- spec.podTemplate.spec.nodeSelector
- spec.podTemplate.spec.env


### How to Delete Objects

`kubectl delete` command will delete an object in `default` namespace by default unless namespace is provided. The following command will delete a Kafka instance `kafka` in demo namespace

```bash
$ kubectl delete kf kafka -n demo
kafka.kubedb.com "kafka" deleted
```

You can also use YAML files to delete objects. The following command will delete an Kafka using the type and name specified in `kafka.yaml`.

```bash
$ kubectl delete -f druid-quickstart.yaml
kafka.kubedb.com "kafka" deleted
```

`kubectl delete` command also takes input from `stdin`.

```bash
cat druid-quickstart.yaml | kubectl delete -f -
```

To delete database with matching labels, use `--selector` flag. The following command will delete kafka with label `app.kubernetes.io/instance=kafka`.

```bash
$ kubectl delete kf -l app.kubernetes.io/instance=kafka
```

## Using Kubectl

You can use Kubectl with KubeDB objects like any other CRDs. Below are some common examples of using Kubectl with KubeDB objects.

```bash
# List objects
$ kubectl get kafka
$ kubectl get kafka.kubedb.com

# Delete objects
$ kubectl delete kafka <name>
```

## Next Steps

- Learn how to use KubeDB to run a Apache Kafka [here](/docs/guides/kafka/README.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
