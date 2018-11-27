package matcher

import (
	api "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	"github.com/onsi/gomega/types"
)

func HavePaused() types.GomegaMatcher {
	return &statusMatcher{
		expected: api.DormantDatabasePhasePaused,
	}
}

func HaveWipedOut() types.GomegaMatcher {
	return &statusMatcher{
		expected: api.DormantDatabasePhaseWipedOut,
	}
}

type statusMatcher struct {
	expected api.DormantDatabasePhase
}

func (matcher *statusMatcher) Match(actual interface{}) (success bool, err error) {
	phase := actual.(api.DormantDatabasePhase)
	return phase == matcher.expected, nil
}

func (matcher *statusMatcher) FailureMessage(actual interface{}) (message string) {
	return "Expected to be Running all Pods"
}

func (matcher *statusMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	return "Expected to be not Running all Pods"
}
