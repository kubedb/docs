package controller

import (
	"fmt"

	"github.com/appscode/go/log"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kutil "kmodules.xyz/client-go"
	core_util "kmodules.xyz/client-go/core/v1"
	meta_util "kmodules.xyz/client-go/meta"
	"kmodules.xyz/monitoring-agent-api/agents"
	mona "kmodules.xyz/monitoring-agent-api/api/v1"
	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha1"
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
	return agent.CreateOrUpdate(postgres.StatsService(), postgres.Spec.Monitor)
}

func (c *Controller) deleteMonitor(postgres *api.Postgres) (kutil.VerbType, error) {
	agent, err := c.newMonitorController(postgres)
	if err != nil {
		return kutil.VerbUnchanged, err
	}
	return agent.Delete(postgres.StatsService())
}

func (c *Controller) getOldAgent(postgres *api.Postgres) mona.Agent {
	service, err := c.Client.CoreV1().Services(postgres.Namespace).Get(postgres.StatsService().ServiceName(), metav1.GetOptions{})
	if err != nil {
		return nil
	}
	oldAgentType, _ := meta_util.GetStringValue(service.Annotations, mona.KeyAgent)
	return agents.New(mona.AgentType(oldAgentType), c.Client, c.ApiExtKubeClient, c.promClient)
}

func (c *Controller) setNewAgent(postgres *api.Postgres) error {
	service, err := c.Client.CoreV1().Services(postgres.Namespace).Get(postgres.StatsService().ServiceName(), metav1.GetOptions{})
	if err != nil {
		return err
	}
	_, _, err = core_util.PatchService(c.Client, service, func(in *core.Service) *core.Service {
		in.Annotations = core_util.UpsertMap(in.Annotations, map[string]string{
			mona.KeyAgent: string(postgres.Spec.Monitor.Agent),
		},
		)
		return in
	})
	return err
}

func (c *Controller) manageMonitor(postgres *api.Postgres) error {
	oldAgent := c.getOldAgent(postgres)
	if postgres.Spec.Monitor != nil {
		if oldAgent != nil &&
			oldAgent.GetType() != postgres.Spec.Monitor.Agent {
			if _, err := oldAgent.Delete(postgres.StatsService()); err != nil {
				log.Errorf("error in deleting Prometheus agent. Reason: %s", err)
			}
		}
		if _, err := c.addOrUpdateMonitor(postgres); err != nil {
			return err
		}
		return c.setNewAgent(postgres)
	} else if oldAgent != nil {
		if _, err := oldAgent.Delete(postgres.StatsService()); err != nil {
			log.Errorf("error in deleting Prometheus agent. Reason: %s", err)
		}
	}
	return nil
}
