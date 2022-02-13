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

package heap

import (
	"fmt"
	"strings"

	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"

	core "k8s.io/api/core/v1"
	core_util "kmodules.xyz/client-go/core/v1"
)

// Ref:
//	- https://www.elastic.co/guide/en/elasticsearch/reference/7.6/heap-size.html#heap-size
//  - no more than 50% of your physical RAM
//  - no more than 32GB that the JVM uses for compressed object pointers (compressed oops)
// 	- no more than 26GB for zero-based compressed oops.

// GetHeapSizeFromMemory takes memory value as input, returns heap size
func GetHeapSizeFromMemory(val int64, percentage *int32) int64 {
	if percentage == nil {
		return api.ElasticsearchMinHeapSize
	}

	ret := (val / 100) * int64(*percentage)
	// 26 GB is safe on most systems
	if ret > api.ElasticsearchMaxHeapSize {
		ret = api.ElasticsearchMaxHeapSize
	} else if ret < api.ElasticsearchMinHeapSize {
		ret = api.ElasticsearchMinHeapSize
	}
	return ret
}

func UpsertJavaOptsEnv(envs []core.EnvVar, envName string, esNode *api.ElasticsearchNode) []core.EnvVar {
	// Calculate Xms, Xmx from memory
	memory := int64(api.ElasticsearchMinHeapSize) // 128mb
	if limit, found := esNode.Resources.Limits[core.ResourceMemory]; found && limit.Value() > 0 {
		memory = limit.Value()
	}
	heapSize := GetHeapSizeFromMemory(memory, esNode.HeapSizePercentage)
	// Elasticsearch bootstrap fails, if -Xms and -Xmx are not equal.
	// Error: initial heap size [X] not equal to maximum heap size [Y]; this can cause resize pauses.
	XmsXmx := fmt.Sprintf("-Xms%d -Xmx%d", heapSize, heapSize)

	var javaOpts core.EnvVar
	for _, e := range envs {
		if e.Name == envName {
			javaOpts = e
			break
		}
	}

	var opts []string
	if javaOpts.Value != "" {
		opts = strings.Split(javaOpts.Value, " ")
	}

	for idx, opt := range opts {
		if strings.HasPrefix(opt, "-Xms") {
			opts = append(opts[:idx], opts[idx+1:]...)
			break
		}
	}

	for idx, opt := range opts {
		if strings.HasPrefix(opt, "-Xmx") {
			opts = append(opts[:idx], opts[idx+1:]...)
			break
		}
	}

	opts = append(opts, XmsXmx)
	envs = core_util.UpsertEnvVars(envs, core.EnvVar{
		Name:  envName,
		Value: strings.Join(opts, " "),
	})

	return envs
}
