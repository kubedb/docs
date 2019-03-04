package admission

import (
	"fmt"
	"strings"

	api "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/mergepatch"
	meta_util "kmodules.xyz/client-go/meta"
)

func ValidateUpdate(obj, oldObj runtime.Object, kind string) error {
	preconditions := getPreconditionFunc(kind)
	_, err := meta_util.CreateStrategicPatch(oldObj, obj, preconditions...)
	if err != nil {
		if mergepatch.IsPreconditionFailed(err) {
			return fmt.Errorf("%v.%v", err, preconditionFailedError(kind))
		}
		return err
	}
	return nil
}

func getPreconditionFunc(kind string) []mergepatch.PreconditionFunc {
	preconditions := []mergepatch.PreconditionFunc{
		mergepatch.RequireKeyUnchanged("apiVersion"),
		mergepatch.RequireKeyUnchanged("kind"),
		mergepatch.RequireMetadataKeyUnchanged("name"),
		mergepatch.RequireMetadataKeyUnchanged("namespace"),
	}

	if fields, found := preconditionSpecField[kind]; found {
		for _, field := range fields {
			preconditions = append(preconditions,
				meta_util.RequireChainKeyUnchanged(field),
			)
		}
	}
	return preconditions
}

var preconditionSpecField = map[string][]string{
	api.ResourceKindDormantDatabase: {
		"spec.origin",
	},
}

func preconditionFailedError(kind string) error {
	str := preconditionSpecField[kind]
	strList := strings.Join(str, "\n\t")
	return fmt.Errorf(strings.Join([]string{`At least one of the following was changed:
	apiVersion
	kind
	name
	namespace
	status`, strList}, "\n\t"))
}
