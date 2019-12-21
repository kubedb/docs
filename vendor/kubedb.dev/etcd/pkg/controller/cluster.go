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
	"crypto/tls"
	"fmt"
	"reflect"
	"time"

	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha1"
	dbutil "kubedb.dev/apimachinery/client/clientset/versioned/typed/kubedb/v1alpha1/util"
	"kubedb.dev/etcd/pkg/util"

	"github.com/appscode/go/log"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

var (
	reconcileInterval         = 8 * time.Second
	podTerminationGracePeriod = int64(5)
)

type clusterEventType string

const (
	eventModifyCluster clusterEventType = "Modify"
)

type clusterEvent struct {
	typ     clusterEventType
	cluster *api.Etcd
}

type Cluster struct {
	logger *logrus.Entry

	cluster *api.Etcd

	// in memory state of the cluster
	// status is the source of truth after Cluster struct is materialized.
	status api.EtcdStatus

	eventCh chan *clusterEvent
	stopCh  chan struct{}

	// members repsersents the members in the etcd cluster.
	// the name of the member is the the name of the pod the member
	// process runs in.
	members util.MemberSet

	tlsConfig *tls.Config

	eventsCli corev1.EventInterface
}

func (c Controller) NewCluster(etcd *api.Etcd) {
	lg := logrus.WithField("pkg", "cluster").WithField("cluster-name", etcd.Name)

	cluster := &Cluster{
		logger:    lg,
		cluster:   etcd,
		eventCh:   make(chan *clusterEvent, 100),
		stopCh:    make(chan struct{}),
		status:    *(etcd.Status.DeepCopy()),
		eventsCli: c.Controller.Client.CoreV1().Events(etcd.Namespace),
	}

	go func() {
		if err := c.setup(cluster); err != nil {
			if cluster.status.Phase != api.DatabasePhaseFailed {
				cluster.status.Reason = err.Error()
				cluster.status.Phase = api.DatabasePhaseFailed
				if err := c.updateCRStatus(cluster); err != nil {
					cluster.logger.Errorf("failed to update cluster phase (%v): %v", api.DatabasePhaseFailed, err)
				}
			}
			return
		}
		c.run(cluster)
	}()

	c.clusters[etcd.Name] = cluster
}

func (c *Controller) setup(cluster *Cluster) error {
	var shouldCreateCluster bool
	switch cluster.status.Phase {
	case "":
		shouldCreateCluster = true
	case api.DatabasePhaseCreating:
		return errors.New("cluster failed to be created")
	case api.DatabasePhaseRunning:
		shouldCreateCluster = false
	default:
		return fmt.Errorf("unexpected cluster phase: %s", cluster.status.Phase)
	}

	/*if cluster.isSecureClient() {
		d, err := util.GetTLSDataFromSecret(c.config.KubeCli, c.cluster.Namespace, c.cluster.Spec.TLS.Static.OperatorSecret)
		if err != nil {
			return err
		}
		cluster.tlsConfig, err = util.NewTLSConfig(d.CertData, d.KeyData, d.CAData)
		if err != nil {
			return err
		}
	}*/

	if shouldCreateCluster {
		return c.create(cluster)
	}
	return nil
}

func (c *Controller) create(cl *Cluster) error {
	cl.status.Phase = api.DatabasePhaseCreating

	if err := c.updateCRStatus(cl); err != nil {
		return fmt.Errorf("cluster create: failed to update cluster phase (%v): %v", api.DatabasePhaseCreating, err)
	}

	return c.prepareSeedMember(cl)
}

func (c *Controller) prepareSeedMember(cl *Cluster) error {
	err := c.bootstrap(cl)
	if err != nil {
		return err
	}

	return nil
}

func (c *Controller) bootstrap(cl *Cluster) error {
	return c.startSeedMember(cl)
}

func (c *Cluster) Update(cl *api.Etcd) {
	c.send(&clusterEvent{
		typ:     eventModifyCluster,
		cluster: cl,
	})
}

func (c *Cluster) Delete() {
	c.logger.Info("cluster is deleted by user")
	close(c.stopCh)
}

func (c *Controller) startSeedMember(cl *Cluster) error {
	m := &util.Member{
		Name:         util.UniqueMemberName(cl.cluster.Name),
		Namespace:    cl.cluster.Namespace,
		SecurePeer:   cl.isSecurePeer(),
		SecureClient: cl.isSecureClient(),
	}
	ms := util.NewMemberSet(m)
	if _, _, err := c.createPod(cl.cluster, ms, m, "new"); err != nil {
		return fmt.Errorf("failed to create seed member (%s): %v", m.Name, err)
	}
	cl.members = ms
	cl.logger.Infof("cluster created with seed member (%s)", m.Name)
	_, err := cl.eventsCli.Create(util.NewMemberAddEvent(m.Name, cl.cluster))
	if err != nil {
		cl.logger.Errorf("failed to create new member add event: %v", err)
	}

	return nil
}

func (c *Controller) run(cluster *Cluster) {
	if err := c.setupServices(cluster); err != nil {
		cluster.logger.Errorf("fail to setup etcd services: %v", err)
	}
	cluster.status.Phase = api.DatabasePhaseRunning
	if err := c.updateCRStatus(cluster); err != nil {
		cluster.logger.Warningf("update initial CR status failed: %v", err)
	}
	if _, err := c.ensureAppBinding(cluster.cluster); err != nil {
		log.Errorln(err)
	}

	var rerr error
	for {
		select {
		case <-cluster.stopCh:
			return
		case event := <-cluster.eventCh:
			switch event.typ {
			case eventModifyCluster:
				err := cluster.handleUpdateEvent(event)
				if err != nil {
					cluster.status.Reason = err.Error()
					//c.reportFailedStatus()
					return
				}
			default:
				panic("unknown event type" + event.typ)
			}

		case <-time.After(reconcileInterval):
			//start := time.Now()

			/*if !c.cluster.Spec.DoNotPause {
				//c.status.PauseControl()
				c.logger.Infof("control is paused, skipping reconciliation")
				continue
			} else {
				//	c.status.Control()
			}*/

			running, pending, err := c.pollPods(cluster)
			if err != nil {
				cluster.logger.Errorf("fail to poll pods: %v", err)
				//	reconcileFailed.WithLabelValues("failed to poll pods").Inc()
				continue
			}

			if len(pending) > 0 {
				// Pod startup might take long, e.g. pulling image. It would deterministically become running or succeeded/failed later.
				cluster.logger.Infof("skip reconciliation: running (%v), pending (%v)", util.GetPodNames(running), util.GetPodNames(pending))
				//	reconcileFailed.WithLabelValues("not all pods are running").Inc()
				continue
			}
			if len(running) == 0 {
				// TODO: how to handle this case?
				//		cluster.Delete()
				//	delete(c.clusters, cluster.cluster.Name)
				cluster.logger.Warningf("all etcd pods are dead.")
				break
			}

			// On controller restore, we could have "members == nil"
			if rerr != nil || cluster.members == nil {
				rerr = cluster.updateMembers(podsToMemberSet(running, cluster.isSecureClient()))
				if rerr != nil {
					cluster.logger.Errorf("failed to update members: %v", rerr)
					break
				}
			}
			rerr = c.reconcile(cluster, running)
			if rerr != nil {
				cluster.logger.Errorf("failed to reconcile: %v", rerr)
				break
			}
			//c.updateMemberStatus(running)
			if err := c.updateCRStatus(cluster); err != nil {
				cluster.logger.Warningf("periodic update CR status failed: %v", err)
			}

			//reconcileHistogram.WithLabelValues(c.name()).Observe(time.Since(start).Seconds())
		}

		if rerr != nil {
			cluster.logger.Infoln(rerr)
			//reconcileFailed.WithLabelValues(rerr.Error()).Inc()
		}

		/*if isFatalError(rerr) {
			c.status.SetReason(rerr.Error())
			c.logger.Errorf("cluster failed: %v", rerr)
			c.reportFailedStatus()
			return
		}*/
	}
}

func (c *Cluster) handleUpdateEvent(event *clusterEvent) error {
	oldSpec := c.cluster.Spec.DeepCopy()
	c.cluster = event.cluster

	if isSpecEqual(event.cluster.Spec, *oldSpec) {
		// We have some fields that once created could not be mutated.
		if !reflect.DeepEqual(event.cluster.Spec, *oldSpec) {
			c.logger.Infof("ignoring update event: %#v", event.cluster.Spec)
		}
		return nil
	}
	// TODO: we can't handle another upgrade while an upgrade is in progress

	//c.logSpecUpdate(*oldSpec, event.cluster.Spec)
	return nil
}

func isSpecEqual(s1, s2 api.EtcdSpec) bool {
	if s1.Replicas != s2.Replicas || s1.Version != s2.Version {
		return false
	}
	return true
}
func (c *Controller) pollPods(cl *Cluster) (running, pending []*v1.Pod, err error) {
	podList, err := c.Controller.Client.CoreV1().Pods(cl.cluster.Namespace).List(metav1.ListOptions{
		LabelSelector: labels.SelectorFromSet(cl.cluster.OffshootLabels()).String(),
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list running pods: %v", err)
	}

	for i := range podList.Items {
		pod := &podList.Items[i]
		// Avoid polling deleted pods. k8s issue where deleted pods would sometimes show the status Pending
		// See https://github.com/coreos/etcd-operator/issues/1693
		if pod.DeletionTimestamp != nil {
			continue
		}
		if len(pod.OwnerReferences) < 1 {
			log.Infof("pollPods: ignore pod %v: no owner", pod.Name)
			continue
		}
		if pod.OwnerReferences[0].UID != cl.cluster.UID {
			log.Infof("pollPods: ignore pod %v: owner (%v) is not %v",
				pod.Name, pod.OwnerReferences[0].UID, cl.cluster.UID)
			continue
		}
		switch pod.Status.Phase {
		case v1.PodRunning:
			running = append(running, pod)
		case v1.PodPending:
			pending = append(pending, pod)
		}
	}

	return running, pending, nil
}

func (c *Controller) removePod(ns, name string) error {
	opts := metav1.NewDeleteOptions(podTerminationGracePeriod)
	err := c.Controller.Client.CoreV1().Pods(ns).Delete(name, opts)
	if err != nil {
		/*if !util.IsKubernetesResourceNotFoundError(err) {
			return err
		}*/
	}
	return nil
}

func (c *Controller) setupServices(cluster *Cluster) error {
	err := c.CreateClientService(cluster)
	if err != nil {
		return err
	}
	return c.CreatePeerService(cluster)
}

func (c *Cluster) send(ev *clusterEvent) {
	select {
	case c.eventCh <- ev:
		l, ecap := len(c.eventCh), cap(c.eventCh)
		if l > int(float64(ecap)*0.8) {
			c.logger.Warningf("eventCh buffer is almost full [%d/%d]", l, ecap)
		}
	case <-c.stopCh:
	}
}

func (c *Cluster) isSecurePeer() bool {
	if c.cluster.Spec.TLS == nil || c.cluster.Spec.TLS.Member == nil {
		return false
	}
	return len(c.cluster.Spec.TLS.Member.PeerSecret) != 0
}

func (c *Cluster) isSecureClient() bool {
	if c.cluster.Spec.TLS == nil {
		return false
	}
	return len(c.cluster.Spec.TLS.OperatorSecret) != 0
}

func (c *Controller) updateCRStatus(cl *Cluster) error {
	if reflect.DeepEqual(cl.cluster.Status, cl.status) {
		return nil
	}
	_, err := dbutil.UpdateEtcdStatus(c.Controller.ExtClient.KubedbV1alpha1(), cl.cluster, func(in *api.EtcdStatus) *api.EtcdStatus {
		in.Phase = cl.status.Phase
		in.ObservedGeneration = cl.cluster.Generation
		return in
	})
	return err
}
