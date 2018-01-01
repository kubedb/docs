package controller

import (
	"fmt"

	"github.com/appscode/kube-mon/agents"
	mona "github.com/appscode/kube-mon/api"
	"github.com/appscode/kutil"
	api "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
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

func (c *Controller) addOrUpdateMonitor(mysql *api.MySQL) (kutil.VerbType, error) {
	agent, err := c.newMonitorController(mysql)
	if err != nil {
		return kutil.VerbUnchanged, err
	}
	return agent.CreateOrUpdate(mysql.StatsAccessor(), mysql.Spec.Monitor)
}

func (c *Controller) deleteMonitor(mysql *api.MySQL) (kutil.VerbType, error) {
	agent, err := c.newMonitorController(mysql)
	if err != nil {
		return kutil.VerbUnchanged, err
	}
	return agent.Delete(mysql.StatsAccessor())
}

// todo: needs to set on status
func (c *Controller) manageMonitor(mysql *api.MySQL) error {
	if mysql.Spec.Monitor != nil {
		_, err := c.addOrUpdateMonitor(mysql)
		if err != nil {
			return err
		}
	} else {
		agent := agents.New(mona.AgentCoreOSPrometheus, c.Client, c.ApiExtKubeClient, c.promClient)
		_, err := agent.CreateOrUpdate(mysql.StatsAccessor(), mysql.Spec.Monitor)
		if err != nil {
			return err
		}
	}
	return nil
}
