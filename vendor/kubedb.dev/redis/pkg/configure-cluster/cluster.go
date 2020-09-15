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
	"strconv"
	"strings"
	"time"

	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha1"

	"github.com/appscode/go/log"
	"github.com/pkg/errors"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/rest"
)

// ConfigureRedisCluster() configures a cluster.
// Here, pods[][] contains the ordered pods. pods[i][j] is the j'th Pod of i'th StatefulSet (where each
// StatefulSet represents a shard. That means if there are 3 shards in cluster, then for each shard
// there is a StatefulSet and 0'th Pod of i'th StatefulSet is the master for i'th shard and 1, 2,
// 3, ....'th Pods are the replicas or slaves of that master).
func ConfigureRedisCluster(
	restConfig *rest.Config, redis *api.Redis, version string, pods [][]*core.Pod) error {

	ver, err := newVersion(version)
	if err != nil {
		return err
	}
	config := Config{
		RestConfig: restConfig,
		Cluster: RedisCluster{
			MasterCnt: int(*redis.Spec.Cluster.Master),
			Replicas:  int(*redis.Spec.Cluster.Replicas),
		},
		redisVersion: ver,
	}

	useTLS := redis.Spec.TLS != nil

	if err := config.waitUntilRedisServersToBeReady(useTLS, pods); err != nil {
		return err
	}

	if err := config.configureClusterState(useTLS, pods); err != nil {
		return err
	}

	return nil
}

// waitUntilRedisServersToBeReady() waits until all of the nodes be ready by running PING command.
// The timeOut limit to wait for each of nodes is 5 minutes. Since PING is run without any message,
// the reply should be the string "PONG" if the node is ready. If any of nodes isn't ready within
// the timeOut, this function returns an error.
// ref: https://redis.io/commands/ping
func (c Config) waitUntilRedisServersToBeReady(useTLS bool, pods [][]*core.Pod) error {
	var err error

	// i is for shards and j is for replicas. So (i, j) pair points to the j'th node (Pod) of i'th shard
	for i := 0; i < c.Cluster.MasterCnt; i++ {
		for j := 0; j <= c.Cluster.Replicas; j++ {
			execPod := pods[i][j]
			pingIP := execPod.Status.PodIP
			if err = wait.PollImmediate(time.Second, time.Minute*5, func() (bool, error) {
				if pong, _ := c.ping(useTLS, execPod, pingIP); pong == "PONG" {
					return true, nil
				}
				return false, nil
			}); err != nil {
				return errors.Wrapf(err, "%q is not ready yet", pingIP)
			}
		}
	}
	log.Infoln("All redis servers are ready")

	return nil
}

// If a node does not know another node then do a `CLUSTER MEET` from the first one.
// Here sender node sends MEET packet to the receiver node.
// ref: https://redis.io/commands/cluster-meet
func (c Config) ensureAllNodesKnowEachOther(useTLS bool, pods [][]*core.Pod) error {
	var (
		err                                error
		execPodSnd, execPodRcv             *core.Pod
		senderIP, receiverIP               string
		senderConf, receiverConf           string
		nodeCntInSender, nodeCntInReceiver int
	)

	// The nodes are represented as follows. Here, M for master and S for slave:
	// node j -> |0   |1   |2...
	//			 |    |    |
	// shard i	 |    |    |
	// |		 |    |    |
	// V		 |    |    |
	// ----------+----+----+-------
	// 0		 |M00 |S01 |S02
	// ----------+----+----+-------
	// 1		 |M10 |S11 |S12
	// ----------+----+----+-------
	// 2		 |M20 |S21 |S22
	// ----------+----+----+-------
	// .
	// .
	// .
	// So, say at any point (i = 1, j = 1), then we check for node with following (x, y) pairs:
	//                 (1, 2), (1, 3), ...
	// (2, 0), (2, 1), (2, 2), (2, 3), ...
	// (3, 0), (3, 1), (3, 2), (3, 3), ...
	for i := 0; i < min(c.Cluster.MasterCnt, len(pods)); i++ {
		for j := 0; j < min(c.Cluster.Replicas+1, len(pods[i])); j++ {

			execPodSnd = pods[i][j]
			senderIP = execPodSnd.Status.PodIP
			if senderConf, err = c.getClusterNodes(useTLS, execPodSnd, senderIP); err != nil {
				return err
			}

			x := i
			y := j
			nodeCntInSender = countNodesInNodesConf(senderConf, nodeFlagMaster) +
				countNodesInNodesConf(senderConf, nodeFlagSlave)

			// if the number of nodes is greater than 1, it means sender node know some other nodes.
			if nodeCntInSender > 1 {
				// find unknown node (receiver node) from sender node
				for ; x < min(c.Cluster.MasterCnt, len(pods)); x++ {
					for ; y < min(c.Cluster.Replicas+1, len(pods[x])); y++ {

						execPodRcv = pods[x][y]
						receiverIP = execPodRcv.Status.PodIP
						if receiverConf, err = c.getClusterNodes(useTLS, execPodRcv, receiverIP); err != nil {
							return err
						}

						nodeCntInReceiver = countNodesInNodesConf(receiverConf, nodeFlagMaster) +
							countNodesInNodesConf(receiverConf, nodeFlagSlave)
						// if node pointed bye (x, y) pair has some other nodes' info that means it is also a member
						// of the cluster. So check whether this node is a member of the cluster and it is unknown to
						// sender node, then send a MEET packet.
						if nodeCntInReceiver > 1 &&
							!strings.Contains(senderConf, receiverIP) {
							if err = c.clusterMeet(useTLS, execPodSnd, senderIP, receiverIP, strconv.Itoa(api.RedisNodePort)); err != nil {
								return err
							}
						}
					}
					y = 0
				}
			}
		}
	}

	return nil
}

// Before configuring the cluster, the operator ensures 1st Pod (node) of each StatefulSet as master.
// ensureFirstPodAsMaster() ensures this.
func (c Config) ensureFirstPodAsMaster(useTLS bool, pods [][]*core.Pod) error {
	log.Infoln("Ensuring 1st pod as master in each statefulSet...")

	var (
		err              error
		nodesConf        string
		contactingNodeIP string
		execPod          *core.Pod
	)

	execPod = pods[0][0]
	contactingNodeIP = execPod.Status.PodIP
	if nodesConf, err = c.getClusterNodes(useTLS, execPod, contactingNodeIP); err != nil {
		return err
	}

	existingMasterCnt := countNodesInNodesConf(nodesConf, nodeFlagMaster)

	// if the number of masters is greater than 1, it means there exists a cluster.
	if existingMasterCnt > 1 {
		if err = c.ensureAllNodesKnowEachOther(useTLS, pods); err != nil {
			return err
		}

		for i := 0; i < c.Cluster.MasterCnt; i++ {
			execPod = pods[i][0]
			contactingNodeIP = execPod.Status.PodIP
			if nodesConf, err = c.getClusterNodes(useTLS, execPod, contactingNodeIP); err != nil {
				return err
			}

			if getNodeRole(getMyConf(nodesConf)) != nodeFlagMaster {
				// role != "master" means this node is not serving as master
				if err = c.clusterFailover(useTLS, execPod, contactingNodeIP); err != nil {
					return err
				}
				// TODO: Need to use a better alternative for successful completion of the above operation.
				time.Sleep(time.Second * 5)
			}
		}
	}

	return nil
}

// getOrderedNodes stores the nodes info into a [][]RedisNode array named orderNodes, where,
//     orderedNodes[i][0] is always master node and pods[i][0] represents this node
//     for j>0, orderedNodes[i][j] are slaves of orderedNodes[i][0] and pods[i][j] represents them
// We order the nodes so that we can process further tasks easily
func (c Config) getOrderedNodes(useTLS bool, pods [][]*core.Pod) ([][]RedisNode, error) {
	var (
		err          error
		nodesConf    string
		nodes        map[string]*RedisNode
		orderedNodes [][]RedisNode
	)

	if err = c.ensureFirstPodAsMaster(useTLS, pods); err != nil {
		return nil, err
	}

Again:
	for {
		execPod := pods[0][0]
		if nodesConf, err = c.getClusterNodes(useTLS, execPod, execPod.Status.PodIP); err != nil {
			return nil, err
		}

		// for j > 0, this ensures pods[i][j] is slave of pods[i][0]
		nodes = processNodesConf(nodesConf)
		for _, master := range nodes {
			for i := 0; i < len(pods); i++ {
				if pods[i][0].Status.PodIP == master.IP {
					for _, slave := range master.Slaves {
						for k := 0; k < len(pods); k++ {
							for j := 1; j < len(pods[k]); j++ {
								if pods[k][j].Status.PodIP == slave.IP && i != k {
									if err = c.clusterReplicate(
										useTLS,
										pods[k][j], pods[k][j].Status.PodIP,
										getNodeId(getNodeConfByIP(nodesConf, pods[k][0].Status.PodIP))); err != nil {
										return nil, err
									}
									time.Sleep(time.Second * 5)
									goto Again
								}
							}
						}
					}
					break
				}
			}
		}
		break
	}

	// order the nodes we got earlier
	orderedNodes = make([][]RedisNode, len(pods))
	gotMasterCnt := 0
	for i := 0; i < len(nodes); i++ {
		if gotMasterCnt >= len(pods) {
			break
		}
		for _, master := range nodes {
			if master.IP == pods[i][0].Status.PodIP {
				gotMasterCnt++
				orderedNodes[i] = make([]RedisNode, len(master.Slaves)+1)
				orderedNodes[i][0] = *master
				for j := 1; j < len(orderedNodes[i]); j++ {
					for _, slave := range master.Slaves {
						if slave.IP == pods[i][j].Status.PodIP {
							orderedNodes[i][j] = *slave

							break
						}
					}
				}

				break
			}
		}
	}

	return orderedNodes, nil
}

func (c Config) resetNode(useTLS bool, execPod *core.Pod, resetNodeIP string) error {
	if err := c.clusterReset(useTLS, execPod, resetNodeIP, string(resetTypeSoft)); err != nil {
		return err
	}
	// TODO: Need to use a better alternative for successful completion of the above operation.
	time.Sleep(time.Second * 5)
	return nil
}

// ensureExtraSlavesBeRemoved() removes the extra slaves
// for each shard i
// 	   if c.Cluster.Replicas is smaller than the number of slaves in i'th shard;
//		   then delete and reset the extra slaves from back
func (c Config) ensureExtraSlavesBeRemoved(useTLS bool, pods [][]*core.Pod) error {
	log.Infoln("Ensuring extra slaves be removed...")

	var (
		err   error
		nodes [][]RedisNode
	)

	if nodes, err = c.getOrderedNodes(useTLS, pods); err != nil {
		return err
	}

	for i := range nodes {
		for j := c.Cluster.Replicas + 1; j < len(nodes[i]); j++ {
			execPod := pods[0][0]
			existingNodeAddr := nodeAddress(nodes[i][0].IP)
			deletingNodeId := nodes[i][j].ID
			if err = c.deleteNode(useTLS, execPod, existingNodeAddr, deletingNodeId); err != nil {
				return err
			}
			// TODO: Need to use a better alternative for successful completion of the above operation.
			time.Sleep(time.Second * 5)

			// reset is needed to ensure that the deleted node forgets all other nodes from its nodes.conf
			// file, so that the operator can not add this node again in later processes.
			// ref: https://redis.io/commands/cluster-reset
			// We use the soft reset here.

			// pods[i][j] represents nodes[i][j] (the node that has been just deleted). So nodes[i][j].IP is
			// the IP of the deleted node that is being reset.
			if err = c.resetNode(useTLS, pods[i][j], nodes[i][j].IP); err != nil {
				return err
			}
		}
	}

	return nil
}

// ensureExtraMastersBeRemoved() removes the extra masters. At first this function makes empty the
// deleting masters and then delete them.
// if c.Cluster.Master is smaller than the number of existing masters or shards; then
// 	   for each i in range [c.Cluster.Master, # of existing master]
//		   delete and reset the slaves of this master
//		   delete and reset the master
func (c Config) ensureExtraMastersBeRemoved(useTLS bool, pods [][]*core.Pod) error {
	log.Infoln("Ensuring extra masters be removed...")

	var (
		err               error
		existingMasterCnt int
		nodes             [][]RedisNode
		slotsPerMaster    int
		slotsRequired     int
	)

	if nodes, err = c.getOrderedNodes(useTLS, pods); err != nil {
		return err
	}

	// count number of masters
	existingMasterCnt = 0
	for i := range nodes {
		if len(nodes[i]) > 0 {
			existingMasterCnt++
		}
	}

	// first the masters being deleted need to be empty. So, we move the slots occupied by them
	// to the other masters.
	if existingMasterCnt > c.Cluster.MasterCnt {
		slotsPerMaster = 16384 / c.Cluster.MasterCnt

		for i := 0; i < c.Cluster.MasterCnt; i++ {
			slotsRequired = slotsPerMaster
			if i == c.Cluster.MasterCnt-1 {
				// this change is only for the last master
				slotsRequired = 16384 - (slotsPerMaster * i)
			}

			if i > 0 {
				// this ensures that we have the latest cluster info
				if nodes, err = c.getOrderedNodes(useTLS, pods); err != nil {
					return err
				}
			}

			to := nodes[i][0]
			for k := c.Cluster.MasterCnt; k < existingMasterCnt; k++ {
				from := nodes[k][0]
				// compare with slotsRequired
				if to.SlotsCnt < slotsRequired {
					// But compare with slotsPerMaster. Existing masters always need slots equal to
					// slotsPerMaster not slotsRequired since slotsRequired may be different for
					// the last master that is being added.
					if from.SlotsCnt > 0 {
						slots := from.SlotsCnt
						if slots > slotsRequired-to.SlotsCnt {
							slots = slotsRequired - to.SlotsCnt
						}

						if err = c.reshard(useTLS, pods[0][0], nodes, k, i, slots); err != nil {
							return err
						}
						// TODO: Need to use a better alternative for successful completion of the above operation.
						time.Sleep(time.Second * 5)
						to.SlotsCnt += slots
						from.SlotsCnt -= slots
					}
				} else {
					break
				}
			}
		}

		// now just delete the nodes
		for i := c.Cluster.MasterCnt; i < existingMasterCnt; i++ {
			for j := 1; j < len(nodes[i]); j++ {

				execPod := pods[0][0]
				existingNodeAddr := nodeAddress(nodes[i][0].IP)
				deletingNodeId := nodes[i][j].ID
				if err = c.deleteNode(useTLS, execPod, existingNodeAddr, deletingNodeId); err != nil {
					return err
				}
				// TODO: Need to use a better alternative for successful completion of the above operation.
				time.Sleep(time.Second * 5)

				// reset is needed to ensure that the deleted node forgets all other nodes from its nodes.conf
				// file, so that the operator can not add this node again in later processes.
				// ref: https://redis.io/commands/cluster-reset
				// We use the soft reset here.

				// pods[i][j] represents nodes[i][j] (the slave node that has been just deleted). So nodes[i][j].IP
				// is the IP of the deleted slave node that is being reset.
				if err = c.resetNode(useTLS, pods[i][j], nodes[i][j].IP); err != nil {
					return err
				}
			}

			execPod := pods[0][0]
			existingNodeAddr := nodeAddress(pods[0][0].Status.PodIP)
			deletingNodeId := nodes[i][0].ID
			if err = c.deleteNode(useTLS, execPod, existingNodeAddr, deletingNodeId); err != nil {
				return err
			}
			// TODO: Need to use a better alternative for successful completion of the above operation.
			time.Sleep(time.Second * 5)

			// here is also same reason for using reset as before

			// pods[i][0] represents nodes[i][0] (the master node that has been just deleted). So nodes[i][0].IP
			// is the IP of the deleted master node that is being reset.
			if err = c.resetNode(useTLS, pods[i][0], nodes[i][0].IP); err != nil {
				return err
			}
		}
	}

	return nil
}

// ensureNewMastersBeAdded() adds new masters. If user wants to add new masters (that means new shards)
// as specified in the Redis CRD, then this info is stored in `c.Cluster.Master`.
// Then the operator creates new StatefulSet (one for each master / shard). And, the pods[] array contains
// those new Pods. Basically pods[i][j] is j'th Pod of i'th shard (i'th StatefulSet)
func (c Config) ensureNewMastersBeAdded(useTLS bool, pods [][]*core.Pod) error {
	log.Infoln("Ensuring new masters be added...")

	var (
		err               error
		existingMasterCnt int
		nodes             [][]RedisNode
	)

	if nodes, err = c.getOrderedNodes(useTLS, pods); err != nil {
		return err
	}

	existingMasterCnt = 0
	for i := range nodes {
		if len(nodes[i]) > 0 {
			existingMasterCnt++
		}
	}

	// if existingMasterCnt is greater than 1, it means there exists a cluster.
	if existingMasterCnt > 1 {
		// add new master(s)
		if existingMasterCnt < c.Cluster.MasterCnt {
			for i := existingMasterCnt; i < c.Cluster.MasterCnt; i++ {
				// ensure node must be empty before adding
				if err = c.resetNode(useTLS, pods[i][0], pods[i][0].Status.PodIP); err != nil {
					return err
				}

				execPod := pods[0][0]
				newAddr := nodeAddress(pods[i][0].Status.PodIP)
				existingAddr := nodeAddress(execPod.Status.PodIP)
				if err = c.addNode(
					useTLS,
					execPod,
					newAddr, existingAddr, ""); err != nil {
					return err
				}
				// TODO: Need to use a better alternative for successful completion of the above operation.
				time.Sleep(time.Second * 5)
			}
		}
	}

	return nil
}

// There are 16384 slots in total. If the number of masters (shard number) is 5, then each master should have
// 16384/5=3276(approximately) slots (the 5th master will keep remaining slots). If any master has less slots
// rebalanceSlots() moves some slots to that master from the master those have extra slots.
func (c Config) rebalanceSlots(useTLS bool, pods [][]*core.Pod) error {
	log.Infoln("Ensuring slots are rebalanced...")

	var (
		err                           error
		existingMasterCnt             int
		nodes                         [][]RedisNode
		slotsPerMaster, slotsRequired int
	)

	if nodes, err = c.getOrderedNodes(useTLS, pods); err != nil {
		return err
	}

	existingMasterCnt = 0
	for i := range nodes {
		if len(nodes[i]) > 0 {
			existingMasterCnt++
		}
	}
	masterIndicesWithLessSlots := make([]int, 0, len(nodes))
	masterIndicesWithExtraSlots := make([]int, 0, len(nodes))

	if existingMasterCnt > 1 {
		slotsPerMaster = 16384 / c.Cluster.MasterCnt
		for i := 0; i < existingMasterCnt; i++ {
			if nodes[i][0].SlotsCnt < slotsPerMaster {
				masterIndicesWithLessSlots = append(masterIndicesWithLessSlots, i)
			} else {
				masterIndicesWithExtraSlots = append(masterIndicesWithExtraSlots, i)
			}
		}

		for i := range masterIndicesWithLessSlots {
			slotsRequired = slotsPerMaster
			if i == len(masterIndicesWithLessSlots)-1 {
				// this change is only for the last master that needs slots
				slotsRequired = 16384 - (slotsPerMaster * i)
			}

			if i > 0 {
				if nodes, err = c.getOrderedNodes(useTLS, pods); err != nil {
					return err
				}
			}

			to := nodes[masterIndicesWithLessSlots[i]][0]
			for k := range masterIndicesWithExtraSlots {
				from := nodes[masterIndicesWithExtraSlots[k]][0]
				// compare with slotsRequired
				if to.SlotsCnt < slotsRequired {
					// But compare with slotsPerMaster. Existing masters always need slots equal to
					// slotsPerMaster not slotsRequired since slotsRequired may change for last master
					// that is being added.
					if from.SlotsCnt > slotsPerMaster {
						slots := from.SlotsCnt - slotsPerMaster
						if slots > slotsRequired-to.SlotsCnt {
							slots = slotsRequired - to.SlotsCnt
						}

						if err = c.reshard(useTLS, pods[0][0], nodes,
							masterIndicesWithExtraSlots[k], masterIndicesWithLessSlots[i], slots); err != nil {
							return err
						}
						// TODO: Need to use a better alternative for successful completion of the above operation.
						time.Sleep(time.Second * 5)
						to.SlotsCnt += slots
						from.SlotsCnt -= slots
					}
				} else {
					break
				}
			}
		}
	}

	return nil
}

// We know that pods[i][j] represents the j'th node of i'th shard (nodes[i][j]) and 0'th node for i'th shard
// nodes[i][0] is the master and the rests are slave nodes. If the c.Cluster.Replicas = 2, then each master
// should have exactly 2 slaves. If in the cluster, i'th master has 1 slave, then nodes[i][2] needs to be added
// to the i'th master. ensureNewSlavesBeAdded() ensures this.
func (c Config) ensureNewSlavesBeAdded(useTLS bool, pods [][]*core.Pod) error {
	log.Infoln("Ensuring new slaves be added...")

	var (
		err               error
		existingMasterCnt int
		nodes             [][]RedisNode
	)

	if nodes, err = c.getOrderedNodes(useTLS, pods); err != nil {
		return err
	}

	existingMasterCnt = 0
	for i := range nodes {
		if len(nodes[i]) > 0 {
			existingMasterCnt++
		}
	}

	if existingMasterCnt > 1 {
		// add new slave(s)
		for i := 0; i < existingMasterCnt; i++ {
			if len(nodes[i])-1 < c.Cluster.Replicas {
				for j := len(nodes[i]); j <= c.Cluster.Replicas; j++ {
					// ensure node must be empty before adding
					if err = c.resetNode(useTLS, pods[i][j], pods[i][j].Status.PodIP); err != nil {
						return err
					}

					execPod := pods[0][0]
					newAddr := nodeAddress(pods[i][j].Status.PodIP)
					existingAddr := nodeAddress(nodes[i][0].IP)
					masterID := nodes[i][0].ID
					if err = c.addNode(
						useTLS,
						execPod,
						newAddr, existingAddr, masterID); err != nil {
						return err
					}
					time.Sleep(time.Second * 5)
				}
			}
		}
	}

	return nil
}

// ensureCluster() ensures that a running cluster exists. If there is no cluster then create one.
func (c Config) ensureCluster(pods [][]*core.Pod, useTLS bool) error {
	log.Infoln("Ensuring new cluster...")

	var (
		masterAddrs   []string
		masterNodeIds []string
		err           error
		nodesConf     string
		nodes         [][]RedisNode
	)
	masterAddrs = make([]string, c.Cluster.MasterCnt)
	masterNodeIds = make([]string, c.Cluster.MasterCnt)

	nodes, err = c.getOrderedNodes(useTLS, pods)
	if err != nil {
		return err
	}

	// count number of masters
	masterCnt := 0
	for i := range nodes {
		if len(nodes[i]) > 0 {
			masterCnt++
		}
	}

	// if the number of masters is greater than 1, it means there exists a cluster.
	if masterCnt > 1 {
		return nil
	}

	// create a cluster using the pods[0][0], pods[1][0], ...., pods[n][0] as master.
	// So first store the master addresses and master ids
	execPod := pods[0][0]
	for i := 0; i < c.Cluster.MasterCnt; i++ {
		masterPod := pods[i][0]
		masterAddrs[i] = nodeAddress(masterPod.Status.PodIP)
		if nodesConf, err = c.getClusterNodes(useTLS, execPod, masterPod.Status.PodIP); err != nil {
			return err
		}
		masterNodeIds[i] = getNodeId(getMyConf(nodesConf))
	}

	// create the cluster with the masters specified
	if err = c.createCluster(useTLS, execPod, masterAddrs...); err != nil {
		return err
	}
	// TODO: Need to use a better alternative for successful completion of the above operation.
	time.Sleep(time.Second * 15)

	// now for each shard i
	//     add pods[i][1], pods[i][2], ... as slaves of pods[i][0]
	for i := 0; i < c.Cluster.MasterCnt; i++ {
		for j := 1; j <= c.Cluster.Replicas; j++ {
			newAddr := nodeAddress(pods[i][j].Status.PodIP)
			existingAddr := masterAddrs[i]
			if err = c.addNode(
				useTLS,
				execPod,
				newAddr, existingAddr, masterNodeIds[i]); err != nil {
				return err
			}
		}
	}
	// TODO: Need to use a better alternative for successful completion of the above operation.
	time.Sleep(time.Second * 15)

	return nil
}

// configureClusterState() creates a Redis cluster if none exists. If there exists a cluster,
// then configure it if needs.
func (c Config) configureClusterState(useTLS bool, pods [][]*core.Pod) error {
	var err error

	if err = c.ensureCluster(pods, useTLS); err != nil {
		return err
	}

	if err = c.ensureExtraSlavesBeRemoved(useTLS, pods); err != nil {
		return err
	}

	if err = c.ensureExtraMastersBeRemoved(useTLS, pods); err != nil {
		return err
	}

	if err = c.ensureNewMastersBeAdded(useTLS, pods); err != nil {
		return err
	}
	if err = c.rebalanceSlots(useTLS, pods); err != nil {
		return err
	}

	if err = c.ensureNewSlavesBeAdded(useTLS, pods); err != nil {
		return err
	}

	if err = c.ensureFirstPodAsMaster(useTLS, pods); err != nil {
		return err
	}

	return nil
}
