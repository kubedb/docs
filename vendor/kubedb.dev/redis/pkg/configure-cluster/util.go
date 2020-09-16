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
	"fmt"
	"strconv"
	"strings"

	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha1"
)

const (
	// constants for flags of nodes
	nodeFlagNoAddr = "noaddr"
	nodeFlagMaster = "master"
	nodeFlagSlave  = "slave"
	nodeFlagMyself = "myself"
	nodeFlagFail   = "fail"
)

func getMyConf(nodesConf string) (myConf string) {
	myConf = ""
	nodes := strings.Split(nodesConf, "\n")
	for _, node := range nodes {
		if strings.Contains(node, nodeFlagMyself) {
			myConf = strings.TrimSpace(node)
			break
		}
	}

	return myConf
}

func getNodeConfByIP(nodesConf, ip string) (myConf string) {
	myConf = ""
	nodes := strings.Split(nodesConf, "\n")
	for _, node := range nodes {
		if strings.Contains(node, ip) {
			myConf = strings.TrimSpace(node)
			break
		}
	}

	return myConf
}

func getNodeId(nodeConf string) string {
	return strings.Split(nodeConf, " ")[0]
}

func getNodeRole(nodeConf string) (nodeRole string) {
	nodeRole = ""
	if strings.Contains(nodeConf, nodeFlagMaster) {
		nodeRole = nodeFlagMaster
	} else if strings.Contains(nodeConf, nodeFlagSlave) {
		nodeRole = nodeFlagSlave
	}

	return nodeRole
}

// processNodesConf stores nodes info into a map from nodesConf in the order they are in nodes.conf file
func processNodesConf(nodesConf string) map[string]*RedisNode {
	var (
		slotRange  []string
		start, end int
		nds        map[string]*RedisNode
	)

	nds = make(map[string]*RedisNode)
	nodes := strings.Split(nodesConf, "\n")

	for _, node := range nodes {
		node = strings.TrimSpace(node)
		parts := strings.Split(strings.TrimSpace(node), " ")

		if strings.Contains(parts[2], nodeFlagNoAddr) {
			continue
		}

		if strings.Contains(parts[2], nodeFlagMaster) {
			nd := RedisNode{
				ID:   parts[0],
				IP:   strings.Split(parts[1], ":")[0],
				Port: api.RedisNodePort,
				Role: nodeFlagMaster,
				Down: false,
			}
			if strings.Contains(parts[2], nodeFlagFail) {
				nd.Down = true
			}
			nd.SlotsCnt = 0
			for j := 8; j < len(parts); j++ {
				if parts[j][0] == '[' && parts[j][len(parts[j])-1] == ']' {
					continue
				}

				slotRange = strings.Split(parts[j], "-")
				start, _ = strconv.Atoi(slotRange[0])
				if len(slotRange) == 1 {
					end = start
				} else {
					end, _ = strconv.Atoi(slotRange[1])
				}

				nd.SlotStart = append(nd.SlotStart, start)
				nd.SlotEnd = append(nd.SlotEnd, end)
				nd.SlotsCnt += (end - start) + 1
			}
			nd.Slaves = []*RedisNode{}

			nds[nd.ID] = &nd
		}
	}

	for _, node := range nodes {
		node = strings.TrimSpace(node)
		parts := strings.Split(strings.TrimSpace(node), " ")

		if strings.Contains(parts[2], nodeFlagNoAddr) {
			continue
		}

		if strings.Contains(parts[2], nodeFlagSlave) {
			nd := RedisNode{
				ID:   parts[0],
				IP:   strings.Split(parts[1], ":")[0],
				Port: api.RedisNodePort,
				Role: nodeFlagSlave,
				Down: false,
			}
			if strings.Contains(parts[2], nodeFlagFail) {
				nd.Down = true
			}

			if nds[parts[3]] == nil {
				continue
			}
			nd.Master = nds[parts[3]]
			nds[parts[3]].Slaves = append(nds[parts[3]].Slaves, &nd)
		}
	}

	return nds
}

func nodeAddress(ip string) string {
	return fmt.Sprintf("%s:%d", ip, api.RedisNodePort)
}

func countNodesInNodesConf(nodesConf, nodeFlag string) int {
	count := 0

	nodes := strings.Split(nodesConf, "\n")
	for _, node := range nodes {
		if strings.Contains(node, nodeFlagNoAddr) {
			continue
		}

		if strings.Contains(node, nodeFlag) {
			count++
		}
	}

	return count
}

func min(a, b int) int {
	if a < b {
		return a
	}

	return b
}

func splitOff(input *string, delim string) {
	if parts := strings.SplitN(*input, delim, 2); len(parts) == 2 {
		*input = parts[0]
	}
}
