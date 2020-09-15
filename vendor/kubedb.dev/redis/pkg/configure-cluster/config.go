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

package configure_cluster

import (
	"k8s.io/client-go/rest"
)

// Config contains necessary fields that are needed in the processes to configure.
type Config struct {
	// RestConfig is needed to execute commands inside Pods
	RestConfig *rest.Config

	// Cluster keeps the cluster info
	Cluster RedisCluster

	// redisVersion is set depending on the major part of redis version string
	redisVersion
}

// RedisCluster contains same info as `.spec.cluster` of Redis CRD
type RedisCluster struct {
	// Number of master nodes
	MasterCnt int

	// Number of replica(s) per master node
	Replicas int
}

// RedisNode stores info about a node
type RedisNode struct {
	// If this node is a master then it's slots info are stored. Otherwise, these slots info will take the
	// default values.
	// There is a total of 16384 slots in the cluster. SlotsCnt is the total number of slots, the current
	// node contains out of 16384 slots.
	// The node may have different ranges of slots. Then SlotStart and SlotEnd are the array of starting and
	// ending indices of the range(s) respectively.
	// Say, the node has slots of ranges 0-1000 2000-2500 10000-12000.
	// So SlotCnt is 3500, SlotStart[] is [0, 2000, 10000] and SlotEnd[] is [1000, 2500 12000]
	SlotStart []int
	SlotEnd   []int
	SlotsCnt  int

	// node id
	ID string

	// node ip
	IP string

	// port at which redis server is running in the node
	Port int

	// node role (either master or slave)
	Role string

	// true if the flag of the node is 'fail'
	Down bool

	// If the node role is slave, then Master is set, otherwise it is nil.
	Master *RedisNode

	// If the node role is master, the Slaves contains the array of it's slave nodes. Otherwise, it is empty.
	Slaves []*RedisNode
}
