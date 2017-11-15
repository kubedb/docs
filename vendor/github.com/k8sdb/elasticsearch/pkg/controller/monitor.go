package controller

import (
	"fmt"

	"github.com/appscode/kutil/tools/monitoring/agents"
	mona "github.com/appscode/kutil/tools/monitoring/api"
	api "github.com/k8sdb/apimachinery/apis/kubedb/v1alpha1"
)

func (c *Controller) newMonitorController(elastic *api.Elasticsearch) (mona.Agent, error) {
	monitorSpec := elastic.Spec.Monitor

	if monitorSpec == nil {
		return nil, fmt.Errorf("MonitorSpec not found in %v", elastic.Spec)
	}

	if monitorSpec.Prometheus != nil {
		return agents.New(monitorSpec.Agent, c.Client, c.ApiExtKubeClient, c.promClient), nil
	}

	return nil, fmt.Errorf("monitoring controller not found for %v", monitorSpec)
}

func (c *Controller) addMonitor(elastic *api.Elasticsearch) error {
	agent, err := c.newMonitorController(elastic)
	if err != nil {
		return err
	}
	return agent.Add(elastic.StatsAccessor(), elastic.Spec.Monitor)
}

func (c *Controller) deleteMonitor(elastic *api.Elasticsearch) error {
	agent, err := c.newMonitorController(elastic)
	if err != nil {
		return err
	}
	return agent.Delete(elastic.StatsAccessor(), elastic.Spec.Monitor)
}

func (c *Controller) updateMonitor(oldElastic, updatedElastic *api.Elasticsearch) error {
	var err error
	var agent mona.Agent
	if updatedElastic.Spec.Monitor == nil {
		agent, err = c.newMonitorController(oldElastic)
	} else {
		agent, err = c.newMonitorController(updatedElastic)
	}
	if err != nil {
		return err
	}
	return agent.Update(updatedElastic.StatsAccessor(), oldElastic.Spec.Monitor, updatedElastic.Spec.Monitor)
}
