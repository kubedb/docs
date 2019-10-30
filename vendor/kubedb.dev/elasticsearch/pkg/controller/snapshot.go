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
package controller

import (
	"fmt"

	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha1"
	amv "kubedb.dev/apimachinery/pkg/validator"
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

	return amv.ValidateSnapshotSpec(snapshot.Spec.Backend)
}

func (c *Controller) WipeOutSnapshot(snapshot *api.Snapshot) error {
	// wipeOut not possible for local backend.
	// Ref: https://github.com/kubedb/project/issues/261
	if snapshot.Spec.Local != nil {
		return nil
	}
	return c.DeleteSnapshotData(snapshot)
}
