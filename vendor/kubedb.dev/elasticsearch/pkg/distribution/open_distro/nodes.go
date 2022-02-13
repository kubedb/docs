/*
Copyright AppsCode Inc. and Contributors

Licensed under the AppsCode Community License 1.0.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://github.com/appscode/licenses/raw/1.0.0/AppsCode-Community-1.0.0.md

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package open_distro

import (
	"fmt"
	"strings"

	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"
	"kubedb.dev/elasticsearch/pkg/lib/heap"

	"gomodules.xyz/pointer"
	core "k8s.io/api/core/v1"
	kutil "kmodules.xyz/client-go"
	core_util "kmodules.xyz/client-go/core/v1"
)

func (es *Elasticsearch) EnsureMasterNodes() (kutil.VerbType, error) {
	statefulSetName := es.db.MasterStatefulSetName()
	masterNode := es.db.Spec.Topology.Master
	labels := map[string]string{
		es.db.NodeRoleSpecificLabelKey(api.ElasticsearchNodeRoleTypeMaster): api.ElasticsearchNodeRoleSet,
	}

	// If replicas is not provided, default to 1.
	replicas := pointer.Int32P(1)
	if masterNode.Replicas != nil {
		replicas = masterNode.Replicas
	}

	// Environment variable list for main container.
	// These are node specific, i.e. changes depending on node type.
	// Following are for Master node:
	envList := []core.EnvVar{
		{
			Name:  "node.ingest",
			Value: "false",
		},
		{
			Name:  "node.master",
			Value: "true",
		},
		{
			Name:  "node.data",
			Value: "false",
		},
	}

	// These Env are only required for master nodes to bootstrap
	// for the vary first time. Need to remove from EnvList as
	// soon as the cluster is up and running.
	if strings.HasPrefix(es.esVersion.Spec.Version, "7.") {
		envList = core_util.UpsertEnvVars(envList, core.EnvVar{
			Name:  "cluster.initial_master_nodes",
			Value: strings.Join(es.db.InitialMasterNodes(), ","),
		})
	} else {
		envList = core_util.UpsertEnvVars(envList, core.EnvVar{
			Name:  "discovery.zen.minimum_master_nodes",
			Value: fmt.Sprintf("%v", (*replicas/2)+1),
		})
	}

	// Upsert common environment variables.
	// These are same for all type of node.
	envList = es.upsertContainerEnv(envList)

	// add/overwrite user provided env; these are provided via crd spec
	envList = core_util.UpsertEnvVars(envList, es.db.Spec.PodTemplate.Spec.Env...)

	// if Local.heapSizePercentage is empty, use Global.heapSizePercentage
	if masterNode.HeapSizePercentage == nil {
		masterNode.HeapSizePercentage = es.db.Spec.HeapSizePercentage
	}
	envList = heap.UpsertJavaOptsEnv(envList, api.ElasticsearchJavaOptsEnv, &masterNode)

	// Environment variables for init container (i.e. config-merger)
	initEnvList := []core.EnvVar{
		{
			Name:  "NODE_ROLES",
			Value: "master",
		},
		// TODO:
		// 		For supporting old config-merger version,
		// 		Should be removed soon.
		//
		{
			Name:  "NODE_MASTER",
			Value: "true",
		},
		{
			Name:  "NODE_DATA",
			Value: "false",
		},
		{
			Name:  "NODE_INGEST",
			Value: "false",
		},
	}

	return es.ensureStatefulSet(&masterNode, statefulSetName, labels, replicas, string(api.ElasticsearchNodeRoleTypeMaster), envList, initEnvList)
}

func (es *Elasticsearch) EnsureDataNodes() (kutil.VerbType, error) {
	// Ignore, if nil
	if es.db.Spec.Topology.Data == nil {
		return kutil.VerbUnchanged, nil
	}
	statefulSetName := es.db.DataStatefulSetName()
	dataNode := es.db.Spec.Topology.Data
	labels := map[string]string{
		es.db.NodeRoleSpecificLabelKey(api.ElasticsearchNodeRoleTypeData): api.ElasticsearchNodeRoleSet,
	}

	// Environment variable list for main container.
	// These are node specific, i.e. changes depending on node type.
	// Following are for Data node:
	envList := []core.EnvVar{
		{
			Name:  "node.ingest",
			Value: "false",
		},
		{
			Name:  "node.master",
			Value: "false",
		},
		{
			Name:  "node.data",
			Value: "true",
		},
	}
	// Upsert common environment variables.
	// These are same for all type of node.
	envList = es.upsertContainerEnv(envList)

	// add/overwrite user provided env; these are provided via crd spec
	envList = core_util.UpsertEnvVars(envList, es.db.Spec.PodTemplate.Spec.Env...)

	// if Local.heapSizePercentage is empty, use Global.heapSizePercentage
	if dataNode.HeapSizePercentage == nil {
		dataNode.HeapSizePercentage = es.db.Spec.HeapSizePercentage
	}
	envList = heap.UpsertJavaOptsEnv(envList, api.ElasticsearchJavaOptsEnv, dataNode)

	// Environment variables for init container (i.e. config-merger)
	initEnvList := []core.EnvVar{
		{
			Name:  "NODE_ROLES",
			Value: "data",
		},
		// TODO:
		// 		For supporting old config-merger version,
		// 		Should be removed soon.
		//
		{
			Name:  "NODE_MASTER",
			Value: "false",
		},
		{
			Name:  "NODE_DATA",
			Value: "true",
		},
		{
			Name:  "NODE_INGEST",
			Value: "false",
		},
	}

	replicas := pointer.Int32P(1)
	if dataNode.Replicas != nil {
		replicas = dataNode.Replicas
	}

	return es.ensureStatefulSet(dataNode, statefulSetName, labels, replicas, string(api.ElasticsearchNodeRoleTypeData), envList, initEnvList)

}

func (es *Elasticsearch) EnsureIngestNodes() (kutil.VerbType, error) {
	statefulSetName := es.db.IngestStatefulSetName()
	ingestNode := es.db.Spec.Topology.Ingest
	labels := map[string]string{
		es.db.NodeRoleSpecificLabelKey(api.ElasticsearchNodeRoleTypeIngest): api.ElasticsearchNodeRoleSet,
	}

	// Environment variable list for main container.
	// These are node specific, i.e. changes depending on node type.
	// Following are for Ingest node:
	envList := []core.EnvVar{
		{
			Name:  "node.ingest",
			Value: "true",
		},
		{
			Name:  "node.master",
			Value: "false",
		},
		{
			Name:  "node.data",
			Value: "false",
		},
	}
	// Upsert common environment variables.
	// These are same for all type of node.
	envList = es.upsertContainerEnv(envList)

	// add/overwrite user provided env; these are provided via crd spec
	envList = core_util.UpsertEnvVars(envList, es.db.Spec.PodTemplate.Spec.Env...)

	// if Local.heapSizePercentage is empty, use Global.heapSizePercentage
	if ingestNode.HeapSizePercentage == nil {
		ingestNode.HeapSizePercentage = es.db.Spec.HeapSizePercentage
	}
	envList = heap.UpsertJavaOptsEnv(envList, api.ElasticsearchJavaOptsEnv, &ingestNode)

	// Environment variables for init container (i.e. config-merger)
	initEnvList := []core.EnvVar{
		{
			Name:  "NODE_ROLES",
			Value: "ingest",
		},
		// TODO:
		// 		For supporting old config-merger version,
		// 		Should be removed soon.
		//
		{
			Name:  "NODE_MASTER",
			Value: "false",
		},
		{
			Name:  "NODE_DATA",
			Value: "false",
		},
		{
			Name:  "NODE_INGEST",
			Value: "true",
		},
	}

	replicas := pointer.Int32P(1)
	if ingestNode.Replicas != nil {
		replicas = ingestNode.Replicas
	}

	return es.ensureStatefulSet(&ingestNode, statefulSetName, labels, replicas, string(api.ElasticsearchNodeRoleTypeIngest), envList, initEnvList)
}

func (es *Elasticsearch) EnsureCombinedNode() (kutil.VerbType, error) {
	statefulSetName := es.db.CombinedStatefulSetName()
	combinedNode := es.getCombinedNode()

	// Each node performs all three roles; master, data, and ingest.
	labels := map[string]string{
		es.db.NodeRoleSpecificLabelKey(api.ElasticsearchNodeRoleTypeMaster): api.ElasticsearchNodeRoleSet,
		es.db.NodeRoleSpecificLabelKey(api.ElasticsearchNodeRoleTypeData):   api.ElasticsearchNodeRoleSet,
		es.db.NodeRoleSpecificLabelKey(api.ElasticsearchNodeRoleTypeIngest): api.ElasticsearchNodeRoleSet,
	}

	// If replicas is not provided, default to 1.
	replicas := pointer.Int32P(1)
	if combinedNode.Replicas != nil {
		replicas = combinedNode.Replicas
	}

	// Environment variable list for main container.
	// These are node specific, i.e. changes depending on node type.
	// Followings are for Combined node:
	envList := []core.EnvVar{
		{
			Name:  "node.ingest",
			Value: "true",
		},
		{
			Name:  "node.master",
			Value: "true",
		},
		{
			Name:  "node.data",
			Value: "true",
		},
	}

	// These Envs are only required for master nodes to bootstrap
	// for the very first time. Need to remove from EnvList as
	// soon as the cluster is up and running.
	if strings.HasPrefix(es.esVersion.Spec.Version, "7.") {
		envList = core_util.UpsertEnvVars(envList, core.EnvVar{
			Name:  "cluster.initial_master_nodes",
			Value: strings.Join(es.db.InitialMasterNodes(), ","),
		})
	} else {
		envList = core_util.UpsertEnvVars(envList, core.EnvVar{
			Name:  "discovery.zen.minimum_master_nodes",
			Value: fmt.Sprintf("%v", (*replicas/2)+1),
		})
	}

	// Upsert common environment variables.
	// These are same for all type of node.
	envList = es.upsertContainerEnv(envList)

	// add/overwrite user provided env; these are provided via crd spec
	envList = core_util.UpsertEnvVars(envList, es.db.Spec.PodTemplate.Spec.Env...)

	// if Local.heapSizePercentage is empty, use Global.heapSizePercentage
	if combinedNode.HeapSizePercentage == nil {
		combinedNode.HeapSizePercentage = es.db.Spec.HeapSizePercentage
	}
	envList = heap.UpsertJavaOptsEnv(envList, api.ElasticsearchJavaOptsEnv, combinedNode)

	// Environment variables for init container (i.e. config-merger)
	initEnvList := []core.EnvVar{
		{
			Name:  "NODE_ROLES",
			Value: "master,data,ingest",
		},
		// TODO:
		// 		For supporting old config-merger version,
		// 		Should be removed soon.
		//
		{
			Name:  "NODE_MASTER",
			Value: "true",
		},
		{
			Name:  "NODE_DATA",
			Value: "true",
		},
		{
			Name:  "NODE_INGEST",
			Value: "true",
		},
	}

	// For affinity, NodeRoleIngest is used.
	return es.ensureStatefulSet(combinedNode, statefulSetName, labels, replicas, string(api.ElasticsearchNodeRoleTypeIngest), envList, initEnvList)

}

// Use ElasticsearchNode struct for combined nodes too,
// to maintain the similar code structure.
func (es *Elasticsearch) getCombinedNode() *api.ElasticsearchNode {
	return &api.ElasticsearchNode{
		Replicas:       es.db.Spec.Replicas,
		Storage:        es.db.Spec.Storage,
		Resources:      es.db.Spec.PodTemplate.Spec.Resources,
		MaxUnavailable: es.db.Spec.MaxUnavailable,
	}
}

func (es *Elasticsearch) EnsureDataContentNode() (kutil.VerbType, error) {
	return kutil.VerbUnchanged, nil
}

func (es *Elasticsearch) EnsureDataHotNode() (kutil.VerbType, error) {
	if es.db.Spec.Topology.DataHot == nil {
		return kutil.VerbUnchanged, nil
	}
	statefulSetName := es.db.DataHotStatefulSetName()
	dataHotNode := es.db.Spec.Topology.DataHot
	labels := map[string]string{
		es.db.NodeRoleSpecificLabelKey(api.ElasticsearchNodeRoleTypeDataHot): api.ElasticsearchNodeRoleSet,
	}
	// Environment variable list for main container.
	// These are node specific, i.e. changes depending on node type.
	// Following are for Data node:
	envList := []core.EnvVar{
		{
			Name:  "node.ingest",
			Value: "false",
		},
		{
			Name:  "node.master",
			Value: "false",
		},
		// OpenDistro doesn't have any data_hot node role type.
		// We will use node role "data" along with "node.attr.temp: hot" feature to mock data_hot feature.
		// Ref:
		// 		- https://opendistro.github.io/for-elasticsearch-docs/docs/elasticsearch/cluster/#advanced-step-7-set-up-a-hot-warm-architecture
		{
			Name:  "node.data",
			Value: "true",
		},
		{
			Name:  "node.attr.temp",
			Value: "hot",
		},
	}
	// Upsert common environment variables.
	// These are same for all type of node.
	envList = es.upsertContainerEnv(envList)

	// add/overwrite user provided env; these are provided via crd spec
	envList = core_util.UpsertEnvVars(envList, es.db.Spec.PodTemplate.Spec.Env...)

	// if Local.heapSizePercentage is empty, use Global.heapSizePercentage
	if dataHotNode.HeapSizePercentage == nil {
		dataHotNode.HeapSizePercentage = es.db.Spec.HeapSizePercentage
	}
	envList = heap.UpsertJavaOptsEnv(envList, api.ElasticsearchJavaOptsEnv, dataHotNode)

	// Environment variables for init container (i.e. config-merger)
	initEnvList := []core.EnvVar{
		{
			Name:  "NODE_ROLES",
			Value: "data_hot",
		},
		// TODO:
		// 		For supporting old config-merger version,
		// 		Should be removed soon.
		//
		{
			Name:  "NODE_MASTER",
			Value: "false",
		},
		{
			Name:  "NODE_DATA",
			Value: "true",
		},
		{
			Name:  "NODE_INGEST",
			Value: "false",
		},
	}

	replicas := pointer.Int32P(1)
	if dataHotNode.Replicas != nil {
		replicas = dataHotNode.Replicas
	}

	return es.ensureStatefulSet(dataHotNode, statefulSetName, labels, replicas, string(api.ElasticsearchNodeRoleTypeDataHot), envList, initEnvList)
}
func (es *Elasticsearch) EnsureDataWarmNode() (kutil.VerbType, error) {
	if es.db.Spec.Topology.DataWarm == nil {
		return kutil.VerbUnchanged, nil
	}
	statefulSetName := es.db.DataWarmStatefulSetName()
	dataWarmNode := es.db.Spec.Topology.DataWarm
	labels := map[string]string{
		es.db.NodeRoleSpecificLabelKey(api.ElasticsearchNodeRoleTypeDataWarm): api.ElasticsearchNodeRoleSet,
	}
	// Environment variable list for main container.
	// These are node specific, i.e. changes depending on node type.
	// Following are for Data node:
	envList := []core.EnvVar{
		{
			Name:  "node.ingest",
			Value: "false",
		},
		{
			Name:  "node.master",
			Value: "false",
		},
		// OpenDistro doesn't have any data_warm node role type.
		// We will use node role "data" along with "node.attr.temp: warm" feature to mock the data_warm feature.
		// Ref:
		// 		- https://opendistro.github.io/for-elasticsearch-docs/docs/elasticsearch/cluster/#advanced-step-7-set-up-a-hot-warm-architecture
		{
			Name:  "node.data",
			Value: "true",
		},
		{
			Name:  "node.attr.temp",
			Value: "warm",
		},
	}
	// Upsert common environment variables.
	// These are same for all type of node.
	envList = es.upsertContainerEnv(envList)

	// add/overwrite user provided env; these are provided via crd spec
	envList = core_util.UpsertEnvVars(envList, es.db.Spec.PodTemplate.Spec.Env...)

	// if Local.heapSizePercentage is empty, use Global.heapSizePercentage
	if dataWarmNode.HeapSizePercentage == nil {
		dataWarmNode.HeapSizePercentage = es.db.Spec.HeapSizePercentage
	}
	envList = heap.UpsertJavaOptsEnv(envList, api.ElasticsearchJavaOptsEnv, dataWarmNode)

	// Environment variables for init container (i.e. config-merger)
	initEnvList := []core.EnvVar{
		{
			Name:  "NODE_ROLES",
			Value: "data_warm",
		},
		// TODO:
		// 		For supporting old config-merger version,
		// 		Should be removed soon.
		//
		{
			Name:  "NODE_MASTER",
			Value: "false",
		},
		{
			Name:  "NODE_DATA",
			Value: "true",
		},
		{
			Name:  "NODE_INGEST",
			Value: "false",
		},
	}

	replicas := pointer.Int32P(1)
	if dataWarmNode.Replicas != nil {
		replicas = dataWarmNode.Replicas
	}

	return es.ensureStatefulSet(dataWarmNode, statefulSetName, labels, replicas, string(api.ElasticsearchNodeRoleTypeDataWarm), envList, initEnvList)
}
func (es *Elasticsearch) EnsureDataColdNode() (kutil.VerbType, error) {
	return kutil.VerbUnchanged, nil
}
func (es *Elasticsearch) EnsureDataFrozenNode() (kutil.VerbType, error) {
	return kutil.VerbUnchanged, nil
}
func (es *Elasticsearch) EnsureMLNode() (kutil.VerbType, error) {
	return kutil.VerbUnchanged, nil
}
func (es *Elasticsearch) EnsureTransformNode() (kutil.VerbType, error) {
	return kutil.VerbUnchanged, nil
}
func (es *Elasticsearch) EnsureCoordinatingNode() (kutil.VerbType, error) {
	return kutil.VerbUnchanged, nil
}
