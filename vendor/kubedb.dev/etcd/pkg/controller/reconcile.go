/*
Copyright The KubeDB Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package controller

import (
	"context"
	"errors"
	"fmt"

	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha1"
	"kubedb.dev/etcd/pkg/util"

	"github.com/appscode/go/log"
	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/etcdserver/api/v3rpc/rpctypes"
	v1 "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/kubernetes"
)

var ErrLostQuorum = errors.New("lost quorum")

func (c *Controller) reconcile(cl *Cluster, pods []*v1.Pod) error {
	sp := cl.cluster.Spec

	running := podsToMemberSet(pods, cl.isSecureClient())
	if !running.IsEqual(cl.members) || int32(cl.members.Size()) != *sp.Replicas {
		return c.reconcileMembers(cl, running)
	}
	//c.status.ClearCondition(api.ClusterConditionScaling)

	if needUpgrade(pods, sp) {
		//c.status.UpgradeVersionTo(sp.Version)

		m := pickOneOldMember(pods, string(sp.Version))
		return c.upgradeOneMember(cl, m)
	}
	//c.status.ClearCondition(api.ClusterConditionUpgrading)

	//c.status.SetVersion(sp.Version)
	//c.status.SetReadyCondition()

	return nil
}

func (c *Controller) reconcileMembers(cl *Cluster, running util.MemberSet) error {
	log.Infof("running members: %v", running)
	log.Infof("cluster membership: %v", cl.members)

	unknownMembers := running.Diff(cl.members)
	if unknownMembers.Size() > 0 {
		log.Infof("removing unexpected pods: %v", unknownMembers)
		for _, m := range unknownMembers {
			if err := c.removePod(cl.cluster.Namespace, m.Name); err != nil {
				return err
			}
		}
	}
	L := running.Diff(unknownMembers)

	if L.Size() == cl.members.Size() {
		return c.resize(cl)
	}

	if L.Size() < cl.members.Size()/2+1 {
		return ErrLostQuorum
	}

	log.Infoln("removing one dead member")
	// remove dead members that doesn't have any running pods before doing resizing.
	return c.removeDeadMember(cl, cl.members.Diff(L).PickOne())
}

func (c *Controller) resize(cl *Cluster) error {
	if cl.members.Size() == int(*cl.cluster.Spec.Replicas) {
		return nil
	}

	if cl.members.Size() < int(*cl.cluster.Spec.Replicas) {
		return c.addOneMember(cl)
	}

	return c.removeOneMember(cl)
}

func (c *Controller) addOneMember(cl *Cluster) error {
	cfg := clientv3.Config{
		Endpoints:   cl.members.ClientURLs(),
		DialTimeout: util.DefaultDialTimeout,
		TLS:         cl.tlsConfig,
	}
	etcdcli, err := clientv3.New(cfg)
	if err != nil {
		return fmt.Errorf("add one member failed: creating etcd client failed %v", err)
	}
	defer etcdcli.Close()

	newMember := cl.newMember()
	ctx, cancel := context.WithTimeout(context.Background(), util.DefaultRequestTimeout)
	resp, err := etcdcli.MemberAdd(ctx, []string{newMember.PeerURL()})
	cancel()
	if err != nil {
		return fmt.Errorf("fail to add new member (%s): %v", newMember.Name, err)
	}
	newMember.ID = resp.Member.ID
	cl.members.Add(newMember)

	_, _, err = c.createPod(cl.cluster, cl.members, newMember, "existing")
	if err != nil {
		return fmt.Errorf("fail to create member's pod (%s): %v", newMember.Name, err)
	}
	log.Infof("added member (%s)", newMember.Name)
	_, err = cl.eventsCli.Create(util.NewMemberAddEvent(newMember.Name, cl.cluster))
	if err != nil {
		cl.logger.Errorf("failed to create new member add event: %v", err)
	}
	// Check StatefulSet Pod status
	/*if vt != kutil.VerbUnchanged {
		if err := c.checkStatefulSetPodStatus(statefulSet); err != nil {
			if ref, rerr := reference.GetReference(clientsetscheme.Scheme, etcd); rerr == nil {
				c.recorder.Eventf(
					ref,
					v1.EventTypeWarning,
					eventer.EventReasonFailedToStart,
					`Failed to CreateOrPatch StatefulSet. Reason: %v`,
					err,
				)
			}
			return kutil.VerbUnchanged, err
		}
		if ref, rerr := reference.GetReference(clientsetscheme.Scheme, etcd); rerr == nil {
			c.recorder.Eventf(
				ref,
				v1.EventTypeNormal,
				eventer.EventReasonSuccessful,
				"Successfully %v StatefulSet",
				vt,
			)
		}
	}*/
	return nil
}

func (c *Controller) removeOneMember(cl *Cluster) error {
	return c.removeMember(cl, cl.members.PickOne())
}

func (c *Controller) removeDeadMember(cl *Cluster, toRemove *util.Member) error {
	log.Infof("removing dead member %q", toRemove.Name)
	_, err := cl.eventsCli.Create(util.ReplacingDeadMemberEvent(toRemove.Name, cl.cluster))
	if err != nil {
		cl.logger.Errorf("failed to create replacing dead member event: %v", err)
	}

	return c.removeMember(cl, toRemove)
}

func (c *Controller) removeMember(cl *Cluster, toRemove *util.Member) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("remove member (%s) failed: %v", toRemove.Name, err)
		}
	}()

	err = util.RemoveMember(cl.members.ClientURLs(), cl.tlsConfig, toRemove.ID)
	if err != nil {
		switch err {
		case rpctypes.ErrMemberNotFound:
			log.Infof("etcd member (%v) has been removed with id %d", toRemove.Name, toRemove.ID)
		default:
			return err
		}
	}
	cl.members.Remove(toRemove.Name)

	if err := c.removePod(cl.cluster.Namespace, toRemove.Name); err != nil {
		return err
	}
	if cl.cluster.Spec.Storage != nil {
		err = cl.removePVC(c.Controller.Client, toRemove.Name)
		if err != nil {
			return err
		}
	}
	cl.logger.Infof("removed member (%v) with ID (%d)", toRemove.Name, toRemove.ID)
	return nil
}

func (c *Cluster) removePVC(client kubernetes.Interface, pvcName string) error {
	err := client.CoreV1().PersistentVolumeClaims(c.cluster.Namespace).Delete(pvcName, nil)
	if err != nil && !kerr.IsNotFound(err) {
		return fmt.Errorf("remove pvc (%s) failed: %v", pvcName, err)
	}
	return nil
}

func needUpgrade(pods []*v1.Pod, cs api.EtcdSpec) bool {
	return len(pods) == int(*cs.Replicas) && pickOneOldMember(pods, string(cs.Version)) != nil
}

func pickOneOldMember(pods []*v1.Pod, newVersion string) *util.Member {
	for _, pod := range pods {
		if util.GetEtcdVersion(pod) == newVersion {
			continue
		}
		return &util.Member{Name: pod.Name, Namespace: pod.Namespace}
	}
	return nil
}
