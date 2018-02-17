package controller

import (
	"fmt"

	"github.com/appscode/go/log"
	"github.com/appscode/kube-mon/agents"
	mona "github.com/appscode/kube-mon/api"
	"github.com/appscode/kutil"
	core_util "github.com/appscode/kutil/core/v1"
	meta_util "github.com/appscode/kutil/meta"
	api "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

func (c *Controller) addOrUpdateMonitor(mongodb *api.MongoDB) (kutil.VerbType, error) {
	agent, err := c.newMonitorController(mongodb)
	if err != nil {
		return kutil.VerbUnchanged, err
	}
	return agent.CreateOrUpdate(mongodb.StatsAccessor(), mongodb.Spec.Monitor)
}

func (c *Controller) deleteMonitor(mongodb *api.MongoDB) (kutil.VerbType, error) {
	agent, err := c.newMonitorController(mongodb)
	if err != nil {
		return kutil.VerbUnchanged, err
	}
	return agent.Delete(mongodb.StatsAccessor())
}

func (c *Controller) getOldAgent(mongodb *api.MongoDB) mona.Agent {
	service, err := c.Client.CoreV1().Services(mongodb.Namespace).Get(mongodb.StatsAccessor().ServiceName(), metav1.GetOptions{})
	if err != nil {
		return nil
	}
	oldAgentType, _ := meta_util.GetStringValue(service.Annotations, mona.KeyAgent)
	return agents.New(mona.AgentType(oldAgentType), c.Client, c.ApiExtKubeClient, c.promClient)
}

func (c *Controller) setNewAgent(mongodb *api.MongoDB) error {
	service, err := c.Client.CoreV1().Services(mongodb.Namespace).Get(mongodb.StatsAccessor().ServiceName(), metav1.GetOptions{})
	if err != nil {
		return err
	}
	_, _, err = core_util.PatchService(c.Client, service, func(in *core.Service) *core.Service {
		in.Annotations = core_util.UpsertMap(in.Annotations, map[string]string{
			mona.KeyAgent: string(mongodb.Spec.Monitor.Agent),
		},
		)
		return in
	})
	return err
}

func (c *Controller) manageMonitor(mongodb *api.MongoDB) error {
	oldAgent := c.getOldAgent(mongodb)
	if mongodb.Spec.Monitor != nil {
		if oldAgent != nil &&
			oldAgent.GetType() != mongodb.Spec.Monitor.Agent {
			if _, err := oldAgent.Delete(mongodb.StatsAccessor()); err != nil {
				log.Error("error in deleting Prometheus agent:", err)
			}
		}
		if _, err := c.addOrUpdateMonitor(mongodb); err != nil {
			return err
		}
		return c.setNewAgent(mongodb)
	} else if oldAgent != nil {
		if _, err := oldAgent.Delete(mongodb.StatsAccessor()); err != nil {
			log.Error("error in deleting Prometheus agent:", err)
		}
	}
	return nil
}
