/*
Copyright AppsCode Inc. and Contributors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package builder

import (
	hooks "kmodules.xyz/webhook-runtime/admission/v1"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

var log = logf.Log.WithName("webhook-runtime")

// WebhookBuilder builds a Webhook.
type WebhookBuilder struct {
	apiType runtime.Object
	gk      schema.GroupKind
	scheme  *runtime.Scheme
}

// WebhookManagedBy allows inform its Scheme and RESTMapper.
func WebhookManagedBy(s *runtime.Scheme) *WebhookBuilder {
	return &WebhookBuilder{scheme: s}
}

// TODO(droot): update the GoDoc for conversion.

// For takes a runtime.Object which should be a CR.
// If the given object implements the admission.Defaulter interface, a MutatingWebhook will be wired for this type.
// If the given object implements the admission.Validator interface, a ValidatingWebhook will be wired for this type.
func (blder *WebhookBuilder) For(apiType runtime.Object) *WebhookBuilder {
	blder.apiType = apiType
	return blder
}

// Complete builds the webhook.
func (blder *WebhookBuilder) Complete() (hooks.AdmissionHook, hooks.AdmissionHook, error) {
	// Create webhook(s) for each type
	gvk, err := apiutil.GVKForObject(blder.apiType, blder.scheme)
	if err != nil {
		return nil, nil, err
	}
	blder.gk = gvk.GroupKind()

	mutator, err := blder.registerDefaultingWebhook()
	if err != nil {
		return nil, nil, err
	}
	validator, err := blder.registerValidatingWebhook()
	if err != nil {
		return nil, nil, err
	}
	return mutator, validator, nil
}

// registerDefaultingWebhook registers a defaulting webhook if th.
func (blder *WebhookBuilder) registerDefaultingWebhook() (hooks.AdmissionHook, error) {
	defaulter, isDefaulter := blder.apiType.(admission.Defaulter)
	if !isDefaulter {
		log.Info("skip registering a mutating webhook, admission.Defaulter interface is not implemented", "GK", blder.gk)
		return nil, nil
	}

	mwh := admission.DefaultingWebhookFor(defaulter)
	if err := mwh.InjectScheme(blder.scheme); err != nil {
		return nil, err
	}
	if err := mwh.InjectLogger(log); err != nil {
		return nil, err
	}
	return &webhook{
		prefix: MutatorGroupPrefix,
		gk:     blder.gk,
		w:      mwh,
	}, nil
}

func (blder *WebhookBuilder) registerValidatingWebhook() (hooks.AdmissionHook, error) {
	checker, isValidator := blder.apiType.(admission.Validator)
	if !isValidator {
		log.Info("skip registering a validating webhook, admission.Validator interface is not implemented", "GK", blder.gk)
		return nil, nil
	}

	vwh := admission.ValidatingWebhookFor(checker)
	if err := vwh.InjectScheme(blder.scheme); err != nil {
		return nil, err
	}
	if err := vwh.InjectLogger(log); err != nil {
		return nil, err
	}
	return &webhook{
		prefix: ValidatorGroupPrefix,
		gk:     blder.gk,
		w:      vwh,
	}, nil
}
