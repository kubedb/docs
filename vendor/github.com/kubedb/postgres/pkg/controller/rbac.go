package controller

import (
	api "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	le "github.com/kubedb/postgres/pkg/leader_election"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	policy_v1beta1 "k8s.io/api/policy/v1beta1"
	rbac "k8s.io/api/rbac/v1beta1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientsetscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/reference"
	core_util "kmodules.xyz/client-go/core/v1"
	rbac_util "kmodules.xyz/client-go/rbac/v1beta1"
)

func (c *Controller) ensureRole(db *api.Postgres, pspName string) error {
	ref, rerr := reference.GetReference(clientsetscheme.Scheme, db)
	if rerr != nil {
		return rerr
	}

	// Create new Roles
	_, _, err := rbac_util.CreateOrPatchRole(
		c.Client,
		metav1.ObjectMeta{
			Name:      db.OffshootName(),
			Namespace: db.Namespace,
		},
		func(in *rbac.Role) *rbac.Role {
			core_util.EnsureOwnerReference(&in.ObjectMeta, ref)
			in.Labels = db.OffshootLabels()
			in.Rules = []rbac.PolicyRule{
				{
					APIGroups:     []string{policy_v1beta1.GroupName},
					Resources:     []string{"podsecuritypolicies"},
					Verbs:         []string{"use"},
					ResourceNames: []string{pspName},
				},
				{
					APIGroups:     []string{apps.GroupName},
					Resources:     []string{"statefulsets"},
					Verbs:         []string{"get"},
					ResourceNames: []string{db.OffshootName()},
				},
				{
					APIGroups: []string{core.GroupName},
					Resources: []string{"pods"},
					Verbs:     []string{"list", "patch"},
				},
				{
					APIGroups: []string{core.GroupName},
					Resources: []string{"configmaps"},
					Verbs:     []string{"create"},
				},
				{
					APIGroups:     []string{core.GroupName},
					Resources:     []string{"configmaps"},
					Verbs:         []string{"get", "update"},
					ResourceNames: []string{le.GetLeaderLockName(db.OffshootName())},
				},
			}
			return in
		},
	)
	return err
}

func (c *Controller) ensureSnapshotRole(db *api.Postgres, pspName string) error {
	ref, rerr := reference.GetReference(clientsetscheme.Scheme, db)
	if rerr != nil {
		return rerr
	}
	// Create new Roles
	_, _, err := rbac_util.CreateOrPatchRole(
		c.Client,
		metav1.ObjectMeta{
			Name:      db.SnapshotSAName(),
			Namespace: db.Namespace,
		},
		func(in *rbac.Role) *rbac.Role {
			core_util.EnsureOwnerReference(&in.ObjectMeta, ref)
			in.Labels = db.OffshootLabels()
			in.Rules = []rbac.PolicyRule{
				{
					APIGroups:     []string{policy_v1beta1.GroupName},
					Resources:     []string{"podsecuritypolicies"},
					Verbs:         []string{"use"},
					ResourceNames: []string{pspName},
				},
			}
			return in
		},
	)
	return err
}

func (c *Controller) createServiceAccount(db *api.Postgres, saName string) error {
	ref, rerr := reference.GetReference(clientsetscheme.Scheme, db)
	if rerr != nil {
		return rerr
	}
	// Create new ServiceAccount
	_, _, err := core_util.CreateOrPatchServiceAccount(
		c.Client,
		metav1.ObjectMeta{
			Name:      saName,
			Namespace: db.Namespace,
		},
		func(in *core.ServiceAccount) *core.ServiceAccount {
			core_util.EnsureOwnerReference(&in.ObjectMeta, ref)
			in.Labels = db.OffshootLabels()
			return in
		},
	)
	return err
}

func (c *Controller) createRoleBinding(db *api.Postgres) error {
	ref, rerr := reference.GetReference(clientsetscheme.Scheme, db)
	if rerr != nil {
		return rerr
	}
	// Ensure new RoleBindings
	_, _, err := rbac_util.CreateOrPatchRoleBinding(
		c.Client,
		metav1.ObjectMeta{
			Name:      db.OffshootName(),
			Namespace: db.Namespace,
		},
		func(in *rbac.RoleBinding) *rbac.RoleBinding {
			core_util.EnsureOwnerReference(&in.ObjectMeta, ref)
			in.Labels = db.OffshootLabels()
			in.RoleRef = rbac.RoleRef{
				APIGroup: rbac.GroupName,
				Kind:     "Role",
				Name:     db.OffshootName(),
			}
			in.Subjects = []rbac.Subject{
				{
					Kind:      rbac.ServiceAccountKind,
					Name:      db.OffshootName(),
					Namespace: db.Namespace,
				},
			}
			return in
		},
	)
	return err
}

func (c *Controller) createSnapshotRoleBinding(db *api.Postgres) error {
	ref, rerr := reference.GetReference(clientsetscheme.Scheme, db)
	if rerr != nil {
		return rerr
	}
	// Ensure new RoleBindings
	_, _, err := rbac_util.CreateOrPatchRoleBinding(
		c.Client,
		metav1.ObjectMeta{
			Name:      db.SnapshotSAName(),
			Namespace: db.Namespace,
		},
		func(in *rbac.RoleBinding) *rbac.RoleBinding {
			core_util.EnsureOwnerReference(&in.ObjectMeta, ref)
			in.Labels = db.OffshootLabels()
			in.RoleRef = rbac.RoleRef{
				APIGroup: rbac.GroupName,
				Kind:     "Role",
				Name:     db.SnapshotSAName(),
			}
			in.Subjects = []rbac.Subject{
				{
					Kind:      rbac.ServiceAccountKind,
					Name:      db.SnapshotSAName(),
					Namespace: db.Namespace,
				},
			}
			return in
		},
	)
	return err
}

func (c *Controller) getPolicyNames(db *api.Postgres) (string, string, error) {
	dbVersion, err := c.ExtClient.CatalogV1alpha1().PostgresVersions().Get(string(db.Spec.Version), metav1.GetOptions{})
	if err != nil {
		return "", "", err
	}
	dbPolicyName := dbVersion.Spec.PodSecurityPolicies.DatabasePolicyName
	snapshotPolicyName := dbVersion.Spec.PodSecurityPolicies.SnapshotterPolicyName

	return dbPolicyName, snapshotPolicyName, nil
}

func (c *Controller) ensureRBACStuff(postgres *api.Postgres) error {
	dbPolicyName, snapshotPolicyName, err := c.getPolicyNames(postgres)
	if err != nil {
		return err
	}

	// Create New Role
	if err := c.ensureRole(postgres, dbPolicyName); err != nil {
		return err
	}

	// Create New ServiceAccount
	if err := c.createServiceAccount(postgres, postgres.OffshootName()); err != nil {
		if !kerr.IsAlreadyExists(err) {
			return err
		}
	}

	// Create New RoleBinding
	if err := c.createRoleBinding(postgres); err != nil {
		return err
	}

	//Role for snapshot
	if err := c.ensureSnapshotRole(postgres, snapshotPolicyName); err != nil {
		return err
	}

	// ServiceAccount for snapshot
	if err := c.createServiceAccount(postgres, postgres.SnapshotSAName()); err != nil {
		if !kerr.IsAlreadyExists(err) {
			return err
		}
	}

	// Create New RoleBinding for snapshot
	if err := c.createSnapshotRoleBinding(postgres); err != nil {
		return err
	}

	return nil
}
