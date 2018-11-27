package controller

import (
	"path/filepath"

	core_util "github.com/appscode/kutil/core/v1"
	"github.com/graymeta/stow"
	_ "github.com/graymeta/stow/azure"
	_ "github.com/graymeta/stow/google"
	_ "github.com/graymeta/stow/s3"
	api "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	batch "k8s.io/api/batch/v1"
	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"kmodules.xyz/objectstore-api/osm"
)

func (c *Controller) DeleteSnapshotData(snapshot *api.Snapshot) error {
	cfg, err := osm.NewOSMContext(c.Client, snapshot.Spec.Backend, snapshot.Namespace)
	if err != nil {
		return err
	}

	loc, err := stow.Dial(cfg.Provider, cfg.Config)
	if err != nil {
		return err
	}
	bucket, err := snapshot.Spec.Backend.Container()
	if err != nil {
		return err
	}
	container, err := loc.Container(bucket)
	if err != nil {
		return err
	}

	prefixLocation, _ := snapshot.Location() // error checked by .Container()
	prefix := filepath.Join(prefixLocation, snapshot.Name)
	cursor := stow.CursorStart
	for {
		items, next, err := container.Items(prefix, cursor, 50)
		if err != nil {
			return err
		}
		for _, item := range items {
			if err := container.RemoveItem(item.ID()); err != nil {
				return err
			}
		}
		cursor = next
		if stow.IsCursorEnd(cursor) {
			break
		}
	}

	return nil
}

func (c *Controller) checkGoverningService(name, namespace string) (bool, error) {
	_, err := c.Client.CoreV1().Services(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		if kerr.IsNotFound(err) {
			return false, nil
		} else {
			return false, err
		}
	}

	return true, nil
}

func (c *Controller) CreateGoverningService(name, namespace string) error {
	// Check if service name exists
	found, err := c.checkGoverningService(name, namespace)
	if err != nil {
		return err
	}
	if found {
		return nil
	}

	service := &core.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: core.ServiceSpec{
			Type:      core.ServiceTypeClusterIP,
			ClusterIP: core.ClusterIPNone,
		},
	}
	_, err = c.Client.CoreV1().Services(namespace).Create(service)
	return err
}

func (c *Controller) SetJobOwnerReference(snapshot *api.Snapshot, job *batch.Job) error {
	secret, err := c.Client.CoreV1().Secrets(snapshot.Namespace).Get(snapshot.OSMSecretName(), metav1.GetOptions{})
	if err != nil {
		if !kerr.IsNotFound(err) {
			return err
		}
	} else {
		_, _, err := core_util.PatchSecret(c.Client, secret, func(in *core.Secret) *core.Secret {
			in.SetOwnerReferences([]metav1.OwnerReference{
				{
					APIVersion: batch.SchemeGroupVersion.String(),
					Kind:       "Job",
					Name:       job.Name,
					UID:        job.UID,
				},
			})
			return in
		})
		if err != nil {
			return err
		}
	}

	pvc, err := c.Client.CoreV1().PersistentVolumeClaims(snapshot.Namespace).Get(job.Name, metav1.GetOptions{})
	if err != nil {
		if !kerr.IsNotFound(err) {
			return err
		}
	} else {
		_, _, err := core_util.PatchPVC(c.Client, pvc, func(in *core.PersistentVolumeClaim) *core.PersistentVolumeClaim {
			in.SetOwnerReferences([]metav1.OwnerReference{
				{
					APIVersion: batch.SchemeGroupVersion.String(),
					Kind:       "Job",
					Name:       job.Name,
					UID:        job.UID,
				},
			})
			return in
		})
		if err != nil {
			return err
		}
	}
	return nil
}
