package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	slog "log"
	"net/http"
	"os"
	"strings"

	"github.com/dcu/mongodb_exporter/collector"
	"github.com/dcu/mongodb_exporter/shared"

	"github.com/golang/glog"
	"github.com/prometheus/client_golang/prometheus"
)

func mongodbDefaultURI() string {
	if u := os.Getenv("MONGODB_URL"); u != "" {
		return u
	}
	return "mongodb://localhost:27017"
}

var (
	listenAddressFlag = flag.String("web.listen-address", ":9001", "Address on which to expose metrics and web interface.")
	metricsPathFlag   = flag.String("web.metrics-path", "/metrics", "Path under which to expose metrics.")
	webTLSCert        = flag.String("web.tls-cert", "", "Path to PEM file that conains the certificate (and optionally also the private key in PEM format).\n"+
		"    \tThis should include the whole certificate chain.\n"+
		"    \tIf provided: The web socket will be a HTTPS socket.\n"+
		"    \tIf not provided: Only HTTP.")
	webTLSPrivateKey = flag.String("web.tls-private-key", "", "Path to PEM file that conains the private key (if not contained in web.tls-cert file).")
	webTLSClientCa   = flag.String("web.tls-client-ca", "", "Path to PEM file that conains the CAs that are trused for client connections.\n"+
		"    \tIf provided: Connecting clients should present a certificate signed by one of this CAs.\n"+
		"    \tIf not provided: Every client will be accepted.")

	mongodbURIFlag = flag.String("mongodb.uri", mongodbDefaultURI(), "Mongodb URI, format: [mongodb://][user:pass@]host1[:port1][,host2[:port2],...][/database][?options]")
	mongodbTLSCert = flag.String("mongodb.tls-cert", "", "Path to PEM file that conains the certificate (and optionally also the private key in PEM format).\n"+
		"    \tThis should include the whole certificate chain.\n"+
		"    \tIf provided: The connection will be opened via TLS to the MongoDB server.")
	mongodbTLSPrivateKey = flag.String("mongodb.tls-private-key", "", "Path to PEM file that conains the private key (if not contained in mongodb.tls-cert file).")
	mongodbTLSCa         = flag.String("mongodb.tls-ca", "", "Path to PEM file that conains the CAs that are trused for server connections.\n"+
		"    \tIf provided: MongoDB servers connecting to should present a certificate signed by one of this CAs.\n"+
		"    \tIf not provided: System default CAs are used.")
	mongodbTLSDisableHostnameValidation = flag.Bool("mongodb.tls-disable-hostname-validation", false, "Do hostname validation for server connection.")
	enabledGroupsFlag                   = flag.String("groups.enabled", "asserts,durability,background_flushing,connections,extra_info,global_lock,index_counters,network,op_counters,op_counters_repl,memory,locks,metrics", "Comma-separated list of groups to use, for more info see: docs.mongodb.org/manual/reference/command/serverStatus/")
	authUserFlag                        = flag.String("auth.user", "", "Username for basic auth.")
	authPassFlag                        = flag.String("auth.pass", "", "Password for basic auth.")
	mongodbUserName                     = flag.String("mongodb.username", "", "Username to connect to Mongodb")
	mongodbAuthMechanism                = flag.String("mongodb.mechanism", "", "auth mechanism to connect to Mongodb (ie: MONGODB-X509)")
	mongodbCollectOplog                 = flag.Bool("mongodb.collect.oplog", true, "collect Mongodb Oplog status")
	mongodbCollectOplogTail             = flag.Bool("mongodb.collect.oplog_tail", false, "tail Mongodb Oplog to get stats")
	mongodbCollectReplSet               = flag.Bool("mongodb.collect.replset", true, "collect Mongodb replica set status")
	mongodbCollectTopMetrics            = flag.Bool("mongodb.collect.top", false, "collect Mongodb Top metrics")
	mongodbCollectDatabaseMetrics       = flag.Bool("mongodb.collect.database", false, "collect MongoDB database metrics")
	mongodbCollectCollectionMetrics     = flag.Bool("mongodb.collect.collection", false, "Collect MongoDB collection metrics")
	mongodbCollectProfileMetrics        = flag.Bool("mongodb.collect.profile", false, "Collect MongoDB profile metrics")
	mongodbCollectConnPoolStats         = flag.Bool("mongodb.collect.connpoolstats", false, "Collect MongoDB connpoolstats")
	mongodbSocketTimeout                = flag.Duration("mongodb.socket-timeout", 0, "timeout for socket operations to mongodb")
	version                             = flag.Bool("version", false, "Print mongodb_exporter version")
)

type basicAuthHandler struct {
	handler  http.HandlerFunc
	user     string
	password string
}

func (h *basicAuthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	user, password, ok := r.BasicAuth()
	if !ok || password != h.password || user != h.user {
		w.Header().Set("WWW-Authenticate", "Basic realm=\"metrics\"")
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}
	h.handler(w, r)
	return
}

func hasUserAndPassword() bool {
	return *authUserFlag != "" && *authPassFlag != ""
}

func prometheusHandler() http.Handler {
	handler := prometheus.Handler()
	if hasUserAndPassword() {
		handler = &basicAuthHandler{
			handler:  prometheus.Handler().ServeHTTP,
			user:     *authUserFlag,
			password: *authPassFlag,
		}
	}

	return handler
}

func startWebServer() {
	handler := prometheusHandler()

	registerCollector()

	http.Handle(*metricsPathFlag, handler)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
<head><title>MongoDB Exporter</title></head>
<body>
<h1>MongoDB Exporter</h1>
<p><a href='` + *metricsPathFlag + `'>Metrics</a></p>
</body>
</html>`))
	})

	server := &http.Server{
		Addr:     *listenAddressFlag,
		ErrorLog: createHTTPServerLogWrapper(),
	}

	var err error
	if len(*webTLSCert) > 0 {
		clientValidation := "no"
		if len(*webTLSClientCa) > 0 && len(*webTLSCert) > 0 {
			certificates, err := shared.LoadCertificatesFrom(*webTLSClientCa)
			if err != nil {
				glog.Fatalf("Couldn't load client CAs from %s. Got: %s", *webTLSClientCa, err)
			}
			server.TLSConfig = &tls.Config{
				ClientCAs:  certificates,
				ClientAuth: tls.RequireAndVerifyClientCert,
			}
			clientValidation = "yes"
		}
		targetTLSPrivateKey := *webTLSPrivateKey
		if len(targetTLSPrivateKey) <= 0 {
			targetTLSPrivateKey = *webTLSCert
		}
		fmt.Printf("Listening on %s (scheme=HTTPS, secured=TLS, clientValidation=%s)\n", server.Addr, clientValidation)
		err = server.ListenAndServeTLS(*webTLSCert, targetTLSPrivateKey)
	} else {
		fmt.Printf("Listening on %s (scheme=HTTP, secured=no, clientValidation=no)\n", server.Addr)
		err = server.ListenAndServe()
	}

	if err != nil {
		panic(err)
	}
}

func registerCollector() {
	mongodbCollector := collector.NewMongodbCollector(collector.MongodbCollectorOpts{
		URI:                      *mongodbURIFlag,
		TLSCertificateFile:       *mongodbTLSCert,
		TLSPrivateKeyFile:        *mongodbTLSPrivateKey,
		TLSCaFile:                *mongodbTLSCa,
		TLSHostnameValidation:    !(*mongodbTLSDisableHostnameValidation),
		CollectOplog:             *mongodbCollectOplog,
		TailOplog:                *mongodbCollectOplogTail,
		CollectReplSet:           *mongodbCollectReplSet,
		CollectTopMetrics:        *mongodbCollectTopMetrics,
		CollectDatabaseMetrics:   *mongodbCollectDatabaseMetrics,
		CollectCollectionMetrics: *mongodbCollectCollectionMetrics,
		CollectProfileMetrics:    *mongodbCollectProfileMetrics,
		CollectConnPoolStats:     *mongodbCollectConnPoolStats,
		UserName:                 *mongodbUserName,
		AuthMechanism:            *mongodbAuthMechanism,
		SocketTimeout:            *mongodbSocketTimeout,
	})
	prometheus.MustRegister(mongodbCollector)
}

type bufferedLogWriter struct {
	buf []byte
}

func (w *bufferedLogWriter) Write(p []byte) (n int, err error) {
	glog.Info(strings.TrimSpace(strings.Replace(string(p), "\n", " ", -1)))
	return len(p), nil
}

func createHTTPServerLogWrapper() *slog.Logger {
	return slog.New(&bufferedLogWriter{}, "", 0)
}

func main() {
	flag.Parse()
	if *version {
		fmt.Println("mongodb_exporter version: {{VERSION}}")
		return
	}
	shared.ParseEnabledGroups(*enabledGroupsFlag)

	startWebServer()
}
