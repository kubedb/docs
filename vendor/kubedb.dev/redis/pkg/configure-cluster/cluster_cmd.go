package configure_cluster

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha1"
)

type ResetType string

const (
	// Reset types for Redis node
	resetTypeSoft ResetType = "soft"
)

// redisVersion captures the set of commands for managing redis cluster
type redisVersion interface {
	ClusterCreateCmd(replicas int32, addrs ...string) []string
	AddNodeAsMasterCmd(newAddr, existingAddr string) []string
	AddNodeAsSlaveCmd(newAddr, existingAddr, masterId string) []string
	DeleteNodeCmd(existingAddr, deletingNodeID string) []string
	ReshardCmd(srcIP, srcID, dstIP, dstID string, slotStart, slotEnd int) []string
}

type version4 struct{}
type version5 struct{}

var _ redisVersion = &version4{}
var _ redisVersion = &version5{}

func newVersion(version string) (redisVersion, error) {
	// remove metadata from version string
	splitOff(&version, "+")
	// remove prerelease from version string
	splitOff(&version, "-")
	// version string is now in `major.minor.patch` form
	dotParts := strings.SplitN(version, ".", 3)
	if len(dotParts) == 0 {
		return nil, fmt.Errorf("version %q has no major part", version)
	}
	major, err := strconv.ParseInt(dotParts[0], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("unable to parse major part in version %q: %v", version, err)
	}
	switch major {
	case 4:
		return &version4{}, nil
	case 5:
		return &version5{}, nil
	}

	return nil, errors.New("unknown version for cluster")
}

/**************************************************
	redis-trib cluster managing commands for v4
	ref: https://github.com/antirez/redis/blob/4.0.11/src/redis-trib.rb
**************************************************/

func (v version4) ClusterCreateCmd(replicas int32, addrs ...string) []string {
	command := []string{"redis-trib", "create"}
	if replicas > 0 {
		replica := strconv.Itoa(int(replicas))
		command = append(command, "--replicas", replica)
	}
	return append(command, addrs...)
}

func (v version4) AddNodeAsMasterCmd(newAddr, existingAddr string) []string {
	return []string{"redis-trib", "add-node", newAddr, existingAddr}
}

func (v version4) AddNodeAsSlaveCmd(newAddr, existingAddr, masterId string) []string {
	return []string{"redis-trib", "add-node", "--slave", "--master-id", masterId, newAddr, existingAddr}
}

func (v version4) DeleteNodeCmd(existingAddr, deletingNodeID string) []string {
	return []string{"redis-trib", "del-node", existingAddr, deletingNodeID}
}

func (v version4) ReshardCmd(srcIP, srcID, dstIP, dstID string, slotStart, slotEnd int) []string {
	return []string{"/conf/cluster.sh", "reshard", srcIP, srcID, dstIP, dstID,
		strconv.Itoa(slotStart), strconv.Itoa(slotEnd)}
}

/*************************************************
	redis-cli cluster managing commands for v5
	ref: https://github.com/antirez/redis/blob/5.0.3/src/redis-cli.c
*************************************************/

func (v version5) ClusterCreateCmd(replicas int32, addrs ...string) []string {
	command := []string{"redis-cli", "--cluster", "create"}
	if replicas > 0 {
		replica := strconv.Itoa(int(replicas))
		command = append(command, "--cluster-replicas", replica)
	}
	return append(command, addrs...)
}

func (v version5) AddNodeAsMasterCmd(newAddr, existingAddr string) []string {
	return []string{"redis-cli", "--cluster", "add-node", newAddr, existingAddr}
}

func (v version5) AddNodeAsSlaveCmd(newAddr, existingAddr, masterId string) []string {
	return []string{"redis-cli", "--cluster", "add-node", "--cluster-slave", "--cluster-master-id", masterId, newAddr, existingAddr}
}

func (v version5) DeleteNodeCmd(existingAddr, deletingNodeID string) []string {
	return []string{"redis-cli", "--cluster", "del-node", existingAddr, deletingNodeID}
}

func (v version5) ReshardCmd(srcIP, srcID, dstIP, dstID string, slotStart, slotEnd int) []string {
	existingAddr := fmt.Sprintf("%s:%d", srcIP, api.RedisNodePort)
	return []string{"redis-cli", "--cluster", "reshard", existingAddr,
		"--cluster-from", srcID, "--cluster-to", dstID, "--cluster-slots",
		strconv.Itoa(slotEnd - slotStart + 1), "--cluster-yes"}
}

/*******************************************
	redis-cli cluster commands (general)
*******************************************/
// https://redis.io/commands/cluster-nodes
func ClusterNodesCmd(ip string) []string {
	return []string{"redis-cli", "-c", "-h", ip, "cluster", "nodes"}
}

// https://redis.io/commands/cluster-meet
func ClusterMeetCmd(senderIP, receiverIP, receiverPort string) []string {
	return []string{"redis-cli", "-c", "-h", senderIP, "cluster", "meet", receiverIP, receiverPort}
}

// https://redis.io/commands/cluster-reset
func ClusterResetCmd(ip, resetType string) []string {
	return []string{"redis-cli", "-c", "cluster", "reset", resetType}
}

// https://redis.io/commands/cluster-failover
func ClusterFailoverCmd(ip string) []string {
	return []string{"redis-cli", "-c", "-h", ip, "cluster", "failover"}
}

// https://redis.io/commands/cluster-replicate
func ClusterReplicateCmd(ip, masterNodeID string) []string {
	return []string{"redis-cli", "-c", "-h", ip, "cluster", "replicate", masterNodeID}
}

/**********************************
	redis-cli general commands
**********************************/

// https://redis.io/commands/ping
func PingCmd(ip string) []string {
	return []string{"redis-cli", "-h", ip, "ping"}
}
