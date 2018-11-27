package namespace

import (
	"net/http"
	"testing"

	"github.com/appscode/kutil/meta"
	api "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	"github.com/kubedb/apimachinery/client/clientset/versioned/scheme"
	admission "k8s.io/api/admission/v1beta1"
	authenticationv1 "k8s.io/api/authentication/v1"
	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	fake_dynamic "k8s.io/client-go/dynamic/fake"
	clientsetscheme "k8s.io/client-go/kubernetes/scheme"
)

func init() {
	scheme.AddToScheme(clientsetscheme.Scheme)
}

var requestKind = metav1.GroupVersionKind{
	Group:   core.SchemeGroupVersion.Group,
	Version: core.SchemeGroupVersion.Version,
	Kind:    "Namespace",
}

func TestNamespaceValidator_Admit(t *testing.T) {
	for _, c := range cases {
		t.Run(c.testName, func(t *testing.T) {
			validator := NamespaceValidator{
				Resources: []string{api.ResourcePluralPostgres},
			}
			validator.initialized = true
			validator.dc = fake_dynamic.NewSimpleDynamicClient(clientsetscheme.Scheme)

			objJS, err := meta.MarshalToJson(c.object, core.SchemeGroupVersion)
			if err != nil {
				t.Fatalf("failed create marshal for input %s: %s", c.testName, err)
			}

			req := new(admission.AdmissionRequest)
			req.Kind = c.kind
			req.Name = c.namespace
			req.Operation = c.operation
			req.UserInfo = authenticationv1.UserInfo{}
			req.Object.Raw = objJS

			if c.operation == admission.Delete {
				if _, err := validator.dc.
					Resource(core.SchemeGroupVersion.WithResource("namespaces")).
					Create(c.object, metav1.CreateOptions{}); err != nil && !kerr.IsAlreadyExists(err) {
					t.Fatalf("failed create namespace for input %s: %s", c.testName, err)
				}
			}
			if len(c.heatUp) > 0 {
				for _, u := range c.heatUp {
					if _, err := validator.dc.
						Resource(api.SchemeGroupVersion.WithResource(api.ResourcePluralPostgres)).
						Namespace("demo").
						Create(u, metav1.CreateOptions{}); err != nil && !kerr.IsAlreadyExists(err) {
						t.Fatalf("failed create db for input %s: %s", c.testName, err)
					}
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
	testName  string
	kind      metav1.GroupVersionKind
	namespace string
	operation admission.Operation
	object    *unstructured.Unstructured
	heatUp    []*unstructured.Unstructured
	result    bool
}{
	{"Create Namespace",
		requestKind,
		"demo",
		admission.Create,
		sampleNamespace(),
		nil,
		true,
	},
	{"Delete Namespace containing doNotPause db",
		requestKind,
		"demo",
		admission.Delete,
		sampleNamespace(),
		[]*unstructured.Unstructured{editDoNotPause(sampleDatabase(), true)},
		false,
	},
	{"Delete Namespace containing NO doNotPause db",
		requestKind,
		"demo",
		admission.Delete,
		sampleNamespace(),
		[]*unstructured.Unstructured{editDoNotPause(sampleDatabase(), false)},
		true,
	},
	{"Delete Namespace containing db with terminationPolicy DoNotTerminate",
		requestKind,
		"demo",
		admission.Delete,
		sampleNamespace(),
		[]*unstructured.Unstructured{editTerminationPolicy(sampleDatabase(), api.TerminationPolicyDoNotTerminate)},
		false,
	},
	{"Delete Namespace containing db with terminationPolicy Pause",
		requestKind,
		"demo",
		admission.Delete,
		sampleNamespace(),
		[]*unstructured.Unstructured{editTerminationPolicy(sampleDatabase(), api.TerminationPolicyPause)},
		false,
	},
	{"Delete Namespace containing db with terminationPolicy Delete",
		requestKind,
		"demo",
		admission.Delete,
		sampleNamespace(),
		[]*unstructured.Unstructured{editTerminationPolicy(sampleDatabase(), api.TerminationPolicyDelete)},
		true,
	},
	{"Delete Namespace containing db with terminationPolicy WipeOut",
		requestKind,
		"demo",
		admission.Delete,
		sampleNamespace(),
		[]*unstructured.Unstructured{editTerminationPolicy(sampleDatabase(), api.TerminationPolicyWipeOut)},
		true,
	},
	{"Delete Namespace containing db with NO terminationPolicy",
		requestKind,
		"demo",
		admission.Delete,
		sampleNamespace(),
		[]*unstructured.Unstructured{deleteTerminationPolicy(sampleDatabase())},
		false,
	},
}

func sampleNamespace() *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": core.SchemeGroupVersion.String(),
			"kind":       "Namespace",
			"metadata": map[string]interface{}{
				"name": "demo",
			},
			"spec": map[string]interface{}{},
		},
	}
}

func sampleDatabase() *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": api.SchemeGroupVersion.String(),
			"kind":       "Postgres",
			"metadata": map[string]interface{}{
				"name":      "foo",
				"namespace": "demo",
				"labels": map[string]interface{}{
					api.LabelDatabaseKind: api.ResourceKindPostgres,
				},
			},
			"spec": map[string]interface{}{
				"terminationPolicy": string(api.TerminationPolicyDelete),
			},
			"status": map[string]interface{}{},
		},
	}
}

func editDoNotPause(db *unstructured.Unstructured, doNotPause bool) *unstructured.Unstructured {
	err := unstructured.SetNestedField(db.Object, doNotPause, "spec", "doNotPause")
	if err != nil {
		panic(err)
	}
	return db
}

func editTerminationPolicy(db *unstructured.Unstructured, terminationPolicy api.TerminationPolicy) *unstructured.Unstructured {
	err := unstructured.SetNestedField(db.Object, string(terminationPolicy), "spec", "terminationPolicy")
	if err != nil {
		panic(err)
	}
	return db
}

func deleteTerminationPolicy(db *unstructured.Unstructured) *unstructured.Unstructured {
	unstructured.RemoveNestedField(db.Object, "spec", "terminationPolicy")
	return db
}
