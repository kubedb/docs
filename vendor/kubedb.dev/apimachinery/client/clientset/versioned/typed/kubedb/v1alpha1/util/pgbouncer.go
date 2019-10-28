package util

import (
	"fmt"

	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha1"
	cs "kubedb.dev/apimachinery/client/clientset/versioned/typed/kubedb/v1alpha1"

	"github.com/golang/glog"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/jsonmergepatch"
	"k8s.io/apimachinery/pkg/util/wait"
	kutil "kmodules.xyz/client-go"
)

func CreateOrPatchPgBouncer(c cs.KubedbV1alpha1Interface, meta metav1.ObjectMeta, transform func(*api.PgBouncer) *api.PgBouncer) (*api.PgBouncer, kutil.VerbType, error) {
	cur, err := c.PgBouncers(meta.Namespace).Get(meta.Name, metav1.GetOptions{})
	if kerr.IsNotFound(err) {
		glog.V(3).Infof("Creating PgBouncer %s/%s.", meta.Namespace, meta.Name)
		out, err := c.PgBouncers(meta.Namespace).Create(transform(&api.PgBouncer{
			TypeMeta: metav1.TypeMeta{
				Kind:       "PgBouncer",
				APIVersion: api.SchemeGroupVersion.String(),
			},
			ObjectMeta: meta,
		}))
		return out, kutil.VerbCreated, err
	} else if err != nil {
		return nil, kutil.VerbUnchanged, err
	}
	return PatchPgBouncer(c, cur, transform)
}

func PatchPgBouncer(c cs.KubedbV1alpha1Interface, cur *api.PgBouncer, transform func(*api.PgBouncer) *api.PgBouncer) (*api.PgBouncer, kutil.VerbType, error) {
	return PatchPgBouncerObject(c, cur, transform(cur.DeepCopy()))
}

func PatchPgBouncerObject(c cs.KubedbV1alpha1Interface, cur, mod *api.PgBouncer) (*api.PgBouncer, kutil.VerbType, error) {
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
	glog.V(3).Infof("Patching PgBouncer %s/%s with %s.", cur.Namespace, cur.Name, string(patch))
	out, err := c.PgBouncers(cur.Namespace).Patch(cur.Name, types.MergePatchType, patch)
	return out, kutil.VerbPatched, err
}

func TryUpdatePgBouncer(c cs.KubedbV1alpha1Interface, meta metav1.ObjectMeta, transform func(*api.PgBouncer) *api.PgBouncer) (result *api.PgBouncer, err error) {
	attempt := 0
	err = wait.PollImmediate(kutil.RetryInterval, kutil.RetryTimeout, func() (bool, error) {
		attempt++
		cur, e2 := c.PgBouncers(meta.Namespace).Get(meta.Name, metav1.GetOptions{})
		if kerr.IsNotFound(e2) {
			return false, e2
		} else if e2 == nil {
			result, e2 = c.PgBouncers(cur.Namespace).Update(transform(cur.DeepCopy()))
			return e2 == nil, nil
		}
		glog.Errorf("Attempt %d failed to update PgBouncer %s/%s due to %v.", attempt, cur.Namespace, cur.Name, e2)
		return false, nil
	})

	if err != nil {
		err = fmt.Errorf("failed to update PgBouncer %s/%s after %d attempts due to %v", meta.Namespace, meta.Name, attempt, err)
	}
	return
}

func UpdatePgBouncerStatus(
	c cs.KubedbV1alpha1Interface,
	in *api.PgBouncer,
	transform func(*api.PgBouncerStatus) *api.PgBouncerStatus,
) (result *api.PgBouncer, err error) {
	apply := func(x *api.PgBouncer) *api.PgBouncer {
		return &api.PgBouncer{
			TypeMeta:   x.TypeMeta,
			ObjectMeta: x.ObjectMeta,
			Spec:       x.Spec,
			Status:     *transform(in.Status.DeepCopy()),
		}
	}

	attempt := 0
	cur := in.DeepCopy()
	err = wait.PollImmediate(kutil.RetryInterval, kutil.RetryTimeout, func() (bool, error) {
		attempt++
		var e2 error
		result, e2 = c.PgBouncers(in.Namespace).UpdateStatus(apply(cur))
		if kerr.IsConflict(e2) {
			latest, e3 := c.PgBouncers(in.Namespace).Get(in.Name, metav1.GetOptions{})
			switch {
			case e3 == nil:
				cur = latest
				return false, nil
			case kutil.IsRequestRetryable(e3):
				return false, nil
			default:
				return false, e3
			}
		} else if err != nil && !kutil.IsRequestRetryable(e2) {
			return false, e2
		}
		return e2 == nil, nil
	})

	if err != nil {
		err = fmt.Errorf("failed to update status of PgBouncer %s/%s after %d attempts due to %v", in.Namespace, in.Name, attempt, err)
	}
	return
}
