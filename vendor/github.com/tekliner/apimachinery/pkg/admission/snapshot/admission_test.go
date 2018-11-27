package snapshot

import (
	"net/http"
	"testing"

	"github.com/appscode/go/types"
	"github.com/appscode/kutil/meta"
	api "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	extFake "github.com/kubedb/apimachinery/client/clientset/versioned/fake"
	"github.com/kubedb/apimachinery/client/clientset/versioned/scheme"
	admission "k8s.io/api/admission/v1beta1"
	authenticationV1 "k8s.io/api/authentication/v1"
	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	clientSetScheme "k8s.io/client-go/kubernetes/scheme"
	store "kmodules.xyz/objectstore-api/api/v1"
)

func init() {
	scheme.AddToScheme(clientSetScheme.Scheme)
}

var requestKind = metaV1.GroupVersionKind{
	Group:   api.SchemeGroupVersion.Group,
	Version: api.SchemeGroupVersion.Version,
	Kind:    api.ResourceKindSnapshot,
}

func TestSnapshotValidator_Admit(t *testing.T) {
	for _, c := range cases {
		t.Run(c.testName, func(t *testing.T) {
			validator := SnapshotValidator{}

			validator.initialized = true
			validator.client = fake.NewSimpleClientset()
			validator.extClient = extFake.NewSimpleClientset()

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

			if c.operation == admission.Delete {
				req.Object = runtime.RawExtension{}
			}
			if c.operation != admission.Update {
				req.OldObject = runtime.RawExtension{}
			}

			if c.heatUp {
				if _, err := validator.extClient.KubedbV1alpha1().MongoDBs(c.namespace).Create(sampleMongoDB()); err != nil && !kerr.IsAlreadyExists(err) {
					t.Errorf(err.Error())
				}
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
	object     api.Snapshot
	oldObject  api.Snapshot
	heatUp     bool
	result     bool
}{
	{"Create Valid Snapshot",
		requestKind,
		"foo",
		"default",
		admission.Create,
		sampleSnapshot(),
		api.Snapshot{},
		true,
		true,
	},
	{"Create Invalid Snapshot",
		requestKind,
		"foo",
		"default",
		admission.Create,
		getAwkwardSnapshot(),
		api.Snapshot{},
		false,
		false,
	},
	{"Edit Status",
		requestKind,
		"foo",
		"default",
		admission.Update,
		editStatus(sampleSnapshot()),
		sampleSnapshot(),
		false,
		true,
	},
	{"Delete Snapshot",
		requestKind,
		"foo",
		"default",
		admission.Delete,
		sampleSnapshot(),
		api.Snapshot{},
		false,
		true,
	},
	{"Delete Non Existing Snapshot",
		requestKind,
		"foo",
		"default",
		admission.Delete,
		api.Snapshot{},
		api.Snapshot{},
		false,
		true,
	},
}

func sampleSnapshot() api.Snapshot {
	return api.Snapshot{
		TypeMeta: metaV1.TypeMeta{
			Kind:       api.ResourceKindSnapshot,
			APIVersion: api.SchemeGroupVersion.String(),
		},
		ObjectMeta: metaV1.ObjectMeta{
			Name:      "bar",
			Namespace: "default",
			Labels: map[string]string{
				api.LabelDatabaseKind: api.ResourceKindMongoDB,
			},
		},
		Spec: api.SnapshotSpec{
			DatabaseName: "foo",
			Backend: store.Backend{
				Local: &store.LocalSpec{
					MountPath: "/repo",
					VolumeSource: core.VolumeSource{
						EmptyDir: &core.EmptyDirVolumeSource{},
					},
				},
			},
		},
	}
}

func getAwkwardSnapshot() api.Snapshot {
	redis := sampleSnapshot()
	redis.Spec = api.SnapshotSpec{
		DatabaseName: "foo",
		Backend: store.Backend{
			StorageSecretName: "foo-secret",
			GCS: &store.GCSSpec{
				Bucket: "bar",
			},
		},
	}
	return redis
}

func editStatus(old api.Snapshot) api.Snapshot {
	old.Status = api.SnapshotStatus{
		Phase: api.SnapshotPhaseRunning,
	}
	return old
}

func sampleMongoDB() *api.MongoDB {
	return &api.MongoDB{
		TypeMeta: metaV1.TypeMeta{
			Kind:       api.ResourceKindMongoDB,
			APIVersion: api.SchemeGroupVersion.String(),
		},
		ObjectMeta: metaV1.ObjectMeta{
			Name:      "foo",
			Namespace: "default",
			Labels: map[string]string{
				api.LabelDatabaseKind: api.ResourceKindMongoDB,
			},
		},
		Spec: api.MongoDBSpec{
			Version:           "3.4",
			TerminationPolicy: api.TerminationPolicyDoNotTerminate,
			Storage: &core.PersistentVolumeClaimSpec{
				StorageClassName: types.StringP("standard"),
				Resources: core.ResourceRequirements{
					Requests: core.ResourceList{
						core.ResourceStorage: resource.MustParse("100Mi"),
					},
				},
			},
			Init: &api.InitSpec{
				ScriptSource: &api.ScriptSourceSpec{
					VolumeSource: core.VolumeSource{
						GitRepo: &core.GitRepoVolumeSource{
							Repository: "https://github.com/kubedb/mongodb-init-scripts.git",
							Directory:  ".",
						},
					},
				},
			},
		},
	}
}
