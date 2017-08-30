package controller

import (
	"fmt"
	"time"

	"github.com/appscode/log"
	tapi "github.com/k8sdb/apimachinery/api"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	updateRetryInterval = 10 * 1000 * 1000 * time.Nanosecond
	maxAttempts         = 5
)

func (c *Controller) UpdateElasticsearch(
	meta metav1.ObjectMeta,
	transformer func(tapi.Elasticsearch) tapi.Elasticsearch,
) (*tapi.Elasticsearch, error) {
	attempt := 0
	for ; attempt < maxAttempts; attempt = attempt + 1 {
		cur, err := c.ExtClient.Elasticsearches(meta.Namespace).Get(meta.Name)
		if err != nil {
			return nil, err
		}

		modified := transformer(*cur)
		if elasticsearch, err := c.ExtClient.Elasticsearches(cur.Namespace).Update(&modified); err == nil {
			return elasticsearch, nil
		}

		log.Errorf("Attempt %d failed to update Elasticsearche %s@%s due to %s.", attempt, cur.Name, cur.Namespace, err)
		time.Sleep(updateRetryInterval)
	}

	return nil, fmt.Errorf("Failed to update Elasticsearche %s@%s after %d attempts.", meta.Name, meta.Namespace, attempt)
}
