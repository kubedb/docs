package controller

import (
	"fmt"

	"github.com/appscode/go/crypto/rand"
	api "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	"github.com/kubedb/apimachinery/client/clientset/versioned/typed/kubedb/v1alpha1/util"
	"github.com/kubedb/apimachinery/pkg/eventer"
	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientsetscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/reference"
	storage "kmodules.xyz/objectstore-api/osm"
)

const (
	etcdUser = "root"

	KeyEtcdUser     = "user"
	KeyEtcdPassword = "password"

	//ExporterSecretPath = "/var/run/secrets/kubedb.com/"
)

func (c *Controller) ensureDatabaseSecret(etcd *api.Etcd) error {
	if etcd.Spec.DatabaseSecret == nil {
		secretVolumeSource, err := c.createDatabaseSecret(etcd)
		if err != nil {
			if ref, rerr := reference.GetReference(clientsetscheme.Scheme, etcd); rerr == nil {
				c.recorder.Eventf(
					ref,
					core.EventTypeWarning,
					eventer.EventReasonFailedToCreate,
					`Failed to create Database Secret. Reason: %v`,
					err.Error(),
				)
			}
			return err
		}

		ms, _, err := util.PatchEtcd(c.ExtClient, etcd, func(in *api.Etcd) *api.Etcd {
			in.Spec.DatabaseSecret = secretVolumeSource
			return in
		})
		if err != nil {
			if ref, rerr := reference.GetReference(clientsetscheme.Scheme, etcd); rerr == nil {
				c.recorder.Eventf(
					ref,
					core.EventTypeWarning,
					eventer.EventReasonFailedToUpdate,
					err.Error(),
				)
			}
			return err
		}
		etcd.Spec.DatabaseSecret = ms.Spec.DatabaseSecret
	}
	return nil
}

func (c *Controller) createDatabaseSecret(etcd *api.Etcd) (*core.SecretVolumeSource, error) {
	authSecretName := etcd.Name + "-auth"

	sc, err := c.checkSecret(authSecretName, etcd)
	if err != nil {
		return nil, err
	}
	if sc == nil {
		randPassword := ""

		// if the password starts with "-" it will cause error in bash scripts (in etcd-tools)
		for randPassword = rand.GeneratePassword(); randPassword[0] == '-'; {
		}

		secret := &core.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:   authSecretName,
				Labels: etcd.OffshootSelectors(),
			},
			Type: core.SecretTypeOpaque,
			StringData: map[string]string{
				KeyEtcdUser:     etcdUser,
				KeyEtcdPassword: randPassword,
			},
		}
		if _, err := c.Client.CoreV1().Secrets(etcd.Namespace).Create(secret); err != nil {
			return nil, err
		}
	}
	return &core.SecretVolumeSource{
		SecretName: authSecretName,
	}, nil
}

func (c *Controller) checkSecret(secretName string, etcd *api.Etcd) (*core.Secret, error) {
	secret, err := c.Client.CoreV1().Secrets(etcd.Namespace).Get(secretName, metav1.GetOptions{})
	if err != nil {
		if kerr.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	if secret.Labels[api.LabelDatabaseKind] != api.ResourceKindEtcd ||
		secret.Labels[api.LabelDatabaseName] != etcd.Name {
		return nil, fmt.Errorf(`intended secret "%v" already exists`, secretName)
	}
	return secret, nil
}

func (c *Controller) createOsmSecret(snapshot *api.Snapshot) error {
	secret, err := storage.NewOSMSecret(c.Client, snapshot.OSMSecretName(), snapshot.Namespace, snapshot.Spec.Backend)
	if err != nil {
		return err
	}
	secret, err = c.Client.CoreV1().Secrets(secret.Namespace).Create(secret)
	if err != nil && !kerr.IsAlreadyExists(err) {
		return err
	}
	return nil
}
