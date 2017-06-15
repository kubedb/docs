package monitor

import (
	tapi "github.com/k8sdb/apimachinery/api"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Monitor interface {
	AddMonitor(meta metav1.ObjectMeta, spec *tapi.MonitorSpec) error
	UpdateMonitor(meta metav1.ObjectMeta, old, new *tapi.MonitorSpec) error
	DeleteMonitor(meta metav1.ObjectMeta, spec *tapi.MonitorSpec) error
}
