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

	"github.com/appscode/go/log"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kutil "kmodules.xyz/client-go"
	core_util "kmodules.xyz/client-go/core/v1"
	meta_util "kmodules.xyz/client-go/meta"
	"kmodules.xyz/monitoring-agent-api/agents"
	mona "kmodules.xyz/monitoring-agent-api/api/v1"
)

func (c *Controller) newMonitorController(mongodb *api.MongoDB) (mona.Agent, error) {
	monitorSpec := mongodb.Spec.Monitor

	if monitorSpec == nil {
		return nil, fmt.Errorf("MonitorSpec not found in %v", mongodb.Spec)
	}

	if monitorSpec.Prometheus != nil {
		return agents.New(monitorSpec.Agent, c.Client, c.ApiExtKubeClient, c.promClient), nil
	}

	return nil, fmt.Errorf("monitoring controller not found for %v", monitorSpec)
}

func (c *Controller) addOrUpdateMonitor(mongodb *api.MongoDB) (kutil.VerbType, error) {
	agent, err := c.newMonitorController(mongodb)
	if err != nil {
		return kutil.VerbUnchanged, err
	}
	return agent.CreateOrUpdate(mongodb.StatsService(), mongodb.Spec.Monitor)
}

func (c *Controller) deleteMonitor(mongodb *api.MongoDB) (kutil.VerbType, error) {
	agent, err := c.newMonitorController(mongodb)
	if err != nil {
		return kutil.VerbUnchanged, err
	}
	return agent.Delete(mongodb.StatsService())
}

func (c *Controller) getOldAgent(mongodb *api.MongoDB) mona.Agent {
	service, err := c.Client.CoreV1().Services(mongodb.Namespace).Get(mongodb.StatsService().ServiceName(), metav1.GetOptions{})
	if err != nil {
		return nil
	}
	oldAgentType, _ := meta_util.GetStringValue(service.Annotations, mona.KeyAgent)
	return agents.New(mona.AgentType(oldAgentType), c.Client, c.ApiExtKubeClient, c.promClient)
}

func (c *Controller) setNewAgent(mongodb *api.MongoDB) error {
	service, err := c.Client.CoreV1().Services(mongodb.Namespace).Get(mongodb.StatsService().ServiceName(), metav1.GetOptions{})
	if err != nil {
		return err
	}
	_, _, err = core_util.PatchService(c.Client, service, func(in *core.Service) *core.Service {
		in.Annotations = core_util.UpsertMap(in.Annotations, map[string]string{
			mona.KeyAgent: string(mongodb.Spec.Monitor.Agent),
		},
		)
		return in
	})
	return err
}

func (c *Controller) manageMonitor(mongodb *api.MongoDB) error {
	oldAgent := c.getOldAgent(mongodb)
	if mongodb.Spec.Monitor != nil {
		if oldAgent != nil && oldAgent.GetType() != mongodb.Spec.Monitor.Agent {
			if _, err := oldAgent.Delete(mongodb.StatsService()); err != nil {
				log.Errorf("error in deleting Prometheus agent. Reason: %v", err.Error())
			}
		}
		if _, err := c.addOrUpdateMonitor(mongodb); err != nil {
			return err
		}
		return c.setNewAgent(mongodb)
	} else if oldAgent != nil {
		if _, err := oldAgent.Delete(mongodb.StatsService()); err != nil {
			log.Errorf("error in deleting Prometheus agent. Reason: %v", err.Error())
		}
	}
	return nil
}
