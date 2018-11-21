package controller

import (
	"fmt"

	"github.com/appscode/go/log"
	"github.com/appscode/kutil"
	core_util "github.com/appscode/kutil/core/v1"
	meta_util "github.com/appscode/kutil/meta"
	api "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"kmodules.xyz/monitoring-agent-api/agents"
	mona "kmodules.xyz/monitoring-agent-api/api/v1"
)

func (c *Controller) newMonitorController(mysql *api.MySQL) (mona.Agent, error) {
	monitorSpec := mysql.Spec.Monitor

	if monitorSpec == nil {
		return nil, fmt.Errorf("MonitorSpec not found for MySQL %v/%v in %v", mysql.Namespace, mysql.Name, mysql.Spec)
	}

	if monitorSpec.Prometheus != nil {
		return agents.New(monitorSpec.Agent, c.Client, c.ApiExtKubeClient, c.promClient), nil
	}

	return nil, fmt.Errorf("monitoring controller not found for MySQL %v/%v in %v", mysql.Namespace, mysql.Name, monitorSpec)
}

func (c *Controller) addOrUpdateMonitor(mysql *api.MySQL) (kutil.VerbType, error) {
	agent, err := c.newMonitorController(mysql)
	if err != nil {
		return kutil.VerbUnchanged, err
	}
	return agent.CreateOrUpdate(mysql.StatsService(), mysql.Spec.Monitor)
}

func (c *Controller) deleteMonitor(mysql *api.MySQL) (kutil.VerbType, error) {
	agent, err := c.newMonitorController(mysql)
	if err != nil {
		return kutil.VerbUnchanged, err
	}
	return agent.Delete(mysql.StatsService())
}

func (c *Controller) getOldAgent(mysql *api.MySQL) mona.Agent {
	service, err := c.Client.CoreV1().Services(mysql.Namespace).Get(mysql.StatsService().ServiceName(), metav1.GetOptions{})
	if err != nil {
		return nil
	}
	oldAgentType, _ := meta_util.GetStringValue(service.Annotations, mona.KeyAgent)
	return agents.New(mona.AgentType(oldAgentType), c.Client, c.ApiExtKubeClient, c.promClient)
}

func (c *Controller) setNewAgent(mysql *api.MySQL) error {
	service, err := c.Client.CoreV1().Services(mysql.Namespace).Get(mysql.StatsService().ServiceName(), metav1.GetOptions{})
	if err != nil {
		return err
	}
	_, _, err = core_util.PatchService(c.Client, service, func(in *core.Service) *core.Service {
		in.Annotations = core_util.UpsertMap(in.Annotations, map[string]string{
			mona.KeyAgent: string(mysql.Spec.Monitor.Agent),
		},
		)
		return in
	})
	return err
}

func (c *Controller) manageMonitor(mysql *api.MySQL) error {
	oldAgent := c.getOldAgent(mysql)
	if mysql.Spec.Monitor != nil {
		if oldAgent != nil &&
			oldAgent.GetType() != mysql.Spec.Monitor.Agent {
			if _, err := oldAgent.Delete(mysql.StatsService()); err != nil {
				log.Errorf("error in deleting Prometheus agent. Reason: %s", err)
			}
		}
		if _, err := c.addOrUpdateMonitor(mysql); err != nil {
			return err
		}
		return c.setNewAgent(mysql)
	} else if oldAgent != nil {
		if _, err := oldAgent.Delete(mysql.StatsService()); err != nil {
			log.Errorf("error in deleting Prometheus agent. Reason: %s", err)
		}
	}
	return nil
}
