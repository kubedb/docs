/*
Copyright AppsCode Inc. and Contributors

Licensed under the PolyForm Noncommercial License 1.0.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://github.com/appscode/licenses/raw/1.0.0/PolyForm-Noncommercial-1.0.0.md

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

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
	ClusterCreateCmd(useTLS bool, replicas int32, addrs ...string) []string
	AddNodeAsMasterCmd(useTLS bool, newAddr, existingAddr string) []string
	AddNodeAsSlaveCmd(useTLS bool, newAddr, existingAddr, masterId string) []string
	DeleteNodeCmd(useTLS bool, existingAddr, deletingNodeID string) []string
	ReshardCmd(useTLS bool, srcIP, srcID, dstIP, dstID string, slotStart, slotEnd int) []string
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
	case 5, 6:
		// all commands are same for version 5 and version 6
		return &version5{}, nil
	}

	return nil, errors.New("unknown version for cluster")
}

/**************************************************
	redis-trib cluster managing commands for v4
	ref: https://github.com/antirez/redis/blob/4.0.11/src/redis-trib.rb
**************************************************/

func (v version4) ClusterCreateCmd(useTLS bool, replicas int32, addrs ...string) []string {
	command := []string{"redis-trib", "create"}
	if replicas > 0 {
		replica := strconv.Itoa(int(replicas))
		command = append(command, "--replicas", replica)
	}
	return append(command, addrs...)
}

func (v version4) AddNodeAsMasterCmd(useTLS bool, newAddr, existingAddr string) []string {
	return []string{"redis-trib", "add-node", newAddr, existingAddr}
}

func (v version4) AddNodeAsSlaveCmd(useTLS bool, newAddr, existingAddr, masterId string) []string {
	return []string{"redis-trib", "add-node", "--slave", "--master-id", masterId, newAddr, existingAddr}
}

func (v version4) DeleteNodeCmd(useTLS bool, existingAddr, deletingNodeID string) []string {
	return []string{"redis-trib", "del-node", existingAddr, deletingNodeID}
}

func (v version4) ReshardCmd(useTLS bool, srcIP, srcID, dstIP, dstID string, slotStart, slotEnd int) []string {
	return []string{"/conf/cluster.sh", "reshard", srcIP, srcID, dstIP, dstID,
		strconv.Itoa(slotStart), strconv.Itoa(slotEnd)}
}

/*************************************************
	redis-cli cluster managing commands for v5, v6
	ref: https://github.com/antirez/redis/blob/5.0.3/src/redis-cli.c
*************************************************/
var (
	tlsArgs = []string{
		"--tls",
		"--cert",
		"/certs/client.crt",
		"--key",
		"/certs/client.key",
		"--cacert",
		"/certs/ca.crt",
	}
)

func (v version5) ClusterCreateCmd(useTLS bool, replicas int32, addrs ...string) []string {
	command := []string{"redis-cli"}
	if useTLS {
		command = append(command, tlsArgs...)
	}
	command = append(command, "--cluster", "create")
	if replicas > 0 {
		replica := strconv.Itoa(int(replicas))
		command = append(command, "--cluster-replicas", replica)
	}
	command = append(command, addrs...)
	return command
}

func (v version5) AddNodeAsMasterCmd(useTLS bool, newAddr, existingAddr string) []string {
	command := []string{"redis-cli"}
	if useTLS {
		command = append(command, tlsArgs...)
	}
	return append(command, "--cluster", "add-node", newAddr, existingAddr)
}

func (v version5) AddNodeAsSlaveCmd(useTLS bool, newAddr, existingAddr, masterId string) []string {
	command := []string{"redis-cli"}
	if useTLS {
		command = append(command, tlsArgs...)
	}
	return append(command, "--cluster", "add-node", "--cluster-slave", "--cluster-master-id", masterId, newAddr, existingAddr)
}

func (v version5) DeleteNodeCmd(useTLS bool, existingAddr, deletingNodeID string) []string {
	command := []string{"redis-cli"}
	if useTLS {
		command = append(command, tlsArgs...)
	}
	return append(command, "--cluster", "del-node", existingAddr, deletingNodeID)
}

func (v version5) ReshardCmd(useTLS bool, srcIP, srcID, dstIP, dstID string, slotStart, slotEnd int) []string {
	existingAddr := fmt.Sprintf("%s:%d", srcIP, api.RedisNodePort)
	command := []string{"redis-cli"}
	if useTLS {
		command = append(command, tlsArgs...)
	}
	return append(command, "--cluster", "reshard", existingAddr,
		"--cluster-from", srcID, "--cluster-to", dstID, "--cluster-slots",
		strconv.Itoa(slotEnd-slotStart+1), "--cluster-yes")
}

/*******************************************
	redis-cli cluster commands (general)
*******************************************/
// https://redis.io/commands/cluster-nodes
func ClusterNodesCmd(useTLS bool, ip string) []string {
	command := []string{"redis-cli"}
	if useTLS {
		command = append(command, tlsArgs...)
	}
	return append(command, "-c", "-h", ip, "cluster", "nodes")
}

// https://redis.io/commands/cluster-meet
func ClusterMeetCmd(useTLS bool, senderIP, receiverIP, receiverPort string) []string {
	command := []string{"redis-cli"}
	if useTLS {
		command = append(command, tlsArgs...)
	}
	return append(command, "-c", "-h", senderIP, "cluster", "meet", receiverIP, receiverPort)
}

// https://redis.io/commands/cluster-reset
func ClusterResetCmd(useTLS bool, ip, resetType string) []string {
	command := []string{"redis-cli"}
	if useTLS {
		command = append(command, tlsArgs...)
	}
	return append(command, "-c", "cluster", "reset", resetType)
}

// https://redis.io/commands/cluster-failover
func ClusterFailoverCmd(useTLS bool, ip string) []string {
	command := []string{"redis-cli"}
	if useTLS {
		command = append(command, tlsArgs...)
	}
	return append(command, "-c", "-h", ip, "cluster", "failover")
}

// https://redis.io/commands/cluster-replicate
func ClusterReplicateCmd(useTLS bool, ip, masterNodeID string) []string {
	command := []string{"redis-cli"}
	if useTLS {
		command = append(command, tlsArgs...)
	}
	return append(command, "-c", "-h", ip, "cluster", "replicate", masterNodeID)
}

/**********************************
	redis-cli general commands
**********************************/

// https://redis.io/commands/ping
func PingCmd(useTLS bool, ip string) []string {
	command := []string{"redis-cli"}
	if useTLS {
		command = append(command, tlsArgs...)
	}
	return append(command, "-h", ip, "ping")
}
