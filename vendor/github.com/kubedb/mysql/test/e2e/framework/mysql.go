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

func (f *Invocation) MySQL() *api.MySQL {
	return &api.MySQL{
		ObjectMeta: metav1.ObjectMeta{
			Name:      rand.WithUniqSuffix("mysql"),
			Namespace: f.namespace,
			Labels: map[string]string{
				"app": f.app,
			},
		},
		Spec: api.MySQLSpec{
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

func (f *Framework) CreateMySQL(obj *api.MySQL) error {
	_, err := f.extClient.KubedbV1alpha1().MySQLs(obj.Namespace).Create(obj)
	return err
}

func (f *Framework) GetMySQL(meta metav1.ObjectMeta) (*api.MySQL, error) {
	return f.extClient.KubedbV1alpha1().MySQLs(meta.Namespace).Get(meta.Name, metav1.GetOptions{})
}

func (f *Framework) PatchMySQL(meta metav1.ObjectMeta, transform func(*api.MySQL) *api.MySQL) (*api.MySQL, error) {
	mysql, err := f.extClient.KubedbV1alpha1().MySQLs(meta.Namespace).Get(meta.Name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	mysql, _, err = util.PatchMySQL(f.extClient.KubedbV1alpha1(), mysql, transform)
	return mysql, err
}

func (f *Framework) DeleteMySQL(meta metav1.ObjectMeta) error {
	return f.extClient.KubedbV1alpha1().MySQLs(meta.Namespace).Delete(meta.Name, &metav1.DeleteOptions{})
}

func (f *Framework) EventuallyMySQL(meta metav1.ObjectMeta) GomegaAsyncAssertion {
	return Eventually(
		func() bool {
			_, err := f.extClient.KubedbV1alpha1().MySQLs(meta.Namespace).Get(meta.Name, metav1.GetOptions{})
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

func (f *Framework) EventuallyMySQLRunning(meta metav1.ObjectMeta) GomegaAsyncAssertion {
	return Eventually(
		func() bool {
			mysql, err := f.extClient.KubedbV1alpha1().MySQLs(meta.Namespace).Get(meta.Name, metav1.GetOptions{})
			Expect(err).NotTo(HaveOccurred())
			return mysql.Status.Phase == api.DatabasePhaseRunning
		},
		time.Minute*15,
		time.Second*5,
	)
}

func (f *Framework) CleanMySQL() {
	mysqlList, err := f.extClient.KubedbV1alpha1().MySQLs(f.namespace).List(metav1.ListOptions{})
	if err != nil {
		return
	}
	for _, e := range mysqlList.Items {
		if _, _, err := util.PatchMySQL(f.extClient.KubedbV1alpha1(), &e, func(in *api.MySQL) *api.MySQL {
			in.ObjectMeta.Finalizers = nil
			in.Spec.TerminationPolicy = api.TerminationPolicyWipeOut
			return in
		}); err != nil {
			fmt.Printf("error Patching MySQL. error: %v", err)
		}
	}
	if err := f.extClient.KubedbV1alpha1().MySQLs(f.namespace).DeleteCollection(deleteInForeground(), metav1.ListOptions{}); err != nil {
		fmt.Printf("error in deletion of MySQL. Error: %v", err)
	}
}
