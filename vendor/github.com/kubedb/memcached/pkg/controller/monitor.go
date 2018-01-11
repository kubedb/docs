package controller

import (
	"fmt"

	"github.com/appscode/go/log"
	"github.com/appscode/kube-mon/agents"
	mona "github.com/appscode/kube-mon/api"
	"github.com/appscode/kutil"
	core_util "github.com/appscode/kutil/core/v1"
	meta_util "github.com/appscode/kutil/meta"
	api "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	return agent.CreateOrUpdate(memcached.StatsAccessor(), memcached.Spec.Monitor)
}

func (c *Controller) deleteMonitor(memcached *api.Memcached) (kutil.VerbType, error) {
	agent, err := c.newMonitorController(memcached)
	if err != nil {
		return kutil.VerbUnchanged, err
	}
	return agent.Delete(memcached.StatsAccessor())
}

func (c *Controller) getOldAgent(memcached *api.Memcached) mona.Agent {
	service, err := c.Client.CoreV1().Services(memcached.Namespace).Get(memcached.StatsAccessor().ServiceName(), metav1.GetOptions{})
	if err != nil {
		return nil
	}
	oldAgentType, _ := meta_util.GetString(service.Annotations, mona.KeyAgent)
	return agents.New(mona.AgentType(oldAgentType), c.Client, c.ApiExtKubeClient, c.promClient)
}

func (c *Controller) setNewAgent(memcached *api.Memcached) error {
	service, err := c.Client.CoreV1().Services(memcached.Namespace).Get(memcached.StatsAccessor().ServiceName(), metav1.GetOptions{})
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
		if oldAgent != nil {
			if oldAgent.GetType() != memcached.Spec.Monitor.Agent {
				if _, err := oldAgent.Delete(memcached.StatsAccessor()); err != nil {
					log.Debugf("error in deleting Prometheus agent:", err)
				}
			}
		}
		if _, err := c.addOrUpdateMonitor(memcached); err != nil {
			return err
		}
		return c.setNewAgent(memcached)
	} else if oldAgent != nil {
		if _, err := oldAgent.Delete(memcached.StatsAccessor()); err != nil {
			log.Debugf("error in deleting Prometheus agent:", err)
		}
	}
	return nil
}
