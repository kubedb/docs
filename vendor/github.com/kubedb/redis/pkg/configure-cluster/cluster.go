package configure_cluster

import (
	"strings"
	"time"

	"github.com/appscode/go/log"
	api "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	"github.com/pkg/errors"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/rest"
)

func ConfigureRedisCluster(
	restConfig *rest.Config, redis *api.Redis, pods [][]*core.Pod) error {
	config := Config{
		RestConfig: restConfig,
		Cluster: RedisCluster{
			MasterCnt: int(*redis.Spec.Cluster.Master),
			Replicas:  int(*redis.Spec.Cluster.Replicas),
		},
	}

	if err := config.waitUntilRedisServersToBeReady(pods); err != nil {
		return err
	}
	if err := config.configureClusterState(pods); err != nil {
		return err
	}

	return nil
}

func (c Config) waitUntilRedisServersToBeReady(pods [][]*core.Pod) error {
	var err error

	for i := 0; i < c.Cluster.MasterCnt; i++ {
		for j := 0; j <= c.Cluster.Replicas; j++ {
			if err = wait.PollImmediate(time.Second, time.Minute*5, func() (bool, error) {
				if pong, _ := c.ping(pods[i][j], pods[i][j].Status.PodIP); pong == "PONG" {
					return true, nil
				}

				return false, nil
			}); err != nil {
				return errors.Wrapf(err, "%q is not ready yet", pods[i][j].Status.PodIP)
			}
		}
	}
	log.Infoln("All redis servers are ready")

	return nil
}

func (c Config) ensureFirstPodAsMaster(pods [][]*core.Pod) error {
	log.Infoln("Ensuring 1st pod as master in each statefulSet...")

	var (
		err       error
		nodesConf string
	)

	if nodesConf, err = c.getClusterNodes(pods[0][0], pods[0][0].Status.PodIP); err != nil {
		return err
	}

	if strings.Count(nodesConf, "master") > 1 {
		for i := 0; i < c.Cluster.MasterCnt; i++ {
			if nodesConf, err = c.getClusterNodes(pods[i][0], pods[i][0].Status.PodIP); err != nil {
				return err
			}
			if getNodeRole(getMyConf(nodesConf)) != "master" {
				if err = c.clusterFailover(pods[i][0], pods[i][0].Status.PodIP); err != nil {
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
//     orderedNodes[i][0] is always master node and
//     orderedNodes[i][j] are slaves of orderedNodes[i][0]
// We do this to keep pace with pods array.
func (c Config) getOrderedNodes(pods [][]*core.Pod) ([][]RedisNode, error) {
	var (
		err          error
		nodesConf    string
		nodes        map[string]*RedisNode
		orderedNodes [][]RedisNode
	)

	if err = c.ensureFirstPodAsMaster(pods); err != nil {
		return nil, err
	}

Again:
	for {
		if nodesConf, err = c.getClusterNodes(pods[0][0], pods[0][0].Status.PodIP); err != nil {
			return nil, err
		}

		// ensures pods[i][j] is slave of pods[i][0]
		nodes = processNodesConf(nodesConf)
		for _, master := range nodes {
			for i := 0; i < len(pods); i++ {
				if pods[i][0].Status.PodIP == master.IP {
					for _, slave := range master.Slaves {
						for k := 0; k < len(pods); k++ {
							for j := 1; j < len(pods[k]); j++ {
								if pods[k][j].Status.PodIP == slave.IP && i != k {
									if err = c.clusterReplicate(
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

	// order the nodes we got
	orderedNodes = make([][]RedisNode, len(nodes))
	for i := 0; i < len(nodes); i++ {
		for _, master := range nodes {
			if master.IP == pods[i][0].Status.PodIP {
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

func (c Config) ensureExtraSlavesBeRemoved(pods [][]*core.Pod) error {
	log.Infoln("Ensuring extra slaves be removed...")

	var (
		err   error
		nodes [][]RedisNode
	)

	nodes, err = c.getOrderedNodes(pods)
	for i := range nodes {
		if c.Cluster.Replicas < len(nodes[i])-1 {
			for j := c.Cluster.Replicas + 1; j < len(nodes[i]); j++ {
				if err = c.deleteNode(pods[0][0], nodeAddress(nodes[i][0].IP), nodes[i][j].ID); err != nil {
					return err
				}
				// TODO: Need to use a better alternative for successful completion of the above operation.
				time.Sleep(time.Second * 5)
			}
		}
	}

	return nil
}

func (c Config) ensureExtraMastersBeRemoved(pods [][]*core.Pod) error {
	log.Infoln("Ensuring extra masters be removed...")

	var (
		err               error
		existingMasterCnt int
		nodes             [][]RedisNode
		slotsPerMaster    int
		slotsRequired     int
	)

	nodes, err = c.getOrderedNodes(pods)
	existingMasterCnt = len(nodes)

	// first the masters being deleted need to be empty
	if existingMasterCnt > c.Cluster.MasterCnt {
		slotsPerMaster = 16384 / c.Cluster.MasterCnt

		for i := 0; i < c.Cluster.MasterCnt; i++ {
			slotsRequired = slotsPerMaster
			if i == c.Cluster.MasterCnt-1 {
				// this change is only for the last master that needs slots
				slotsRequired = 16384 - (slotsPerMaster * i)
			}

			if i > 0 {
				if nodes, err = c.getOrderedNodes(pods); err != nil {
					return err
				}
			}

			to := nodes[i][0]
			for k := c.Cluster.MasterCnt; k < existingMasterCnt; k++ {
				from := nodes[k][0]
				// compare with slotsRequired
				if to.SlotsCnt < slotsRequired {
					// But compare with slotsPerMaster. Existing masters always need slots equal to
					// slotsPerMaster not slotsRequired since slotsRequired may change for last master
					// that is being added.
					if from.SlotsCnt > 0 {
						slots := from.SlotsCnt
						if slots > slotsRequired-to.SlotsCnt {
							slots = slotsRequired - to.SlotsCnt
						}

						if err = c.reshard(pods[0][0], nodes, k, i, slots); err != nil {
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
				if err = c.deleteNode(pods[0][0], nodeAddress(nodes[i][0].IP), nodes[i][j].ID); err != nil {
					return err
				}
				// TODO: Need to use a better alternative for successful completion of the above operation.
				time.Sleep(time.Second * 5)
			}
			if err = c.deleteNode(pods[0][0], nodeAddress(pods[0][0].Status.PodIP), nodes[i][0].ID); err != nil {
				return err
			}
			// TODO: Need to use a better alternative for successful completion of the above operation.
			time.Sleep(time.Second * 5)
		}
	}

	return nil
}

func (c Config) ensureNewMastersBeAdded(pods [][]*core.Pod) error {
	log.Infoln("Ensuring new masters be added...")

	var (
		err               error
		existingMasterCnt int
		nodes             [][]RedisNode
	)

	nodes, err = c.getOrderedNodes(pods)
	existingMasterCnt = len(nodes)

	if existingMasterCnt > 1 {
		// add new master(s)
		if existingMasterCnt < c.Cluster.MasterCnt {
			for i := existingMasterCnt; i < c.Cluster.MasterCnt; i++ {
				// ensure node must be empty before adding
				if err = c.clusterReset(pods[i][0], pods[i][0].Status.PodIP); err != nil {
					return err
				}
				// TODO: Need to use a better alternative for successful completion of the above operation.
				time.Sleep(time.Second * 5)

				if err = c.addNode(
					pods[0][0],
					nodeAddress(pods[i][0].Status.PodIP), nodeAddress(pods[0][0].Status.PodIP), ""); err != nil {
					return err
				}
				// TODO: Need to use a better alternative for successful completion of the above operation.
				time.Sleep(time.Second * 5)
			}
		}
	}

	return nil
}

func (c Config) rebalanceSlots(pods [][]*core.Pod) error {
	log.Infoln("Ensuring slots are rebalanced...")

	var (
		err                           error
		existingMasterCnt             int
		nodes                         [][]RedisNode
		masterIndicesWithLessSlots    []int
		masterIndicesWithExtraSlots   []int
		slotsPerMaster, slotsRequired int
	)

	nodes, err = c.getOrderedNodes(pods)

	existingMasterCnt = len(nodes)

	if existingMasterCnt > 1 {
		slotsPerMaster = 16384 / c.Cluster.MasterCnt
		for i := range nodes {
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
				if nodes, err = c.getOrderedNodes(pods); err != nil {
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

						if err = c.reshard(pods[0][0], nodes,
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

func (c Config) ensureNewSlavesBeAdded(pods [][]*core.Pod) error {
	log.Infoln("Ensuring new slaves be added...")

	var (
		err               error
		existingMasterCnt int
		nodes             [][]RedisNode
	)

	nodes, err = c.getOrderedNodes(pods)
	existingMasterCnt = len(nodes)

	if existingMasterCnt > 1 {
		// add new slave(s)
		for i := range nodes {
			if len(nodes[i])-1 < c.Cluster.Replicas {
				for j := len(nodes[i]); j <= c.Cluster.Replicas; j++ {
					// ensure node must be empty before adding
					if err = c.clusterReset(pods[i][j], pods[i][j].Status.PodIP); err != nil {
						return err
					}
					// TODO: Need to use a better alternative for successful completion of the above operation.
					time.Sleep(time.Second * 5)

					if err = c.addNode(
						pods[0][0],
						nodeAddress(pods[i][j].Status.PodIP), nodeAddress(nodes[i][0].IP), nodes[i][0].ID); err != nil {
						return err
					}
					time.Sleep(time.Second * 5)
				}
			}
		}
	}

	return nil
}

func (c Config) ensureCluster(pods [][]*core.Pod) error {
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

	nodes, err = c.getOrderedNodes(pods)
	if err != nil {
		return err
	}
	if len(nodes) > 1 {
		return nil
	}

	// create a cluster using the pods[0][0], pods[1][0], ...., pods[n][0] as master
	for i := 0; i < c.Cluster.MasterCnt; i++ {
		masterAddrs[i] = nodeAddress(pods[i][0].Status.PodIP)
		if nodesConf, err = c.getClusterNodes(pods[0][0], pods[i][0].Status.PodIP); err != nil {
			return err
		}
		masterNodeIds[i] = getNodeId(getMyConf(nodesConf))
	}
	if err = c.createCluster(pods[0][0], masterAddrs...); err != nil {
		return err
	}
	// TODO: Need to use a better alternative for successful completion of the above operation.
	time.Sleep(time.Second * 15)

	// now for each shard i
	//     add pods[i][1], pods[i][2], ... as slaves of pods[i][0]
	for i := 0; i < c.Cluster.MasterCnt; i++ {
		for j := 1; j <= c.Cluster.Replicas; j++ {
			if err = c.addNode(
				pods[0][0],
				nodeAddress(pods[i][j].Status.PodIP), masterAddrs[i], masterNodeIds[i]); err != nil {
				return err
			}
		}
	}
	// TODO: Need to use a better alternative for successful completion of the above operation.
	time.Sleep(time.Second * 15)

	return nil
}

func (c Config) configureClusterState(pods [][]*core.Pod) error {
	var err error

	if err = c.ensureCluster(pods); err != nil {
		return err
	}

	if err = c.ensureExtraSlavesBeRemoved(pods); err != nil {
		return err
	}

	if err = c.ensureExtraMastersBeRemoved(pods); err != nil {
		return err
	}

	if err = c.ensureNewMastersBeAdded(pods); err != nil {
		return err
	}
	if err = c.rebalanceSlots(pods); err != nil {
		return err
	}

	if err = c.ensureNewSlavesBeAdded(pods); err != nil {
		return err
	}

	return nil
}
