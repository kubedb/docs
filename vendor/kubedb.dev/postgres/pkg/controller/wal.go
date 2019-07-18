package controller

import (
	"fmt"
	"os"
	"path/filepath"

	"gomodules.xyz/stow"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"kmodules.xyz/objectstore-api/osm"
	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha1"
)

func WalDataDir(postgres *api.Postgres) string {
	spec := postgres.Spec.Archiver.Storage
	if spec.S3 != nil {
		return filepath.Join(spec.S3.Prefix, api.DatabaseNamePrefix, postgres.Namespace, postgres.Name, "archive")
	} else if spec.GCS != nil {
		return filepath.Join(spec.GCS.Prefix, api.DatabaseNamePrefix, postgres.Namespace, postgres.Name, "archive")
	} else if spec.Azure != nil {
		return filepath.Join(spec.Azure.Prefix, api.DatabaseNamePrefix, postgres.Namespace, postgres.Name, "archive")
	} else if spec.Swift != nil {
		return filepath.Join(spec.Swift.Prefix, api.DatabaseNamePrefix, postgres.Namespace, postgres.Name, "archive")
	} else if spec.Local != nil {
		return os.Getenv("RESTORE_FILE_PREFIX") //never gets called
	}
	return ""
}

func (c *Controller) wipeOutWalData(meta metav1.ObjectMeta, spec *api.PostgresSpec) error {
	if spec == nil {
		return fmt.Errorf("wipeout wal data failed. Reason: invalid postgres spec")
	}

	postgres := &api.Postgres{
		ObjectMeta: meta,
		Spec:       *spec,
	}

	if postgres.Spec.Archiver == nil {
		// no archiver was configured. nothing to remove.
		return nil
	}
	if postgres.Spec.Archiver.Storage.Local != nil {
		// Do not remove local Data.
		return nil
	}
	cfg, err := osm.NewOSMContext(c.Client, *postgres.Spec.Archiver.Storage, postgres.Namespace)
	if err != nil {
		return err
	}

	loc, err := stow.Dial(cfg.Provider, cfg.Config)
	if err != nil {
		return err
	}
	bucket, err := postgres.Spec.Archiver.Storage.Container()
	if err != nil {
		return err
	}
	container, err := loc.Container(bucket)
	if err != nil {
		return err
	}

	prefix := WalDataDir(postgres)
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
