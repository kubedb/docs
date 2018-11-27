package framework

import (
	"time"

	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (f *Framework) EventuallyCRD() GomegaAsyncAssertion {
	return Eventually(
		func() error {
			// Check Memcached TPR
			if _, err := f.extClient.KubedbV1alpha1().Memcacheds(core.NamespaceAll).List(metav1.ListOptions{}); err != nil {
				return errors.New("CRD Memcached is not ready")
			}

			// Check DormantDatabases TPR
			if _, err := f.extClient.KubedbV1alpha1().DormantDatabases(core.NamespaceAll).List(metav1.ListOptions{}); err != nil {
				return errors.New("CRD DormantDatabase is not ready")
			}

			return nil
		},
		time.Minute*2,
		time.Second*10,
	)
}
