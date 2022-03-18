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
	"strings"

	"github.com/pkg/errors"
	core "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
	"kmodules.xyz/client-go/tools/exec"
)

func (c Config) createCluster(useTLS bool, pod *core.Pod, addrs ...string) error {
	options := []func(options *exec.Options){
		exec.Input("yes"),
		exec.Command(c.ClusterCreateCmd(useTLS, 0, addrs...)...),
		exec.Container("redis"),
	}
	_, err := exec.ExecIntoPod(c.RestConfig, pod, options...)
	if err != nil {
		return errors.Wrapf(err, "Failed to create cluster using (%v)", addrs)
	}

	return nil
}

func (c Config) addNode(useTLS bool, pod *core.Pod, newAddr, existingAddr, masterId string) error {
	var err error

	if masterId == "" {
		options := []func(options *exec.Options){
			exec.Command(c.AddNodeAsMasterCmd(useTLS, newAddr, existingAddr)...),
			exec.Container("redis"),
		}
		if _, err = exec.ExecIntoPod(c.RestConfig, pod, options...); err != nil {
			return errors.Wrapf(err, "Failed to add %q as a master", newAddr)
		}
	} else {
		options := []func(options *exec.Options){
			exec.Command(c.AddNodeAsSlaveCmd(useTLS, newAddr, existingAddr, masterId)...),
			exec.Container("redis"),
		}
		if _, err = exec.ExecIntoPod(c.RestConfig, pod, options...); err != nil {
			return errors.Wrapf(err, "Failed to add %q as a slave of master with id %q", newAddr, masterId)
		}
	}

	return nil
}

func (c Config) deleteNode(useTLS bool, pod *core.Pod, existingAddr, deletingNodeID string) error {
	options := []func(options *exec.Options){
		exec.Command(c.DeleteNodeCmd(useTLS, existingAddr, deletingNodeID)...),
		exec.Container("redis"),
	}
	_, err := exec.ExecIntoPod(c.RestConfig, pod, options...)
	if err != nil && !strings.Contains(err.Error(), "command terminated with exit code 1") {
		return errors.Wrapf(err, "Failed to delete node with ID %q", deletingNodeID)
	}

	return nil
}

func (c Config) ping(useTLS bool, pod *core.Pod, ip string) (string, error) {
	options := []func(options *exec.Options){
		exec.Command(PingCmd(useTLS, ip)...),
		exec.Container("redis"),
	}
	pong, err := exec.ExecIntoPod(c.RestConfig, pod, options...)
	if err != nil {
		return "", errors.Wrapf(err, "Failed to ping %q", pod.Status.PodIP)
	}

	return strings.TrimSpace(pong), nil
}

func (c Config) getClusterNodes(useTLS bool, pod *core.Pod, ip string) (string, error) {
	options := []func(options *exec.Options){
		exec.Command(ClusterNodesCmd(useTLS, ip)...),
		exec.Container("redis"),
	}
	out, err := exec.ExecIntoPod(c.RestConfig, pod, options...)
	if err != nil {
		return "", errors.Wrapf(err, "Failed to get cluster nodes from %q", ip)
	}

	return strings.TrimSpace(out), nil
}

func (c Config) clusterMeet(useTLS bool, pod *core.Pod, senderIP, receiverIP, receiverPort string) error {
	options := []func(options *exec.Options){
		exec.Command(ClusterMeetCmd(useTLS, senderIP, receiverIP, receiverPort)...),
		exec.Container("redis"),
	}
	_, err := exec.ExecIntoPod(c.RestConfig, pod, options...)
	if err != nil {
		return errors.Wrapf(err, "Failed to meet node %q with node %q", senderIP, receiverIP)
	}

	return nil
}

func (c Config) clusterReset(useTLS bool, pod *core.Pod, ip, resetType string) error {
	options := []func(options *exec.Options){
		exec.Command(ClusterResetCmd(useTLS, ip, resetType)...),
		exec.Container("redis"),
	}
	_, err := exec.ExecIntoPod(c.RestConfig, pod, options...)
	if err != nil {
		return errors.Wrapf(err, "Failed to reset node %q", ip)
	}

	return nil
}

func (c Config) clusterFailover(useTLS bool, pod *core.Pod, ip string) error {
	options := []func(options *exec.Options){
		exec.Command(ClusterFailoverCmd(useTLS, ip)...),
		exec.Container("redis"),
	}
	_, err := exec.ExecIntoPod(c.RestConfig, pod, options...)
	if err != nil {
		return errors.Wrapf(err, "Failed to failover node %q", ip)
	}

	return nil
}

func (c Config) clusterReplicate(useTLS bool, pod *core.Pod, receivingNodeIP, masterNodeID string) error {
	options := []func(options *exec.Options){
		exec.Command(ClusterReplicateCmd(useTLS, receivingNodeIP, masterNodeID)...),
		exec.Container("redis"),
	}
	_, err := exec.ExecIntoPod(c.RestConfig, pod, options...)
	if err != nil {
		return errors.Wrapf(err, "Failed to replicate node %q of node with ID %s",
			receivingNodeIP, masterNodeID)
	}

	return nil
}

func (c Config) reshard(useTLS bool, pod *core.Pod, nodes [][]RedisNode, src, dst, requstedSlotsCount int) error {
	klog.Infof("Resharding %d slots from %q to %q...", requstedSlotsCount, nodes[src][0].IP, nodes[dst][0].IP)

	var (
		need int
		err  error
	)

	need = requstedSlotsCount

	for i := range nodes[src][0].SlotStart {
		if need <= 0 {
			break
		}

		start := nodes[src][0].SlotStart[i]
		end := nodes[src][0].SlotEnd[i]
		if end-start+1 > need {
			end = start + need - 1
		}
		cmd := c.ReshardCmd(useTLS, nodes[src][0].IP, nodes[src][0].ID, nodes[dst][0].IP, nodes[dst][0].ID, start, end)
		options := []func(options *exec.Options){
			exec.Command(cmd...),
			exec.Container("redis"),
		}
		_, err = exec.ExecIntoPod(c.RestConfig, pod, options...)
		if err != nil {
			return errors.Wrapf(err, "Failed to reshard %d slots from %q to %q",
				requstedSlotsCount, nodes[src][0].IP, nodes[dst][0].IP)
		}

		need -= (end - start + 1)
	}

	return nil
}
