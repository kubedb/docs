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
	"context"
	"fmt"

	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"

	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	kutil "kmodules.xyz/client-go"
	core_util "kmodules.xyz/client-go/core/v1"
	meta_util "kmodules.xyz/client-go/meta"
	"kmodules.xyz/monitoring-agent-api/agents"
	mona "kmodules.xyz/monitoring-agent-api/api/v1"
)

func (c *Controller) newMonitorController(db *api.MongoDB) (mona.Agent, error) {
	monitorSpec := db.Spec.Monitor

	if monitorSpec == nil {
		return nil, fmt.Errorf("MonitorSpec not found in %v", db.Spec)
	}

	if monitorSpec.Prometheus != nil {
		return agents.New(monitorSpec.Agent, c.Client, c.promClient)
	}

	return nil, fmt.Errorf("monitoring controller not found for %v", monitorSpec)
}

func (c *Controller) addOrUpdateMonitor(db *api.MongoDB) (kutil.VerbType, error) {
	agent, err := c.newMonitorController(db)
	if err != nil {
		return kutil.VerbUnchanged, err
	}
	return agent.CreateOrUpdate(db.StatsService(), db.Spec.Monitor)
}

func (c *Controller) deleteMonitor(db *api.MongoDB) error {
	agent, err := c.newMonitorController(db)
	if err != nil {
		return err
	}
	_, err = agent.Delete(db.StatsService())
	return err
}

func (c *Controller) getOldAgent(db *api.MongoDB) mona.Agent {
	service, err := c.Client.CoreV1().Services(db.Namespace).Get(context.TODO(), db.StatsService().ServiceName(), metav1.GetOptions{})
	if err != nil {
		return nil
	}
	oldAgentType, _ := meta_util.GetStringValue(service.Annotations, mona.KeyAgent)
	agent, _ := agents.New(mona.AgentType(oldAgentType), c.Client, c.promClient)
	return agent
}

func (c *Controller) setNewAgent(db *api.MongoDB) error {
	service, err := c.Client.CoreV1().Services(db.Namespace).Get(context.TODO(), db.StatsService().ServiceName(), metav1.GetOptions{})
	if err != nil {
		return err
	}
	_, _, err = core_util.PatchService(context.TODO(), c.Client, service, func(in *core.Service) *core.Service {
		in.Annotations = core_util.UpsertMap(in.Annotations, map[string]string{
			mona.KeyAgent: string(db.Spec.Monitor.Agent),
		},
		)
		return in
	}, metav1.PatchOptions{})
	return err
}

func (c *Controller) manageMonitor(db *api.MongoDB) error {
	oldAgent := c.getOldAgent(db)
	if db.Spec.Monitor != nil {
		if oldAgent != nil && oldAgent.GetType() != db.Spec.Monitor.Agent {
			if _, err := oldAgent.Delete(db.StatsService()); err != nil {
				klog.Errorf("error in deleting Prometheus agent. Reason: %v", err.Error())
			}
		}
		if _, err := c.addOrUpdateMonitor(db); err != nil {
			return err
		}
		return c.setNewAgent(db)
	} else if oldAgent != nil {
		if _, err := oldAgent.Delete(db.StatsService()); err != nil {
			klog.Errorf("error in deleting Prometheus agent. Reason: %v", err.Error())
		}
	}
	return nil
}
