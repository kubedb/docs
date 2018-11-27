package admission

import (
	"net/http"
	"testing"

	"github.com/appscode/go/types"
	"github.com/appscode/kutil/meta"
	catalog "github.com/kubedb/apimachinery/apis/catalog/v1alpha1"
	api "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	extFake "github.com/kubedb/apimachinery/client/clientset/versioned/fake"
	"github.com/kubedb/apimachinery/client/clientset/versioned/scheme"
	admission "k8s.io/api/admission/v1beta1"
	apps "k8s.io/api/apps/v1"
	authenticationV1 "k8s.io/api/authentication/v1"
	core "k8s.io/api/core/v1"
	storageV1beta1 "k8s.io/api/storage/v1beta1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	clientSetScheme "k8s.io/client-go/kubernetes/scheme"
	mona "kmodules.xyz/monitoring-agent-api/api/v1"
)

func init() {
	scheme.AddToScheme(clientSetScheme.Scheme)
}

var requestKind = metaV1.GroupVersionKind{
	Group:   api.SchemeGroupVersion.Group,
	Version: api.SchemeGroupVersion.Version,
	Kind:    api.ResourceKindRedis,
}

func TestRedisValidator_Admit(t *testing.T) {
	for _, c := range cases {
		t.Run(c.testName, func(t *testing.T) {
			validator := RedisValidator{}

			validator.initialized = true
			validator.extClient = extFake.NewSimpleClientset(
				&catalog.RedisVersion{
					ObjectMeta: metaV1.ObjectMeta{
						Name: "4.0",
					},
				},
			)
			validator.client = fake.NewSimpleClientset(
				&storageV1beta1.StorageClass{
					ObjectMeta: metaV1.ObjectMeta{
						Name: "standard",
					},
				},
			)

			objJS, err := meta.MarshalToJson(&c.object, api.SchemeGroupVersion)
			if err != nil {
				panic(err)
			}
			oldObjJS, err := meta.MarshalToJson(&c.oldObject, api.SchemeGroupVersion)
			if err != nil {
				panic(err)
			}

			req := new(admission.AdmissionRequest)

			req.Kind = c.kind
			req.Name = c.objectName
			req.Namespace = c.namespace
			req.Operation = c.operation
			req.UserInfo = authenticationV1.UserInfo{}
			req.Object.Raw = objJS
			req.OldObject.Raw = oldObjJS

			if c.heatUp {
				if _, err := validator.extClient.KubedbV1alpha1().Redises(c.namespace).Create(&c.object); err != nil && !kerr.IsAlreadyExists(err) {
					t.Errorf(err.Error())
				}
			}
			if c.operation == admission.Delete {
				req.Object = runtime.RawExtension{}
			}
			if c.operation != admission.Update {
				req.OldObject = runtime.RawExtension{}
			}

			response := validator.Admit(req)
			if c.result == true {
				if response.Allowed != true {
					t.Errorf("expected: 'Allowed=true'. but got response: %v", response)
				}
			} else if c.result == false {
				if response.Allowed == true || response.Result.Code == http.StatusInternalServerError {
					t.Errorf("expected: 'Allowed=false', but got response: %v", response)
				}
			}
		})
	}

}

var cases = []struct {
	testName   string
	kind       metaV1.GroupVersionKind
	objectName string
	namespace  string
	operation  admission.Operation
	object     api.Redis
	oldObject  api.Redis
	heatUp     bool
	result     bool
}{
	{"Create Valid Redis",
		requestKind,
		"foo",
		"default",
		admission.Create,
		sampleRedis(),
		api.Redis{},
		false,
		true,
	},
	{"Create Invalid Redis",
		requestKind,
		"foo",
		"default",
		admission.Create,
		getAwkwardRedis(),
		api.Redis{},
		false,
		false,
	},
	{"Edit Redis Spec.Version",
		requestKind,
		"foo",
		"default",
		admission.Update,
		editSpecVersion(sampleRedis()),
		sampleRedis(),
		false,
		false,
	},
	{"Edit Status",
		requestKind,
		"foo",
		"default",
		admission.Update,
		editStatus(sampleRedis()),
		sampleRedis(),
		false,
		true,
	},
	{"Edit Spec.Monitor",
		requestKind,
		"foo",
		"default",
		admission.Update,
		editSpecMonitor(sampleRedis()),
		sampleRedis(),
		false,
		true,
	},
	{"Edit Invalid Spec.Monitor",
		requestKind,
		"foo",
		"default",
		admission.Update,
		editSpecInvalidMonitor(sampleRedis()),
		sampleRedis(),
		false,
		false,
	},
	{"Edit Spec.TerminationPolicy",
		requestKind,
		"foo",
		"default",
		admission.Update,
		pauseDatabase(sampleRedis()),
		sampleRedis(),
		false,
		true,
	},
	{"Delete Redis when Spec.TerminationPolicy = DoNotTerminate",
		requestKind,
		"foo",
		"default",
		admission.Delete,
		sampleRedis(),
		api.Redis{},
		true,
		false,
	},
	{"Delete Redis when Spec.TerminationPolicy = Pause",
		requestKind,
		"foo",
		"default",
		admission.Delete,
		pauseDatabase(sampleRedis()),
		api.Redis{},
		true,
		true,
	},
	{"Delete Non Existing Redis",
		requestKind,
		"foo",
		"default",
		admission.Delete,
		api.Redis{},
		api.Redis{},
		false,
		true,
	},
}

func sampleRedis() api.Redis {
	return api.Redis{
		TypeMeta: metaV1.TypeMeta{
			Kind:       api.ResourceKindRedis,
			APIVersion: api.SchemeGroupVersion.String(),
		},
		ObjectMeta: metaV1.ObjectMeta{
			Name:      "foo",
			Namespace: "default",
			Labels: map[string]string{
				api.LabelDatabaseKind: api.ResourceKindRedis,
			},
		},
		Spec: api.RedisSpec{
			Version:     "4.0",
			Replicas:    types.Int32P(1),
			StorageType: api.StorageTypeDurable,
			Storage: &core.PersistentVolumeClaimSpec{
				StorageClassName: types.StringP("standard"),
				Resources: core.ResourceRequirements{
					Requests: core.ResourceList{
						core.ResourceStorage: resource.MustParse("100Mi"),
					},
				},
			},
			UpdateStrategy: apps.StatefulSetUpdateStrategy{
				Type: apps.RollingUpdateStatefulSetStrategyType,
			},
			TerminationPolicy: api.TerminationPolicyDoNotTerminate,
		},
	}
}

func getAwkwardRedis() api.Redis {
	redis := sampleRedis()
	redis.Spec.Version = "3.0"
	return redis
}

func editSpecVersion(old api.Redis) api.Redis {
	old.Spec.Version = "4.4"
	return old
}

func editStatus(old api.Redis) api.Redis {
	old.Status = api.RedisStatus{
		Phase: api.DatabasePhaseCreating,
	}
	return old
}

func editSpecMonitor(old api.Redis) api.Redis {
	old.Spec.Monitor = &mona.AgentSpec{
		Agent: mona.AgentPrometheusBuiltin,
		Prometheus: &mona.PrometheusSpec{
			Port: 4567,
		},
	}
	return old
}

// should be failed because more fields required for COreOS Monitoring
func editSpecInvalidMonitor(old api.Redis) api.Redis {
	old.Spec.Monitor = &mona.AgentSpec{
		Agent: mona.AgentCoreOSPrometheus,
	}
	return old
}

func pauseDatabase(old api.Redis) api.Redis {
	old.Spec.TerminationPolicy = api.TerminationPolicyPause
	return old
}
