package controller

import (
	"fmt"

	"github.com/appscode/kutil/tools/monitoring/agents"
	mona "github.com/appscode/kutil/tools/monitoring/api"
	api "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
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

func (c *Controller) addMonitor(mongodb *api.MongoDB) error {
	agent, err := c.newMonitorController(mongodb)
	if err != nil {
		return err
	}
	return agent.Add(mongodb.StatsAccessor(), mongodb.Spec.Monitor)
}

func (c *Controller) deleteMonitor(mongodb *api.MongoDB) error {
	agent, err := c.newMonitorController(mongodb)
	if err != nil {
		return err
	}
	return agent.Delete(mongodb.StatsAccessor(), mongodb.Spec.Monitor)
}

func (c *Controller) updateMonitor(oldMongoDB, updatedMongoDB *api.MongoDB) error {
	var err error
	var agent mona.Agent
	if updatedMongoDB.Spec.Monitor == nil {
		agent, err = c.newMonitorController(oldMongoDB)
	} else {
		agent, err = c.newMonitorController(updatedMongoDB)
	}
	if err != nil {
		return err
	}
	return agent.Update(updatedMongoDB.StatsAccessor(), oldMongoDB.Spec.Monitor, updatedMongoDB.Spec.Monitor)
}
