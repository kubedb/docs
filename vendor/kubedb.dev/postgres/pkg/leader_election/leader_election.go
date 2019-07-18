package leader_election

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/user"
	"strconv"
	"strings"
	"time"

	"github.com/appscode/go/ioutil"
	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

	fmt.Println(fmt.Sprintf(`We want "%v" as our leader`, hostname))

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
	if _, err := kubeClient.CoreV1().ConfigMaps(namespace).Create(configMap); err != nil && !kerr.IsAlreadyExists(err) {
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

	go func() {
		leaderelection.RunOrDie(context.Background(), leaderelection.LeaderElectionConfig{
			Lock: resLock,
			// ref: https://github.com/kubernetes/apiserver/blob/kubernetes-1.12.0/pkg/apis/config/v1alpha1/defaults.go#L26-L52
			LeaseDuration: time.Duration(leaseDuration) * time.Second,
			RenewDeadline: time.Duration(renewDeadline) * time.Second,
			RetryPeriod:   time.Duration(retryPeriod) * time.Second,
			Callbacks: leaderelection.LeaderCallbacks{
				OnStartedLeading: func(ctx context.Context) {
					fmt.Println("Got leadership, now do your jobs")
				},
				OnStoppedLeading: func() {
					fmt.Println("Lost leadership, now quit")
					os.Exit(1)
				},
				OnNewLeader: func(identity string) {
					statefulSet, err := kubeClient.AppsV1().StatefulSets(namespace).Get(statefulSetName, metav1.GetOptions{})
					if err != nil {
						log.Fatalln(err)
					}

					pods, err := kubeClient.CoreV1().Pods(namespace).List(metav1.ListOptions{
						LabelSelector: metav1.FormatLabelSelector(statefulSet.Spec.Selector),
					})
					if err != nil {
						log.Fatalln(err)
					}

					for _, pod := range pods.Items {
						role := RoleReplica
						if pod.Name == identity {
							role = RolePrimary
						}
						_, _, err = core_util.PatchPod(kubeClient, &pod, func(in *core.Pod) *core.Pod {
							in.Labels["kubedb.com/role"] = role
							return in
						})
					}

					role := RoleReplica
					if identity == hostname {
						role = RolePrimary
					}

					if runningFirstTime {
						runningFirstTime = false
						go func() {
							// su-exec postgres /scripts/primary/run.sh
							cmd := exec.Command("su-exec", "postgres", fmt.Sprintf("/scripts/%s/run.sh", role))
							cmd.Stdout = os.Stdout
							cmd.Stderr = os.Stderr

							if err = cmd.Run(); err != nil {
								log.Println(err)
							}
							os.Exit(1)
						}()
					} else {
						if identity == hostname {
							if !ioutil.WriteString("/tmp/pg-failover-trigger", "") {
								log.Fatalln("Failed to create trigger file")
							}
						}
					}
				},
			},
		})
	}()

	select {}
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
