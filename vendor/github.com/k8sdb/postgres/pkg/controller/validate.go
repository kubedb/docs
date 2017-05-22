package controller

import (
	"fmt"

	tapi "github.com/k8sdb/apimachinery/api"
	amc "github.com/k8sdb/apimachinery/pkg/controller"
)

func (c *Controller) validatePostgres(postgres *tapi.Postgres) error {
	if postgres.Spec.Version == "" {
		return fmt.Errorf(`Object 'Version' is missing in '%v'`, postgres.Spec)
	}

	if err := amc.CheckDockerImageVersion(ImagePostgres, postgres.Spec.Version); err != nil {
		return fmt.Errorf(`Image %v:%v not found`, ImagePostgres, postgres.Spec.Version)
	}

	if postgres.Spec.Storage != nil {
		var err error
		if _, err = c.ValidateStorageSpec(postgres.Spec.Storage); err != nil {
			return err
		}
	}

	backupScheduleSpec := postgres.Spec.BackupSchedule
	if postgres.Spec.BackupSchedule != nil {
		if err := c.ValidateBackupSchedule(backupScheduleSpec); err != nil {
			return err
		}

		if err := c.CheckBucketAccess(backupScheduleSpec.SnapshotStorageSpec, postgres.Namespace); err != nil {
			return err
		}
	}
	return nil
}
