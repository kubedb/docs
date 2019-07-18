package util

import (
	"fmt"
	"strings"
	"time"
)

const DefaultTimeoutSecond = 5 * time.Second

type Member struct {
	Name      string
	Namespace string
	Service   string

	ID uint64

	SecureClient bool
	SecurePeer   bool
}

func NewMember(name, namespace, service string) *Member {
	return &Member{
		Name:         name,
		Namespace:    namespace,
		Service:      service,
		SecureClient: false,
		SecurePeer:   false,
	}
}
func (m *Member) BuildEtcdArgs() []string {
	return []string{
		fmt.Sprintf("--data-dir=%s", "/data/db"),
		fmt.Sprintf("--name=%s", m.Name),
		fmt.Sprintf("--initial-advertise-peer-urls=%s", m.PeerURL()),
		fmt.Sprintf("--listen-peer-urls=%s", m.ListenPeerURL()),
		fmt.Sprintf("--listen-client-urls=%s", m.ListenClientURL()),
		fmt.Sprintf("--advertise-client-urls=%s", m.ClientURL()),
		//	fmt.Sprintf("--initial-cluster=%s", strings.Join(initialCluster, ",")),
		//fmt.Sprintf("--initial-cluster-token=%s", etcd.OffshootName()),
	}
}

func (m *Member) Addr() string {
	return fmt.Sprintf("%s.%s.%s.svc", m.Name, clusterNameFromMemberName(m.Name), m.Namespace)
}

func (m *Member) clientScheme() string {
	if m.SecureClient {
		return "https"
	}
	return "http"
}

func (m *Member) peerScheme() string {
	if m.SecurePeer {
		return "https"
	}
	return "http"
}

func (m *Member) ListenClientURL() string {
	return fmt.Sprintf("%s://0.0.0.0:2379", m.clientScheme())
}
func (m *Member) ListenPeerURL() string {
	return fmt.Sprintf("%s://0.0.0.0:2380", m.peerScheme())
}

func (m *Member) PeerURL() string {
	return fmt.Sprintf("%s://%s:2380", m.peerScheme(), m.Addr())
}

func (m *Member) ClientURL() string {
	return fmt.Sprintf("%s://%s:2379", m.clientScheme(), m.Addr())
}

type MemberSet map[string]*Member

func NewMemberSet(ms ...*Member) MemberSet {
	res := MemberSet{}
	for _, m := range ms {
		res[m.Name] = m
	}
	return res
}

func (ms MemberSet) Add(m *Member) {
	ms[m.Name] = m
}

func (ms MemberSet) Remove(name string) {
	delete(ms, name)
}

func (ms MemberSet) PickOne() *Member {
	for _, m := range ms {
		return m
	}
	panic("empty")
}

func (ms MemberSet) PeerURLPairs() []string {
	ps := make([]string, 0)
	for _, m := range ms {
		ps = append(ps, fmt.Sprintf("%s=%s", m.Name, m.PeerURL()))
	}
	return ps
}

func (ms MemberSet) ClientURLs() []string {
	endpoints := make([]string, 0, len(ms))
	for _, m := range ms {
		endpoints = append(endpoints, m.ClientURL())
	}
	return endpoints
}

func clusterNameFromMemberName(mn string) string {
	i := strings.LastIndex(mn, "-")
	if i == -1 {
		panic(fmt.Sprintf("unexpected member name: %s", mn))
	}
	return mn[:i]
}
