package controller

import (
	"fmt"
	"time"

	"github.com/appscode/log"
	tapi "github.com/k8sdb/apimachinery/apis/kubedb/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	updateRetryInterval = 10 * 1000 * 1000 * time.Nanosecond
	maxAttempts         = 5
)

func (c *Controller) UpdatePostgres(
	meta metav1.ObjectMeta,
	transformer func(tapi.Postgres) tapi.Postgres,
) (*tapi.Postgres, error) {
	attempt := 0
	for ; attempt < maxAttempts; attempt = attempt + 1 {
		cur, err := c.ExtClient.Postgreses(meta.Namespace).Get(meta.Name, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}

		modified := transformer(*cur)
		if postgres, err := c.ExtClient.Postgreses(cur.Namespace).Update(&modified); err == nil {
			return postgres, nil
		}

		log.Errorf("Attempt %d failed to update Postgres %s@%s due to %s.", attempt, cur.Name, cur.Namespace, err)
		time.Sleep(updateRetryInterval)
	}

	return nil, fmt.Errorf("Failed to update Postgres %s@%s after %d attempts.", meta.Name, meta.Namespace, attempt)
}
