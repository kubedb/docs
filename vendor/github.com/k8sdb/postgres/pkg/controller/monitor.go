package controller

import (
	"fmt"

	tapi "github.com/k8sdb/apimachinery/apis/kubedb/v1alpha1"
	"github.com/k8sdb/apimachinery/pkg/monitor"
)

func (c *Controller) newMonitorController(postgres *tapi.Postgres) (monitor.Monitor, error) {
	monitorSpec := postgres.Spec.Monitor

	if monitorSpec == nil {
		return nil, fmt.Errorf("MonitorSpec not found in %v", postgres.Spec)
	}

	if monitorSpec.Prometheus != nil {
		return monitor.NewPrometheusController(c.Client, c.ApiExtKubeClient, c.promClient, c.opt.OperatorNamespace), nil
	}

	return nil, fmt.Errorf("Monitoring controller not found for %v", monitorSpec)
}

func (c *Controller) addMonitor(postgres *tapi.Postgres) error {
	ctrl, err := c.newMonitorController(postgres)
	if err != nil {
		return err
	}
	return ctrl.AddMonitor(postgres.ObjectMeta, postgres.Spec.Monitor)
}

func (c *Controller) deleteMonitor(postgres *tapi.Postgres) error {
	ctrl, err := c.newMonitorController(postgres)
	if err != nil {
		return err
	}
	return ctrl.DeleteMonitor(postgres.ObjectMeta, postgres.Spec.Monitor)
}

func (c *Controller) updateMonitor(oldPostgres, updatedPostgres *tapi.Postgres) error {
	var err error
	var ctrl monitor.Monitor
	if updatedPostgres.Spec.Monitor == nil {
		ctrl, err = c.newMonitorController(oldPostgres)
	} else {
		ctrl, err = c.newMonitorController(updatedPostgres)
	}
	if err != nil {
		return err
	}
	return ctrl.UpdateMonitor(updatedPostgres.ObjectMeta, oldPostgres.Spec.Monitor, updatedPostgres.Spec.Monitor)
}
