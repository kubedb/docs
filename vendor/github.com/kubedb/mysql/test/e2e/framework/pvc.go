package framework

import (
	"time"

	api "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	. "github.com/onsi/gomega"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

func (f *Framework) EventuallyPVCCount(meta metav1.ObjectMeta) GomegaAsyncAssertion {
	labelMap := map[string]string{
		api.LabelDatabaseKind: api.ResourceKindMySQL,
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

func (f *Invocation) GetPersistentVolumeClaim() *core.PersistentVolumeClaim {
	return &core.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      f.app,
			Namespace: f.namespace,
		},
		Spec: core.PersistentVolumeClaimSpec{
			AccessModes: []core.PersistentVolumeAccessMode{
				core.ReadWriteOnce,
			},
			StorageClassName: &f.StorageClass,
			Resources: core.ResourceRequirements{
				Requests: core.ResourceList{
					core.ResourceName(core.ResourceStorage): resource.MustParse("50Mi"),
				},
			},
		},
	}
}

func (f *Invocation) CreatePersistentVolumeClaim(pvc *core.PersistentVolumeClaim) error {
	_, err := f.kubeClient.CoreV1().PersistentVolumeClaims(pvc.Namespace).Create(pvc)
	return err
}

func (f *Invocation) DeletePersistentVolumeClaim(meta metav1.ObjectMeta) error {
	return f.kubeClient.CoreV1().PersistentVolumeClaims(meta.Namespace).Delete(meta.Name, deleteInForeground())
}
