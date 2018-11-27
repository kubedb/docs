package meta_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/appscode/kutil/meta"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

var lblAphlict = map[string]string{
	"app": "AppAphlictserver",
}

func TestMarshalToYAML(t *testing.T) {
	obj := &core.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "AppAphlictserver",
			Namespace: core.NamespaceDefault,
			Labels:    lblAphlict,
		},
		Spec: core.ServiceSpec{
			Selector: lblAphlict,
			Type:     core.ServiceTypeNodePort,
			Ports: []core.ServicePort{
				{
					Port:       int32(22280),
					Protocol:   core.ProtocolTCP,
					TargetPort: intstr.FromString("client-server"),
					Name:       "client-server",
				},
				{
					Port:       int32(22281),
					Protocol:   core.ProtocolTCP,
					TargetPort: intstr.FromString("admin-server"),
					Name:       "admin-server",
				},
			},
		},
	}

	b, err := meta.MarshalToYAML(obj, core.SchemeGroupVersion)
	fmt.Println(err)
	fmt.Println(string(b))
}

const domainKey = "kubedb.com"

func TestFilterKeys(t *testing.T) {
	cases := []struct {
		name string
		in   map[string]string
		out  map[string]string
	}{
		{
			"IndexRune < 0",
			map[string]string{
				"k": "v",
			},
			map[string]string{
				"k": "v",
			},
		},
		{
			"IndexRune == 0",
			map[string]string{
				"/k": "v",
			},
			map[string]string{
				"/k": "v",
			},
		},
		{
			"IndexRune < n - xyz.abc/w1",
			map[string]string{
				"xyz.abc/w1": "v1",
				"w2":         "v2",
			},
			map[string]string{
				"xyz.abc/w1": "v1",
				"w2":         "v2",
			},
		},
		{
			"IndexRune < n - .abc/w1",
			map[string]string{
				".abc/w1": "v1",
				"w2":      "v2",
			},
			map[string]string{
				".abc/w1": "v1",
				"w2":      "v2",
			},
		},
		{
			"IndexRune == n - matching_domain",
			map[string]string{
				domainKey + "/w1": "v1",
				"w2":              "v2",
			},
			map[string]string{
				"w2": "v2",
			},
		},
		{
			"IndexRune > n - matching_subdomain",
			map[string]string{
				"xyz." + domainKey + "/w1": "v1",
				"w2":                       "v2",
			},
			map[string]string{
				"w2": "v2",
			},
		},
		{
			"IndexRune > n - matching_subdomain-2",
			map[string]string{
				"." + domainKey + "/w1": "v1",
				"w2":                    "v2",
			},
			map[string]string{
				"w2": "v2",
			},
		},
		{
			"IndexRune == n - unmatched_domain",
			map[string]string{
				"cubedb.com/w1": "v1",
				"w2":            "v2",
			},
			map[string]string{
				"cubedb.com/w1": "v1",
				"w2":            "v2",
			},
		},
		{
			"IndexRune > n - unmatched_subdomain",
			map[string]string{
				"xyz.cubedb.com/w1": "v1",
				"w2":                "v2",
			},
			map[string]string{
				"xyz.cubedb.com/w1": "v1",
				"w2":                "v2",
			},
		},
		{
			"IndexRune > n - unmatched_subdomain-2",
			map[string]string{
				".cubedb.com/w1": "v1",
				"w2":             "v2",
			},
			map[string]string{
				".cubedb.com/w1": "v1",
				"w2":             "v2",
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			result := meta.FilterKeys(domainKey, nil, c.in)
			if !reflect.DeepEqual(c.out, result) {
				t.Errorf("Failed filterTag test for '%v': expected %+v, got %+v", c.in, c.out, result)
			}
		})
	}
}
