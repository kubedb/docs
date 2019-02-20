package controller

import (
	"fmt"

	core_util "github.com/appscode/kutil/core/v1"
	"github.com/kubedb/apimachinery/apis"
	api "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	"github.com/kubedb/apimachinery/client/clientset/versioned/typed/kubedb/v1alpha1/util"
	amv "github.com/kubedb/apimachinery/pkg/validator"
	batch "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func (c *Controller) GetDatabase(meta metav1.ObjectMeta) (runtime.Object, error) {
	mysql, err := c.myLister.MySQLs(meta.Namespace).Get(meta.Name)
	if err != nil {
		return nil, err
	}

	return mysql, nil
}

func (c *Controller) SetDatabaseStatus(meta metav1.ObjectMeta, phase api.DatabasePhase, reason string) error {
	mysql, err := c.myLister.MySQLs(meta.Namespace).Get(meta.Name)
	if err != nil {
		return err
	}
	_, err = util.UpdateMySQLStatus(c.ExtClient.KubedbV1alpha1(), mysql, func(in *api.MySQLStatus) *api.MySQLStatus {
		in.Phase = phase
		in.Reason = reason
		return in
	}, apis.EnableStatusSubresource)
	return err
}

func (c *Controller) UpsertDatabaseAnnotation(meta metav1.ObjectMeta, annotation map[string]string) error {
	mysql, err := c.myLister.MySQLs(meta.Namespace).Get(meta.Name)
	if err != nil {
		return err
	}

	_, _, err = util.PatchMySQL(c.ExtClient.KubedbV1alpha1(), mysql, func(in *api.MySQL) *api.MySQL {
		in.Annotations = core_util.UpsertMap(mysql.Annotations, annotation)
		return in
	})
	return err
}

func (c *Controller) ValidateSnapshot(snapshot *api.Snapshot) error {
	// Database name can't empty
	databaseName := snapshot.Spec.DatabaseName
	if databaseName == "" {
		return fmt.Errorf(`object 'DatabaseName' is missing in '%v'`, snapshot.Spec)
	}

	if _, err := c.myLister.MySQLs(snapshot.Namespace).Get(databaseName); err != nil {
		return err
	}

	return amv.ValidateSnapshotSpec(snapshot.Spec.Backend)
}

func (c *Controller) GetSnapshotter(snapshot *api.Snapshot) (*batch.Job, error) {
	return c.getSnapshotterJob(snapshot)
}

func (c *Controller) WipeOutSnapshot(snapshot *api.Snapshot) error {
	// wipeOut not possible for local backend.
	// Ref: https://github.com/kubedb/project/issues/261
	if snapshot.Spec.Local != nil {
		return nil
	}
	return c.DeleteSnapshotData(snapshot)
}
