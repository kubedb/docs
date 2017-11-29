package controller

import (
	"fmt"

	"github.com/appscode/kutil/tools/monitoring/agents"
	mona "github.com/appscode/kutil/tools/monitoring/api"
	api "github.com/k8sdb/apimachinery/apis/kubedb/v1alpha1"
)

func (c *Controller) newMonitorController(mysql *api.MySQL) (mona.Agent, error) {
	monitorSpec := mysql.Spec.Monitor

	if monitorSpec == nil {
		return nil, fmt.Errorf("MonitorSpec not found in %v", mysql.Spec)
	}

	if monitorSpec.Prometheus != nil {
		return agents.New(monitorSpec.Agent, c.Client, c.ApiExtKubeClient, c.promClient), nil
	}

	return nil, fmt.Errorf("monitoring controller not found for %v", monitorSpec)
}

func (c *Controller) addMonitor(mysql *api.MySQL) error {
	agent, err := c.newMonitorController(mysql)
	if err != nil {
		return err
	}
	return agent.Add(mysql.StatsAccessor(), mysql.Spec.Monitor)
}

func (c *Controller) deleteMonitor(mysql *api.MySQL) error {
	agent, err := c.newMonitorController(mysql)
	if err != nil {
		return err
	}
	return agent.Delete(mysql.StatsAccessor(), mysql.Spec.Monitor)
}

func (c *Controller) updateMonitor(oldMySQL, updatedMySQL *api.MySQL) error {
	var err error
	var agent mona.Agent
	if updatedMySQL.Spec.Monitor == nil {
		agent, err = c.newMonitorController(oldMySQL)
	} else {
		agent, err = c.newMonitorController(updatedMySQL)
	}
	if err != nil {
		return err
	}
	return agent.Update(updatedMySQL.StatsAccessor(), oldMySQL.Spec.Monitor, updatedMySQL.Spec.Monitor)
}
