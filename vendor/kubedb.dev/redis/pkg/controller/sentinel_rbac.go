/*
Copyright AppsCode Inc. and Contributors

Licensed under the AppsCode Community License 1.0.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://github.com/appscode/licenses/raw/1.0.0/AppsCode-Community-1.0.0.md

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"

	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"

	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	policy_v1beta1 "k8s.io/api/policy/v1beta1"
	rbac "k8s.io/api/rbac/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	core_util "kmodules.xyz/client-go/core/v1"
	rbac_util "kmodules.xyz/client-go/rbac/v1"
)

func (c *Controller) createSentinelServiceAccount(db *api.RedisSentinel, saName string) error {
	owner := metav1.NewControllerRef(db, api.SchemeGroupVersion.WithKind(api.ResourceKindRedisSentinel))

	// Create new ServiceAccount
	_, _, err := core_util.CreateOrPatchServiceAccount(
		context.TODO(),
		c.Client,
		metav1.ObjectMeta{
			Name:      saName,
			Namespace: db.Namespace,
		},
		func(in *core.ServiceAccount) *core.ServiceAccount {
			core_util.EnsureOwnerReference(&in.ObjectMeta, owner)
			in.Labels = db.OffshootLabels()
			return in
		},
		metav1.PatchOptions{},
	)
	return err
}

func (c *Controller) ensureSentinelRole(db *api.RedisSentinel, name string, pspName string) error {
	owner := metav1.NewControllerRef(db, api.SchemeGroupVersion.WithKind(api.ResourceKindRedisSentinel))

	// Create new Role for Redis and it's Snapshot
	_, _, err := rbac_util.CreateOrPatchRole(
		context.TODO(),
		c.Client,
		metav1.ObjectMeta{
			Name:      name,
			Namespace: db.Namespace,
		},
		func(in *rbac.Role) *rbac.Role {
			core_util.EnsureOwnerReference(&in.ObjectMeta, owner)
			in.Labels = db.OffshootLabels()
			in.Rules = []rbac.PolicyRule{
				{
					APIGroups: []string{apps.GroupName},
					Resources: []string{"statefulsets"},
					Verbs:     []string{"get"},
				},
				{
					APIGroups: []string{core.GroupName},
					Resources: []string{"pods"},
					Verbs:     []string{"get", "list", "patch", "delete"},
				},
			}
			if pspName != "" {
				pspRule := rbac.PolicyRule{
					APIGroups:     []string{policy_v1beta1.GroupName},
					Resources:     []string{"podsecuritypolicies"},
					Verbs:         []string{"use"},
					ResourceNames: []string{pspName},
				}
				in.Rules = append(in.Rules, pspRule)
			}
			return in
		},
		metav1.PatchOptions{},
	)
	return err
}

// we are not going to override the existing subject list for role binding of the sentinel. just append new one with the existing subjects.
// this is the case when sentinel is in different namespace from redis. in that case we need to add the redis SAName also in subject list.
func (c *Controller) createSentinelRoleBinding(db *api.RedisSentinel, roleName string, saName string) error {
	var subjects []rbac.Subject
	roleBinding, err := c.Client.RbacV1().RoleBindings(db.Namespace).Get(context.TODO(), roleName, metav1.GetOptions{})
	if err != nil && !kerr.IsNotFound(err) {
		return err
	}
	if err == nil {
		subjects = roleBinding.Subjects
	}
	subjects = UpsertSentinelRoleBindingSubject(subjects, rbac.Subject{
		Kind:      rbac.ServiceAccountKind,
		Name:      saName,
		Namespace: db.Namespace,
	})

	owner := metav1.NewControllerRef(db, api.SchemeGroupVersion.WithKind(api.ResourceKindRedisSentinel))
	// Ensure new RoleBindings for Redis sentinel and it's Snapshot
	_, _, err = rbac_util.CreateOrPatchRoleBinding(
		context.TODO(),
		c.Client,
		metav1.ObjectMeta{
			Name:      roleName,
			Namespace: db.Namespace,
		},
		func(in *rbac.RoleBinding) *rbac.RoleBinding {
			core_util.EnsureOwnerReference(&in.ObjectMeta, owner)
			in.Labels = db.OffshootLabels()
			in.RoleRef = rbac.RoleRef{
				APIGroup: rbac.GroupName,
				Kind:     "Role",
				Name:     roleName,
			}
			in.Subjects = subjects
			return in
		},
		metav1.PatchOptions{},
	)
	return err
}

func (c *Controller) getSentinelPolicyNames(db *api.RedisSentinel) (string, error) {
	dbVersion, err := c.DBClient.CatalogV1alpha1().RedisVersions().Get(context.TODO(), string(db.Spec.Version), metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	dbPolicyName := dbVersion.Spec.PodSecurityPolicies.DatabasePolicyName

	return dbPolicyName, nil
}

func (c *Controller) ensureSentinelRBACStuff(db *api.RedisSentinel) error {
	saName := db.Spec.PodTemplate.Spec.ServiceAccountName
	if saName == "" {
		saName = db.OffshootName()
		db.Spec.PodTemplate.Spec.ServiceAccountName = saName
	}

	sa, err := c.Client.CoreV1().ServiceAccounts(db.Namespace).Get(context.TODO(), saName, metav1.GetOptions{})
	if kerr.IsNotFound(err) {
		// create service account, since it does not exist
		if err = c.createSentinelServiceAccount(db, saName); err != nil {
			if !kerr.IsAlreadyExists(err) {
				return err
			}
		}
	} else if err != nil {
		return err
	} else if owned, _ := core_util.IsOwnedBy(sa, db); !owned {
		// user provided the service account, so do nothing.
		return nil
	}

	// Create New Role
	pspName, err := c.getSentinelPolicyNames(db)
	if err != nil {
		return err
	}
	if err := c.ensureSentinelRole(db, db.OffshootName(), pspName); err != nil {
		return err
	}

	// Create New RoleBinding
	if err := c.createSentinelRoleBinding(db, db.OffshootName(), saName); err != nil {
		return err
	}

	return nil
}

// UpsertSentinelRoleBindingSubject will upsert new rbac subject in the old one
func UpsertSentinelRoleBindingSubject(subjects []rbac.Subject, newSubjects ...rbac.Subject) []rbac.Subject {
	upsert := func(subject rbac.Subject) {
		for i, sub := range subjects {
			if sub.Name == subject.Name {
				subjects[i] = subject
				return
			}
		}
		subjects = append(subjects, subject)
	}

	for _, subject := range newSubjects {
		upsert(subject)
	}
	return subjects
}
