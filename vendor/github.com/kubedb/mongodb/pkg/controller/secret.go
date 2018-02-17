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
	mongodbUser = "root"

	KeyMongoDBUser     = "user"
	KeyMongoDBPassword = "password"

	ExporterSecretPath = "/var/run/secrets/kubedb.com/"
)

func (c *Controller) ensureDatabaseSecret(mongodb *api.MongoDB) error {
	if mongodb.Spec.DatabaseSecret == nil {
		secretVolumeSource, err := c.createDatabaseSecret(mongodb)
		if err != nil {
			c.recorder.Eventf(
				mongodb.ObjectReference(),
				core.EventTypeWarning,
				eventer.EventReasonFailedToCreate,
				`Failed to create Database Secret. Reason: %v`,
				err.Error(),
			)
			return err
		}

		ms, _, err := util.PatchMongoDB(c.ExtClient, mongodb, func(in *api.MongoDB) *api.MongoDB {
			in.Spec.DatabaseSecret = secretVolumeSource
			return in
		})
		if err != nil {
			c.recorder.Eventf(
				mongodb.ObjectReference(),
				core.EventTypeWarning,
				eventer.EventReasonFailedToUpdate,
				err.Error(),
			)
			return err
		}
		mongodb.Spec.DatabaseSecret = ms.Spec.DatabaseSecret
	}
	return nil
}

func (c *Controller) createDatabaseSecret(mongodb *api.MongoDB) (*core.SecretVolumeSource, error) {
	authSecretName := mongodb.Name + "-auth"

	sc, err := c.checkSecret(authSecretName, mongodb)
	if err != nil {
		return nil, err
	}
	if sc == nil {
		secret := &core.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name: authSecretName,
				Labels: map[string]string{
					api.LabelDatabaseKind: api.ResourceKindMongoDB,
					api.LabelDatabaseName: mongodb.Name,
				},
			},
			Type: core.SecretTypeOpaque,
			StringData: map[string]string{
				KeyMongoDBUser:     mongodbUser,
				KeyMongoDBPassword: rand.GeneratePassword(),
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

func (c *Controller) checkSecret(secretName string, mongodb *api.MongoDB) (*core.Secret, error) {
	secret, err := c.Client.CoreV1().Secrets(mongodb.Namespace).Get(secretName, metav1.GetOptions{})
	if err != nil {
		if kerr.IsNotFound(err) {
			return nil, nil
		} else {
			return nil, err
		}
	}
	if secret.Labels[api.LabelDatabaseKind] != api.ResourceKindMongoDB ||
		secret.Labels[api.LabelDatabaseName] != mongodb.Name {
		return nil, fmt.Errorf(`intended secret "%v" already exists`, secretName)
	}

	return secret, nil
}

func (c *Controller) deleteSecret(dormantDb *api.DormantDatabase, secretVolume *core.SecretVolumeSource) error {
	secretFound := false
	mongodbList, err := c.ExtClient.MongoDBs(dormantDb.Namespace).List(metav1.ListOptions{})
	if err != nil {
		return err
	}

	for _, mongodb := range mongodbList.Items {
		databaseSecret := mongodb.Spec.DatabaseSecret
		if databaseSecret != nil {
			if databaseSecret.SecretName == secretVolume.SecretName {
				secretFound = true
				break
			}
		}
	}

	if !secretFound {
		labelMap := map[string]string{
			api.LabelDatabaseKind: api.ResourceKindMongoDB,
		}
		dormantDatabaseList, err := c.ExtClient.DormantDatabases(dormantDb.Namespace).List(
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

			databaseSecret := ddb.Spec.Origin.Spec.MongoDB.DatabaseSecret
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
