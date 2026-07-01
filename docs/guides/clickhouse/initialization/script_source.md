---
title: Initialize ClickHouse using Script Source
menu:
  docs_{{ .version }}:
    identifier: ch-script-source-initialization
    name: Using Script
    parent: ch-initialization-clickhouse
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Initialize ClickHouse with Script

KubeDB supports ClickHouse database initialization. This tutorial will show you how to use KubeDB to initialize a ClickHouse database from a script.

## Before You Begin

At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created

$ kubectl get ns demo
NAME    STATUS  AGE
demo    Active  5s
```

> Note: YAML files used in this tutorial are stored in [docs/examples/clickhouse](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/clickhouse) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Prepare Initialization Scripts

ClickHouse supports initialization with `.sh` and `.sql` files. In this tutorial, we will use an `init.sql` script to create a database `init_script` with a table `kubedb_table` and insert some sample data.

We will use a ConfigMap as the script source. You can use any Kubernetes supported [volume](https://kubernetes.io/docs/concepts/storage/volumes) as a script source.

At first, we will create a ConfigMap from an `init.sql` file. Then, we will provide this ConfigMap as a script source in `init.script` of the ClickHouse CRD spec.

Let's create a ConfigMap with the initialization script:

```bash
$ kubectl create configmap -n demo ch-init-script \
--from-literal=init.sql="$(curl -fsSL https://raw.githubusercontent.com/Bonusree/init_script/main/clickhouse_init.sql)"
configmap/ch-init-script created
```

## Create ClickHouse with Script Source

Following YAML describes the ClickHouse object with `init.script`:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: ClickHouse
metadata:
  name: script-clickhouse
  namespace: demo
spec:
  version: "24.4.1"
  storageType: Durable
  storage:
    storageClassName: "standard"
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  init:
    script:
      configMap:
        name: ch-init-script
  deletionPolicy: WipeOut
```

Here,

- `init.script` specifies the scripts used to initialize the database when it is being created.

VolumeSource provided in `init.script` will be mounted in the Pod and will be executed while creating ClickHouse.

Now, let's create the ClickHouse CRD using the YAML shown above:

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/clickhouse/initialization/script-clickhouse.yaml
clickhouse.kubedb.com/script-clickhouse created
```

Now, wait until ClickHouse goes in `Ready` state. Verify that the database is in `Ready` state using the following command:

```bash
$ kubectl-dba describe ch -n demo script-clickhouse
Name:         script-clickhouse
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  kubedb.com/v1alpha2
Kind:         ClickHouse
Metadata:
  Creation Timestamp:  2026-07-01T10:38:15Z
  Finalizers:
    kubedb.com/clickhouse
  Generation:        3
  Resource Version:  1099945
  UID:               89f5c0b9-5388-434a-81ad-96650fca95d0
Spec:
  Auth Secret:
    API Group:  
    Kind:       Secret
    Name:       script-clickhouse-auth
  Auto Ops:
  Deletion Policy:  WipeOut
  Health Checker:
    Failure Threshold:  3
    Period Seconds:     20
    Timeout Seconds:    10
  Init:
    Script:
      Config Map:
        Name:  ch-init-script
  Pod Template:
    Controller:
    Metadata:
    Spec:
      Containers:
        Name:  clickhouse
        Resources:
          Limits:
            Memory:  4Gi
          Requests:
            Cpu:     1
            Memory:  4Gi
        Security Context:
          Allow Privilege Escalation:  false
          Capabilities:
            Drop:
              ALL
          Run As Non Root:  true
          Run As User:      101
          Seccomp Profile:
            Type:  RuntimeDefault
      Init Containers:
        Name:  clickhouse-init
        Resources:
        Security Context:
          Allow Privilege Escalation:  false
          Capabilities:
            Drop:
              ALL
          Run As Non Root:  true
          Run As User:      101
          Seccomp Profile:
            Type:  RuntimeDefault
      Pod Placement Policy:
        Name:  default
      Security Context:
        Fs Group:  101
  Replicas:        1
  Storage:
    Access Modes:
      ReadWriteOnce
    Resources:
      Requests:
        Storage:         1Gi
    Storage Class Name:  local-path
  Storage Type:          Durable
  Version:               24.4.1
Status:
  Conditions:
    Last Transition Time:  2026-07-01T10:38:15Z
    Message:               The KubeDB operator has started the provisioning of ClickHouse: demo/script-clickhouse
    Observed Generation:   2
    Reason:                ProvisioningStarted
    Status:                True
    Type:                  ProvisioningStarted
    Last Transition Time:  2026-07-01T10:38:48Z
    Message:               All desired replicas are ready
    Observed Generation:   3
    Reason:                AllReplicasReady
    Status:                True
    Type:                  ReplicaReady
    Last Transition Time:  2026-07-01T10:39:05Z
    Message:               The Clickhouse: demo/script-clickhouse is accepting client requests
    Observed Generation:   3
    Reason:                AcceptingConnection
    Status:                True
    Type:                  AcceptingConnection
    Last Transition Time:  2026-07-01T10:39:05Z
    Message:               database demo/script-clickhouse is ready
    Observed Generation:   3
    Reason:                AllReplicasReady
    Status:                True
    Type:                  Ready
    Last Transition Time:  2026-07-01T10:39:05Z
    Message:               The ClickHouse: demo/script-clickhouse is successfully provisioned.
    Observed Generation:   3
    Reason:                DatabaseSuccessfullyProvisioned
    Status:                True
    Type:                  Provisioned
  Phase:                   Ready
Events:                    <none>
```


## Verify Initialization

Now let's connect to our ClickHouse instance to verify that the database has been initialized successfully.

**Connection Information:**

- Host name/address: you can use any of these
  - Service: `script-clickhouse.demo`
  - Pod IP: (`$ kubectl get pods script-clickhouse-0 -n demo -o yaml | grep podIP`)
- Port: `9000` (native TCP) or `8123` (HTTP)

- Username: Run the following command to get the *username*:

  ```bash
  $ kubectl get secret -n demo script-clickhouse-auth -o jsonpath='{.data.username}' | base64 -d
  admin
  ```

- Password: Run the following command to get the *password*:

  ```bash
  $ kubectl get secret -n demo script-clickhouse-auth -o jsonpath='{.data.password}' | base64 -d
  NkBpF0IQRCZ2isMb
  ```

Now, connect to ClickHouse using the `clickhouse-client` and run the following query to confirm initialization:

```bash
$ kubectl exec -it -n demo script-clickhouse-0 -- clickhouse-client --user=admin --password=NkBpF0IQRCZ2isMb --query "SHOW TABLES FROM init_script"
kubedb_table
```

You can also verify that the table was populated correctly:

```bash
$ kubectl exec -it -n demo script-clickhouse-0 -- clickhouse-client --user=admin --password=NkBpF0IQRCZ2isMb --query "SELECT * FROM init_script.kubedb_table"
1	name1
```

We can see that the table `kubedb_table` in the `init_script` database was created and populated through the initialization script.

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
$ kubectl delete -n demo clickhouse/script-clickhouse
$ kubectl delete -n demo configmap/ch-init-script
$ kubectl delete ns demo
```

## Next Steps

- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
