package v1alpha1

import (
	"fmt"
	"strings"

	crdutils "github.com/appscode/kutil/apiextensions/v1beta1"
	apiextensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	mona "kmodules.xyz/monitoring-agent-api/api/v1"
)

func (p MongoDB) OffshootName() string {
	return p.Name
}

func (p MongoDB) OffshootLabels() map[string]string {
	return map[string]string{
		LabelDatabaseName: p.Name,
		LabelDatabaseKind: ResourceKindMongoDB,
	}
}

func (p MongoDB) StatefulSetLabels() map[string]string {
	labels := p.OffshootLabels()
	for key, val := range p.Labels {
		if !strings.HasPrefix(key, GenericKey+"/") && !strings.HasPrefix(key, MongoDBKey+"/") {
			labels[key] = val
		}
	}
	return labels
}

func (p MongoDB) StatefulSetAnnotations() map[string]string {
	annotations := make(map[string]string)
	for key, val := range p.Annotations {
		if !strings.HasPrefix(key, GenericKey+"/") && !strings.HasPrefix(key, MongoDBKey+"/") {
			annotations[key] = val
		}
	}
	return annotations
}

func (p MongoDB) ResourceShortCode() string {
	return ResourceCodeMongoDB
}

func (p MongoDB) ResourceKind() string {
	return ResourceKindMongoDB
}

func (p MongoDB) ResourceSingular() string {
	return ResourceSingularMongoDB
}

func (p MongoDB) ResourcePlural() string {
	return ResourcePluralMongoDB
}

func (p MongoDB) ServiceName() string {
	return p.OffshootName()
}

func (p MongoDB) ServiceMonitorName() string {
	return fmt.Sprintf("kubedb-%s-%s", p.Namespace, p.Name)
}

func (p MongoDB) Path() string {
	return fmt.Sprintf("/kubedb.com/v1alpha1/namespaces/%s/%s/%s/metrics", p.Namespace, p.ResourcePlural(), p.Name)
}

func (p MongoDB) Scheme() string {
	return ""
}

func (p *MongoDB) StatsAccessor() mona.StatsAccessor {
	return p
}

func (m *MongoDB) GetMonitoringVendor() string {
	if m.Spec.Monitor != nil {
		return m.Spec.Monitor.Agent.Vendor()
	}
	return ""
}

func (p MongoDB) CustomResourceDefinition() *apiextensions.CustomResourceDefinition {
	return crdutils.NewCustomResourceDefinition(crdutils.Config{
		Group:         SchemeGroupVersion.Group,
		Plural:        ResourcePluralMongoDB,
		Singular:      ResourceSingularMongoDB,
		Kind:          ResourceKindMongoDB,
		ShortNames:    []string{ResourceCodeMongoDB},
		ResourceScope: string(apiextensions.NamespaceScoped),
		Versions: []apiextensions.CustomResourceDefinitionVersion{
			{
				Name:    SchemeGroupVersion.Version,
				Served:  true,
				Storage: true,
			},
		},
		Labels: crdutils.Labels{
			LabelsMap: map[string]string{"app": "kubedb"},
		},
		SpecDefinitionName:      "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1.MongoDB",
		EnableValidation:        true,
		GetOpenAPIDefinitions:   GetOpenAPIDefinitions,
		EnableStatusSubresource: EnableStatusSubresource,
	}, setNameSchema)
}
