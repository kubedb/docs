package controller

import (
	"fmt"

	api "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	"github.com/kubedb/apimachinery/pkg/eventer"
	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

var (
	NodeRole = "kubedb.com/role"
)

func (c *Controller) ensureService(postgres *api.Postgres) error {
	name := postgres.OffshootName()
	// Check if service name exists
	found, err := c.findService(postgres, name)
	if err != nil {
		return err
	}
	if !found {
		// create database Service
		if err := c.createService(postgres); err != nil {
			c.recorder.Eventf(
				postgres.ObjectReference(),
				core.EventTypeWarning,
				eventer.EventReasonFailedToCreate,
				"Failed to create Service. Reason: %v",
				err,
			)
			return err
		}
	}

	primaryService := postgres.PrimaryName()
	found, err = c.findService(postgres, primaryService)
	if err != nil {
		return err
	}
	if !found {
		// create database Discovery Service
		if err := c.createPrimaryService(postgres); err != nil {
			c.recorder.Eventf(
				postgres.ObjectReference(),
				core.EventTypeWarning,
				eventer.EventReasonFailedToCreate,
				"Failed to create Discovery Service. Reason: %v",
				err,
			)
			return err
		}
	}
	return nil
}

func (c *Controller) findService(postgres *api.Postgres, name string) (bool, error) {
	postgresName := postgres.OffshootName()
	service, err := c.Client.CoreV1().Services(postgres.Namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		if kerr.IsNotFound(err) {
			return false, nil
		} else {
			return false, err
		}
	}

	if service.Spec.Selector[api.LabelDatabaseName] != postgresName {
		return false, fmt.Errorf(`intended service "%v" already exists`, name)
	}

	return true, nil
}

func (c *Controller) createService(postgres *api.Postgres) error {
	svc := &core.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:   postgres.OffshootName(),
			Labels: postgres.OffshootLabels(),
		},
		Spec: core.ServiceSpec{
			Ports: []core.ServicePort{
				{
					Name:       "api",
					Port:       5432,
					TargetPort: intstr.FromString("api"),
				},
			},
			Selector: postgres.OffshootLabels(),
		},
	}

	if postgres.Spec.Monitor != nil &&
		postgres.Spec.Monitor.Agent == api.AgentCoreosPrometheus &&
		postgres.Spec.Monitor.Prometheus != nil {
		svc.Spec.Ports = append(svc.Spec.Ports, core.ServicePort{
			Name:       api.PrometheusExporterPortName,
			Port:       api.PrometheusExporterPortNumber,
			TargetPort: intstr.FromString(api.PrometheusExporterPortName),
		})
	}

	if _, err := c.Client.CoreV1().Services(postgres.Namespace).Create(svc); err != nil {
		return err
	}

	return nil
}

func (c *Controller) createPrimaryService(postgres *api.Postgres) error {
	svc := &core.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:   postgres.PrimaryName(),
			Labels: postgres.OffshootLabels(),
		},
		Spec: core.ServiceSpec{
			Ports: []core.ServicePort{
				{
					Name:       "api",
					Port:       5432,
					TargetPort: intstr.FromString("api"),
				},
			},
			Selector: postgres.OffshootLabels(),
		},
	}
	svc.Spec.Selector[NodeRole] = "primary"

	if _, err := c.Client.CoreV1().Services(postgres.Namespace).Create(svc); err != nil {
		return err
	}

	return nil
}
