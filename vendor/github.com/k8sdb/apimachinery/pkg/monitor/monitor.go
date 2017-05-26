package monitor

import (
	tapi "github.com/k8sdb/apimachinery/api"
	kapi "k8s.io/kubernetes/pkg/api"
)

type Monitor interface {
	AddMonitor(meta kapi.ObjectMeta, spec *tapi.MonitorSpec) error
	UpdateMonitor(meta kapi.ObjectMeta, old, new *tapi.MonitorSpec) error
	DeleteMonitor(meta kapi.ObjectMeta, spec *tapi.MonitorSpec) error
}
