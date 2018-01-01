package controller

import (
	"fmt"

	"github.com/appscode/kube-mon/agents"
	mona "github.com/appscode/kube-mon/api"
	"github.com/appscode/kutil"
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

func (c *Controller) addOrUpdateMonitor(postgres *api.Postgres) (kutil.VerbType, error) {
	agent, err := c.newMonitorController(postgres)
	if err != nil {
		return kutil.VerbUnchanged, err
	}
	return agent.CreateOrUpdate(postgres.StatsAccessor(), postgres.Spec.Monitor)
}

func (c *Controller) deleteMonitor(postgres *api.Postgres) (kutil.VerbType, error) {
	agent, err := c.newMonitorController(postgres)
	if err != nil {
		return kutil.VerbUnchanged, err
	}
	return agent.Delete(postgres.StatsAccessor())
}

// todo: needs to set on status
func (c *Controller) manageMonitor(postgres *api.Postgres) error {
	if postgres.Spec.Monitor != nil {
		_, err := c.addOrUpdateMonitor(postgres)
		if err != nil {
			return err
		}
	} else {
		agent := agents.New(mona.AgentCoreOSPrometheus, c.Client, c.ApiExtKubeClient, c.promClient)
		_, err := agent.CreateOrUpdate(postgres.StatsAccessor(), postgres.Spec.Monitor)
		if err != nil {
			return err
		}
	}
	return nil
}
