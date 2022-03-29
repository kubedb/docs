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
	"fmt"

	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"

	_ "github.com/lib/pq"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	"kmodules.xyz/client-go/tools/certholder"
	"xorm.io/xorm"
)

func (c *Controller) GetPostgresClient(ctx context.Context, db *api.Postgres, dnsName string, port int32) (*xorm.Engine, error) {
	user, pass, err := c.GetPostgresAuthCredentials(ctx, db)
	if err != nil {
		return nil, fmt.Errorf("DB basic auth is not found for PostgreSQL %v/%v", db.Namespace, db.Name)
	}
	cnnstr := ""
	sslMode := db.Spec.SSLMode

	//  sslMode == "prefer" and sslMode == "allow"  don't have support for github.com/lib/pq postgres client. as we are using
	// github.com/lib/pq postgres client utils for connecting our server we need to access with  any of require , verify-ca, verify-full or disable.
	// here we have chosen "require" sslmode to connect postgres as a client
	if sslMode == "prefer" || sslMode == "allow" {
		sslMode = "require"
	}
	if db.Spec.TLS != nil {
		secretName := db.GetCertSecretName(api.PostgresClientCert)

		certSecret, err := c.Client.CoreV1().Secrets(db.Namespace).Get(ctx, secretName, metav1.GetOptions{})
		if err != nil {
			klog.Error(err, "failed to get certificate secret.", secretName)
			return nil, err
		}

		certs, _ := certholder.DefaultHolder.ForResource(api.SchemeGroupVersion.WithResource(api.ResourcePluralPostgres), db.ObjectMeta)
		paths, err := certs.Save(certSecret)
		if err != nil {
			klog.Error(err, "failed to save certificate")
			return nil, err
		}
		if db.Spec.ClientAuthMode == api.ClientAuthModeCert {
			cnnstr = fmt.Sprintf("user=%s password=%s host=%s port=%d connect_timeout=15 dbname=postgres sslmode=%s sslrootcert=%s sslcert=%s sslkey=%s", user, pass, dnsName, port, sslMode, paths.CACert, paths.Cert, paths.Key)
		} else {
			cnnstr = fmt.Sprintf("user=%s password=%s host=%s port=%d connect_timeout=15 dbname=postgres sslmode=%s sslrootcert=%s", user, pass, dnsName, port, sslMode, paths.CACert)
		}
	} else {
		cnnstr = fmt.Sprintf("user=%s password=%s host=%s port=%d connect_timeout=15 dbname=postgres sslmode=%s", user, pass, dnsName, port, sslMode)
	}

	eng, err := xorm.NewEngine("postgres", cnnstr)
	if err != nil {
		return nil, fmt.Errorf("failed to create xorm engine")
	}
	eng.SetDefaultContext(ctx)
	return eng, nil
}
