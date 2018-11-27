package discovery_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/appscode/go/log"
	"github.com/appscode/kutil/discovery"
	apps "k8s.io/api/apps/v1"
	autoscaling "k8s.io/api/autoscaling/v1"
	batch_v1 "k8s.io/api/batch/v1"
	batch_v1beta1 "k8s.io/api/batch/v1beta1"
	certificates "k8s.io/api/certificates/v1beta1"
	core "k8s.io/api/core/v1"
	extensions "k8s.io/api/extensions/v1beta1"
	networking "k8s.io/api/networking/v1"
	policy "k8s.io/api/policy/v1beta1"
	rbac "k8s.io/api/rbac/v1"
	settings "k8s.io/api/settings/v1alpha1"
	storage "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func skipIfNoMinikube(t *testing.T) {
	if _, ok := os.LookupEnv("MINIKUBE"); !ok {
		t.Skip("minikube not found")
	}
}

func TestRestMapper(t *testing.T) {
	skipIfNoMinikube(t)

	masterURL := ""
	kubeconfigPath := filepath.Join(homedir.HomeDir(), ".kube", "config")

	config, err := clientcmd.BuildConfigFromFlags(masterURL, kubeconfigPath)
	if err != nil {
		log.Fatalf("Could not get Kubernetes config: %s", err)
	}

	kc := kubernetes.NewForConfigOrDie(config)

	restmapper, err := discovery.LoadRestMapper(kc.Discovery())
	if err != nil {
		t.Fatal(err)
	}

	data := []struct {
		in  interface{}
		out schema.GroupVersionResource
	}{
		{&apps.ControllerRevision{}, apps.SchemeGroupVersion.WithResource("controllerrevisions")},
		{&apps.Deployment{}, apps.SchemeGroupVersion.WithResource("deployments")},
		{&apps.ReplicaSet{}, apps.SchemeGroupVersion.WithResource("replicasets")},
		{&apps.StatefulSet{}, apps.SchemeGroupVersion.WithResource("statefulsets")},
		{&autoscaling.HorizontalPodAutoscaler{}, autoscaling.SchemeGroupVersion.WithResource("horizontalpodautoscalers")},
		{&batch_v1.Job{}, batch_v1.SchemeGroupVersion.WithResource("jobs")},
		{&batch_v1beta1.CronJob{}, batch_v1beta1.SchemeGroupVersion.WithResource("cronjobs")},
		{&certificates.CertificateSigningRequest{}, certificates.SchemeGroupVersion.WithResource("certificatesigningrequests")},
		{&core.ComponentStatus{}, core.SchemeGroupVersion.WithResource("componentstatuses")},
		{&core.ConfigMap{}, core.SchemeGroupVersion.WithResource("configmaps")},
		{&core.Endpoints{}, core.SchemeGroupVersion.WithResource("endpoints")},
		{&core.Event{}, core.SchemeGroupVersion.WithResource("events")},
		{&core.LimitRange{}, core.SchemeGroupVersion.WithResource("limitranges")},
		{&core.Namespace{}, core.SchemeGroupVersion.WithResource("namespaces")},
		{&core.Node{}, core.SchemeGroupVersion.WithResource("nodes")},
		{&core.PersistentVolumeClaim{}, core.SchemeGroupVersion.WithResource("persistentvolumeclaims")},
		{&core.PersistentVolume{}, core.SchemeGroupVersion.WithResource("persistentvolumes")},
		{&core.PodTemplate{}, core.SchemeGroupVersion.WithResource("podtemplates")},
		{&core.Pod{}, core.SchemeGroupVersion.WithResource("pods")},
		{&core.ReplicationController{}, core.SchemeGroupVersion.WithResource("replicationcontrollers")},
		{&core.ResourceQuota{}, core.SchemeGroupVersion.WithResource("resourcequotas")},
		{&core.Secret{}, core.SchemeGroupVersion.WithResource("secrets")},
		{&core.ServiceAccount{}, core.SchemeGroupVersion.WithResource("serviceaccounts")},
		{&core.Service{}, core.SchemeGroupVersion.WithResource("services")},
		{&extensions.DaemonSet{}, extensions.SchemeGroupVersion.WithResource("daemonsets")},
		{&extensions.Ingress{}, extensions.SchemeGroupVersion.WithResource("ingresses")},
		{&networking.NetworkPolicy{}, networking.SchemeGroupVersion.WithResource("networkpolicies")},
		{&policy.PodDisruptionBudget{}, policy.SchemeGroupVersion.WithResource("poddisruptionbudgets")},
		{&rbac.ClusterRoleBinding{}, rbac.SchemeGroupVersion.WithResource("clusterrolebindings")},
		{&rbac.ClusterRole{}, rbac.SchemeGroupVersion.WithResource("clusterroles")},
		{&rbac.RoleBinding{}, rbac.SchemeGroupVersion.WithResource("rolebindings")},
		{&rbac.Role{}, rbac.SchemeGroupVersion.WithResource("roles")},
		{&settings.PodPreset{}, settings.SchemeGroupVersion.WithResource("podpresets")},
		{&storage.StorageClass{}, storage.SchemeGroupVersion.WithResource("storageclasses")},
	}

	for _, tt := range data {
		gvr, err := discovery.DetectResource(restmapper, tt.in)
		if err != nil {
			t.Error(err)
		}
		if gvr != tt.out {
			t.Errorf("Failed to DetectResource: expected %+v, got %+v", tt.out, gvr)
		}
	}
}

func TestResourceForGVK(t *testing.T) {
	skipIfNoMinikube(t)

	masterURL := ""
	kubeconfigPath := filepath.Join(homedir.HomeDir(), ".kube", "config")

	config, err := clientcmd.BuildConfigFromFlags(masterURL, kubeconfigPath)
	if err != nil {
		log.Fatalf("Could not get Kubernetes config: %s", err)
	}
	kc := kubernetes.NewForConfigOrDie(config)

	appsGV := apps.SchemeGroupVersion
	appsData := []struct {
		in  schema.GroupVersionKind
		out schema.GroupVersionResource
	}{
		{appsGV.WithKind("Deployment"), appsGV.WithResource("deployments")},
		{appsGV.WithKind("ReplicaSet"), appsGV.WithResource("replicasets")},
		{appsGV.WithKind("StatefulSet"), appsGV.WithResource("statefulsets")},
	}
	for _, tt := range appsData {
		gvr, err := discovery.ResourceForGVK(kc.Discovery(), tt.in)
		if err != nil {
			t.Fatal(err)
		}
		if gvr != tt.out {
			t.Errorf("Failed to DetectResource: expected %+v, got %+v", tt.out, gvr)
		}
	}

	coreGV := core.SchemeGroupVersion
	coreData := []struct {
		in  schema.GroupVersionKind
		out schema.GroupVersionResource
	}{
		{coreGV.WithKind("Pod"), coreGV.WithResource("pods")},
		{coreGV.WithKind("ConfigMap"), coreGV.WithResource("configmaps")},
		{coreGV.WithKind("Secret"), coreGV.WithResource("secrets")},
	}
	for _, tt := range coreData {
		gvr, err := discovery.ResourceForGVK(kc.Discovery(), tt.in)
		if err != nil {
			t.Fatal(err)
		}
		if gvr != tt.out {
			t.Errorf("Failed to DetectResource: expected %+v, got %+v", tt.out, gvr)
		}
	}
}
