package configure_cluster

import (
	"strconv"
)

type ResetType string

const (
	resetTypeHard ResetType = "hard"
	resetTypeSoft ResetType = "soft"
)

/**********************************
		redis-trib commands
***********************************/
func ClusterCreateCmd(replicas int32, addrs ...string) []string {
	command := []string{"redis-trib", "create"}
	if replicas > 0 {
		replica := strconv.Itoa(int(replicas))
		command = append(command, "--replicas", replica)
	}
	return append(command, addrs...)
}

func AddNodeAsMasterCmd(newAddr, existingAddr string) []string {
	return []string{"redis-trib", "add-node", newAddr, existingAddr}
}

func AddNodeAsSlaveCmd(newAddr, existingAddr, masterId string) []string {
	return []string{"redis-trib", "add-node", "--slave", "--master-id", masterId, newAddr, existingAddr}
}

func DeleteNodeCmd(existingAddr, deletingNodeID string) []string {
	return []string{"redis-trib", "del-node", existingAddr, deletingNodeID}
}

/**********************************
	redis-cli cluster commands
***********************************/
func ClusterInfoCmd() []string {
	return []string{"redis-cli", "-c", "cluster", "info"}
}

func ClusterNodesCmd(ip string) []string {
	return []string{"redis-cli", "-c", "-h", ip, "cluster", "nodes"}
}

func ClusterMeetCmd(ip, meetIP, meetPort string) []string {
	return []string{"redis-cli", "-c", "-h", ip, "cluster", "meet", meetIP, meetPort}
}

func ClusterResetCmd(ip, resetType string) []string {
	return []string{"redis-cli", "-c", "cluster", "reset", resetType}
}

func ClusterFailoverCmd(ip string) []string {
	return []string{"redis-cli", "-c", "-h", ip, "cluster", "failover"}
}

func ClusterSetSlotImportingCmd(dstNodeIP, slot, srcNodeID string) []string {
	return []string{"redis-cli", "-c", "-h", dstNodeIP, "cluster", "setslot", slot, "importing", srcNodeID}
}

func ClusterSetSlotMigratingCmd(srcNodeIP, slot, dstNodeID string) []string {
	return []string{"redis-cli", "-c", "-h", srcNodeIP, "cluster", "setslot", slot, "migrating", dstNodeID}
}

func ClusterSetSlotNodeCmd(toNodeIP, slot, dstNodeID string) []string {
	return []string{"redis-cli", "-c", "-h", toNodeIP, "cluster", "setslot", slot, "node", dstNodeID}
}

func ClusterGetKeysInSlotCmd(srcNodeIP, slot string) []string {
	return []string{"redis-cli", "-c", "-h", srcNodeIP, "cluster", "getkeysinslot", slot, "1"}
}

func ClusterReplicateCmd(ip, masterNodeID string) []string {
	return []string{"redis-cli", "-c", "-h", ip, "cluster", "replicate", masterNodeID}
}

func ReplicationInfoCmd() []string {
	return []string{"redis-cli", "-c", "info", "replication"}
}

func DebugSegfaultCmd() []string {
	return []string{"redis-cli", "-c", "debug", "segfault"}
}

/**********************************
	redis-cli general commands
***********************************/
func PingCmd(ip string) []string {
	return []string{"redis-cli", "-h", ip, "ping"}
}

func MigrateKeyCmd(srcNodeIP, dstNodeIP, dstNodePort, key, dbID, timeout string) []string {
	return []string{"redis-cli", "-h", srcNodeIP, "migrate", dstNodeIP, dstNodePort, key, dbID, timeout}
}
