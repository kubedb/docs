package util

import (
	"fmt"
	"os"
	"time"

	"github.com/appscode/go/encoding/json/types"
	api "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewMemberAddEvent(memberName string, cl *api.Etcd) *v1.Event {
	event := newClusterEvent(cl)
	event.Type = v1.EventTypeNormal
	event.Reason = "New Member Added"
	event.Message = fmt.Sprintf("New member %s added to cluster", memberName)
	return event
}

func MemberRemoveEvent(memberName string, cl *api.Etcd) *v1.Event {
	event := newClusterEvent(cl)
	event.Type = v1.EventTypeNormal
	event.Reason = "Member Removed"
	event.Message = fmt.Sprintf("Existing member %s removed from the cluster", memberName)
	return event
}

func ReplacingDeadMemberEvent(memberName string, cl *api.Etcd) *v1.Event {
	event := newClusterEvent(cl)
	event.Type = v1.EventTypeNormal
	event.Reason = "Replacing Dead Member"
	event.Message = fmt.Sprintf("The dead member %s is being replaced", memberName)
	return event
}

func MemberUpgradedEvent(memberName string, oldVersion, newVersion types.StrYo, cl *api.Etcd) *v1.Event {
	event := newClusterEvent(cl)
	event.Type = v1.EventTypeNormal
	event.Reason = "Member Upgraded"
	event.Message = fmt.Sprintf("Member %s upgraded from %v to %v ", memberName, oldVersion, newVersion)
	return event
}

func newClusterEvent(cl *api.Etcd) *v1.Event {
	t := time.Now()
	return &v1.Event{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: cl.Name + "-",
			Namespace:    cl.Namespace,
		},
		InvolvedObject: v1.ObjectReference{
			APIVersion:      api.SchemeGroupVersion.String(),
			Kind:            api.ResourceKindEtcd,
			Name:            cl.Name,
			Namespace:       cl.Namespace,
			UID:             cl.UID,
			ResourceVersion: cl.ResourceVersion,
		},
		Source: v1.EventSource{
			Component: os.Getenv(EnvOperatorPodName),
		},
		// Each cluster event is unique so it should not be collapsed with other events
		FirstTimestamp: metav1.Time{Time: t},
		LastTimestamp:  metav1.Time{Time: t},
		Count:          int32(1),
	}
}
