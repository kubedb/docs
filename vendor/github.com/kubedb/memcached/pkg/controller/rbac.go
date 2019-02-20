package controller

import (
	core_util "github.com/appscode/kutil/core/v1"
	api "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientsetscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/reference"
)

func (c *Controller) createServiceAccount(memcached *api.Memcached) error {
	ref, rerr := reference.GetReference(clientsetscheme.Scheme, memcached)
	if rerr != nil {
		return rerr
	}
	// Create new ServiceAccount
	_, _, err := core_util.CreateOrPatchServiceAccount(
		c.Client,
		metav1.ObjectMeta{
			Name:      memcached.OffshootName(),
			Namespace: memcached.Namespace,
		},
		func(in *core.ServiceAccount) *core.ServiceAccount {
			core_util.EnsureOwnerReference(&in.ObjectMeta, ref)
			return in
		},
	)
	return err
}

func (c *Controller) ensureRBACStuff(memcached *api.Memcached) error {
	// Create New ServiceAccount
	if err := c.createServiceAccount(memcached); err != nil {
		if !kerr.IsAlreadyExists(err) {
			return err
		}
	}

	return nil
}
