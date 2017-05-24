package controller

import (
	"fmt"

	tapi "github.com/k8sdb/apimachinery/api"
	amc "github.com/k8sdb/apimachinery/pkg/controller"
)

func (c *Controller) validateElastic(elastic *tapi.Elastic) error {
	if elastic.Spec.Version == "" {
		return fmt.Errorf(`Object 'Version' is missing in '%v'`, elastic.Spec)
	}

	if err := amc.CheckDockerImageVersion(ImageElasticsearch, elastic.Spec.Version); err != nil {
		return fmt.Errorf(`Image %v:%v not found`, ImageElasticsearch, elastic.Spec.Version)
	}

	if err := amc.CheckDockerImageVersion(ImageOperatorElasticsearch, c.operatorTag); err != nil {
		return fmt.Errorf(`Image %v:%v not found`, ImageOperatorElasticsearch, c.operatorTag)
	}

	if elastic.Spec.Storage != nil {
		var err error
		if _, err = c.ValidateStorageSpec(elastic.Spec.Storage); err != nil {
			return err
		}
	}

	backupScheduleSpec := elastic.Spec.BackupSchedule
	if elastic.Spec.BackupSchedule != nil {
		if err := c.ValidateBackupSchedule(backupScheduleSpec); err != nil {
			return err
		}

		if err := c.CheckBucketAccess(backupScheduleSpec.SnapshotStorageSpec, elastic.Namespace); err != nil {
			return err
		}
	}
	return nil
}
