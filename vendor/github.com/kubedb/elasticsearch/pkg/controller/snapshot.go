package controller

import (
	"fmt"

	api "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	amv "github.com/kubedb/apimachinery/pkg/validator"
)

func (c *Controller) ValidateSnapshot(snapshot *api.Snapshot) error {
	// Database name can't empty
	databaseName := snapshot.Spec.DatabaseName
	if databaseName == "" {
		return fmt.Errorf(`object 'DatabaseName' is missing in '%v'`, snapshot.Spec)
	}

	if _, err := c.esLister.Elasticsearches(snapshot.Namespace).Get(databaseName); err != nil {
		return err
	}

	return amv.ValidateSnapshotSpec(c.Client, snapshot.Spec.SnapshotStorageSpec, snapshot.Namespace)
}

func (c *Controller) WipeOutSnapshot(snapshot *api.Snapshot) error {
	if snapshot.Spec.Local != nil {
		local := snapshot.Spec.Local
		if local.VolumeSource.EmptyDir != nil {
			return nil
		}
	}
	return c.DeleteSnapshotData(snapshot)
}
