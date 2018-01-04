package eventer

import (
	"github.com/appscode/go/log"
	core "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/record"
)

const (
	EventReasonPausing              string = "Pausing"
	EventReasonWipingOut            string = "WipingOut"
	EventReasonFailedToCreate       string = "Failed"
	EventReasonFailedToPause        string = "Failed"
	EventReasonFailedToDelete       string = "Failed"
	EventReasonFailedToWipeOut      string = "Failed"
	EventReasonFailedToGet          string = "Failed"
	EventReasonFailedToInitialize   string = "Failed"
	EventReasonFailedToList         string = "Failed"
	EventReasonFailedToResume       string = "Failed"
	EventReasonFailedToSchedule     string = "Failed"
	EventReasonFailedToStart        string = "Failed"
	EventReasonFailedToUpdate       string = "Failed"
	EventReasonIgnoredSnapshot      string = "IgnoredSnapshot"
	EventReasonInitializing         string = "Initializing"
	EventReasonInvalid              string = "Invalid"
	EventReasonResuming             string = "Resuming"
	EventReasonSnapshotFailed       string = "SnapshotFailed"
	EventReasonStarting             string = "Starting"
	EventReasonSuccessful           string = "Successful"
	EventReasonSuccessfulCreate     string = "SuccessfulCreate"
	EventReasonSuccessfulPause      string = "SuccessfulPause"
	EventReasonSuccessfulWipeOut    string = "SuccessfulWipeOut"
	EventReasonSuccessfulSnapshot   string = "SuccessfulSnapshot"
	EventReasonSuccessfulInitialize string = "SuccessfulInitialize"
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
