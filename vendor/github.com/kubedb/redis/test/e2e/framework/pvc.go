package framework

import (
	"time"

	api "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

func (f *Framework) EventuallyPVCCount(meta metav1.ObjectMeta) GomegaAsyncAssertion {
	labelMap := map[string]string{
		api.LabelDatabaseKind: api.ResourceKindMemcached,
		api.LabelDatabaseName: meta.Name,
	}
	labelSelector := labels.SelectorFromSet(labelMap)

	return Eventually(
		func() int {
			pvcList, err := f.kubeClient.CoreV1().PersistentVolumeClaims(meta.Namespace).List(
				metav1.ListOptions{
					LabelSelector: labelSelector.String(),
				},
			)
			Expect(err).NotTo(HaveOccurred())

			return len(pvcList.Items)
		},
		time.Minute*5,
		time.Second*5,
	)
}
