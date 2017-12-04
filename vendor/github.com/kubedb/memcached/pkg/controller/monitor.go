package controller

import (
	"fmt"

	"github.com/appscode/kutil/tools/monitoring/agents"
	mona "github.com/appscode/kutil/tools/monitoring/api"
	api "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
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

func (c *Controller) addMonitor(memcached *api.Memcached) error {
	agent, err := c.newMonitorController(memcached)
	if err != nil {
		return err
	}
	return agent.Add(memcached.StatsAccessor(), memcached.Spec.Monitor)
}

func (c *Controller) deleteMonitor(memcached *api.Memcached) error {
	agent, err := c.newMonitorController(memcached)
	if err != nil {
		return err
	}
	return agent.Delete(memcached.StatsAccessor(), memcached.Spec.Monitor)
}

func (c *Controller) updateMonitor(oldMemcached, updatedMemcached *api.Memcached) error {
	var err error
	var agent mona.Agent
	if updatedMemcached.Spec.Monitor == nil {
		agent, err = c.newMonitorController(oldMemcached)
	} else {
		agent, err = c.newMonitorController(updatedMemcached)
	}
	if err != nil {
		return err
	}
	return agent.Update(updatedMemcached.StatsAccessor(), oldMemcached.Spec.Monitor, updatedMemcached.Spec.Monitor)
}
