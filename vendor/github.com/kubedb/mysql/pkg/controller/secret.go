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
)

const (
	mysqlUser = "root"

	KeyMySQLUser     = "user"
	KeyMySQLPassword = "password"

	ExporterSecretPath = "/var/run/secrets/kubedb.com/"
)

func (c *Controller) ensureDatabaseSecret(mysql *api.MySQL) error {
	if mysql.Spec.DatabaseSecret == nil {
		secretVolumeSource, err := c.createDatabaseSecret(mysql)
		if err != nil {
			c.recorder.Eventf(
				mysql,
				core.EventTypeWarning,
				eventer.EventReasonFailedToCreate,
				`Failed to create Database Secret. Reason: %v`,
				err.Error(),
			)
			return err
		}

		ms, _, err := util.PatchMySQL(c.ExtClient, mysql, func(in *api.MySQL) *api.MySQL {
			in.Spec.DatabaseSecret = secretVolumeSource
			return in
		})
		if err != nil {
			c.recorder.Eventf(
				mysql,
				core.EventTypeWarning,
				eventer.EventReasonFailedToUpdate,
				err.Error(),
			)
			return err
		}
		mysql.Spec.DatabaseSecret = ms.Spec.DatabaseSecret
	}
	return nil
}

func (c *Controller) createDatabaseSecret(mysql *api.MySQL) (*core.SecretVolumeSource, error) {
	authSecretName := mysql.Name + "-auth"

	sc, err := c.checkSecret(authSecretName, mysql)
	if err != nil {
		return nil, err
	}
	if sc == nil {
		randPassword := ""

		// if the password starts with "-", it will cause error in bash scripts (in mysql-tools)
		for randPassword = rand.GeneratePassword(); randPassword[0] == '-'; {
		}

		secret := &core.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:   authSecretName,
				Labels: mysql.OffshootSelectors(),
			},
			Type: core.SecretTypeOpaque,
			StringData: map[string]string{
				KeyMySQLUser:     mysqlUser,
				KeyMySQLPassword: randPassword,
			},
		}
		if _, err := c.Client.CoreV1().Secrets(mysql.Namespace).Create(secret); err != nil {
			return nil, err
		}
	}
	return &core.SecretVolumeSource{
		SecretName: authSecretName,
	}, nil
}

func (c *Controller) checkSecret(secretName string, mysql *api.MySQL) (*core.Secret, error) {
	secret, err := c.Client.CoreV1().Secrets(mysql.Namespace).Get(secretName, metav1.GetOptions{})
	if err != nil {
		if kerr.IsNotFound(err) {
			return nil, nil
		}
		return nil, err

	}
	if secret.Labels[api.LabelDatabaseKind] != api.ResourceKindMySQL ||
		secret.Labels[api.LabelDatabaseName] != mysql.Name {
		return nil, fmt.Errorf(`intended secret "%v" already exists`, secretName)
	}
	return secret, nil
}
