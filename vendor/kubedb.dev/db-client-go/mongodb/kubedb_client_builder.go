/*
Copyright AppsCode Inc. and Contributors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package mongodb

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"

	"go.mongodb.org/mongo-driver/mongo"
	mgoptions "go.mongodb.org/mongo-driver/mongo/options"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
	"kmodules.xyz/client-go/tools/certholder"
)

type KubeDBClientBuilder struct {
	kubeClient kubernetes.Interface
	db         *api.MongoDB
	url        string
	podName    string
	repSetName string
	direct     bool
	certs      *certholder.ResourceCerts
	ctx        context.Context
}

func NewKubeDBClientBuilder(db *api.MongoDB, kubeClient kubernetes.Interface) *KubeDBClientBuilder {
	return &KubeDBClientBuilder{
		kubeClient: kubeClient,
		db:         db,
		direct:     false,
	}
}

func (o *KubeDBClientBuilder) WithURL(url string) *KubeDBClientBuilder {
	o.url = url
	return o
}

func (o *KubeDBClientBuilder) WithPod(podName string) *KubeDBClientBuilder {
	o.podName = podName
	o.direct = true
	return o
}

func (o *KubeDBClientBuilder) WithReplSet(replSetName string) *KubeDBClientBuilder {
	o.repSetName = replSetName
	return o
}

func (o *KubeDBClientBuilder) WithContext(ctx context.Context) *KubeDBClientBuilder {
	o.ctx = ctx
	return o
}

func (o *KubeDBClientBuilder) WithDirect() *KubeDBClientBuilder {
	o.direct = true
	return o
}

func (o *KubeDBClientBuilder) WithCerts(certs *certholder.ResourceCerts) *KubeDBClientBuilder {
	o.certs = certs
	return o
}

func (o *KubeDBClientBuilder) GetMongoClient() (*Client, error) {
	db := o.db

	if o.podName != "" {
		o.url = o.getURL()
	}

	if o.podName == "" && o.url == "" {
		if db.Spec.ShardTopology != nil {
			// Shard
			o.url = strings.Join(db.MongosHosts(), ",")
		} else {
			// Standalone or ReplicaSet
			o.url = strings.Join(db.Hosts(), ",")
		}
	}

	clientOpts, err := o.getMongoDBClientOpts()
	if err != nil {
		return nil, err
	}

	client, err := mongo.Connect(o.ctx, clientOpts)
	if err != nil {
		return nil, err
	}

	err = client.Ping(o.ctx, nil)
	if err != nil {
		return nil, err
	}

	return &Client{
		Client: client,
	}, nil
}

func (o *KubeDBClientBuilder) getURL() string {
	nodeType := o.podName[:strings.LastIndex(o.podName, "-")]
	return fmt.Sprintf("%s.%s.%s.svc", o.podName, o.db.GoverningServiceName(nodeType), o.db.Namespace)
}

func (o *KubeDBClientBuilder) getMongoDBClientOpts() (*mgoptions.ClientOptions, error) {
	db := o.db
	repSetConfig := ""
	if o.repSetName != "" {
		repSetConfig = "replicaSet=" + o.repSetName + "&"
	}

	user, pass, err := o.getMongoDBRootCredentials()
	if err != nil {
		return nil, err
	}
	var clientOpts *mgoptions.ClientOptions
	if db.Spec.TLS != nil {
		secretName := db.GetCertSecretName(api.MongoDBClientCert, "")
		var paths *certholder.Paths
		if o.certs == nil {
			certSecret, err := o.kubeClient.CoreV1().Secrets(db.Namespace).Get(context.TODO(), secretName, metav1.GetOptions{})
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

			paths, err = certs.Get(secretName)
			if err != nil {
				return nil, err
			}
		} else {
			paths, err = o.certs.Get(secretName)
			if err != nil {
				return nil, err
			}
		}

		uri := fmt.Sprintf("mongodb://%s:%s@%s/admin?%vtls=true&tlsCAFile=%v&tlsCertificateKeyFile=%v", user, pass, o.url, repSetConfig, paths.CACert, paths.Pem)
		clientOpts = mgoptions.Client().ApplyURI(uri)
	} else {
		clientOpts = mgoptions.Client().ApplyURI(fmt.Sprintf("mongodb://%s:%s@%s/admin?%v", user, pass, o.url, repSetConfig))
	}

	clientOpts.SetDirect(o.direct)
	clientOpts.SetConnectTimeout(5 * time.Second)

	return clientOpts, nil
}

func (o *KubeDBClientBuilder) getMongoDBRootCredentials() (string, string, error) {
	db := o.db
	if db.Spec.AuthSecret == nil {
		return "", "", errors.New("no database secret")
	}
	secret, err := o.kubeClient.CoreV1().Secrets(db.Namespace).Get(context.TODO(), db.Spec.AuthSecret.Name, metav1.GetOptions{})
	if err != nil {
		return "", "", err
	}
	return string(secret.Data[core.BasicAuthUsernameKey]), string(secret.Data[core.BasicAuthPasswordKey]), nil
}
