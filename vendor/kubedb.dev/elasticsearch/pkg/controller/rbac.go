package controller

import (
	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha1"

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

func (c *Controller) ensureRole(db *api.Elasticsearch, name string, pspName string) error {
	ref, rerr := reference.GetReference(clientsetscheme.Scheme, db)
	if rerr != nil {
		return rerr
	}

	// Create new Roles
	_, _, err := rbac_util.CreateOrPatchRole(
		c.Client,
		metav1.ObjectMeta{
			Name:      name,
			Namespace: db.Namespace,
		},
		func(in *rbac.Role) *rbac.Role {
			core_util.EnsureOwnerReference(&in.ObjectMeta, ref)
			in.Labels = db.OffshootLabels()
			in.Rules = []rbac.PolicyRule{}
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
	)
	return err
}

func (c *Controller) createRoleBinding(db *api.Elasticsearch, roleName string, saName string) error {
	ref, rerr := reference.GetReference(clientsetscheme.Scheme, db)
	if rerr != nil {
		return rerr
	}
	// Ensure new RoleBindings
	_, _, err := rbac_util.CreateOrPatchRoleBinding(
		c.Client,
		metav1.ObjectMeta{
			Name:      roleName,
			Namespace: db.Namespace,
		},
		func(in *rbac.RoleBinding) *rbac.RoleBinding {
			core_util.EnsureOwnerReference(&in.ObjectMeta, ref)
			in.Labels = db.OffshootLabels()
			in.RoleRef = rbac.RoleRef{
				APIGroup: rbac.GroupName,
				Kind:     "Role",
				Name:     roleName,
			}
			in.Subjects = []rbac.Subject{
				{
					Kind:      rbac.ServiceAccountKind,
					Name:      saName,
					Namespace: db.Namespace,
				},
			}
			return in
		},
	)
	return err
}

func (c *Controller) createServiceAccount(db *api.Elasticsearch, saName string) error {
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

func (c *Controller) getPolicyNames(db *api.Elasticsearch) (string, string, error) {
	dbVersion, err := c.esVersionLister.Get(string(db.Spec.Version))
	if err != nil {
		return "", "", err
	}
	dbPolicyName := dbVersion.Spec.PodSecurityPolicies.DatabasePolicyName
	snapshotPolicyName := dbVersion.Spec.PodSecurityPolicies.SnapshotterPolicyName

	return dbPolicyName, snapshotPolicyName, nil
}

func (c *Controller) ensureDatabaseRBAC(elasticsearch *api.Elasticsearch) error {
	saName := elasticsearch.Spec.PodTemplate.Spec.ServiceAccountName
	if saName == "" {
		saName = elasticsearch.OffshootName()
		elasticsearch.Spec.PodTemplate.Spec.ServiceAccountName = saName
	}

	sa, err := c.Client.CoreV1().ServiceAccounts(elasticsearch.Namespace).Get(saName, metav1.GetOptions{})
	if kerr.IsNotFound(err) {
		// create service account, since it does not exist
		if err = c.createServiceAccount(elasticsearch, saName); err != nil {
			if !kerr.IsAlreadyExists(err) {
				return err
			}
		}
	} else if err != nil {
		return err
	} else if !core_util.IsOwnedBy(sa, elasticsearch) {
		// user provided the service account, so do nothing.
		return nil
	}

	// Create New Role
	pspName, _, err := c.getPolicyNames(elasticsearch)
	if err != nil {
		return err
	}
	if err := c.ensureRole(elasticsearch, elasticsearch.OffshootName(), pspName); err != nil {
		return err
	}

	// Create New RoleBinding
	if err := c.createRoleBinding(elasticsearch, elasticsearch.OffshootName(), saName); err != nil {
		return err
	}

	return nil
}

func (c *Controller) ensureSnapshotRBAC(elasticsearch *api.Elasticsearch) error {
	_, snapshotPolicyName, err := c.getPolicyNames(elasticsearch)
	if err != nil {
		return err
	}

	// Create New Role for Snapshot
	if err := c.ensureRole(elasticsearch, elasticsearch.SnapshotSAName(), snapshotPolicyName); err != nil {
		return err
	}

	// Create New RoleBinding for Snapshot
	if err := c.createRoleBinding(elasticsearch, elasticsearch.SnapshotSAName(), elasticsearch.SnapshotSAName()); err != nil {
		return err
	}

	// Create New Snapshot ServiceAccount
	if err := c.createServiceAccount(elasticsearch, elasticsearch.SnapshotSAName()); err != nil {
		if !kerr.IsAlreadyExists(err) {
			return err
		}
	}

	return nil
}
