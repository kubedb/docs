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

package elastic_stack

import (
	"errors"
	"fmt"
	"strings"

	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"
	"kubedb.dev/elasticsearch/pkg/lib/heap"

	"github.com/blang/semver"
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

	heapSize := int64(api.ElasticsearchMinHeapSize) // 128mb
	if limit, found := masterNode.Resources.Limits[core.ResourceMemory]; found && limit.Value() > 0 {
		heapSize = heap.GetHeapSizeFromMemory(limit.Value())
	}

	// Environment variable list for main container.
	// These are node specific, i.e. changes depending on node type.
	// Following are for Master node:
	envList := []core.EnvVar{
		{
			Name:  "ES_JAVA_OPTS",
			Value: fmt.Sprintf("-Xms%v -Xmx%v", heapSize, heapSize),
		},
	}
	dbVersion, err := semver.Parse(es.esVersion.Spec.Version)
	if err != nil {
		return kutil.VerbUnchanged, err
	}

	// For Elasticsearch version 7.x.x
	if dbVersion.Major >= 7 {
		// This Env is only required for master nodes to bootstrap
		// for the vary first time. Need to remove from EnvList as
		// soon as the cluster is up and running.
		envList = core_util.UpsertEnvVars(envList, core.EnvVar{
			Name:  "cluster.initial_master_nodes",
			Value: strings.Join(es.db.InitialMasterNodes(), ","),
		})
	} else {
		// For Elasticsearch version >= 6.8.x, < 7.0.0
		envList = core_util.UpsertEnvVars(envList, core.EnvVar{
			Name:  "discovery.zen.minimum_master_nodes",
			Value: fmt.Sprintf("%v", (*replicas/2)+1),
		})
	}

	// For Elasticsearch version >= 7.9.x
	// The legacy node role setting is deprecated.
	if dbVersion.Major > 7 || (dbVersion.Major == 7 && dbVersion.Minor >= 9) {
		// Set "NODE_ROLES" env,
		// It is used while generating elasticsearch.yml file.
		envList = core_util.UpsertEnvVars(envList, core.EnvVar{
			Name:  "NODE_ROLES",
			Value: "master",
		})

	} else {
		// For Elasticsearch version >=7.6.x, <7.9.x
		// For master node, only master role is true.
		envList = core_util.UpsertEnvVars(envList, []core.EnvVar{
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
			{
				Name:  "node.ml",
				Value: "false",
			},
		}...)
	}

	// Upsert common environment variables.
	// These are same for all type of node.
	envList = es.upsertContainerEnv(envList)

	// add/overwrite user provided env; these are provided via crd spec
	envList = core_util.UpsertEnvVars(envList, es.db.Spec.PodTemplate.Spec.Env...)

	// Environment variables for init container (i.e. config-merger)
	initEnvList := []core.EnvVar{
		{
			Name:  "NODE_ROLES",
			Value: "master",
		},
		// TODO:
		// 		For supporting old config-merger version,
		// 		Should be removed soon.
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
	// If missing, do nothing
	if es.db.Spec.Topology.Data == nil {
		return kutil.VerbUnchanged, nil
	}

	statefulSetName := es.db.DataStatefulSetName()
	dataNode := es.db.Spec.Topology.Data
	labels := map[string]string{
		es.db.NodeRoleSpecificLabelKey(api.ElasticsearchNodeRoleTypeData): api.ElasticsearchNodeRoleSet,
	}

	heapSize := int64(api.ElasticsearchMinHeapSize) // 128mb
	if limit, found := dataNode.Resources.Limits[core.ResourceMemory]; found && limit.Value() > 0 {
		heapSize = heap.GetHeapSizeFromMemory(limit.Value())
	}

	// Environment variable list for main container.
	// These are node specific, i.e. changes depending on node type.
	// Following are for Data node:
	envList := []core.EnvVar{
		{
			Name:  "ES_JAVA_OPTS",
			Value: fmt.Sprintf("-Xms%v -Xmx%v", heapSize, heapSize),
		},
	}
	dbVersion, err := semver.Parse(es.esVersion.Spec.Version)
	if err != nil {
		return kutil.VerbUnchanged, err
	}

	// For Elasticsearch version >= 7.9.x
	// The legacy node role setting is deprecated.
	if dbVersion.Major > 7 || (dbVersion.Major == 7 && dbVersion.Minor >= 9) {
		// Set "NODE_ROLES" env,
		// It is used while generating elasticsearch.yml file.
		envList = core_util.UpsertEnvVars(envList, core.EnvVar{
			Name:  "NODE_ROLES",
			Value: "data",
		})

	} else {
		// For Elasticsearch version >=6.8.0, <7.9.x
		// For data node, only data role is true.
		envList = core_util.UpsertEnvVars(envList, []core.EnvVar{
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
			{
				Name:  "node.ml",
				Value: "false",
			},
		}...)
	}

	// Upsert common environment variables.
	// These are same for all type of node.
	envList = es.upsertContainerEnv(envList)

	// add/overwrite user provided env; these are provided via crd spec
	envList = core_util.UpsertEnvVars(envList, es.db.Spec.PodTemplate.Spec.Env...)

	// Environment variables for init container (i.e. config-merger)
	initEnvList := []core.EnvVar{
		{
			Name:  "NODE_ROLES",
			Value: "data",
		},
		// TODO:
		// 		For supporting old config-merger version,
		// 		Should be removed soon.
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

	heapSize := int64(api.ElasticsearchMinHeapSize) // 128mb
	if limit, found := ingestNode.Resources.Limits[core.ResourceMemory]; found && limit.Value() > 0 {
		heapSize = heap.GetHeapSizeFromMemory(limit.Value())
	}

	// Environment variable list for main container.
	// These are node specific, i.e. changes depending on node type.
	// Following are for Ingest node:
	envList := []core.EnvVar{
		{
			Name:  "ES_JAVA_OPTS",
			Value: fmt.Sprintf("-Xms%v -Xmx%v", heapSize, heapSize),
		},
	}
	dbVersion, err := semver.Parse(es.esVersion.Spec.Version)
	if err != nil {
		return kutil.VerbUnchanged, err
	}

	// For Elasticsearch version >= 7.9.x
	// The legacy node role setting is deprecated.
	if dbVersion.Major > 7 || (dbVersion.Major == 7 && dbVersion.Minor >= 9) {
		// Set "NODE_ROLES" env,
		// It is used in elasticsearch.yml file.
		envList = core_util.UpsertEnvVars(envList, core.EnvVar{
			Name:  "NODE_ROLES",
			Value: "ingest",
		})

	} else {
		// For Elasticsearch version >=6.8.x, <7.9.x
		// For ingest node, only ingest role is true.
		envList = core_util.UpsertEnvVars(envList, []core.EnvVar{
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
			{
				Name:  "node.ml",
				Value: "false",
			},
		}...)
	}

	// Upsert common environment variables.
	// These are same for all type of node.
	envList = es.upsertContainerEnv(envList)

	// add/overwrite user provided env; these are provided via crd spec
	envList = core_util.UpsertEnvVars(envList, es.db.Spec.PodTemplate.Spec.Env...)

	// Environment variables for init container (i.e. config-merger)
	initEnvList := []core.EnvVar{
		{
			Name:  "NODE_ROLES",
			Value: "ingest",
		},
		// TODO:
		// 		For supporting old config-merger version,
		// 		Should be removed soon.
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

	heapSize := int64(api.ElasticsearchMinHeapSize) // 128mb
	if limit, found := combinedNode.Resources.Limits[core.ResourceMemory]; found && limit.Value() > 0 {
		heapSize = heap.GetHeapSizeFromMemory(limit.Value())
	}

	// Environment variable list for main container.
	// These are node specific, i.e. changes depending on node type.
	// Followings are for Combined node:
	envList := []core.EnvVar{
		{
			Name:  "ES_JAVA_OPTS",
			Value: fmt.Sprintf("-Xms%v -Xmx%v", heapSize, heapSize),
		},
	}
	dbVersion, err := semver.Parse(es.esVersion.Spec.Version)
	if err != nil {
		return kutil.VerbUnchanged, err
	}

	// These Env are only required for master nodes to bootstrap
	// for the vary first time. Need to remove from EnvList as
	// soon as the cluster is up and running.
	if dbVersion.Major >= 7 {
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

	// For Elasticsearch version >= 7.9.x
	// The legacy node role setting is deprecated.
	if dbVersion.Major > 7 || (dbVersion.Major == 7 && dbVersion.Minor >= 9) {
		// Set "NODE_ROLES" env,
		// It is used in elasticsearch.yml file.
		envList = core_util.UpsertEnvVars(envList, core.EnvVar{
			Name:  "NODE_ROLES",
			Value: "master, data, ingest",
		})
	} else {
		// For Elasticsearch version >=6.8.0, <7.9.x
		// For combined node, all master, data, ingest are ture.
		envList = core_util.UpsertEnvVars(envList, []core.EnvVar{
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
		}...)
	}

	// Upsert common environment variables.
	// These are same for all type of node.
	envList = es.upsertContainerEnv(envList)

	// add/overwrite user provided env; these are provided via crd spec
	envList = core_util.UpsertEnvVars(envList, es.db.Spec.PodTemplate.Spec.Env...)

	// Environment variables for init container (i.e. config-merger)
	initEnvList := []core.EnvVar{
		{
			Name:  "NODE_ROLES",
			Value: "master, data, ingest",
		},
		// TODO:
		// 		For supporting old config-merger version,
		// 		Should be removed soon.
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

func (es *Elasticsearch) EnsureDataContentNode() (kutil.VerbType, error) {
	// If missing, do nothing
	if es.db.Spec.Topology.DataContent == nil {
		return kutil.VerbUnchanged, nil
	}
	dbVersion, err := semver.Parse(es.esVersion.Spec.Version)
	if err != nil {
		return kutil.VerbUnchanged, err
	}
	// Data-Content node is introduced at ES version 7.10
	// Otherwise return error
	if !(dbVersion.Major >= 7 && dbVersion.Minor >= 10) {
		return kutil.VerbUnchanged, errors.New("data-content node isn't supported; The data-content node is introduced at version 7.10")
	}

	statefulSetName := es.db.DataContentStatefulSetName()
	dataContentNode := es.db.Spec.Topology.DataContent
	labels := map[string]string{
		es.db.NodeRoleSpecificLabelKey(api.ElasticsearchNodeRoleTypeDataContent): api.ElasticsearchNodeRoleSet,
	}

	heapSize := int64(api.ElasticsearchMinHeapSize) // 128mb
	if limit, found := dataContentNode.Resources.Limits[core.ResourceMemory]; found && limit.Value() > 0 {
		heapSize = heap.GetHeapSizeFromMemory(limit.Value())
	}

	// Environment variable list for main container.
	// These are node specific, i.e. changes depending on node type.
	// Following are for Data-Content node:
	envList := []core.EnvVar{
		{
			Name:  "ES_JAVA_OPTS",
			Value: fmt.Sprintf("-Xms%v -Xmx%v", heapSize, heapSize),
		},
	}

	// Set "NODE_ROLES" env,
	// It is used while generating elasticsearch.yml file.
	envList = core_util.UpsertEnvVars(envList, core.EnvVar{
		Name:  "NODE_ROLES",
		Value: "data_content",
	})

	// Upsert common environment variables.
	// These are same for all type of node.
	envList = es.upsertContainerEnv(envList)

	// add/overwrite user provided env; these are provided via crd spec
	envList = core_util.UpsertEnvVars(envList, es.db.Spec.PodTemplate.Spec.Env...)

	// Environment variables for init container (i.e. config-merger)
	initEnvList := []core.EnvVar{
		{
			Name:  "NODE_ROLES",
			Value: "data_content",
		},
	}

	replicas := pointer.Int32P(1)
	if dataContentNode.Replicas != nil {
		replicas = dataContentNode.Replicas
	}

	return es.ensureStatefulSet(dataContentNode, statefulSetName, labels, replicas, string(api.ElasticsearchNodeRoleTypeDataContent), envList, initEnvList)
}

func (es *Elasticsearch) EnsureDataHotNode() (kutil.VerbType, error) {
	// If missing, do nothing
	if es.db.Spec.Topology.DataHot == nil {
		return kutil.VerbUnchanged, nil
	}
	dbVersion, err := semver.Parse(es.esVersion.Spec.Version)
	if err != nil {
		return kutil.VerbUnchanged, err
	}
	// Data-Hot node is introduced at ES version 7.10
	// Otherwise return error
	if !(dbVersion.Major >= 7 && dbVersion.Minor >= 10) {
		return kutil.VerbUnchanged, errors.New("data-hot node isn't supported; The data-hot node is introduced at version 7.10")
	}

	statefulSetName := es.db.DataHotStatefulSetName()
	dataHotNode := es.db.Spec.Topology.DataHot
	labels := map[string]string{
		es.db.NodeRoleSpecificLabelKey(api.ElasticsearchNodeRoleTypeDataHot): api.ElasticsearchNodeRoleSet,
	}

	heapSize := int64(api.ElasticsearchMinHeapSize) // 128mb
	if limit, found := dataHotNode.Resources.Limits[core.ResourceMemory]; found && limit.Value() > 0 {
		heapSize = heap.GetHeapSizeFromMemory(limit.Value())
	}

	// Environment variable list for main container.
	// These are node specific, i.e. changes depending on node type.
	// Following are for Data-HOT node:
	envList := []core.EnvVar{
		{
			Name:  "ES_JAVA_OPTS",
			Value: fmt.Sprintf("-Xms%v -Xmx%v", heapSize, heapSize),
		},
	}

	// Set "NODE_ROLES" env,
	// It is used while generating elasticsearch.yml file.
	envList = core_util.UpsertEnvVars(envList, core.EnvVar{
		Name:  "NODE_ROLES",
		Value: "data_hot",
	})

	// Upsert common environment variables.
	// These are same for all type of node.
	envList = es.upsertContainerEnv(envList)

	// add/overwrite user provided env; these are provided via crd spec
	envList = core_util.UpsertEnvVars(envList, es.db.Spec.PodTemplate.Spec.Env...)

	// Environment variables for init container (i.e. config-merger)
	initEnvList := []core.EnvVar{
		{
			Name:  "NODE_ROLES",
			Value: "data_hot",
		},
	}

	replicas := pointer.Int32P(1)
	if dataHotNode.Replicas != nil {
		replicas = dataHotNode.Replicas
	}

	return es.ensureStatefulSet(dataHotNode, statefulSetName, labels, replicas, string(api.ElasticsearchNodeRoleTypeDataHot), envList, initEnvList)
}

func (es *Elasticsearch) EnsureDataWarmNode() (kutil.VerbType, error) {
	// If missing, do nothing
	if es.db.Spec.Topology.DataWarm == nil {
		return kutil.VerbUnchanged, nil
	}
	dbVersion, err := semver.Parse(es.esVersion.Spec.Version)
	if err != nil {
		return kutil.VerbUnchanged, err
	}
	// Data-Hot node is introduced at ES version 7.10
	// Otherwise return error
	if !(dbVersion.Major >= 7 && dbVersion.Minor >= 10) {
		return kutil.VerbUnchanged, errors.New("data-warm node isn't supported; The data-warm node is introduced at version 7.10")
	}

	statefulSetName := es.db.DataWarmStatefulSetName()
	dataWarmNode := es.db.Spec.Topology.DataWarm
	labels := map[string]string{
		es.db.NodeRoleSpecificLabelKey(api.ElasticsearchNodeRoleTypeDataWarm): api.ElasticsearchNodeRoleSet,
	}

	heapSize := int64(api.ElasticsearchMinHeapSize) // 128mb
	if limit, found := dataWarmNode.Resources.Limits[core.ResourceMemory]; found && limit.Value() > 0 {
		heapSize = heap.GetHeapSizeFromMemory(limit.Value())
	}

	// Environment variable list for main container.
	// These are node specific, i.e. changes depending on node type.
	// Following are for Data-WARM node:
	envList := []core.EnvVar{
		{
			Name:  "ES_JAVA_OPTS",
			Value: fmt.Sprintf("-Xms%v -Xmx%v", heapSize, heapSize),
		},
	}

	// Set "NODE_ROLES" env,
	// It is used while generating elasticsearch.yml file.
	envList = core_util.UpsertEnvVars(envList, core.EnvVar{
		Name:  "NODE_ROLES",
		Value: "data_warm",
	})

	// Upsert common environment variables.
	// These are same for all type of node.
	envList = es.upsertContainerEnv(envList)

	// add/overwrite user provided env; these are provided via crd spec
	envList = core_util.UpsertEnvVars(envList, es.db.Spec.PodTemplate.Spec.Env...)

	// Environment variables for init container (i.e. config-merger)
	initEnvList := []core.EnvVar{
		{
			Name:  "NODE_ROLES",
			Value: "data_warm",
		},
	}

	replicas := pointer.Int32P(1)
	if dataWarmNode.Replicas != nil {
		replicas = dataWarmNode.Replicas
	}

	return es.ensureStatefulSet(dataWarmNode, statefulSetName, labels, replicas, string(api.ElasticsearchNodeRoleTypeDataWarm), envList, initEnvList)
}

func (es *Elasticsearch) EnsureDataColdNode() (kutil.VerbType, error) {
	// If missing, do nothing
	if es.db.Spec.Topology.DataCold == nil {
		return kutil.VerbUnchanged, nil
	}
	dbVersion, err := semver.Parse(es.esVersion.Spec.Version)
	if err != nil {
		return kutil.VerbUnchanged, err
	}
	// Data-Hot node is introduced at ES version 7.10
	// Otherwise return error
	if !(dbVersion.Major >= 7 && dbVersion.Minor >= 10) {
		return kutil.VerbUnchanged, errors.New("data-cold node isn't supported; The data-cold node is introduced at version 7.10")
	}

	statefulSetName := es.db.DataColdStatefulSetName()
	dataColdNode := es.db.Spec.Topology.DataCold
	labels := map[string]string{
		es.db.NodeRoleSpecificLabelKey(api.ElasticsearchNodeRoleTypeDataCold): api.ElasticsearchNodeRoleSet,
	}

	heapSize := int64(api.ElasticsearchMinHeapSize) // 128mb
	if limit, found := dataColdNode.Resources.Limits[core.ResourceMemory]; found && limit.Value() > 0 {
		heapSize = heap.GetHeapSizeFromMemory(limit.Value())
	}

	// Environment variable list for main container.
	// These are node specific, i.e. changes depending on node type.
	// Following are for Data-COLD node:
	envList := []core.EnvVar{
		{
			Name:  "ES_JAVA_OPTS",
			Value: fmt.Sprintf("-Xms%v -Xmx%v", heapSize, heapSize),
		},
	}

	// Set "NODE_ROLES" env,
	// It is used while generating elasticsearch.yml file.
	envList = core_util.UpsertEnvVars(envList, core.EnvVar{
		Name:  "NODE_ROLES",
		Value: "data_cold",
	})

	// Upsert common environment variables.
	// These are same for all type of node.
	envList = es.upsertContainerEnv(envList)

	// add/overwrite user provided env; these are provided via crd spec
	envList = core_util.UpsertEnvVars(envList, es.db.Spec.PodTemplate.Spec.Env...)

	// Environment variables for init container (i.e. config-merger)
	initEnvList := []core.EnvVar{
		{
			Name:  "NODE_ROLES",
			Value: "data_cold",
		},
	}

	replicas := pointer.Int32P(1)
	if dataColdNode.Replicas != nil {
		replicas = dataColdNode.Replicas
	}

	return es.ensureStatefulSet(dataColdNode, statefulSetName, labels, replicas, string(api.ElasticsearchNodeRoleTypeDataCold), envList, initEnvList)
}

func (es *Elasticsearch) EnsureDataFrozenNode() (kutil.VerbType, error) {
	// If missing, do nothing
	if es.db.Spec.Topology.DataFrozen == nil {
		return kutil.VerbUnchanged, nil
	}
	dbVersion, err := semver.Parse(es.esVersion.Spec.Version)
	if err != nil {
		return kutil.VerbUnchanged, err
	}
	// Data-Frozen node is introduced at ES version 7.12
	// Otherwise return error
	if !(dbVersion.Major >= 7 && dbVersion.Minor >= 12) {
		return kutil.VerbUnchanged, errors.New("data-frozen node isn't supported; The data-frozen node is introduced at version 7.12")
	}

	statefulSetName := es.db.DataFrozenStatefulSetName()
	dataFrozenNode := es.db.Spec.Topology.DataFrozen
	labels := map[string]string{
		es.db.NodeRoleSpecificLabelKey(api.ElasticsearchNodeRoleTypeDataFrozen): api.ElasticsearchNodeRoleSet,
	}

	heapSize := int64(api.ElasticsearchMinHeapSize) // 128mb
	if limit, found := dataFrozenNode.Resources.Limits[core.ResourceMemory]; found && limit.Value() > 0 {
		heapSize = heap.GetHeapSizeFromMemory(limit.Value())
	}

	// Environment variable list for main container.
	// These are node specific, i.e. changes depending on node type.
	// Following are for Data-Frozen node:
	envList := []core.EnvVar{
		{
			Name:  "ES_JAVA_OPTS",
			Value: fmt.Sprintf("-Xms%v -Xmx%v", heapSize, heapSize),
		},
	}

	// Set "NODE_ROLES" env,
	// It is used while generating elasticsearch.yml file.
	envList = core_util.UpsertEnvVars(envList, core.EnvVar{
		Name:  "NODE_ROLES",
		Value: "data_frozen",
	})

	// Upsert common environment variables.
	// These are same for all type of node.
	envList = es.upsertContainerEnv(envList)

	// add/overwrite user provided env; these are provided via crd spec
	envList = core_util.UpsertEnvVars(envList, es.db.Spec.PodTemplate.Spec.Env...)

	// Environment variables for init container (i.e. config-merger)
	initEnvList := []core.EnvVar{
		{
			Name:  "NODE_DATA_FROZEN",
			Value: "true",
		},
	}

	replicas := pointer.Int32P(1)
	if dataFrozenNode.Replicas != nil {
		replicas = dataFrozenNode.Replicas
	}

	return es.ensureStatefulSet(dataFrozenNode, statefulSetName, labels, replicas, string(api.ElasticsearchNodeRoleTypeDataFrozen), envList, initEnvList)
}

func (es *Elasticsearch) EnsureMLNode() (kutil.VerbType, error) {
	// If missing, do nothing
	if es.db.Spec.Topology.ML == nil {
		return kutil.VerbUnchanged, nil
	}
	dbVersion, err := semver.Parse(es.esVersion.Spec.Version)
	if err != nil {
		return kutil.VerbUnchanged, err
	}

	statefulSetName := es.db.MLStatefulSetName()
	mlNode := es.db.Spec.Topology.ML
	labels := map[string]string{
		es.db.NodeRoleSpecificLabelKey(api.ElasticsearchNodeRoleTypeML): api.ElasticsearchNodeRoleSet,
	}

	heapSize := int64(api.ElasticsearchMinHeapSize) // 128mb
	if limit, found := mlNode.Resources.Limits[core.ResourceMemory]; found && limit.Value() > 0 {
		heapSize = heap.GetHeapSizeFromMemory(limit.Value())
	}

	// Environment variable list for main container.
	// These are node specific, i.e. changes depending on node type.
	// Following are for ML node:
	envList := []core.EnvVar{
		{
			Name:  "ES_JAVA_OPTS",
			Value: fmt.Sprintf("-Xms%v -Xmx%v", heapSize, heapSize),
		},
	}

	// For Elasticsearch version >= 7.9.x
	// The legacy node role setting is deprecated.
	if dbVersion.Major > 7 || (dbVersion.Major == 7 && dbVersion.Minor >= 9) {
		// Set "NODE_ROLES" env,
		// It is used while generating elasticsearch.yml file.
		// The remote_cluster_client role is optional but strongly recommended.
		// Otherwise, cross-cluster search fails when used in machine learning jobs or datafeeds.
		//	Ref:
		//		- https://www.elastic.co/guide/en/elasticsearch/reference/7.13/modules-node.html#ml-node
		envList = core_util.UpsertEnvVars(envList, core.EnvVar{
			Name:  "NODE_ROLES",
			Value: "ml, remote_cluster_client",
		})

	} else {
		// For Elasticsearch version >=6.8.0, <7.9.x
		// For data node, only ml role is true.
		envList = core_util.UpsertEnvVars(envList, []core.EnvVar{
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
				Value: "false",
			},
			{
				Name:  "node.ml",
				Value: "true",
			},
		}...)
	}
	// Upsert common environment variables.
	// These are same for all type of node.
	envList = es.upsertContainerEnv(envList)

	// add/overwrite user provided env; these are provided via crd spec
	envList = core_util.UpsertEnvVars(envList, es.db.Spec.PodTemplate.Spec.Env...)

	// Environment variables for init container (i.e. config-merger)
	initEnvList := []core.EnvVar{
		{
			Name:  "NODE_ROLES",
			Value: "ml,remote_cluster_client",
		},
	}

	replicas := pointer.Int32P(1)
	if mlNode.Replicas != nil {
		replicas = mlNode.Replicas
	}

	return es.ensureStatefulSet(mlNode, statefulSetName, labels, replicas, string(api.ElasticsearchNodeRoleTypeML), envList, initEnvList)
}

func (es *Elasticsearch) EnsureTransformNode() (kutil.VerbType, error) {
	// If missing, do nothing
	if es.db.Spec.Topology.Transform == nil {
		return kutil.VerbUnchanged, nil
	}
	dbVersion, err := semver.Parse(es.esVersion.Spec.Version)
	if err != nil {
		return kutil.VerbUnchanged, err
	}
	// Transform node is introduced at ES version 7.11
	// Otherwise return error
	if !(dbVersion.Major >= 7 && dbVersion.Minor >= 11) {
		return kutil.VerbUnchanged, errors.New("transform node isn't supported; The transform node is introduced at version 7.11")
	}

	statefulSetName := es.db.TransformStatefulSetName()
	transformNode := es.db.Spec.Topology.Transform
	labels := map[string]string{
		es.db.NodeRoleSpecificLabelKey(api.ElasticsearchNodeRoleTypeTransform): api.ElasticsearchNodeRoleSet,
	}

	heapSize := int64(api.ElasticsearchMinHeapSize) // 128mb
	if limit, found := transformNode.Resources.Limits[core.ResourceMemory]; found && limit.Value() > 0 {
		heapSize = heap.GetHeapSizeFromMemory(limit.Value())
	}

	// Environment variable list for main container.
	// These are node specific, i.e. changes depending on node type.
	// Following are for Transform node:
	envList := []core.EnvVar{
		{
			Name:  "ES_JAVA_OPTS",
			Value: fmt.Sprintf("-Xms%v -Xmx%v", heapSize, heapSize),
		},
	}

	// Set "NODE_ROLES" env,
	// It is used while generating elasticsearch.yml file.
	// The remote_cluster_client role is optional but strongly recommended.
	// Otherwise, cross-cluster search fails when used in transforms.
	envList = core_util.UpsertEnvVars(envList, core.EnvVar{
		Name:  "NODE_ROLES",
		Value: "transform, remote_cluster_client",
	})

	// Upsert common environment variables.
	// These are same for all type of node.
	envList = es.upsertContainerEnv(envList)

	// add/overwrite user provided env; these are provided via crd spec
	envList = core_util.UpsertEnvVars(envList, es.db.Spec.PodTemplate.Spec.Env...)

	// Environment variables for init container (i.e. config-merger)
	initEnvList := []core.EnvVar{
		{
			Name:  "NODE_ROLES",
			Value: "transform, remote_cluster_client",
		},
	}

	replicas := pointer.Int32P(1)
	if transformNode.Replicas != nil {
		replicas = transformNode.Replicas
	}

	return es.ensureStatefulSet(transformNode, statefulSetName, labels, replicas, string(api.ElasticsearchNodeRoleTypeTransform), envList, initEnvList)
}

// TODO:
//		- Why CoordinatingNode is not supported?
// 			- https://discuss.elastic.co/t/how-to-set-default-value-while-using-env-in-elasticsearch-yml-config-file/275440
//			- https://stackoverflow.com/questions/67936479/how-to-set-default-value-while-using-env-in-elasticsearch-yml-config-file
func (es *Elasticsearch) EnsureCoordinatingNode() (kutil.VerbType, error) {
	// If missing, do nothing
	if es.db.Spec.Topology.Coordinating == nil {
		return kutil.VerbUnchanged, nil
	}
	dbVersion, err := semver.Parse(es.esVersion.Spec.Version)
	if err != nil {
		return kutil.VerbUnchanged, err
	}
	// Transform node is introduced at ES version 7.11
	// Otherwise return error
	if !(dbVersion.Major >= 7 && dbVersion.Minor >= 11) {
		return kutil.VerbUnchanged, errors.New("transform node isn't supported; The transform node is introduced at version 7.11")
	}

	statefulSetName := es.db.CoordinatingStatefulSetName()
	coordinatingNode := es.db.Spec.Topology.Coordinating
	labels := map[string]string{
		es.db.NodeRoleSpecificLabelKey(api.ElasticsearchNodeRoleTypeCoordinating): api.ElasticsearchNodeRoleSet,
	}

	heapSize := int64(api.ElasticsearchMinHeapSize) // 128mb
	if limit, found := coordinatingNode.Resources.Limits[core.ResourceMemory]; found && limit.Value() > 0 {
		heapSize = heap.GetHeapSizeFromMemory(limit.Value())
	}

	// Environment variable list for main container.
	// These are node specific, i.e. changes depending on node type.
	// Following are for Transform node:
	envList := []core.EnvVar{
		{
			Name:  "ES_JAVA_OPTS",
			Value: fmt.Sprintf("-Xms%v -Xmx%v", heapSize, heapSize),
		},
	}

	// For Elasticsearch version >= 7.9.x
	// The legacy node role setting is deprecated.
	if dbVersion.Major > 7 || (dbVersion.Major == 7 && dbVersion.Minor >= 9) {
		// Set "NODE_ROLES" env,
		// It is used while generating elasticsearch.yml file.
		// Every node is implicitly a coordinating node. This means that a node that has
		// an explicit empty list of roles via node.roles will only act as a coordinating node, which cannot be disabled.
		envList = core_util.UpsertEnvVars(envList, core.EnvVar{
			Name:  "NODE_ROLES",
			Value: "",
		})

	} else {
		// For Elasticsearch version >=6.8.0, <7.9.x
		// Every node is implicitly a coordinating node.
		// This means that a node that has all three node.master, node.data and node.ingest set to false
		// will only act as a coordinating node, which cannot be disabled.
		envList = core_util.UpsertEnvVars(envList, []core.EnvVar{
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
				Value: "false",
			},
			{
				Name:  "node.ml",
				Value: "false",
			},
		}...)
	}

	// Upsert common environment variables.
	// These are same for all type of node.
	envList = es.upsertContainerEnv(envList)

	// add/overwrite user provided env; these are provided via crd spec
	envList = core_util.UpsertEnvVars(envList, es.db.Spec.PodTemplate.Spec.Env...)

	// Environment variables for init container (i.e. config-merger)
	initEnvList := []core.EnvVar{
		{
			Name:  "NODE_ROLES",
			Value: "",
		},
	}

	replicas := pointer.Int32P(1)
	if coordinatingNode.Replicas != nil {
		replicas = coordinatingNode.Replicas
	}

	return es.ensureStatefulSet(coordinatingNode, statefulSetName, labels, replicas, string(api.ElasticsearchNodeRoleTypeCoordinating), envList, initEnvList)
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
