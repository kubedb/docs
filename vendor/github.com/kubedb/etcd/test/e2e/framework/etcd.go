package framework

import (
	"fmt"
	"time"

	"github.com/appscode/go/crypto/rand"
	"github.com/appscode/go/encoding/json/types"
	api "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	"github.com/kubedb/apimachinery/client/clientset/versioned/typed/kubedb/v1alpha1/util"
	. "github.com/onsi/gomega"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (f *Invocation) Etcd() *api.Etcd {
	return &api.Etcd{
		ObjectMeta: metav1.ObjectMeta{
			Name:      rand.WithUniqSuffix("etcd"),
			Namespace: f.namespace,
			Labels: map[string]string{
				"app": f.app,
			},
		},
		Spec: api.EtcdSpec{
			Version: types.StrYo(DBVersion),
		},
	}
}

func (f *Framework) CreateEtcd(obj *api.Etcd) error {
	_, err := f.extClient.KubedbV1alpha1().Etcds(obj.Namespace).Create(obj)
	return err
}

func (f *Framework) GetEtcd(meta metav1.ObjectMeta) (*api.Etcd, error) {
	return f.extClient.KubedbV1alpha1().Etcds(meta.Namespace).Get(meta.Name, metav1.GetOptions{})
}

func (f *Framework) PatchEtcd(meta metav1.ObjectMeta, transform func(*api.Etcd) *api.Etcd) (*api.Etcd, error) {
	etcd, err := f.extClient.KubedbV1alpha1().Etcds(meta.Namespace).Get(meta.Name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	etcd, _, err = util.PatchEtcd(f.extClient.KubedbV1alpha1(), etcd, transform)
	return etcd, err
}

func (f *Framework) DeleteEtcd(meta metav1.ObjectMeta) error {
	return f.extClient.KubedbV1alpha1().Etcds(meta.Namespace).Delete(meta.Name, deleteInBackground())
}

func (f *Framework) EventuallyEtcd(meta metav1.ObjectMeta) GomegaAsyncAssertion {
	return Eventually(
		func() bool {
			_, err := f.extClient.KubedbV1alpha1().Etcds(meta.Namespace).Get(meta.Name, metav1.GetOptions{})
			if err != nil {
				if kerr.IsNotFound(err) {
					return false
				}
				Expect(err).NotTo(HaveOccurred())
			}
			return true
		},
		time.Minute*10,
		time.Second*5,
	)
}

func (f *Framework) EventuallyEtcdRunning(meta metav1.ObjectMeta) GomegaAsyncAssertion {
	return Eventually(
		func() bool {
			etcd, err := f.extClient.KubedbV1alpha1().Etcds(meta.Namespace).Get(meta.Name, metav1.GetOptions{})
			Expect(err).NotTo(HaveOccurred())
			return etcd.Status.Phase == api.DatabasePhaseRunning
		},
		time.Minute*15,
		time.Second*5,
	)
}

func (f *Framework) CleanEtcd() {
	etcdList, err := f.extClient.KubedbV1alpha1().Etcds(f.namespace).List(metav1.ListOptions{})
	if err != nil {
		return
	}
	for _, e := range etcdList.Items {
		if _, _, err := util.PatchEtcd(f.extClient.KubedbV1alpha1(), &e, func(in *api.Etcd) *api.Etcd {
			in.ObjectMeta.Finalizers = nil
			return in
		}); err != nil {
			fmt.Printf("error Patching Etcd. error: %v", err)
		}
	}
	if err := f.extClient.KubedbV1alpha1().Etcds(f.namespace).DeleteCollection(deleteInBackground(), metav1.ListOptions{}); err != nil {
		fmt.Printf("error in deletion of Etcd. Error: %v", err)
	}
}
