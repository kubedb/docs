package framework

import (
	"fmt"
	"strings"
	"time"

	"github.com/appscode/kutil/tools/portforward"
	"github.com/go-redis/redis"
	. "github.com/onsi/gomega"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (f *Framework) GetDatabasePod(meta metav1.ObjectMeta) (*core.Pod, error) {
	return f.kubeClient.CoreV1().Pods(meta.Namespace).Get(meta.Name+"-0", metav1.GetOptions{})
}

func (f *Framework) GetRedisClient(meta metav1.ObjectMeta) (*redis.Client, error) {
	pod, err := f.GetDatabasePod(meta)
	if err != nil {
		return nil, err
	}

	f.tunnel = portforward.NewTunnel(
		f.kubeClient.CoreV1().RESTClient(),
		f.restConfig,
		meta.Namespace,
		pod.Name,
		6379,
	)

	if err := f.tunnel.ForwardPort(); err != nil {
		return nil, err
	}

	return redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("localhost:%v", f.tunnel.Local),
		Password: "", // no password set
		DB:       0,  // use default DB
	}), nil
}

func (f *Framework) EventuallyRedisConfig(meta metav1.ObjectMeta, config string) GomegaAsyncAssertion {
	configPair := strings.Split(config, " ")

	return Eventually(
		func() string {

			client, err := f.GetRedisClient(meta)
			Expect(err).NotTo(HaveOccurred())

			defer f.tunnel.Close()

			// ping database to check if it is ready
			pong, err := client.Ping().Result()
			if err != nil {
				return ""
			}

			if !strings.Contains(pong, "PONG") {
				return ""
			}

			// get configuration
			response := client.ConfigGet(configPair[0])
			result := response.Val()
			ret := make([]string, 0)
			for _, r := range result {
				ret = append(ret, r.(string))
			}
			return strings.Join(ret, " ")
		},
		time.Minute*5,
		time.Second*5,
	)
}

func (f *Framework) EventuallySetItem(meta metav1.ObjectMeta, key, value string) GomegaAsyncAssertion {
	return Eventually(
		func() bool {
			client, err := f.GetRedisClient(meta)
			Expect(err).NotTo(HaveOccurred())

			defer f.tunnel.Close()

			err = client.Set(key, value, 0).Err()
			if err != nil {
				return false
			}
			return true
		},
		time.Minute*5,
		time.Second*5,
	)
}

func (f *Framework) EventuallyGetItem(meta metav1.ObjectMeta, key string) GomegaAsyncAssertion {
	return Eventually(
		func() string {
			client, err := f.GetRedisClient(meta)
			Expect(err).NotTo(HaveOccurred())

			defer f.tunnel.Close()

			val, err := client.Get(key).Result()
			if err != nil {
				return ""
			}
			return string(val)
		},
		time.Minute*5,
		time.Second*5,
	)
}
