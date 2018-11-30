package util

import (
	"fmt"

	"github.com/appscode/kutil"
	"github.com/golang/glog"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/jsonmergepatch"
	"k8s.io/apimachinery/pkg/util/wait"
	api "kmodules.xyz/custom-resources/apis/appcatalog/v1alpha1"
	cs "kmodules.xyz/custom-resources/client/clientset/versioned/typed/appcatalog/v1alpha1"
)

func CreateOrPatchAppBinding(c cs.AppcatalogV1alpha1Interface, meta metav1.ObjectMeta, transform func(*api.AppBinding) *api.AppBinding) (*api.AppBinding, kutil.VerbType, error) {
	cur, err := c.AppBindings(meta.Namespace).Get(meta.Name, metav1.GetOptions{})
	if kerr.IsNotFound(err) {
		glog.V(3).Infof("Creating AppBinding %s/%s.", meta.Namespace, meta.Name)
		out, err := c.AppBindings(meta.Namespace).Create(transform(&api.AppBinding{
			TypeMeta: metav1.TypeMeta{
				Kind:       "AppBinding",
				APIVersion: api.SchemeGroupVersion.String(),
			},
			ObjectMeta: meta,
		}))
		return out, kutil.VerbCreated, err
	} else if err != nil {
		return nil, kutil.VerbUnchanged, err
	}
	return PatchAppBinding(c, cur, transform)
}

func PatchAppBinding(c cs.AppcatalogV1alpha1Interface, cur *api.AppBinding, transform func(*api.AppBinding) *api.AppBinding) (*api.AppBinding, kutil.VerbType, error) {
	return PatchAppBindingObject(c, cur, transform(cur.DeepCopy()))
}

func PatchAppBindingObject(c cs.AppcatalogV1alpha1Interface, cur, mod *api.AppBinding) (*api.AppBinding, kutil.VerbType, error) {
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
	out, err := c.AppBindings(cur.Namespace).Patch(cur.Name, types.MergePatchType, patch)
	return out, kutil.VerbPatched, err
}

func TryUpdateApp(c cs.AppcatalogV1alpha1Interface, meta metav1.ObjectMeta, transform func(*api.AppBinding) *api.AppBinding) (result *api.AppBinding, err error) {
	attempt := 0
	err = wait.PollImmediate(kutil.RetryInterval, kutil.RetryTimeout, func() (bool, error) {
		attempt++
		cur, e2 := c.AppBindings(meta.Namespace).Get(meta.Name, metav1.GetOptions{})
		if kerr.IsNotFound(e2) {
			return false, e2
		} else if e2 == nil {

			result, e2 = c.AppBindings(cur.Namespace).Update(transform(cur.DeepCopy()))
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
