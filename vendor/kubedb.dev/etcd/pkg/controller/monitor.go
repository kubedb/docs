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

func (c *Controller) newMonitorController(etcd *api.Etcd) (mona.Agent, error) {
	monitorSpec := etcd.Spec.Monitor

	if monitorSpec == nil {
		return nil, fmt.Errorf("MonitorSpec not found in %v", etcd.Spec)
	}

	if monitorSpec.Prometheus != nil {
		return agents.New(monitorSpec.Agent, c.Client, c.ApiExtKubeClient, c.promClient), nil
	}

	return nil, fmt.Errorf("monitoring controller not found for %v", monitorSpec)
}

func (c *Controller) addOrUpdateMonitor(etcd *api.Etcd) (kutil.VerbType, error) {
	agent, err := c.newMonitorController(etcd)
	if err != nil {
		return kutil.VerbUnchanged, err
	}
	return agent.CreateOrUpdate(etcd.StatsService(), etcd.Spec.Monitor)
}

func (c *Controller) deleteMonitor(etcd *api.Etcd) (kutil.VerbType, error) {
	agent, err := c.newMonitorController(etcd)
	if err != nil {
		return kutil.VerbUnchanged, err
	}
	return agent.Delete(etcd.StatsService())
}

func (c *Controller) getOldAgent(etcd *api.Etcd) mona.Agent {
	service, err := c.Client.CoreV1().Services(etcd.Namespace).Get(etcd.StatsService().ServiceName(), metav1.GetOptions{})
	if err != nil {
		return nil
	}
	oldAgentType, _ := meta_util.GetStringValue(service.Annotations, mona.KeyAgent)
	return agents.New(mona.AgentType(oldAgentType), c.Client, c.ApiExtKubeClient, c.promClient)
}

func (c *Controller) setNewAgent(etcd *api.Etcd) error {
	service, err := c.Client.CoreV1().Services(etcd.Namespace).Get(etcd.StatsService().ServiceName(), metav1.GetOptions{})
	if err != nil {
		return err
	}
	_, _, err = core_util.PatchService(c.Client, service, func(in *core.Service) *core.Service {
		in.Annotations = core_util.UpsertMap(in.Annotations, map[string]string{
			mona.KeyAgent: string(etcd.Spec.Monitor.Agent),
		},
		)
		return in
	})
	return err
}

func (c *Controller) manageMonitor(etcd *api.Etcd) error {
	oldAgent := c.getOldAgent(etcd)
	if etcd.Spec.Monitor != nil {
		if oldAgent != nil &&
			oldAgent.GetType() != etcd.Spec.Monitor.Agent {
			if _, err := oldAgent.Delete(etcd.StatsService()); err != nil {
				log.Error("error in deleting Prometheus agent:", err)
			}
		}
		if _, err := c.addOrUpdateMonitor(etcd); err != nil {
			return err
		}
		return c.setNewAgent(etcd)
	} else if oldAgent != nil {
		if _, err := oldAgent.Delete(etcd.StatsService()); err != nil {
			log.Error("error in deleting Prometheus agent:", err)
		}
	}
	return nil
}
