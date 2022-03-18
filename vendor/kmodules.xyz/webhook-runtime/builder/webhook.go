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
	"context"

	hooks "kmodules.xyz/webhook-runtime/admission/v1"

	v1 "k8s.io/api/admission/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

type webhook struct {
	gk     schema.GroupKind
	prefix GroupPrefix
	w      *admission.Webhook
}

var _ hooks.AdmissionHook = &webhook{}

func (m *webhook) Initialize(_ *rest.Config, _ <-chan struct{}) error {
	return nil
}

func (m *webhook) Resource() (plural schema.GroupVersionResource, singular string) {
	return resource(m.prefix, m.gk)
}

func (m *webhook) Admit(admissionSpec *v1.AdmissionRequest) *v1.AdmissionResponse {
	req := admission.Request{AdmissionRequest: *admissionSpec}
	resp := m.w.Handle(context.TODO(), req)
	return &resp.AdmissionResponse
}
