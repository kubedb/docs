/*
Copyright AppsCode Inc. and Contributors

Licensed under the AppsCode Community License 1.0.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://github.com/appscode/licenses/raw/1.0.0/AppsCode-Community-1.0.0.md

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	amc "kubedb.dev/apimachinery/pkg/controller"

	pcm "github.com/prometheus-operator/prometheus-operator/pkg/client/versioned/typed/monitoring/v1"
)

type Reconciler struct {
	amc.Config
	*amc.Controller

	// Prometheus client
	PromClient pcm.MonitoringV1Interface
	enableIPv6 bool
}

func (c *Controller) getReconciler() Reconciler {
	return Reconciler{
		Config:     c.Config,
		Controller: c.Controller,
		PromClient: c.promClient,
	}
}
