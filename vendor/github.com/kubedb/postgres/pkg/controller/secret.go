package controller

import (
	"fmt"

	"github.com/appscode/go/crypto/rand"
	api "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	kutildb "github.com/kubedb/apimachinery/client/typed/kubedb/v1alpha1/util"
	"github.com/kubedb/apimachinery/pkg/eventer"
	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c *Controller) ensureDatabaseSecret(postgres *api.Postgres) error {
	databaseSecretVolume := postgres.Spec.DatabaseSecret
	if databaseSecretVolume == nil {
		var err error
		if databaseSecretVolume, err = c.createDatabaseSecret(postgres); err != nil {
			return err
		}
		pg, err := kutildb.PatchPostgres(c.ExtClient, postgres, func(in *api.Postgres) *api.Postgres {
			// This will ignore processing all kind of Update while creating
			if in.Annotations == nil {
				in.Annotations = make(map[string]string)
			}
			in.Annotations["kubedb.com/ignore"] = "set"
			in.Spec.DatabaseSecret = databaseSecretVolume
			return in
		})
		if err != nil {
			c.recorder.Eventf(postgres.ObjectReference(), core.EventTypeWarning, eventer.EventReasonFailedToUpdate, err.Error())
			return err
		}
		*postgres = *pg
	}
	return nil
}

func (c *Controller) findDatabaseSecret(postgres *api.Postgres) (*core.Secret, error) {
	name := postgres.OffshootName() + "-auth"

	secret, err := c.Client.CoreV1().Secrets(postgres.Namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		if kerr.IsNotFound(err) {
			return nil, nil
		} else {
			return nil, err
		}
	}

	if secret.Labels[api.LabelDatabaseKind] != api.ResourceKindPostgres ||
		secret.Labels[api.LabelDatabaseName] != postgres.Name {
		return nil, fmt.Errorf(`intended secret "%v" already exists`, name)
	}

	return secret, nil
}

func (c *Controller) createDatabaseSecret(postgres *api.Postgres) (*core.SecretVolumeSource, error) {
	databaseSecret, err := c.findDatabaseSecret(postgres)
	if err != nil {
		return nil, err
	}
	if databaseSecret != nil {
		return &core.SecretVolumeSource{
			SecretName: databaseSecret.Name,
		}, nil
	}

	postgresPassword := fmt.Sprintf("POSTGRES_PASSWORD=%s\n", rand.GeneratePassword())

	data := map[string][]byte{
		".admin": []byte(postgresPassword),
	}

	name := fmt.Sprintf("%v-auth", postgres.OffshootName())
	secret := &core.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Labels: map[string]string{
				api.LabelDatabaseKind: api.ResourceKindPostgres,
				api.LabelDatabaseName: postgres.OffshootName(),
			},
		},
		Type: core.SecretTypeOpaque,
		Data: data,
	}
	if _, err := c.Client.CoreV1().Secrets(postgres.Namespace).Create(secret); err != nil {
		return nil, err
	}

	return &core.SecretVolumeSource{
		SecretName: secret.Name,
	}, nil
}
