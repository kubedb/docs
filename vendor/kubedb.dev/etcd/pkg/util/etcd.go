package util

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/util/rand"
)

const (
	DefaultDialTimeout    = 5 * time.Second
	DefaultRequestTimeout = 5 * time.Second

	randomSuffixLength = 10
	maxNameLength      = 63 - randomSuffixLength - 1

	EnvOperatorPodName      = "MY_POD_NAME"
	EnvOperatorPodNamespace = "MY_POD_NAMESPACE"

	CliCertFile = "etcd-client.crt"
	CliKeyFile  = "etcd-client.key"
	CliCAFile   = "etcd-client-ca.crt"
)

func ListMembers(clientURLs []string, tc *tls.Config) (*clientv3.MemberListResponse, error) {
	cfg := clientv3.Config{
		Endpoints:   clientURLs,
		DialTimeout: DefaultDialTimeout,
		TLS:         tc,
	}
	etcdcli, err := clientv3.New(cfg)
	if err != nil {
		return nil, fmt.Errorf("list members failed: creating etcd client failed: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), DefaultRequestTimeout)
	resp, err := etcdcli.MemberList(ctx)
	cancel()
	etcdcli.Close()
	return resp, err
}

func RemoveMember(clientURLs []string, tc *tls.Config, id uint64) error {
	cfg := clientv3.Config{
		Endpoints:   clientURLs,
		DialTimeout: DefaultDialTimeout,
		TLS:         tc,
	}
	etcdcli, err := clientv3.New(cfg)
	if err != nil {
		return err
	}
	defer etcdcli.Close()

	ctx, cancel := context.WithTimeout(context.Background(), DefaultRequestTimeout)

	_, err = etcdcli.Cluster.MemberRemove(ctx, id)
	cancel()
	return err
}

func (ms MemberSet) IsEqual(other MemberSet) bool {
	if ms.Size() != other.Size() {
		return false
	}
	for n := range ms {
		if _, ok := other[n]; !ok {
			return false
		}
	}
	return true
}

func (ms MemberSet) Diff(other MemberSet) MemberSet {
	diff := MemberSet{}
	for n, m := range ms {
		if _, ok := other[n]; !ok {
			diff[n] = m
		}
	}
	return diff
}

func (ms MemberSet) Size() int {
	return len(ms)
}

var validPeerURL = regexp.MustCompile(`^\w+:\/\/[\w\.\-]+(:\d+)?$`)

func MemberNameFromPeerURL(pu string) (string, error) {
	// url.Parse has very loose validation. We do our own validation.
	if !validPeerURL.MatchString(pu) {
		return "", errors.New("invalid PeerURL format")
	}
	u, err := url.Parse(pu)
	if err != nil {
		return "", err
	}
	path := strings.Split(u.Host, ":")[0]
	name := strings.Split(path, ".")[0]
	return name, err
}

func UniqueMemberName(clusterName string) string {
	suffix := rand.String(randomSuffixLength)
	if len(clusterName) > maxNameLength {
		clusterName = clusterName[:maxNameLength]
	}
	return clusterName + "-" + suffix
}
