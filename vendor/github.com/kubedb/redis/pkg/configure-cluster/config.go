package configure_cluster

import (
	"k8s.io/client-go/rest"
)

type Config struct {
	RestConfig *rest.Config
	Cluster    RedisCluster
}

type RedisCluster struct {
	MasterCnt int
	Replicas  int
}

type RedisNode struct {
	SlotStart []int
	SlotEnd   []int
	SlotsCnt  int

	ID   string
	IP   string
	Port int
	Role string
	Down bool

	Master *RedisNode
	Slaves []*RedisNode
}
