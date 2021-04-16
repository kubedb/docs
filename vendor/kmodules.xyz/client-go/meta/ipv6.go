package meta

import (
	"context"
	"io/ioutil"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func IPv6EnabledInCluster(kc kubernetes.Interface) (bool, error) {
	svc, err := kc.CoreV1().Services(metav1.NamespaceDefault).Get(context.TODO(), "kubernetes", metav1.GetOptions{})
	if err != nil {
		return false, err
	}
	clusterIPs := []string{svc.Spec.ClusterIP}
	for _, ip := range clusterIPs {
		if strings.ContainsRune(ip, ':') {
			return true, nil
		}
	}
	return false, nil
}

func IPv6EnabledInKernel() (bool, error) {
	content, err := ioutil.ReadFile("/sys/module/ipv6/parameters/disable")
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(string(content)) == "0", nil
}
