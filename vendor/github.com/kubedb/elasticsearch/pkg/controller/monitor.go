package controller

import (
	"fmt"

	"github.com/appscode/kube-mon/agents"
	mona "github.com/appscode/kube-mon/api"
	"github.com/appscode/kutil"
	api "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
)

func (c *Controller) newMonitorController(elasticsearch *api.Elasticsearch) (mona.Agent, error) {
	monitorSpec := elasticsearch.Spec.Monitor

	if monitorSpec == nil {
		return nil, fmt.Errorf("MonitorSpec not found in %v", elasticsearch.Spec)
	}

	if monitorSpec.Prometheus != nil {
		return agents.New(monitorSpec.Agent, c.Client, c.ApiExtKubeClient, c.promClient), nil
	}

	return nil, fmt.Errorf("monitoring controller not found for %v", monitorSpec)
}

func (c *Controller) addOrUpdateMonitor(elasticsearch *api.Elasticsearch) (kutil.VerbType, error) {
	agent, err := c.newMonitorController(elasticsearch)
	if err != nil {
		return kutil.VerbUnchanged, err
	}
	return agent.CreateOrUpdate(elasticsearch.StatsAccessor(), elasticsearch.Spec.Monitor)
}

func (c *Controller) deleteMonitor(elasticsearch *api.Elasticsearch) (kutil.VerbType, error) {
	agent, err := c.newMonitorController(elasticsearch)
	if err != nil {
		return kutil.VerbUnchanged, err
	}
	return agent.Delete(elasticsearch.StatsAccessor())
}

// todo: needs to set on status
func (c *Controller) manageMonitor(elasticsearch *api.Elasticsearch) error {
	if elasticsearch.Spec.Monitor != nil {
		_, err := c.addOrUpdateMonitor(elasticsearch)
		if err != nil {
			return err
		}
	} else {
		agent := agents.New(mona.AgentCoreOSPrometheus, c.Client, c.ApiExtKubeClient, c.promClient)
		_, err := agent.CreateOrUpdate(elasticsearch.StatsAccessor(), elasticsearch.Spec.Monitor)
		if err != nil {
			return err
		}
	}
	return nil
}
