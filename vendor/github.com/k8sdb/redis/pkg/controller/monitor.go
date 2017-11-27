package controller

import (
	"fmt"

	"github.com/appscode/kutil/tools/monitoring/agents"
	mona "github.com/appscode/kutil/tools/monitoring/api"
	api "github.com/k8sdb/apimachinery/apis/kubedb/v1alpha1"
)

func (c *Controller) newMonitorController(redis *api.Redis) (mona.Agent, error) {
	monitorSpec := redis.Spec.Monitor

	if monitorSpec == nil {
		return nil, fmt.Errorf("MonitorSpec not found in %v", redis.Spec)
	}

	if monitorSpec.Prometheus != nil {
		return agents.New(monitorSpec.Agent, c.Client, c.ApiExtKubeClient, c.promClient), nil
	}

	return nil, fmt.Errorf("monitoring controller not found for %v", monitorSpec)
}

func (c *Controller) addMonitor(redis *api.Redis) error {
	agent, err := c.newMonitorController(redis)
	if err != nil {
		return err
	}
	return agent.Add(redis.StatsAccessor(), redis.Spec.Monitor)
}

func (c *Controller) deleteMonitor(redis *api.Redis) error {
	agent, err := c.newMonitorController(redis)
	if err != nil {
		return err
	}
	return agent.Delete(redis.StatsAccessor(), redis.Spec.Monitor)
}

func (c *Controller) updateMonitor(oldRedis, updatedRedis *api.Redis) error {
	var err error
	var agent mona.Agent
	if updatedRedis.Spec.Monitor == nil {
		agent, err = c.newMonitorController(oldRedis)
	} else {
		agent, err = c.newMonitorController(updatedRedis)
	}
	if err != nil {
		return err
	}
	return agent.Update(updatedRedis.StatsAccessor(), oldRedis.Spec.Monitor, updatedRedis.Spec.Monitor)
}
