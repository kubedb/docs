package controller

import (
	"fmt"
	"time"

	"github.com/appscode/log"
	tapi "github.com/k8sdb/apimachinery/api"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	updateRetryInterval = 10 * time.Second
	maxAttempts         = 5
)

func (c *SnapshotController) UpdateSnapshot(meta metav1.ObjectMeta, transformer func(postgres tapi.Snapshot) tapi.Snapshot) error {
	attempt := 0
	for ; attempt < maxAttempts; attempt = attempt + 1 {
		cur, err := c.extClient.Snapshots(meta.Namespace).Get(meta.Name)
		if err != nil {
			return err
		}

		modified := transformer(*cur)
		if _, err := c.extClient.Snapshots(cur.Namespace).Update(&modified); err == nil {
			return nil
		}

		log.Errorf("Attempt %d failed to update Snapshot %s@%s due to %s.", attempt, cur.Name, cur.Namespace, err)
		time.Sleep(updateRetryInterval)
	}

	return fmt.Errorf("Failed to update Snapshot %s@%s after %d attempts.", meta.Name, meta.Namespace, attempt)
}

func (c *DormantDbController) UpdateDormantDatabase(meta metav1.ObjectMeta, transformer func(postgres tapi.DormantDatabase) tapi.DormantDatabase) error {
	attempt := 0
	for ; attempt < maxAttempts; attempt = attempt + 1 {
		cur, err := c.extClient.DormantDatabases(meta.Namespace).Get(meta.Name)
		if err != nil {
			return err
		}

		modified := transformer(*cur)
		if _, err := c.extClient.DormantDatabases(cur.Namespace).Update(&modified); err == nil {
			return nil
		}

		log.Errorf("Attempt %d failed to update DormantDatabase %s@%s due to %s.", attempt, cur.Name, cur.Namespace, err)
		time.Sleep(updateRetryInterval)
	}

	return fmt.Errorf("Failed to update DormantDatabase %s@%s after %d attempts.", meta.Name, meta.Namespace, attempt)
}
