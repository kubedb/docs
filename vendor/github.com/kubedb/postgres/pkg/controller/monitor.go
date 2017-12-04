package controller

import (
	"fmt"

	"github.com/appscode/kutil/tools/monitoring/agents"
	mona "github.com/appscode/kutil/tools/monitoring/api"
	api "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
)

func (c *Controller) newMonitorController(postgres *api.Postgres) (mona.Agent, error) {
	monitorSpec := postgres.Spec.Monitor

	if monitorSpec == nil {
		return nil, fmt.Errorf("MonitorSpec not found in %v", postgres.Spec)
	}

	if monitorSpec.Prometheus != nil {
		return agents.New(monitorSpec.Agent, c.Client, c.ApiExtKubeClient, c.promClient), nil
	}

	return nil, fmt.Errorf("monitoring controller not found for %v", monitorSpec)
}

func (c *Controller) addMonitor(postgres *api.Postgres) error {
	agent, err := c.newMonitorController(postgres)
	if err != nil {
		return err
	}
	return agent.Add(postgres.StatsAccessor(), postgres.Spec.Monitor)
}

func (c *Controller) deleteMonitor(postgres *api.Postgres) error {
	agent, err := c.newMonitorController(postgres)
	if err != nil {
		return err
	}
	return agent.Delete(postgres.StatsAccessor(), postgres.Spec.Monitor)
}

func (c *Controller) updateMonitor(oldPostgres, updatedPostgres *api.Postgres) error {
	var err error
	var agent mona.Agent
	if updatedPostgres.Spec.Monitor == nil {
		agent, err = c.newMonitorController(oldPostgres)
	} else {
		agent, err = c.newMonitorController(updatedPostgres)
	}
	if err != nil {
		return err
	}
	return agent.Update(updatedPostgres.StatsAccessor(), oldPostgres.Spec.Monitor, updatedPostgres.Spec.Monitor)
}
