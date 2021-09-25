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
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"time"

	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"
	"kubedb.dev/apimachinery/client/clientset/versioned/typed/kubedb/v1alpha2/util"
	"kubedb.dev/apimachinery/pkg/eventer"
	validator "kubedb.dev/redis/pkg/admission"

	"github.com/Masterminds/semver/v3"
	rd "github.com/go-redis/redis"
	"github.com/pkg/errors"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	kutil "kmodules.xyz/client-go"
	kmapi "kmodules.xyz/client-go/api/v1"
	dynamic_util "kmodules.xyz/client-go/dynamic"
)

func (c *Controller) createSentinel(db *api.RedisSentinel) error {
	if err := validator.ValidateRedisSentinel(c.Client, c.DBClient, db, true); err != nil {
		c.Recorder.Event(
			db,
			core.EventTypeWarning,
			eventer.EventReasonInvalid,
			err.Error(),
		)
		klog.Errorln(err)
		return nil // user error so just record error and don't retry.
	}

	// ensure Governing Service
	if err := c.ensureSentinelGoverningService(db); err != nil {
		return fmt.Errorf(`failed to create governing Service for : "%v/%v". Reason: %v`, db.Namespace, db.Name, err)
	}

	// ensure auth require for redis
	if err := c.ensureSentinelAuthSecret(db); err != nil {
		return err
	}

	// Ensure ClusterRoles for statefulsets
	if err := c.ensureSentinelRBACStuff(db); err != nil {
		return err
	}
	// ensure database Service
	vt1, err := c.ensureSentinelService(db)
	if err != nil {
		return err
	}
	// wait for  Certificates secrets
	if db.Spec.TLS != nil {
		ok, err := dynamic_util.ResourcesExists(
			c.DynamicClient,
			core.SchemeGroupVersion.WithResource("secrets"),
			db.Namespace,
			db.GetCertSecretName(api.RedisServerCert),
			db.GetCertSecretName(api.RedisClientCert),
			db.GetCertSecretName(api.RedisMetricsExporterCert),
		)
		if err != nil {
			return err
		}
		if !ok {
			klog.Infof("wait for all certificate secrets for Redis %s/%s", db.Namespace, db.Name)
			return nil
		}
	}

	// ensure database StatefulSet
	vt, err := c.ensureSentinelStatefulSet(db)
	if err != nil && err != ErrStsNotReady {
		return err
	}
	if err == ErrStsNotReady {
		return nil
	}

	if vt1 == kutil.VerbCreated && vt == kutil.VerbCreated {
		c.Recorder.Event(
			db,
			core.EventTypeNormal,
			eventer.EventReasonSuccessful,
			"Successfully created Redis Sentinel",
		)
	} else if vt1 == kutil.VerbPatched || vt == kutil.VerbPatched {
		c.Recorder.Event(
			db,
			core.EventTypeNormal,
			eventer.EventReasonSuccessful,
			"Successfully patched Redis Sentinel",
		)
	}

	_, err = c.ensureSentinelAppBinding(db)
	if err != nil {
		klog.Errorln(err)
		return err
	}

	// ensure StatsService for desired monitoring
	if _, err := c.ensureSentinelStatsService(db); err != nil {
		c.Recorder.Eventf(
			db,
			core.EventTypeWarning,
			eventer.EventReasonFailedToCreate,
			"Failed to manage monitoring system. Reason: %v",
			err,
		)
		klog.Errorf("failed to manage monitoring system. Reason: %v", err)
		return nil
	}

	if err := c.manageSentinelMonitor(db); err != nil {
		c.Recorder.Eventf(
			db,
			core.EventTypeWarning,
			eventer.EventReasonFailedToCreate,
			"Failed to manage monitoring system. Reason: %v",
			err,
		)
		klog.Errorf("failed to manage monitoring system. Reason: %v", err)
		return nil
	}

	// Check: ReplicaReady --> AcceptingConnection --> Ready --> Provisioned
	// If spec.Init.WaitForInitialRestore is true, but data wasn't restored successfully,
	// process won't reach here (returned nil at the beginning). As it is here, that means data was restored successfully.
	// No need to check for IsConditionTrue(DataRestored).
	if kmapi.IsConditionTrue(db.Status.Conditions, api.DatabaseReplicaReady) &&
		kmapi.IsConditionTrue(db.Status.Conditions, api.DatabaseAcceptingConnection) &&
		kmapi.IsConditionTrue(db.Status.Conditions, api.DatabaseReady) &&
		!kmapi.IsConditionTrue(db.Status.Conditions, api.DatabaseProvisioned) {

		_, err := util.UpdateRedisSentinelStatus(
			context.TODO(),
			c.DBClient.KubedbV1alpha2(),
			db.ObjectMeta,
			func(in *api.RedisSentinelStatus) (types.UID, *api.RedisSentinelStatus) {
				in.Conditions = kmapi.SetCondition(in.Conditions,
					kmapi.Condition{
						Type:               api.DatabaseProvisioned,
						Status:             core.ConditionTrue,
						Reason:             api.DatabaseSuccessfullyProvisioned,
						ObservedGeneration: db.Generation,
						Message:            fmt.Sprintf("The Redis: %s/%s is successfully provisioned.", db.Namespace, db.Name),
					})
				return db.UID, in
			},
			metav1.UpdateOptions{},
		)
		if err != nil {
			return err
		}
	}

	// If the database is successfully provisioned,
	// Set spec.Init.Initialized to true, if init!=nil.
	// This will prevent the operator from re-initializing the database.
	if kmapi.IsConditionTrue(db.Status.Conditions, api.DatabaseProvisioned) {
		_, _, err := util.CreateOrPatchRedisSentinel(context.TODO(), c.DBClient.KubedbV1alpha2(), db.ObjectMeta, func(in *api.RedisSentinel) *api.RedisSentinel {
			return in
		}, metav1.PatchOptions{})

		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Controller) haltSentinel(db *api.RedisSentinel) error {
	if db.Spec.Halted && db.Spec.TerminationPolicy != api.TerminationPolicyHalt {
		return errors.New("can't halt db. 'spec.terminationPolicy' is not 'Halt'")
	}
	klog.Infof("Halting Redis Sentinel %v/%v", db.Namespace, db.Name)
	if err := c.haltSentinelDatabase(db); err != nil {
		return err
	}
	if err := c.waitUntilSentinelHalted(db); err != nil {
		return err
	}
	klog.Infof("update status of Redis Sentinel %v/%v to Halted.", db.Namespace, db.Name)
	if _, err := util.UpdateRedisStatus(context.TODO(), c.DBClient.KubedbV1alpha2(), db.ObjectMeta, func(in *api.RedisStatus) (types.UID, *api.RedisStatus) {
		in.Conditions = kmapi.SetCondition(in.Conditions, kmapi.Condition{
			Type:               api.DatabaseHalted,
			Status:             core.ConditionTrue,
			Reason:             api.DatabaseHaltedSuccessfully,
			ObservedGeneration: db.Generation,
			Message:            fmt.Sprintf("Redis Sentinel %s/%s successfully halted.", db.Namespace, db.Name),
		})
		return db.UID, in
	}, metav1.UpdateOptions{}); err != nil {
		return err
	}
	return nil
}

func (c *Controller) terminateSentinel(db *api.RedisSentinel) error {
	// If TerminationPolicy is "halt", keep PVCs,Secrets intact.
	if db.Spec.TerminationPolicy == api.TerminationPolicyHalt {
		if err := c.removeOwnerReferenceFromOffshootsForSentinel(db); err != nil {
			return err
		}
	} else {
		// If TerminationPolicy is "wipeOut", delete everything (ie, PVCs,Secrets,Snapshots).
		// If TerminationPolicy is "delete", delete PVCs and keep snapshots,secrets intact.
		// In both these cases, don't create dormantdatabase
		if err := c.setOwnerReferenceToOffshootsForSentinel(db); err != nil {
			return err
		}
	}

	if db.Spec.Monitor != nil {
		if err := c.deleteSentinelMonitor(db); err != nil {
			klog.Errorln(err)
			return nil
		}
	}
	return nil
}

func (c *Controller) setOwnerReferenceToOffshootsForSentinel(db *api.RedisSentinel) error {
	owner := metav1.NewControllerRef(db, api.SchemeGroupVersion.WithKind(api.ResourceKindRedisSentinel))
	selector := labels.SelectorFromSet(db.OffshootSelectors())
	secrets := db.GetPersistentSecrets()
	secrets = append(secrets, c.GetRedisSentinelSecrets(db)...)
	// If TerminationPolicy is "wipeOut", delete snapshots and secrets,
	// else, keep it intact.
	if db.Spec.TerminationPolicy == api.TerminationPolicyWipeOut {
		if err := c.wipeOutSentinel(db.ObjectMeta, secrets, owner); err != nil {
			return errors.Wrap(err, "error in wiping out database.")
		}
	} else {
		// Make sure secret's ownerreference is removed.
		if err := dynamic_util.RemoveOwnerReferenceForItems(
			context.TODO(),
			c.DynamicClient,
			core.SchemeGroupVersion.WithResource("secrets"),
			db.Namespace,
			secrets,
			db); err != nil {
			return err
		}
	}

	// delete PVC for both "wipeOut" and "delete" TerminationPolicy.
	return dynamic_util.EnsureOwnerReferenceForSelector(
		context.TODO(),
		c.DynamicClient,
		core.SchemeGroupVersion.WithResource("persistentvolumeclaims"),
		db.Namespace,
		selector,
		owner)
}

func (c *Controller) removeOwnerReferenceFromOffshootsForSentinel(db *api.RedisSentinel) error {
	// First, Get LabelSelector for Other Components
	labelSelector := labels.SelectorFromSet(db.OffshootSelectors())
	secrets := db.GetPersistentSecrets()
	secrets = append(secrets, c.GetRedisSentinelSecrets(db)...)
	if err := dynamic_util.RemoveOwnerReferenceForItems(
		context.TODO(),
		c.DynamicClient,
		core.SchemeGroupVersion.WithResource("secrets"),
		db.Namespace,
		secrets,
		db); err != nil {
		return err
	}
	if err := dynamic_util.RemoveOwnerReferenceForSelector(
		context.TODO(),
		c.DynamicClient,
		core.SchemeGroupVersion.WithResource("persistentvolumeclaims"),
		db.Namespace,
		labelSelector,
		db); err != nil {
		return err
	}
	return nil
}

func (c *Controller) getRedisSentinelClient(db *api.RedisSentinel, dnsName string, port int) (*rd.SentinelClient, error) {
	if db.Spec.AuthSecret == nil {
		return nil, errors.New("no database secret")
	}
	redisVersion, err := c.DBClient.CatalogV1alpha1().RedisVersions().Get(context.TODO(), string(db.Spec.Version), metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	curVersion, err := semver.NewVersion(redisVersion.Spec.Version)
	if err != nil {
		return nil, fmt.Errorf("can't get the version from RedisVersion spec")
	}
	rdOpts := &rd.Options{
		DialTimeout: 15 * time.Second,
		IdleTimeout: 3 * time.Second,
		PoolSize:    1,
		Addr:        fmt.Sprintf("%s:%v", dnsName, port),
	}
	if curVersion.Major() > 4 {
		authSecret, err := c.Client.CoreV1().Secrets(db.Namespace).Get(context.TODO(), db.Spec.AuthSecret.Name, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		rdOpts.Password = string(authSecret.Data[core.BasicAuthPasswordKey])
	}
	if db.Spec.TLS != nil {
		sec, err := c.Client.CoreV1().Secrets(db.Namespace).Get(context.TODO(), db.CertificateName(api.RedisClientCert), metav1.GetOptions{})
		if err != nil {
			klog.Error(err, "error in getting the secret")
			return nil, err
		}
		pool := x509.NewCertPool()
		pool.AppendCertsFromPEM(sec.Data["ca.crt"])
		cert, err := tls.X509KeyPair(sec.Data["tls.crt"], sec.Data["tls.key"])
		if err != nil {
			klog.Error(err, "error in making certificate")
			return nil, err
		}
		rdOpts.TLSConfig = &tls.Config{
			Certificates: []tls.Certificate{
				cert,
			},
			ClientCAs: pool,
			RootCAs:   pool,
		}
	}
	rdClient := rd.NewSentinelClient(rdOpts)
	return rdClient, nil
}
