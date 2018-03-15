package v1

import (
	"sync"

	"github.com/appscode/kutil"
	"github.com/appscode/kutil/admission/api"
	"github.com/appscode/kutil/meta"
	admission "k8s.io/api/admission/v1beta1"
	"k8s.io/api/apps/v1"
	"k8s.io/api/apps/v1beta2"
	extensions "k8s.io/api/extensions/v1beta1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/kubernetes/pkg/api/legacyscheme"
)

type ReplicaSetWebhook struct {
	client   kubernetes.Interface
	handler  api.ResourceHandler
	plural   schema.GroupVersionResource
	singular string

	initialized bool
	lock        sync.RWMutex
}

var _ api.AdmissionHook = &ReplicaSetWebhook{}

func NewReplicaSetWebhook(plural schema.GroupVersionResource, singular string, handler api.ResourceHandler) *ReplicaSetWebhook {
	return &ReplicaSetWebhook{
		plural:   plural,
		singular: singular,
		handler:  handler,
	}
}

func (a *ReplicaSetWebhook) Resource() (schema.GroupVersionResource, string) {
	return a.plural, a.singular
}

func (a *ReplicaSetWebhook) Initialize(config *rest.Config, stopCh <-chan struct{}) error {
	a.lock.Lock()
	defer a.lock.Unlock()

	a.initialized = true

	var err error
	a.client, err = kubernetes.NewForConfig(config)
	return err
}

func (a *ReplicaSetWebhook) Admit(req *admission.AdmissionRequest) *admission.AdmissionResponse {
	status := &admission.AdmissionResponse{}

	if a.handler == nil ||
		(req.Operation != admission.Create && req.Operation != admission.Update && req.Operation != admission.Delete) ||
		len(req.SubResource) != 0 ||
		(req.Kind.Group != v1.GroupName && req.Kind.Group != extensions.GroupName) ||
		req.Kind.Kind != "ReplicaSet" {
		status.Allowed = true
		return status
	}

	a.lock.RLock()
	defer a.lock.RUnlock()
	if !a.initialized {
		return api.StatusUninitialized()
	}
	gv := schema.GroupVersion{Group: req.Kind.Group, Version: req.Kind.Version}

	switch req.Operation {
	case admission.Delete:
		// req.Object.Raw = nil, so read from kubernetes
		obj, err := a.client.AppsV1().ReplicaSets(req.Namespace).Get(req.Name, metav1.GetOptions{})
		if err != nil && !kerr.IsNotFound(err) {
			return api.StatusInternalServerError(err)
		} else if err == nil {
			err2 := a.handler.OnDelete(obj)
			if err2 != nil {
				return api.StatusBadRequest(err)
			}
		}
	case admission.Create:
		v1Obj, originalObj, err := convert_to_v1_replicaset(gv, req.Object.Raw)
		if err != nil {
			return api.StatusBadRequest(err)
		}

		v1Mod, err := a.handler.OnCreate(v1Obj)
		if err != nil {
			return api.StatusForbidden(err)
		} else if v1Mod != nil {
			patch, err := create_replicaset_patch(gv, originalObj, v1Mod)
			if err != nil {
				return api.StatusInternalServerError(err)
			}
			status.Patch = patch
			patchType := admission.PatchTypeJSONPatch
			status.PatchType = &patchType
		}
	case admission.Update:
		v1Obj, originalObj, err := convert_to_v1_replicaset(gv, req.Object.Raw)
		if err != nil {
			return api.StatusBadRequest(err)
		}
		v1OldObj, _, err := convert_to_v1_replicaset(gv, req.OldObject.Raw)
		if err != nil {
			return api.StatusBadRequest(err)
		}

		v1Mod, err := a.handler.OnUpdate(v1OldObj, v1Obj)
		if err != nil {
			return api.StatusForbidden(err)
		} else if v1Mod != nil {
			patch, err := create_replicaset_patch(gv, originalObj, v1Mod)
			if err != nil {
				return api.StatusInternalServerError(err)
			}
			status.Patch = patch
			patchType := admission.PatchTypeJSONPatch
			status.PatchType = &patchType
		}
	}

	status.Allowed = true
	return status
}

func convert_to_v1_replicaset(gv schema.GroupVersion, raw []byte) (*v1.ReplicaSet, runtime.Object, error) {
	switch gv {
	case v1.SchemeGroupVersion:
		v1Obj := &v1.ReplicaSet{}
		err := json.Unmarshal(raw, v1Obj)
		if err != nil {
			return nil, nil, err
		}
		return v1Obj, v1Obj, nil

	case v1beta2.SchemeGroupVersion:
		v1beta2Obj := &v1beta2.ReplicaSet{}
		err := json.Unmarshal(raw, v1beta2Obj)
		if err != nil {
			return nil, nil, err
		}

		v1Obj := &v1.ReplicaSet{}
		err = scheme.Scheme.Convert(v1beta2Obj, v1Obj, nil)
		if err != nil {
			return nil, nil, err
		}
		return v1Obj, v1beta2Obj, nil

	case extensions.SchemeGroupVersion:
		extObj := &extensions.ReplicaSet{}
		err := json.Unmarshal(raw, extObj)
		if err != nil {
			return nil, nil, err
		}

		v1Obj := &v1.ReplicaSet{}
		err = scheme.Scheme.Convert(extObj, v1Obj, nil)
		if err != nil {
			return nil, nil, err
		}
		return v1Obj, extObj, nil
	}
	return nil, nil, kutil.ErrUnknown
}

func create_replicaset_patch(gv schema.GroupVersion, originalObj, v1Mod interface{}) ([]byte, error) {
	switch gv {
	case v1.SchemeGroupVersion:
		v1Obj := v1Mod.(runtime.Object)
		legacyscheme.Scheme.Default(v1Obj)
		return meta.CreateJSONPatch(originalObj.(runtime.Object), v1Obj)

	case v1beta2.SchemeGroupVersion:
		v1beta2Mod := &v1beta2.ReplicaSet{}
		err := scheme.Scheme.Convert(v1Mod, v1beta2Mod, nil)
		if err != nil {
			return nil, err
		}
		legacyscheme.Scheme.Default(v1beta2Mod)
		return meta.CreateJSONPatch(originalObj.(runtime.Object), v1beta2Mod)

	case extensions.SchemeGroupVersion:
		extMod := &extensions.ReplicaSet{}
		err := scheme.Scheme.Convert(v1Mod, extMod, nil)
		if err != nil {
			return nil, err
		}
		legacyscheme.Scheme.Default(extMod)
		return meta.CreateJSONPatch(originalObj.(runtime.Object), extMod)
	}
	return nil, kutil.ErrUnknown
}
