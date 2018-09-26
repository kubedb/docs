package controller

import (
	"encoding/base64"
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
	mongodbUser = "root"

	KeyMongoDBUser     = "user"
	KeyMongoDBPassword = "password"
	KeyForKeyFile      = "key.txt"

	DatabaseSecretSuffix = "-auth"
	KeyFileSecretSuffix  = "-keyfile"

	ExporterSecretPath = "/var/run/secrets/kubedb.com/"
)

func (c *Controller) ensureDatabaseSecret(mongodb *api.MongoDB) error {
	if mongodb.Spec.DatabaseSecret == nil {
		secretVolumeSource, err := c.createDatabaseSecret(mongodb)
		if err != nil {
			c.recorder.Eventf(
				mongodb,
				core.EventTypeWarning,
				eventer.EventReasonFailedToCreate,
				`Failed to create Database Secret. Reason: %v`,
				err.Error(),
			)

			return err
		}

		ms, _, err := util.PatchMongoDB(c.ExtClient.KubedbV1alpha1(), mongodb, func(in *api.MongoDB) *api.MongoDB {
			in.Spec.DatabaseSecret = secretVolumeSource
			return in
		})
		if err != nil {
			c.recorder.Eventf(
				mongodb,
				core.EventTypeWarning,
				eventer.EventReasonFailedToUpdate,
				err.Error(),
			)

			return err
		}
		mongodb.Spec.DatabaseSecret = ms.Spec.DatabaseSecret
	}

	// keyfile secret for mongodb replication
	if mongodb.Spec.ReplicaSet != nil &&
		mongodb.Spec.ReplicaSet.KeyFile == nil {

		secretVolumeSource, err := c.createKeyFileSecret(mongodb)
		if err != nil {
			c.recorder.Eventf(
				mongodb,
				core.EventTypeWarning,
				eventer.EventReasonFailedToCreate,
				`Failed to create KeyFile Secret. Reason: %v`,
				err.Error(),
			)

			return err
		}

		ms, _, err := util.PatchMongoDB(c.ExtClient.KubedbV1alpha1(), mongodb, func(in *api.MongoDB) *api.MongoDB {
			in.Spec.ReplicaSet.KeyFile = secretVolumeSource
			return in
		})
		if err != nil {
			c.recorder.Eventf(
				mongodb,
				core.EventTypeWarning,
				eventer.EventReasonFailedToUpdate,
				err.Error(),
			)
			return err
		}
		mongodb.Spec.ReplicaSet.KeyFile = ms.Spec.ReplicaSet.KeyFile
	}

	return nil
}

func (c *Controller) createDatabaseSecret(mongodb *api.MongoDB) (*core.SecretVolumeSource, error) {
	authSecretName := mongodb.Name + DatabaseSecretSuffix

	sc, err := c.checkSecret(authSecretName, mongodb)
	if err != nil {
		return nil, err
	}
	if sc == nil {
		randPassword := ""

		// if the password starts with "-" it will cause error in bash scripts (in mongodb-tools)
		for randPassword = rand.GeneratePassword(); randPassword[0] == '-'; {
		}

		secret := &core.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:   authSecretName,
				Labels: mongodb.OffshootSelectors(),
			},
			Type: core.SecretTypeOpaque,
			StringData: map[string]string{
				KeyMongoDBUser:     mongodbUser,
				KeyMongoDBPassword: randPassword,
			},
		}
		if _, err := c.Client.CoreV1().Secrets(mongodb.Namespace).Create(secret); err != nil {
			return nil, err
		}
	}
	return &core.SecretVolumeSource{
		SecretName: authSecretName,
	}, nil
}

func (c *Controller) createKeyFileSecret(mongodb *api.MongoDB) (*core.SecretVolumeSource, error) {
	tokenSecretName := mongodb.Name + KeyFileSecretSuffix

	sc, err := c.checkSecret(tokenSecretName, mongodb)
	if err != nil {
		return nil, err
	}
	if sc == nil {
		randToken := rand.GenerateTokenWithLength(756)
		base64Token := base64.StdEncoding.EncodeToString([]byte(randToken))

		secret := &core.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:   tokenSecretName,
				Labels: mongodb.OffshootLabels(),
			},
			Type: core.SecretTypeOpaque,
			StringData: map[string]string{
				KeyForKeyFile: base64Token,
			},
		}
		if _, err := c.Client.CoreV1().Secrets(mongodb.Namespace).Create(secret); err != nil {
			return nil, err
		}
	}
	return &core.SecretVolumeSource{
		SecretName: tokenSecretName,
	}, nil
}

func (c *Controller) checkSecret(secretName string, mongodb *api.MongoDB) (*core.Secret, error) {
	secret, err := c.Client.CoreV1().Secrets(mongodb.Namespace).Get(secretName, metav1.GetOptions{})
	if err != nil {
		if kerr.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	if secret.Labels[api.LabelDatabaseKind] != api.ResourceKindMongoDB ||
		secret.Labels[api.LabelDatabaseName] != mongodb.Name {
		return nil, fmt.Errorf(`intended secret "%v" already exists`, secretName)
	}
	return secret, nil
}
