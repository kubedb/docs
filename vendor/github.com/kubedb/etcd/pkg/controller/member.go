package controller

import (
	"fmt"

	"github.com/coreos/etcd/etcdserver/etcdserverpb"
	"github.com/kubedb/etcd/pkg/util"
	"github.com/pkg/errors"
	"k8s.io/api/core/v1"
)

func (c *Cluster) updateMembers(known util.MemberSet) error {
	resp, err := util.ListMembers(known.ClientURLs(), c.tlsConfig)
	if err != nil {
		return err
	}
	members := util.MemberSet{}
	for _, m := range resp.Members {
		name, err := getMemberName(m, c.cluster.GetName())
		if err != nil {
			return errors.Wrap(err, "get member name failed")
		}

		members[name] = &util.Member{
			Name:         name,
			Namespace:    c.cluster.Namespace,
			ID:           m.ID,
			SecurePeer:   c.isSecurePeer(),
			SecureClient: c.isSecureClient(),
		}
	}
	c.members = members
	return nil
}

func (c *Cluster) newMember() *util.Member {
	name := util.UniqueMemberName(c.cluster.Name)
	return &util.Member{
		Name:         name,
		Namespace:    c.cluster.Namespace,
		SecurePeer:   c.isSecurePeer(),
		SecureClient: c.isSecureClient(),
	}
}

func podsToMemberSet(pods []*v1.Pod, sc bool) util.MemberSet {
	members := util.MemberSet{}
	for _, pod := range pods {
		m := &util.Member{Name: pod.Name, Namespace: pod.Namespace, SecureClient: sc}
		members.Add(m)
	}
	return members
}

func getMemberName(m *etcdserverpb.Member, clusterName string) (string, error) {
	name, err := util.MemberNameFromPeerURL(m.PeerURLs[0])
	if err != nil {
		return "", fmt.Errorf("invalid member peerURL (%s): %v", m.PeerURLs[0], err)
	}
	return name, nil
}
