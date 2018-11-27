package framework

import (
	"strings"

	string_util "github.com/appscode/go/strings"
	"github.com/ghodss/yaml"
	"github.com/kubedb/elasticsearch/pkg/util/es"
	. "github.com/onsi/gomega"
	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (f *Invocation) GetCommonConfig() string {
	commonSetting := es.Setting{
		Path: &es.PathSetting{
			Logs: "/data/elasticsearch/common-logdir",
		},
	}
	data, err := yaml.Marshal(commonSetting)
	Expect(err).NotTo(HaveOccurred())
	return string(data)
}

func (f *Invocation) GetMasterConfig() string {
	masterSetting := es.Setting{
		Node: &es.NodeSetting{
			Name: "es-node-master",
		},
		Path: &es.PathSetting{
			Data: []string{"/data/elasticsearch/master-datadir"},
		},
	}
	data, err := yaml.Marshal(masterSetting)
	Expect(err).NotTo(HaveOccurred())
	return string(data)
}

func (f *Invocation) GetClientConfig() string {
	clientSetting := es.Setting{
		Node: &es.NodeSetting{
			Name: "es-node-client",
		},
		Path: &es.PathSetting{
			Data: []string{"/data/elasticsearch/client-datadir"},
		},
	}
	data, err := yaml.Marshal(clientSetting)
	Expect(err).NotTo(HaveOccurred())
	return string(data)
}

func (f *Invocation) GetDataConfig() string {
	dataSetting := es.Setting{
		Node: &es.NodeSetting{
			Name: "es-node-data",
		},
		Path: &es.PathSetting{
			Data: []string{"/data/elasticsearch/data-datadir"},
		},
	}
	data, err := yaml.Marshal(dataSetting)
	Expect(err).NotTo(HaveOccurred())
	return string(data)
}

func (f *Invocation) GetCustomConfig() *core.ConfigMap {
	return &core.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      f.app,
			Namespace: f.namespace,
		},
		Data: map[string]string{},
	}
}

func (f *Invocation) IsUsingProvidedConfig(nodeInfo []es.NodeInfo) bool {
	for _, node := range nodeInfo {
		if string_util.Contains(node.Roles, "master") || strings.HasSuffix(node.Name, "master") {
			masterConfig := &es.Setting{}
			err := yaml.Unmarshal([]byte(f.GetMasterConfig()), masterConfig)
			Expect(err).NotTo(HaveOccurred())

			if node.Name != masterConfig.Node.Name {
				return false
			}

			if !string_util.EqualSlice(node.Settings.Path.Data, masterConfig.Path.Data) {
				return false
			}
		}

		if (string_util.Contains(node.Roles, "ingest") &&
			!string_util.Contains(node.Roles, "master")) ||
			strings.HasSuffix(node.Name, "client") { // master config has higher precedence

			clientConfig := &es.Setting{}
			err := yaml.Unmarshal([]byte(f.GetClientConfig()), clientConfig)
			Expect(err).NotTo(HaveOccurred())

			if node.Name != clientConfig.Node.Name {
				return false
			}
			if !string_util.EqualSlice(node.Settings.Path.Data, clientConfig.Path.Data) {
				return false
			}
		}

		if (string_util.Contains(node.Roles, "data") &&
			!(string_util.Contains(node.Roles, "master") || string_util.Contains(node.Roles, "ingest"))) ||
			strings.HasSuffix(node.Name, "data") { //master and ingest config has higher precedence

			dataConfig := &es.Setting{}
			err := yaml.Unmarshal([]byte(f.GetDataConfig()), dataConfig)
			Expect(err).NotTo(HaveOccurred())

			if node.Name != dataConfig.Node.Name {
				return false
			}
			if !string_util.EqualSlice(node.Settings.Path.Data, dataConfig.Path.Data) {
				return false
			}
		}

		// check for common config
		commonConfig := &es.Setting{}
		err := yaml.Unmarshal([]byte(f.GetCommonConfig()), commonConfig)
		Expect(err).NotTo(HaveOccurred())

		if node.Settings.Path.Logs != commonConfig.Path.Logs {
			return false
		}

	}
	return true
}

func (f *Invocation) CreateConfigMap(obj *core.ConfigMap) error {
	_, err := f.kubeClient.CoreV1().ConfigMaps(obj.Namespace).Create(obj)
	return err
}

func (f *Invocation) DeleteConfigMap(meta metav1.ObjectMeta) error {
	err := f.kubeClient.CoreV1().ConfigMaps(meta.Namespace).Delete(meta.Name, deleteInForeground())
	if err != nil && !kerr.IsNotFound(err) {
		return err
	}
	return nil
}
