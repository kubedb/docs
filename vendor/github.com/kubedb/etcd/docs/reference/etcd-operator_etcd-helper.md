## etcd-operator etcd-helper

Run etcd helper

### Synopsis

Run etcd helper

```
etcd-operator etcd-helper [flags]
```

### Options

```
      --Config IgnoredFlag[=true]                    
      --Config-file string                           Path to the server configuration file
      --advertise-client-urls URLs                   List of this member's client URLs to advertise to the public. (default http://localhost:2379)
      --auth-token string                            Specify auth token specific options. (default "simple")
      --auto-compaction-retention int                Auto compaction retention for mvcc key value store in hour. 0 means disable auto compaction.
      --auto-tls                                     Client TLS using generated certificates
      --ca-file string                               DEPRECATED: Path to the client server TLS CA file.
      --cert-file string                             Path to the client server TLS cert file.
      --client-cert-auth                             Enable client cert authentication.
      --cluster-active-size IgnoredFlag[=true]       
      --cluster-remove-delay IgnoredFlag[=true]      
      --cluster-sync-interval IgnoredFlag[=true]     
      --cors CORSInfo                                Comma-separated white list of origins for CORS (cross-origin resource sharing).
      --data-dir string                              Path to the data directory.
      --debug                                        Enable debug-level logging for etcd.
      --discovery string                             Discovery URL used to bootstrap the cluster.
      --discovery-fallback StringsFlag               Valid values include exit, proxy (default proxy)
      --discovery-proxy string                       HTTP proxy to use for traffic to discovery service.
      --discovery-srv string                         DNS domain used to bootstrap initial cluster.
      --election-timeout uint                        Time (in milliseconds) for an election to timeout. (default 1000)
      --enable-pprof                                 Enable runtime profiling data via HTTP server. Address is at client URL + "/debug/pprof/"
      --enable-v2                                    Accept etcd V2 client requests. (default true)
      --force IgnoredFlag[=true]                     
      --force-new-cluster                            Force to create a new one member cluster.
      --grpc-keepalive-interval duration             Frequency duration of server-to-client ping to check if a connection is alive (0 to disable). (default 2h0m0s)
      --grpc-keepalive-min-time duration             Minimum interval duration that a client should wait before pinging server. (default 5s)
      --grpc-keepalive-timeout duration              Additional duration of wait before closing a non-responsive connection (0 to disable). (default 20s)
      --heartbeat-interval uint                      Time (in milliseconds) of a heartbeat interval. (default 100)
  -h, --help                                         help for etcd-helper
      --initial-advertise-peer-urls URLs             List of this member's peer URLs to advertise to the rest of the cluster. (default http://localhost:2380)
      --initial-cluster string                       Initial cluster configuration for bootstrapping. (default "default=http://localhost:2380")
      --initial-cluster-state StringsFlag            Initial cluster state ('new' or 'existing'). (default new)
      --initial-cluster-token string                 Initial cluster token for the etcd cluster during bootstrap. (default "etcd-cluster")
      --key-file string                              Path to the client server TLS key file.
      --listen-client-urls URLs                      List of URLs to listen on for client traffic. (default http://localhost:2379)
      --listen-peer-urls URLs                        List of URLs to listen on for peer traffic. (default http://localhost:2380)
      --log-output string                            Specify 'stdout' or 'stderr' to skip journald logging even when running under systemd. (default "default")
      --log-package-levels string                    Specify a particular log level for each etcd package (eg: 'etcdmain=CRITICAL,etcdserver=DEBUG').
      --max-request-bytes uint                       Maximum client request size in bytes the server will accept. (default 1572864)
      --max-result-buffer IgnoredFlag[=true]         
      --max-retry-attempts IgnoredFlag[=true]        
      --max-snapshots uint                           Maximum number of snapshot files to retain (0 is unlimited). (default 5)
      --max-wals uint                                Maximum number of wal files to retain (0 is unlimited). (default 5)
      --metrics string                               Set level of detail for exported metrics, specify 'extensive' to include histogram metrics (default "basic")
      --name string                                  Human-readable name for this member. (default "default")
      --peer-auto-tls                                Peer TLS using generated certificates
      --peer-ca-file string                          DEPRECATED: Path to the peer server TLS CA file.
      --peer-cert-file string                        Path to the peer server TLS cert file.
      --peer-client-cert-auth                        Enable peer client cert authentication.
      --peer-election-timeout IgnoredFlag[=true]     
      --peer-heartbeat-interval IgnoredFlag[=true]   
      --peer-key-file string                         Path to the peer server TLS key file.
      --peer-trusted-ca-file string                  Path to the peer server TLS trusted CA file.
      --proxy StringsFlag                            Valid values include off, readonly, on (default off)
      --proxy-dial-timeout uint                      Time (in milliseconds) for a dial to timeout. (default 1000)
      --proxy-failure-wait uint                      Time (in milliseconds) an endpoint will be held in a failed state. (default 5000)
      --proxy-read-timeout uint                      Time (in milliseconds) for a read to timeout.
      --proxy-refresh-interval uint                  Time (in milliseconds) of the endpoints refresh interval. (default 30000)
      --proxy-write-timeout uint                     Time (in milliseconds) for a write to timeout. (default 5000)
      --quota-backend-bytes int                      Raise alarms when backend size exceeds the given quota. 0 means use the default quota.
      --retry-interval IgnoredFlag[=true]            
      --snapshot IgnoredFlag[=true]                  
      --snapshot-count uint                          Number of committed transactions to trigger a snapshot to disk. (default 100000)
      --strict-reconfig-check                        Reject reconfiguration requests that would cause quorum loss. (default true)
      --test.coverprofile IgnoredFlag[=true]         
      --test.outputdir IgnoredFlag[=true]            
      --trusted-ca-file string                       Path to the client server TLS trusted CA key file.
      --version                                      Print the version and exit.
      --vv IgnoredFlag[=true]                        
      --wal-dir string                               Path to the dedicated wal directory.
```

### Options inherited from parent commands

```
      --alsologtostderr                  log to standard error as well as files
      --bypass-validating-webhook-xray   if true, bypasses validating webhook xray checks
      --enable-analytics                 Send analytical events to Google Analytics (default true)
      --log_backtrace_at traceLocation   when logging hits line file:N, emit a stack trace (default :0)
      --log_dir string                   If non-empty, write log files in this directory
      --logtostderr                      log to standard error instead of files
      --stderrthreshold severity         logs at or above this threshold go to stderr
      --use-kubeapiserver-fqdn-for-aks   if true, uses kube-apiserver FQDN for AKS cluster to workaround https://github.com/Azure/AKS/issues/522 (default true)
  -v, --v Level                          log level for V logs
      --vmodule moduleSpec               comma-separated list of pattern=N settings for file-filtered logging
```

### SEE ALSO

* [etcd-operator](etcd-operator.md)	 - 

