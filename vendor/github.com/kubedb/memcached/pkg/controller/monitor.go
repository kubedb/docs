package controller

import (
	"fmt"

	"github.com/appscode/go/log"
	api "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kutil "kmodules.xyz/client-go"
	core_util "kmodules.xyz/client-go/core/v1"
	meta_util "kmodules.xyz/client-go/meta"
	"kmodules.xyz/monitoring-agent-api/agents"
	mona "kmodules.xyz/monitoring-agent-api/api/v1"
)

func (c *Controller) newMonitorController(memcached *api.Memcached) (mona.Agent, error) {
	monitorSpec := memcached.Spec.Monitor

	if monitorSpec == nil {
		return nil, fmt.Errorf("MonitorSpec not found in %v", memcached.Spec)
	}

	if monitorSpec.Prometheus != nil {
		return agents.New(monitorSpec.Agent, c.Client, c.ApiExtKubeClient, c.promClient), nil
	}

	return nil, fmt.Errorf("monitoring controller not found for %v", monitorSpec)
}

func (c *Controller) addOrUpdateMonitor(memcached *api.Memcached) (kutil.VerbType, error) {
	agent, err := c.newMonitorController(memcached)
	if err != nil {
		return kutil.VerbUnchanged, err
	}
	return agent.CreateOrUpdate(memcached.StatsService(), memcached.Spec.Monitor)
}

func (c *Controller) deleteMonitor(memcached *api.Memcached) (kutil.VerbType, error) {
	agent, err := c.newMonitorController(memcached)
	if err != nil {
		return kutil.VerbUnchanged, err
	}
	return agent.Delete(memcached.StatsService())
}

func (c *Controller) getOldAgent(memcached *api.Memcached) mona.Agent {
	service, err := c.Client.CoreV1().Services(memcached.Namespace).Get(memcached.StatsService().ServiceName(), metav1.GetOptions{})
	if err != nil {
		return nil
	}
	oldAgentType, _ := meta_util.GetStringValue(service.Annotations, mona.KeyAgent)
	return agents.New(mona.AgentType(oldAgentType), c.Client, c.ApiExtKubeClient, c.promClient)
}

func (c *Controller) setNewAgent(memcached *api.Memcached) error {
	service, err := c.Client.CoreV1().Services(memcached.Namespace).Get(memcached.StatsService().ServiceName(), metav1.GetOptions{})
	if err != nil {
		return err
	}
	_, _, err = core_util.PatchService(c.Client, service, func(in *core.Service) *core.Service {
		in.Annotations = core_util.UpsertMap(in.Annotations, map[string]string{
			mona.KeyAgent: string(memcached.Spec.Monitor.Agent),
		},
		)
		return in
	})
	return err
}

func (c *Controller) manageMonitor(memcached *api.Memcached) error {
	oldAgent := c.getOldAgent(memcached)
	if memcached.Spec.Monitor != nil {
		if oldAgent != nil &&
			oldAgent.GetType() != memcached.Spec.Monitor.Agent {
			if _, err := oldAgent.Delete(memcached.StatsService()); err != nil {
				log.Error("error in deleting Prometheus agent:", err)
			}
		}
		if _, err := c.addOrUpdateMonitor(memcached); err != nil {
			return err
		}
		return c.setNewAgent(memcached)
	} else if oldAgent != nil {
		if _, err := oldAgent.Delete(memcached.StatsService()); err != nil {
			log.Error("error in deleting Prometheus agent:", err)
		}
	}
	return nil
}
