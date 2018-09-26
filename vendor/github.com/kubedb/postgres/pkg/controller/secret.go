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
	"k8s.io/apimachinery/pkg/labels"
)

const (
	PostgresUser       = "POSTGRES_USER"
	PostgresPassword   = "POSTGRES_PASSWORD"
	ExporterSecretPath = "/var/run/secrets/kubedb.com/"
)

func (c *Controller) ensureDatabaseSecret(postgres *api.Postgres) error {
	databaseSecretVolume := postgres.Spec.DatabaseSecret
	if databaseSecretVolume == nil {
		var err error
		if databaseSecretVolume, err = c.createDatabaseSecret(postgres); err != nil {
			return err
		}
		pg, _, err := util.PatchPostgres(c.ExtClient.KubedbV1alpha1(), postgres, func(in *api.Postgres) *api.Postgres {
			in.Spec.DatabaseSecret = databaseSecretVolume
			return in
		})
		if err != nil {
			c.recorder.Eventf(
				postgres,
				core.EventTypeWarning,
				eventer.EventReasonFailedToUpdate,
				err.Error(),
			)
			return err
		}
		postgres.Spec.DatabaseSecret = pg.Spec.DatabaseSecret
	}
	return nil
}

func (c *Controller) findDatabaseSecret(postgres *api.Postgres) (*core.Secret, error) {
	name := postgres.OffshootName() + "-auth"

	secret, err := c.Client.CoreV1().Secrets(postgres.Namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		if kerr.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
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
		Data: map[string][]byte{
			PostgresUser:     []byte("postgres"),
			PostgresPassword: []byte(rand.GeneratePassword()),
		},
	}
	if _, err := c.Client.CoreV1().Secrets(postgres.Namespace).Create(secret); err != nil {
		return nil, err
	}

	return &core.SecretVolumeSource{
		SecretName: secret.Name,
	}, nil
}

func (c *Controller) deleteSecret(dormantDb *api.DormantDatabase, secretVolume *core.SecretVolumeSource) error {
	secretFound := false
	postgresList, err := c.ExtClient.KubedbV1alpha1().Postgreses(dormantDb.Namespace).List(metav1.ListOptions{})
	if err != nil {
		return err
	}

	for _, postgres := range postgresList.Items {
		databaseSecret := postgres.Spec.DatabaseSecret
		if databaseSecret != nil {
			if databaseSecret.SecretName == secretVolume.SecretName {
				secretFound = true
				break
			}
		}
	}

	if !secretFound {
		labelMap := map[string]string{
			api.LabelDatabaseKind: api.ResourceKindPostgres,
		}
		dormantDatabaseList, err := c.ExtClient.KubedbV1alpha1().DormantDatabases(dormantDb.Namespace).List(
			metav1.ListOptions{
				LabelSelector: labels.SelectorFromSet(labelMap).String(),
			},
		)
		if err != nil {
			return err
		}

		for _, ddb := range dormantDatabaseList.Items {
			if ddb.Name == dormantDb.Name {
				continue
			}

			databaseSecret := ddb.Spec.Origin.Spec.Postgres.DatabaseSecret
			if databaseSecret != nil {
				if databaseSecret.SecretName == secretVolume.SecretName {
					secretFound = true
					break
				}
			}
		}
	}

	if !secretFound {
		if err := c.Client.CoreV1().Secrets(dormantDb.Namespace).Delete(secretVolume.SecretName, nil); !kerr.IsNotFound(err) {
			return err
		}
	}

	return nil
}
