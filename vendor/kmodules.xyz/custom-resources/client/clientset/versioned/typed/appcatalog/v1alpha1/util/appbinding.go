/*
Copyright The Kmodules Authors.

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

package util

import (
	"context"
	"fmt"

	kutil "kmodules.xyz/client-go"
	api "kmodules.xyz/custom-resources/apis/appcatalog/v1alpha1"
	cs "kmodules.xyz/custom-resources/client/clientset/versioned/typed/appcatalog/v1alpha1"

	"github.com/golang/glog"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/jsonmergepatch"
	"k8s.io/apimachinery/pkg/util/wait"
)

func CreateOrPatchAppBinding(ctx context.Context, c cs.AppcatalogV1alpha1Interface, meta metav1.ObjectMeta, transform func(*api.AppBinding) *api.AppBinding, opts metav1.PatchOptions) (*api.AppBinding, kutil.VerbType, error) {
	cur, err := c.AppBindings(meta.Namespace).Get(ctx, meta.Name, metav1.GetOptions{})
	if kerr.IsNotFound(err) {
		glog.V(3).Infof("Creating AppBinding %s/%s.", meta.Namespace, meta.Name)
		out, err := c.AppBindings(meta.Namespace).Create(ctx, transform(&api.AppBinding{
			TypeMeta: metav1.TypeMeta{
				Kind:       "AppBinding",
				APIVersion: api.SchemeGroupVersion.String(),
			},
			ObjectMeta: meta,
		}), metav1.CreateOptions{
			DryRun:       opts.DryRun,
			FieldManager: opts.FieldManager,
		})
		return out, kutil.VerbCreated, err
	} else if err != nil {
		return nil, kutil.VerbUnchanged, err
	}
	return PatchAppBinding(ctx, c, cur, transform, opts)
}

func PatchAppBinding(ctx context.Context, c cs.AppcatalogV1alpha1Interface, cur *api.AppBinding, transform func(*api.AppBinding) *api.AppBinding, opts metav1.PatchOptions) (*api.AppBinding, kutil.VerbType, error) {
	return PatchAppBindingObject(ctx, c, cur, transform(cur.DeepCopy()), opts)
}

func PatchAppBindingObject(ctx context.Context, c cs.AppcatalogV1alpha1Interface, cur, mod *api.AppBinding, opts metav1.PatchOptions) (*api.AppBinding, kutil.VerbType, error) {
	curJson, err := json.Marshal(cur)
	if err != nil {
		return nil, kutil.VerbUnchanged, err
	}

	modJson, err := json.Marshal(mod)
	if err != nil {
		return nil, kutil.VerbUnchanged, err
	}

	patch, err := jsonmergepatch.CreateThreeWayJSONMergePatch(curJson, modJson, curJson)
	if err != nil {
		return nil, kutil.VerbUnchanged, err
	}
	if len(patch) == 0 || string(patch) == "{}" {
		return cur, kutil.VerbUnchanged, nil
	}
	glog.V(3).Infof("Patching AppBinding %s/%s with %s.", cur.Namespace, cur.Name, string(patch))
	out, err := c.AppBindings(cur.Namespace).Patch(ctx, cur.Name, types.MergePatchType, patch, opts)
	return out, kutil.VerbPatched, err
}

func TryUpdateAppBinding(ctx context.Context, c cs.AppcatalogV1alpha1Interface, meta metav1.ObjectMeta, transform func(*api.AppBinding) *api.AppBinding, opts metav1.UpdateOptions) (result *api.AppBinding, err error) {
	attempt := 0
	err = wait.PollImmediate(kutil.RetryInterval, kutil.RetryTimeout, func() (bool, error) {
		attempt++
		cur, e2 := c.AppBindings(meta.Namespace).Get(ctx, meta.Name, metav1.GetOptions{})
		if kerr.IsNotFound(e2) {
			return false, e2
		} else if e2 == nil {

			result, e2 = c.AppBindings(cur.Namespace).Update(ctx, transform(cur.DeepCopy()), opts)
			return e2 == nil, nil
		}
		glog.Errorf("Attempt %d failed to update AppBinding %s/%s due to %v.", attempt, cur.Namespace, cur.Name, e2)
		return false, nil
	})

	if err != nil {
		err = fmt.Errorf("failed to update AppBinding %s/%s after %d attempts due to %v", meta.Namespace, meta.Name, attempt, err)
	}
	return
}
