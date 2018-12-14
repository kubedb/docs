package eventer

import (
	"fmt"
	"time"

	"github.com/appscode/go/log"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/tools/reference"
)

const (
	EventReasonPausing                      string = "Pausing"
	EventReasonWipingOut                    string = "WipingOut"
	EventReasonFailedToCreate               string = "Failed"
	EventReasonFailedToPause                string = "Failed"
	EventReasonFailedToDelete               string = "Failed"
	EventReasonFailedToWipeOut              string = "Failed"
	EventReasonFailedToGet                  string = "Failed"
	EventReasonFailedToInitialize           string = "Failed"
	EventReasonFailedToList                 string = "Failed"
	EventReasonFailedToResume               string = "Failed"
	EventReasonFailedToSchedule             string = "Failed"
	EventReasonFailedToStart                string = "Failed"
	EventReasonFailedToUpdate               string = "Failed"
	EventReasonIgnoredSnapshot              string = "IgnoredSnapshot"
	EventReasonInitializing                 string = "Initializing"
	EventReasonInvalid                      string = "Invalid"
	EventReasonResuming                     string = "Resuming"
	EventReasonSnapshotFailed               string = "SnapshotFailed"
	EventReasonSnapshotError                string = "SnapshotError"
	EventReasonStarting                     string = "Starting"
	EventReasonSuccessful                   string = "Successful"
	EventReasonSuccessfulCreate             string = "SuccessfulCreate"
	EventReasonSuccessfulPause              string = "SuccessfulPause"
	EventReasonSuccessfulWipeOut            string = "SuccessfulWipeOut"
	EventReasonSuccessfulSnapshot           string = "SuccessfulSnapshot"
	EventReasonSuccessfulInitialize         string = "SuccessfulInitialize"
	EventReasonAdmissionWebhookNotActivated string = "AdmissionWebhookNotActivated"
)

func NewEventRecorder(client kubernetes.Interface, component string) record.EventRecorder {
	// Event Broadcaster
	broadcaster := record.NewBroadcaster()
	broadcaster.StartEventWatcher(
		func(event *core.Event) {
			if _, err := client.Core().Events(event.Namespace).Create(event); err != nil {
				log.Errorln(err)
			}
		},
	)

	return broadcaster.NewRecorder(scheme.Scheme, core.EventSource{Component: component})
}

func CreateEvent(client kubernetes.Interface, component string, obj runtime.Object, eventType, reason, message string) (*core.Event, error) {
	ref, err := reference.GetReference(scheme.Scheme, obj)
	if err != nil {
		return nil, err
	}

	t := metav1.Time{Time: time.Now()}

	return client.CoreV1().Events(ref.Namespace).Create(&core.Event{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%v.%x", ref.Name, t.UnixNano()),
			Namespace: ref.Namespace,
		},
		InvolvedObject: *ref,
		Reason:         reason,
		Message:        message,
		FirstTimestamp: t,
		LastTimestamp:  t,
		Count:          1,
		Type:           eventType,
		Source:         core.EventSource{Component: component},
	})
}

func CreateEventWithLog(client kubernetes.Interface, component string, obj runtime.Object, eventType, reason, message string) {
	event, err := CreateEvent(client, component, obj, eventType, reason, message)
	if err != nil {
		log.Errorln("Failed to write event, reason: ", err)
	} else {
		log.Infoln("Event created: ", event.Name)
	}
}
