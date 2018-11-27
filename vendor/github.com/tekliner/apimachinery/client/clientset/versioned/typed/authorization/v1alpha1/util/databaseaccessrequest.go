package util

import (
	"encoding/json"
	"fmt"

	"github.com/appscode/kutil"
	"github.com/evanphx/json-patch"
	"github.com/golang/glog"
	api "github.com/kubedb/apimachinery/apis/authorization/v1alpha1"
	cs "github.com/kubedb/apimachinery/client/clientset/versioned/typed/authorization/v1alpha1"
	"github.com/pkg/errors"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
)

func CreateOrPatchDatabaseAccessRequest(c cs.AuthorizationV1alpha1Interface, meta metav1.ObjectMeta, transform func(alert *api.DatabaseAccessRequest) *api.DatabaseAccessRequest) (*api.DatabaseAccessRequest, kutil.VerbType, error) {
	cur, err := c.DatabaseAccessRequests(meta.Namespace).Get(meta.Name, metav1.GetOptions{})
	if kerr.IsNotFound(err) {
		glog.V(3).Infof("Creating DatabaseAccessRequest %s/%s.", meta.Namespace, meta.Name)
		out, err := c.DatabaseAccessRequests(meta.Namespace).Create(transform(&api.DatabaseAccessRequest{
			TypeMeta: metav1.TypeMeta{
				Kind:       api.ResourceKindDatabaseAccessRequest,
				APIVersion: api.SchemeGroupVersion.String(),
			},
			ObjectMeta: meta,
		}))
		return out, kutil.VerbCreated, err
	} else if err != nil {
		return nil, kutil.VerbUnchanged, err
	}
	return PatchDatabaseAccessRequest(c, cur, transform)
}

func PatchDatabaseAccessRequest(c cs.AuthorizationV1alpha1Interface, cur *api.DatabaseAccessRequest, transform func(*api.DatabaseAccessRequest) *api.DatabaseAccessRequest) (*api.DatabaseAccessRequest, kutil.VerbType, error) {
	return PatchDatabaseAccessRequestObject(c, cur, transform(cur.DeepCopy()))
}

func PatchDatabaseAccessRequestObject(c cs.AuthorizationV1alpha1Interface, cur, mod *api.DatabaseAccessRequest) (*api.DatabaseAccessRequest, kutil.VerbType, error) {
	curJson, err := json.Marshal(cur)
	if err != nil {
		return nil, kutil.VerbUnchanged, err
	}

	modJson, err := json.Marshal(mod)
	if err != nil {
		return nil, kutil.VerbUnchanged, err
	}

	patch, err := jsonpatch.CreateMergePatch(curJson, modJson)
	if err != nil {
		return nil, kutil.VerbUnchanged, err
	}
	if len(patch) == 0 || string(patch) == "{}" {
		return cur, kutil.VerbUnchanged, nil
	}
	glog.V(3).Infof("Patching DatabaseAccessRequest %s/%s with %s.", cur.Namespace, cur.Name, string(patch))
	out, err := c.DatabaseAccessRequests(cur.Namespace).Patch(cur.Name, types.MergePatchType, patch)
	return out, kutil.VerbPatched, err
}

func TryUpdateDatabaseAccessRequest(c cs.AuthorizationV1alpha1Interface, meta metav1.ObjectMeta, transform func(*api.DatabaseAccessRequest) *api.DatabaseAccessRequest) (result *api.DatabaseAccessRequest, err error) {
	attempt := 0
	err = wait.PollImmediate(kutil.RetryInterval, kutil.RetryTimeout, func() (bool, error) {
		attempt++
		cur, e2 := c.DatabaseAccessRequests(meta.Namespace).Get(meta.Name, metav1.GetOptions{})
		if kerr.IsNotFound(e2) {
			return false, e2
		} else if e2 == nil {
			result, e2 = c.DatabaseAccessRequests(cur.Namespace).Update(transform(cur.DeepCopy()))
			return e2 == nil, nil
		}
		glog.Errorf("Attempt %d failed to update DatabaseAccessRequest %s/%s due to %v.", attempt, cur.Namespace, cur.Name, e2)
		return false, nil
	})

	if err != nil {
		err = errors.Errorf("failed to update DatabaseAccessRequest %s/%s after %d attempts due to %v", meta.Namespace, meta.Name, attempt, err)
	}
	return
}

func UpdateDatabaseAccessRequestStatus(
	c cs.AuthorizationV1alpha1Interface,
	in *api.DatabaseAccessRequest,
	transform func(*api.DatabaseAccessRequestStatus) *api.DatabaseAccessRequestStatus,
	useSubresource ...bool,
) (result *api.DatabaseAccessRequest, err error) {
	if len(useSubresource) > 1 {
		return nil, errors.Errorf("invalid value passed for useSubresource: %v", useSubresource)
	}

	apply := func(x *api.DatabaseAccessRequest) *api.DatabaseAccessRequest {
		return &api.DatabaseAccessRequest{
			TypeMeta:   x.TypeMeta,
			ObjectMeta: x.ObjectMeta,
			Spec:       x.Spec,
			Status:     *transform(in.Status.DeepCopy()),
		}
	}

	if len(useSubresource) == 1 && useSubresource[0] {
		attempt := 0
		cur := in.DeepCopy()
		err = wait.PollImmediate(kutil.RetryInterval, kutil.RetryTimeout, func() (bool, error) {
			attempt++
			var e2 error
			result, e2 = c.DatabaseAccessRequests(in.Namespace).UpdateStatus(apply(cur))
			if kerr.IsConflict(e2) {
				latest, e3 := c.DatabaseAccessRequests(in.Namespace).Get(in.Name, metav1.GetOptions{})
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
			err = fmt.Errorf("failed to update status of DatabaseAccessRequest %s/%s after %d attempts due to %v", in.Namespace, in.Name, attempt, err)
		}
		return
	}

	result, _, err = PatchDatabaseAccessRequestObject(c, in, apply(in))
	return
}
