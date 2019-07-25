# KubeDB Operator Design Document

KubeDB Operator is a wapper of all DB services supported by [kubedb](https://kubedb.com). 

## Functions In KubeDB Operator

1. Wapper DB Operators and APIServer as a application 
2. Implement KubeDB APIServer and Register Admission Webhooks(For Mutating and Validating)
3. Config and Start Operators

## Implement Details

### Wapper DB Operators and APIServer as a application

KubeDB Operator is a CLI applications based on [cobra](https://github.com/spf13/cobra).

KubeDB Operator defined two commands, "version" and "run"

#### **From the [cobra](https://github.com/spf13/cobra) framework view**

All the code referenced [cobra](https://github.com/spf13/cobra) is under [root.go](https://github.com/kubedb/operator/blob/master/pkg/cmds/root.go)

Thay do the following three things:

1. Create a rootCmd 
```go
var rootCmd = &cobra.Command{
		Use:               "kubedb-operator [command]",
		Short:             "KubeDB operator by AppsCode",
		DisableAutoGenTag: true,
		PersistentPreRun: func(c *cobra.Command, args []string) {
			flags.DumpAll(c.Flags())
			//log.Infoln(c.Flags())
			cli.SendAnalytics(c, version)

			scheme.AddToScheme(clientsetscheme.Scheme)
			appcatscheme.AddToScheme(clientsetscheme.Scheme)
			cli.LoggerOptions = golog.ParseFlags(c.Flags())

		},
	}
```
2. Create a child versionCMD 
```go
rootCmd.AddCommand(v.NewCmdVersion())
```

3. Create a child runCMD

```go
rootCmd.AddCommand(NewCmdRun(version, os.Stdout, os.Stderr, stopCh))
```

But how the flags(kubernetes autorize, etcd param) are specified?

Thay are specified by 

1. NewRootCmd: ```go rootCmd.AddCommand(NewCmdRun(version, os.Stdout, os.Stderr, stopCh))```
2. NewCmdRun: ```go o.AddFlags(cmd.Flags())```
3. AddFlags: ```go o.RecommendedOptions.AddFlags(fs) ```(All flags concerned about kubernetes is specified)
4. AddFlags:  ```go o.ExtraOptions.AddFlags(fs)```(All KubeDB concerned is specified)
```go
type ExtraOptions struct {
	EnableRBAC                  bool
	OperatorNamespace           string
	RestrictToOperatorNamespace bool
	GoverningService            string
	QPS                         float64
	Burst                       int
	ResyncPeriod                time.Duration
	MaxNumRequeues              int
	NumThreads                  int

	EnableMutatingWebhook   bool
	EnableValidatingWebhook bool
}
```

### Implement KubeDB APIServer

![API Server implement Details](/images/APIServer-implement-details.png)

### Config and Start Operators

For more details you can read [controller souce code](https://github.com/kubedb/operator/tree/master/pkg/controller). It's simple and directive. Two function is finished in this package.

1. ```EnsureCustomResourceDefinitions``` when init operator controller
2. Config and initialize operators.

## How To Debug KubeDB Operator

```console
# clone kubedb-operator project
$ git clone https://github.com/kubedb/operator

# cd to operator project
$ cd $GOPATH/src/github.com/kubedb/operator

# deploy KubeDB resources and run KubeDB operator in localhost
$ ./hack/deploy/setup.sh --minikube --run

# deploy Kubedb Catalog, *KubeDB operator will run in a message loop, you should not stop it immediately, To install Kubedb Catalog you can just start another terminal.* 
$ ./hack/deploy/install-catalog.sh
```

After deploy KubeDB successfully, you can debug it using any IDE you like, You should start your KubeDB operator using the same param in *./hack/deploy/setup.sh*. 

![Debug using atom](/images/debug-using-atom.gif)

