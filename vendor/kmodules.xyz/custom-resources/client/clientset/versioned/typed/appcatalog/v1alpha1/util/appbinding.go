package util

import (
	"bytes"
	"fmt"

	"github.com/appscode/kutil"
	"github.com/golang/glog"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/jsonmergepatch"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/util/jsonpath"
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

// ref: https://github.com/kubernetes-incubator/service-catalog/blob/37b874716ad709a175e426f5f5638322a600849f/pkg/controller/controller_binding.go#L588
func TransformCredentials(kc kubernetes.Interface, transforms []api.SecretTransform, credentials map[string]interface{}) error {
	for _, t := range transforms {
		switch {
		case t.AddKey != nil:
			var value interface{}
			if t.AddKey.JSONPathExpression != nil {
				result, err := evaluateJSONPath(*t.AddKey.JSONPathExpression, credentials)
				if err != nil {
					return err
				}
				value = result
			} else if t.AddKey.StringValue != nil {
				value = *t.AddKey.StringValue
			} else {
				value = t.AddKey.Value
			}
			credentials[t.AddKey.Key] = value
		case t.RenameKey != nil:
			value, ok := credentials[t.RenameKey.From]
			if ok {
				credentials[t.RenameKey.To] = value
				delete(credentials, t.RenameKey.From)
			}
		case t.AddKeysFrom != nil:
			secret, err := kc.CoreV1().
				Secrets(t.AddKeysFrom.SecretRef.Namespace).
				Get(t.AddKeysFrom.SecretRef.Name, metav1.GetOptions{})
			if err != nil {
				return err // TODO: if the Secret doesn't exist yet, can we perform the transform when it does?
			}
			for k, v := range secret.Data {
				credentials[k] = v
			}
		case t.RemoveKey != nil:
			delete(credentials, t.RemoveKey.Key)
		}
	}
	return nil
}

func evaluateJSONPath(jsonPath string, credentials map[string]interface{}) (string, error) {
	j := jsonpath.New("expression")
	buf := new(bytes.Buffer)
	if err := j.Parse(jsonPath); err != nil {
		return "", err
	}
	if err := j.Execute(buf, credentials); err != nil {
		return "", err
	}
	return buf.String(), nil
}
