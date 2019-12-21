/*
Copyright The KubeDB Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package job

import (
	"github.com/appscode/go/log"
	"github.com/appscode/go/types"
	batch "k8s.io/api/batch/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
	"kmodules.xyz/client-go/tools/queue"
)

func (c *Controller) addEventHandler(selector labels.Selector) {
	c.JobQueue = queue.New("Job", c.MaxNumRequeues, c.NumThreads, c.runJob)
	c.jobLister = c.KubeInformerFactory.Batch().V1().Jobs().Lister()
	c.JobInformer.AddEventHandler(queue.NewFilteredHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			job := obj.(*batch.Job)
			if job.Status.Succeeded > 0 || job.Status.Failed > types.Int32(job.Spec.BackoffLimit) {
				queue.Enqueue(c.JobQueue.GetQueue(), obj)
			}
		},
		UpdateFunc: func(old interface{}, new interface{}) {
			oldObj := old.(*batch.Job)
			newObj := new.(*batch.Job)
			if isJobCompleted(oldObj, newObj) {
				queue.Enqueue(c.JobQueue.GetQueue(), new)
			}
		},
		DeleteFunc: func(obj interface{}) {
			job, ok := obj.(*batch.Job)
			if !ok {
				log.Warningln("Invalid Job object")
				return
			}
			if job.Status.Succeeded == 0 && job.Status.Failed <= types.Int32(job.Spec.BackoffLimit) {
				queue.Enqueue(c.JobQueue.GetQueue(), obj)
			}
		},
	}, selector))
}

func isJobCompleted(old, new *batch.Job) bool {
	if old.Status.Succeeded == 0 && new.Status.Succeeded > 0 {
		return true
	}
	if old.Status.Failed < types.Int32(old.Spec.BackoffLimit) && new.Status.Failed >= types.Int32(new.Spec.BackoffLimit) {
		return true
	}
	return false
}

func (c *Controller) runJob(key string) error {
	log.Debugf("started processing, key: %v", key)
	obj, exists, err := c.JobInformer.GetIndexer().GetByKey(key)
	if err != nil {
		log.Errorf("Fetching object with key %s from store failed with %v", key, err)
		return err
	}

	if !exists {
		log.Debugf("Job %s does not exist anymore", key)
	} else {
		// Note that you also have to check the uid if you have a local controlled resource, which
		// is dependent on the actual instance, to detect that a Job was recreated with the same name
		job := obj.(*batch.Job).DeepCopy()
		if err := c.completeJob(job); err != nil {
			log.Errorln(err)
			return err
		}
	}
	return nil
}
