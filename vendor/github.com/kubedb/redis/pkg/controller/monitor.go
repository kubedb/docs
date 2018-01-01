package controller

import (
	"fmt"

	"github.com/appscode/kube-mon/agents"
	mona "github.com/appscode/kube-mon/api"
	"github.com/appscode/kutil"
	api "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
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

func (c *Controller) addOrUpdateMonitor(redis *api.Redis) (kutil.VerbType, error) {
	agent, err := c.newMonitorController(redis)
	if err != nil {
		return kutil.VerbUnchanged, err
	}
	return agent.CreateOrUpdate(redis.StatsAccessor(), redis.Spec.Monitor)
}

func (c *Controller) deleteMonitor(redis *api.Redis) (kutil.VerbType, error) {
	agent, err := c.newMonitorController(redis)
	if err != nil {
		return kutil.VerbUnchanged, err
	}
	return agent.Delete(redis.StatsAccessor())
}

// todo: needs to set on status
func (c *Controller) manageMonitor(redis *api.Redis) error {
	if redis.Spec.Monitor != nil {
		_, err := c.addOrUpdateMonitor(redis)
		if err != nil {
			return err
		}
	} else {
		agent := agents.New(mona.AgentCoreOSPrometheus, c.Client, c.ApiExtKubeClient, c.promClient)
		_, err := agent.CreateOrUpdate(redis.StatsAccessor(), redis.Spec.Monitor)
		if err != nil {
			return err
		}
	}
	return nil
}
