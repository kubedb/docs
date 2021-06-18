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
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"strings"
	"time"

	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"

	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	mgoptions "go.mongodb.org/mongo-driver/mongo/options"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	"kmodules.xyz/client-go/tools/certholder"
)

func (c *Controller) GetMongoClient(ctx context.Context, db *api.MongoDB, url, repSetName string) (*mongo.Client, error) {
	clientOpts, err := c.GetMongoDBClientOpts(db, url, repSetName)
	if err != nil {
		return nil, err
	}

	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		return nil, err
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		return nil, err
	}

	return client, nil
}

func (c *Controller) GetURL(db *api.MongoDB, clientPodName string) string {
	nodeType := clientPodName[:strings.LastIndex(clientPodName, "-")]
	return fmt.Sprintf("%s.%s.%s.svc", clientPodName, db.GoverningServiceName(nodeType), db.Namespace)
}

func (c *Controller) GetMongoDBClientOpts(db *api.MongoDB, url, repSetName string) (*mgoptions.ClientOptions, error) {
	repSetConfig := ""
	if repSetName != "" {
		repSetConfig = "replicaSet=" + repSetName + "&"
	}

	user, pass, err := c.GetMongoDBRootCredentials(db)
	if err != nil {
		return nil, err
	}
	var clientOpts *mgoptions.ClientOptions
	if db.Spec.TLS != nil {
		secretName := db.GetCertSecretName(api.MongoDBClientCert, "")
		certSecret, err := c.Client.CoreV1().Secrets(db.Namespace).Get(context.TODO(), secretName, metav1.GetOptions{})
		if err != nil {
			klog.Error(err, "failed to get certificate secret. ", secretName)
			return nil, err
		}

		certs, _ := certholder.DefaultHolder.
			ForResource(api.SchemeGroupVersion.WithResource(api.ResourcePluralMongoDB), db.ObjectMeta)
		_, err = certs.Save(certSecret)
		if err != nil {
			klog.Error(err, "failed to save certificate")
			return nil, err
		}

		paths, err := certs.Get(secretName)
		if err != nil {
			return nil, err
		}

		uri := fmt.Sprintf("mongodb://%s:%s@%s/admin?%vtls=true&tlsCAFile=%v&tlsCertificateKeyFile=%v", user, pass, url, repSetConfig, paths.CACert, paths.Pem)
		clientOpts = mgoptions.Client().ApplyURI(uri)
	} else {
		clientOpts = mgoptions.Client().ApplyURI(fmt.Sprintf("mongodb://%s:%s@%s/admin?%v", user, pass, url, repSetConfig))
	}

	clientOpts.SetDirect(false)
	clientOpts.SetConnectTimeout(5 * time.Second)

	return clientOpts, nil
}

func (c *Controller) GetMongoDBRootCredentials(db *api.MongoDB) (string, string, error) {
	if db.Spec.AuthSecret == nil {
		return "", "", errors.New("no database secret")
	}
	secret, err := c.Client.CoreV1().Secrets(db.Namespace).Get(context.TODO(), db.Spec.AuthSecret.Name, metav1.GetOptions{})
	if err != nil {
		return "", "", err
	}
	return string(secret.Data[core.BasicAuthUsernameKey]), string(secret.Data[core.BasicAuthPasswordKey]), nil
}

func (c *Controller) CreateTLSUsers(db *api.MongoDB) error {
	secretName := db.GetCertSecretName(api.MongoDBClientCert, "")
	certSecret, err := c.Client.CoreV1().Secrets(db.Namespace).Get(context.TODO(), secretName, metav1.GetOptions{})
	if err != nil {
		klog.Error(err, "failed to get certificate secret", "Secret", secretName)
		return err
	}

	blk, _ := pem.Decode(certSecret.Data[core.TLSCertKey])

	clientCert, err := x509.ParseCertificate(blk.Bytes)
	if err != nil {
		klog.Error(err, "failed to get certificate secret", "Secret", secretName)
		return err
	}
	tlsUserName := clientCert.Subject.String()

	if db.Spec.ShardTopology != nil {
		err = c.CreateTLSUser(db, strings.Join(db.MongosHosts(), ","), "", tlsUserName)
		if err != nil {
			return err
		}

		for i := int32(0); i < db.Spec.ShardTopology.Shard.Shards; i++ {
			err = c.CreateTLSUser(db, strings.Join(db.ShardHosts(i), ","), db.ShardRepSetName(i), tlsUserName)
			if err != nil {
				return err
			}
		}
	} else {
		err = c.CreateTLSUser(db, strings.Join(db.Hosts(), ","), db.RepSetName(), tlsUserName)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Controller) CreateTLSUser(db *api.MongoDB, url, repSetName, tlsUserName string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()
	dbClient, err := c.GetMongoClient(ctx, db, url, repSetName)
	if err != nil {
		return err
	}
	defer func() {
		err = dbClient.Disconnect(context.TODO())
		if err != nil {
			klog.Errorf("Failed to disconnect client for mongodb %s/%s. error: %v", db.Namespace, db.Name, err)
		}
	}()

	res := make(map[string]interface{})
	err = dbClient.Database("$external").RunCommand(context.Background(), bson.D{{Key: "usersInfo", Value: tlsUserName}}).Decode(&res)
	if err != nil {
		klog.Error("Failed to get user info. error: ", err)
		return err
	}
	users, ok := res["users"].(primitive.A)
	if ok && len(users) == 0 {
		klog.Info("Creating TLS user with name ", tlsUserName)
		err = dbClient.Database("$external").RunCommand(context.Background(),
			bson.D{
				{
					Key:   "createUser",
					Value: tlsUserName,
				},
				{
					Key: "roles",
					Value: bson.A{
						bson.D{
							{Key: "role", Value: "root"},
							{Key: "db", Value: "admin"},
						},
					},
				},
			}).Decode(&res)
		if err != nil {
			klog.Error("Failed to create tls user. error: ", err)
			return err
		}
	}

	return nil
}
