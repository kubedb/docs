/*
Copyright AppsCode Inc. and Contributors

Licensed under the AppsCode Community License 1.0.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://github.com/appscode/licenses/raw/1.0.0/AppsCode-Community-1.0.0.md

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package leader_election

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"os/user"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/appscode/go/ioutil"
	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
	"k8s.io/client-go/tools/record"
	core_util "kmodules.xyz/client-go/core/v1"
	"kmodules.xyz/client-go/tools/clientcmd"
)

const (
	RolePrimary = "primary"
	RoleReplica = "replica"

	LeaseDurationEnv = "LEASE_DURATION"
	RenewDeadlineEnv = "RENEW_DEADLINE"
	RetryPeriodEnv   = "RETRY_PERIOD"
)

func RunLeaderElection() {
	namespace, leaseDuration, renewDeadline, retryPeriod := loadEnvVariables()

	// Change owner of Postgres data directory
	if err := setPermission(); err != nil {
		log.Fatalln(err)
	}

	hostname, err := os.Hostname()
	if err != nil {
		log.Fatalln(err)
	}

	parts := strings.Split(hostname, "-")
	statefulSetName := strings.Join(parts[:len(parts)-1], "-")

	log.Printf("We want \"%v\" as our leader\n", hostname)

	config, err := restclient.InClusterConfig()
	if err != nil {
		log.Fatalln(err)
	}
	clientcmd.Fix(config)

	kubeClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalln(err)
	}

	configMap := &core.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      GetLeaderLockName(statefulSetName),
			Namespace: namespace,
		},
	}
	if _, err := kubeClient.CoreV1().ConfigMaps(namespace).Create(context.TODO(), configMap, metav1.CreateOptions{}); err != nil && !kerr.IsAlreadyExists(err) {
		log.Fatalln(err)
	}

	resLock := &resourcelock.ConfigMapLock{
		ConfigMapMeta: configMap.ObjectMeta,
		Client:        kubeClient.CoreV1(),
		LockConfig: resourcelock.ResourceLockConfig{
			Identity:      hostname,
			EventRecorder: &record.FakeRecorder{},
		},
	}

	runningFirstTime := true
	var cmd *exec.Cmd
	lastLeader := ""

	runWrapperUntilExit := func(role string) {
		log.Printf("Starting database wrapper script as %s\n", role)
		// su-exec postgres /scripts/primary/run.sh
		cmd = exec.Command("su-exec", "postgres", fmt.Sprintf("/scripts/%s/run.sh", role))
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err = cmd.Run()
		log.Println("DB Wrapper terminated, exiting too. If there was an error, find it below.")
		cmd = nil
		if err != nil {
			log.Println(err)
			os.Exit(1)
		}
		os.Exit(0)
	}
	terminateDB := func() {
		log.Println("Terminating postgres server and operator")
		if cmd != nil && cmd.Process != nil {
			log.Println("Postgres is running, sending SIGTERM")
			utilruntime.Must(cmd.Process.Signal(syscall.SIGTERM))
			log.Println("Waiting for postgres to terminate")
			select {}
		} else {
			log.Println("Postgres already exited, nothing to do")
		}
	}

	go func() {
		leaderelection.RunOrDie(context.Background(), leaderelection.LeaderElectionConfig{
			Lock: resLock,
			// ref: https://github.com/kubernetes/apiserver/blob/kubernetes-1.12.0/pkg/apis/config/v1alpha1/defaults.go#L26-L52
			LeaseDuration: time.Duration(leaseDuration) * time.Second,
			RenewDeadline: time.Duration(renewDeadline) * time.Second,
			RetryPeriod:   time.Duration(retryPeriod) * time.Second,
			Callbacks: leaderelection.LeaderCallbacks{
				OnStartedLeading: func(ctx context.Context) {
					log.Println("Got leadership, creating the trigger file")
					if !ioutil.WriteString("/tmp/pg-failover-trigger", "") {
						log.Fatalln("Failed to create trigger file")
					}
				},
				OnStoppedLeading: func() {
					log.Println("Lost leadership, initiating a restart to correctly signal the database")
					terminateDB()
				},
				OnNewLeader: func(identity string) {
					log.Printf("Leader changed from '%s' to '%s'\n", lastLeader, identity)
					lastLeader = identity
					statefulSet, err := kubeClient.AppsV1().StatefulSets(namespace).Get(context.TODO(), statefulSetName, metav1.GetOptions{})
					if err != nil {
						log.Fatalln(err)
					}

					pods, err := kubeClient.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{
						LabelSelector: metav1.FormatLabelSelector(statefulSet.Spec.Selector),
					})
					if err != nil {
						log.Fatalln(err)
					}

					log.Println("Annotating pods for statefulset")
					for _, pod := range pods.Items {
						role := RoleReplica
						if pod.Name == identity {
							role = RolePrimary
						}
						if _, _, err := core_util.PatchPod(context.TODO(), kubeClient, &pod, func(in *core.Pod) *core.Pod {
							in.Labels["kubedb.com/role"] = role
							return in
						}, metav1.PatchOptions{}); err != nil {
							// not sure if panic-ing will make the situation better or worse. but, as we are going to
							// reimplement the postgres clustering part, lets keep it as it is right now
							fmt.Println("got error while updating pod label:", err)
						}

					}

					role := RoleReplica
					if identity == hostname {
						role = RolePrimary
					}

					log.Printf("This pod is now a %s\n", role)

					if runningFirstTime {
						runningFirstTime = false
						go runWrapperUntilExit(role)
					}
				},
			},
		})
		log.Println("Leader election died, exiting DB")
		terminateDB()
	}()

	doneChan := make(chan os.Signal, 1)
	signal.Notify(doneChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	recvSig := <-doneChan
	log.Printf("Received signal: %s, exiting\n", recvSig)
	terminateDB()
}

func loadEnvVariables() (namespace string, leaseDuration, renewDeadline, retryPeriod int) {
	var err error

	namespace = os.Getenv("NAMESPACE")
	if namespace == "" {
		namespace = "default"
	}

	leaseDurationStr := os.Getenv(LeaseDurationEnv)
	if leaseDuration, err = strconv.Atoi(leaseDurationStr); err != nil || leaseDuration == 0 {
		leaseDuration = 15
	}

	renewDeadlineStr := os.Getenv(RenewDeadlineEnv)
	if renewDeadline, err = strconv.Atoi(renewDeadlineStr); err != nil || renewDeadline == 0 {
		renewDeadline = 10
	}

	retryPeriodStr := os.Getenv(RetryPeriodEnv)
	if retryPeriod, err = strconv.Atoi(retryPeriodStr); err != nil || retryPeriod == 0 {
		retryPeriod = 2
	}

	return
}

func setPermission() error {
	u, err := user.Lookup("postgres")
	if err != nil {
		return err
	}
	uid, err := strconv.Atoi(u.Uid)
	if err != nil {
		return err
	}
	gid, err := strconv.Atoi(u.Gid)
	if err != nil {
		return err
	}
	err = os.Chown("/var/pv", uid, gid)
	if err != nil {
		return err
	}
	return nil
}

func GetLeaderLockName(offshootName string) string {
	return fmt.Sprintf("%s-leader-lock", offshootName)
}
