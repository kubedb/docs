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
	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"
)

// Ref:
//	- https://www.elastic.co/guide/en/elasticsearch/reference/7.6/heap-size.html#heap-size
//  - no more than 50% of your physical RAM
//  - no more than 32GB that the JVM uses for compressed object pointers (compressed oops)
// 	- no more than 26GB for zero-based compressed oops.

// GetHeapSizeFromMemory takes memory value as input, returns heap size
func GetHeapSizeFromMemory(val int64) int64 {
	// no more than 50% of main memory (RAM)
	ret := (val / 100) * 50

	// 26 GB is safe on most systems
	if ret > api.ElasticsearchMaxHeapSize {
		ret = api.ElasticsearchMaxHeapSize
	} else if ret < api.ElasticsearchMinHeapSize {
		ret = api.ElasticsearchMinHeapSize
	}
	return ret
}
