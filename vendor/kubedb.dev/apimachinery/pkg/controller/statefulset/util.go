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

package statefulset

import (
	"context"
	"fmt"

	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha1"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	kmapi "kmodules.xyz/client-go/api/v1"
	dmcond "kmodules.xyz/client-go/dynamic/conditions"
)

func (c *Controller) extractDatabaseInfo(sts *appsv1.StatefulSet) (*databaseInfo, error) {
	// read the controlling owner
	owner := metav1.GetControllerOf(sts)
	if owner == nil {
		return nil, fmt.Errorf("StatefulSet %s/%s has no controlling owner", sts.Namespace, sts.Name)
	}
	gv, err := schema.ParseGroupVersion(owner.APIVersion)
	if err != nil {
		return nil, err
	}
	dbInfo := &databaseInfo{
		do: dmcond.DynamicOptions{
			Client:    c.DynamicClient,
			Kind:      owner.Kind,
			Name:      owner.Name,
			Namespace: sts.Namespace,
		},
	}
	dbInfo.do.GVR = schema.GroupVersionResource{
		Group:   gv.Group,
		Version: gv.Version,
	}
	switch owner.Kind {
	case api.ResourceKindElasticsearch:
		dbInfo.do.GVR.Resource = api.ResourcePluralElasticsearch
		es, err := c.DBClient.KubedbV1alpha1().Elasticsearches(dbInfo.do.Namespace).Get(context.TODO(), dbInfo.do.Name, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		dbInfo.replicasReady, dbInfo.msg, err = es.IsReplicasReady(c.StsLister)
		if err != nil {
			return nil, err
		}

	case api.ResourceKindMongoDB:
		dbInfo.do.GVR.Resource = api.ResourcePluralMongoDB
		mg, err := c.DBClient.KubedbV1alpha1().MongoDBs(dbInfo.do.Namespace).Get(context.TODO(), dbInfo.do.Name, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		dbInfo.replicasReady, dbInfo.msg, err = mg.IsReplicasReady(c.StsLister)
		if err != nil {
			return nil, err
		}

	case api.ResourceKindMySQL:
		dbInfo.do.GVR.Resource = api.ResourcePluralMySQL
		my, err := c.DBClient.KubedbV1alpha1().MySQLs(dbInfo.do.Namespace).Get(context.TODO(), dbInfo.do.Name, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		dbInfo.replicasReady, dbInfo.msg, err = my.IsReplicasReady(c.StsLister)
		if err != nil {
			return nil, err
		}

	case api.ResourceKindPerconaXtraDB:
		dbInfo.do.GVR.Resource = api.ResourcePluralPerconaXtraDB
		px, err := c.DBClient.KubedbV1alpha1().PerconaXtraDBs(dbInfo.do.Namespace).Get(context.TODO(), dbInfo.do.Name, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		dbInfo.replicasReady, dbInfo.msg, err = px.IsReplicasReady(c.StsLister)
		if err != nil {
			return nil, err
		}

	case api.ResourceKindMariaDB:
		dbInfo.do.GVR.Resource = api.ResourcePluralMariaDB
		mr, err := c.DBClient.KubedbV1alpha1().MariaDBs(dbInfo.do.Namespace).Get(context.TODO(), dbInfo.do.Name, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		dbInfo.replicasReady, dbInfo.msg, err = mr.IsReplicasReady(c.StsLister)
		if err != nil {
			return nil, err
		}

	case api.ResourceKindPostgres:
		dbInfo.do.GVR.Resource = api.ResourcePluralPostgres
		pg, err := c.DBClient.KubedbV1alpha1().Postgreses(dbInfo.do.Namespace).Get(context.TODO(), dbInfo.do.Name, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		dbInfo.replicasReady, dbInfo.msg, err = pg.IsReplicasReady(c.StsLister)
		if err != nil {
			return nil, err
		}

	case api.ResourceKindRedis:
		dbInfo.do.GVR.Resource = api.ResourcePluralRedis
		rd, err := c.DBClient.KubedbV1alpha1().Redises(dbInfo.do.Namespace).Get(context.TODO(), dbInfo.do.Name, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		dbInfo.replicasReady, dbInfo.msg, err = rd.IsReplicasReady(c.StsLister)
		if err != nil {
			return nil, err
		}

	case api.ResourceKindMemcached:
		dbInfo.do.GVR.Resource = api.ResourcePluralMemcached
		mc, err := c.DBClient.KubedbV1alpha1().Memcacheds(dbInfo.do.Namespace).Get(context.TODO(), dbInfo.do.Name, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		dbInfo.replicasReady, dbInfo.msg, err = mc.IsReplicasReady(c.StsLister)
		if err != nil {
			return nil, err
		}

	case api.ResourceKindProxySQL:
		dbInfo.do.GVR.Resource = api.ResourcePluralProxySQL
		pxql, err := c.DBClient.KubedbV1alpha1().ProxySQLs(dbInfo.do.Namespace).Get(context.TODO(), dbInfo.do.Name, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		dbInfo.replicasReady, dbInfo.msg, err = pxql.IsReplicasReady(c.StsLister)
		if err != nil {
			return nil, err
		}

	case api.ResourceKindPgBouncer:
		dbInfo.do.GVR.Resource = api.ResourcePluralPgBouncer
		pgb, err := c.DBClient.KubedbV1alpha1().PgBouncers(dbInfo.do.Namespace).Get(context.TODO(), dbInfo.do.Name, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		dbInfo.replicasReady, dbInfo.msg, err = pgb.IsReplicasReady(c.StsLister)
		if err != nil {
			return nil, err
		}

	case api.ResourceKindEtcd:
		dbInfo.do.GVR.Resource = api.ResourcePluralEtcd
		etcd, err := c.DBClient.KubedbV1alpha1().Etcds(dbInfo.do.Namespace).Get(context.TODO(), dbInfo.do.Name, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		dbInfo.replicasReady, dbInfo.msg, err = etcd.IsReplicasReady(c.StsLister)
		if err != nil {
			return nil, err
		}

	default:
		return nil, fmt.Errorf("unknown database kind: %s", owner.Kind)
	}
	return dbInfo, nil
}

func (c *Controller) ensureReadyReplicasCond(dbInfo *databaseInfo) error {
	dbCond := kmapi.Condition{
		Type:    api.DatabaseReplicaReady,
		Message: dbInfo.msg,
	}

	if dbInfo.replicasReady {
		dbCond.Status = kmapi.ConditionTrue
		dbCond.Reason = api.AllReplicasAreReady
	} else {
		dbCond.Status = kmapi.ConditionFalse
		dbCond.Reason = api.SomeReplicasAreNotReady
	}

	// Add "ReplicasReady" condition to the respective database CR
	return dbInfo.do.SetCondition(dbCond)
}
