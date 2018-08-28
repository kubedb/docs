package util

import (
	"k8s.io/api/core/v1"
)

const (
	EtcdVersionAnnotationKey = "etcd.version"

	EtcdClientPort = 2379
)

func GetEtcdVersion(pod *v1.Pod) string {
	return pod.Annotations[EtcdVersionAnnotationKey]
}

func GetPodNames(pods []*v1.Pod) []string {
	if len(pods) == 0 {
		return nil
	}
	res := []string{}
	for _, p := range pods {
		res = append(res, p.Name)
	}
	return res
}
