---
title: Kubectl-Dba Connect Elasticsearch
menu:
  docs_{{ .version }}:
    identifier: kubectl-dba-connect-elasticsearch
    name: Kubectl-Dba Connect Elasticsearch
    parent: reference-cli
menu_name: docs_{{ .version }}
section_menu_id: reference
---
## kubectl-dba connect elasticsearch

Connect to a shell to run elasticsearch api calls

### Synopsis

Use this cmd to run api calls to your elasticsearch database. 

This command connects you to a shell to run curl commands. 

It exports the following environment variables to run api calls to your database:
  $USERNAME
  $PASSWORD
  $ADDRESS
  $CACERT
  $CERT
  $KEY

Example connect command:
  # connect to a shell with curl access to the database of name es-demo in demo namespace
  kubectl dba connect es es-demo -n demo

Example curl commands:
  # curl command to run on the connected elasticsearch database:
  curl -u $USERNAME:$PASSWORD $ADDRESS/_cluster/health?pretty

  # curl command to run on the connected tls secured elasticsearch database:
  curl --cacert $CACERT --cert $CERT --key $KEY  -u $USERNAME:$PASSWORD $ADDRESS/_cluster/health?pretty

```
kubectl-dba connect elasticsearch [flags]
```

### Options

```
  -h, --help   help for elasticsearch
```

### Options inherited from parent commands

```
      --as string                      Username to impersonate for the operation
      --as-group stringArray           Group to impersonate for the operation, this flag can be repeated to specify multiple groups.
      --cache-dir string               Default cache directory (default "/home/runner/.kube/cache")
      --certificate-authority string   Path to a cert file for the certificate authority
      --client-certificate string      Path to a client certificate file for TLS
      --client-key string              Path to a client key file for TLS
      --cluster string                 The name of the kubeconfig cluster to use
      --context string                 The name of the kubeconfig context to use
      --enable-analytics               Send analytical events to Google Analytics (default true)
      --insecure-skip-tls-verify       If true, the server's certificate will not be checked for validity. This will make your HTTPS connections insecure
      --kubeconfig string              Path to the kubeconfig file to use for CLI requests.
      --match-server-version           Require server version to match client version
  -n, --namespace string               If present, the namespace scope for this CLI request
      --request-timeout string         The length of time to wait before giving up on a single server request. Non-zero values should contain a corresponding time unit (e.g. 1s, 2m, 3h). A value of zero means don't timeout requests. (default "0")
  -s, --server string                  The address and port of the Kubernetes API server
      --tls-server-name string         Server name to use for server certificate validation. If it is not provided, the hostname used to contact the server is used
      --token string                   Bearer token for authentication to the API server
      --user string                    The name of the kubeconfig user to use
```

### SEE ALSO

* [kubectl-dba connect](/docs/reference/cli/kubectl-dba_connect.md)	 - Connect to a database.

