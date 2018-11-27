package framework

import (
	"fmt"
	"time"

	"github.com/appscode/go/crypto/rand"
	jsonTypes "github.com/appscode/go/encoding/json/types"
	"github.com/appscode/go/types"
	api "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	"github.com/kubedb/apimachinery/client/clientset/versioned/typed/kubedb/v1alpha1/util"
	. "github.com/onsi/gomega"
	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (f *Invocation) Redis() *api.Redis {
	return &api.Redis{
		ObjectMeta: metav1.ObjectMeta{
			Name:      rand.WithUniqSuffix("redis"),
			Namespace: f.namespace,
			Labels: map[string]string{
				"app": f.app,
			},
		},
		Spec: api.RedisSpec{
			Version: jsonTypes.StrYo(DBVersion),
			Storage: &core.PersistentVolumeClaimSpec{
				Resources: core.ResourceRequirements{
					Requests: core.ResourceList{
						core.ResourceStorage: resource.MustParse("1Gi"),
					},
				},
				StorageClassName: types.StringP(f.StorageClass),
			},
		},
	}
}

func (f *Framework) CreateRedis(obj *api.Redis) error {
	_, err := f.extClient.KubedbV1alpha1().Redises(obj.Namespace).Create(obj)
	return err
}

func (f *Framework) GetRedis(meta metav1.ObjectMeta) (*api.Redis, error) {
	return f.extClient.KubedbV1alpha1().Redises(meta.Namespace).Get(meta.Name, metav1.GetOptions{})
}

func (f *Framework) TryPatchRedis(meta metav1.ObjectMeta, transform func(*api.Redis) *api.Redis) (*api.Redis, error) {
	redis, err := f.extClient.KubedbV1alpha1().Redises(meta.Namespace).Get(meta.Name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	redis, _, err = util.PatchRedis(f.extClient.KubedbV1alpha1(), redis, transform)
	return redis, err
}

func (f *Framework) DeleteRedis(meta metav1.ObjectMeta) error {
	return f.extClient.KubedbV1alpha1().Redises(meta.Namespace).Delete(meta.Name, &metav1.DeleteOptions{})
}

func (f *Framework) EventuallyRedis(meta metav1.ObjectMeta) GomegaAsyncAssertion {
	return Eventually(
		func() bool {
			_, err := f.extClient.KubedbV1alpha1().Redises(meta.Namespace).Get(meta.Name, metav1.GetOptions{})
			if err != nil {
				if kerr.IsNotFound(err) {
					return false
				}
				Expect(err).NotTo(HaveOccurred())
			}
			return true
		},
		time.Minute*5,
		time.Second*5,
	)
}

func (f *Framework) EventuallyRedisRunning(meta metav1.ObjectMeta) GomegaAsyncAssertion {
	return Eventually(
		func() bool {
			redis, err := f.extClient.KubedbV1alpha1().Redises(meta.Namespace).Get(meta.Name, metav1.GetOptions{})
			Expect(err).NotTo(HaveOccurred())
			return redis.Status.Phase == api.DatabasePhaseRunning
		},
		time.Minute*5,
		time.Second*5,
	)
}

func (f *Framework) CleanRedis() {
	redisList, err := f.extClient.KubedbV1alpha1().Redises(f.namespace).List(metav1.ListOptions{})
	if err != nil {
		return
	}
	for _, e := range redisList.Items {
		if _, _, err := util.PatchRedis(f.extClient.KubedbV1alpha1(), &e, func(in *api.Redis) *api.Redis {
			in.ObjectMeta.Finalizers = nil
			in.Spec.TerminationPolicy = api.TerminationPolicyWipeOut
			return in
		}); err != nil {
			fmt.Printf("error Patching Redis. error: %v", err)
		}
	}
	if err := f.extClient.KubedbV1alpha1().Redises(f.namespace).DeleteCollection(deleteInForeground(), metav1.ListOptions{}); err != nil {
		fmt.Printf("error in deletion of Redis. Error: %v", err)
	}
}
