package controller

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/appscode/go/crypto/rand"
	"github.com/appscode/go/log"
	"github.com/appscode/go/types"
	api "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	"github.com/kubedb/apimachinery/client/typed/kubedb/v1alpha1/util"
	"github.com/kubedb/apimachinery/pkg/docker"
	"github.com/kubedb/apimachinery/pkg/eventer"
	"github.com/kubedb/apimachinery/pkg/storage"
	apps "k8s.io/api/apps/v1beta1"
	batch "k8s.io/api/batch/v1"
	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	// Duration in Minute
	// Check whether pod under StatefulSet is running or not
	// Continue checking for this duration until failure
	durationCheckStatefulSet = time.Minute * 30
)

func (c *Controller) findService(mongodb *api.MongoDB) (bool, error) {
	name := mongodb.OffshootName()
	service, err := c.Client.CoreV1().Services(mongodb.Namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		if kerr.IsNotFound(err) {
			return false, nil
		} else {
			return false, err
		}
	}

	if service.Spec.Selector[api.LabelDatabaseName] != name {
		return false, fmt.Errorf(`Intended service "%v" already exists`, name)
	}

	return true, nil
}

func (c *Controller) createService(mongodb *api.MongoDB) error {
	svc := &core.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:   mongodb.OffshootName(),
			Labels: mongodb.OffshootLabels(),
		},
		Spec: core.ServiceSpec{
			Ports: []core.ServicePort{
				{
					Name:       "db",
					Port:       27017,
					TargetPort: intstr.FromString("db"),
				},
			},
			Selector: mongodb.OffshootLabels(),
		},
	}
	if mongodb.Spec.Monitor != nil &&
		mongodb.Spec.Monitor.Agent == api.AgentCoreosPrometheus &&
		mongodb.Spec.Monitor.Prometheus != nil {
		svc.Spec.Ports = append(svc.Spec.Ports, core.ServicePort{
			Name:       api.PrometheusExporterPortName,
			Port:       mongodb.Spec.Monitor.Prometheus.Port,
			TargetPort: intstr.FromString(api.PrometheusExporterPortName),
		})
	}

	if _, err := c.Client.CoreV1().Services(mongodb.Namespace).Create(svc); err != nil {
		return err
	}

	return nil
}

func (c *Controller) findStatefulSet(mongodb *api.MongoDB) (bool, error) {
	// SatatefulSet for MongoDB database
	statefulSet, err := c.Client.AppsV1beta1().StatefulSets(mongodb.Namespace).Get(mongodb.OffshootName(), metav1.GetOptions{})
	if err != nil {
		if kerr.IsNotFound(err) {
			return false, nil
		} else {
			return false, err
		}
	}

	if statefulSet.Labels[api.LabelDatabaseKind] != api.ResourceKindMongoDB {
		return false, fmt.Errorf(`Intended statefulSet "%v" already exists`, mongodb.OffshootName())
	}

	return true, nil
}

func (c *Controller) createStatefulSet(mongodb *api.MongoDB) (*apps.StatefulSet, error) {
	// SatatefulSet for MongoDB database
	statefulSet := &apps.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:        mongodb.OffshootName(),
			Namespace:   mongodb.Namespace,
			Labels:      mongodb.StatefulSetLabels(),
			Annotations: mongodb.StatefulSetAnnotations(),
		},
		Spec: apps.StatefulSetSpec{
			Replicas:    types.Int32P(1),
			ServiceName: c.opt.GoverningService,
			Template: core.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: mongodb.OffshootLabels(),
				},
				Spec: core.PodSpec{
					Containers: []core.Container{
						{
							Name:            api.ResourceNameMongoDB,
							Image:           fmt.Sprintf("%s:%s", docker.ImageMongoDB, mongodb.Spec.Version),
							ImagePullPolicy: core.PullIfNotPresent,
							Ports: []core.ContainerPort{
								{
									Name:          "db",
									ContainerPort: 27017,
								},
							},
							Resources: mongodb.Spec.Resources,
							VolumeMounts: []core.VolumeMount{
								{
									Name:      "data",
									MountPath: "/data/db", //Data files path of mongodb, ref: https://github.com/docker-library/docs/tree/master/mongo#where-to-store-data
								},
							},
							Args: []string{
								"--auth",
							},
							Env: []core.EnvVar{
								{
									Name:  "MONGO_INITDB_ROOT_USERNAME",
									Value: "root",
								},
							},
						},
					},
					NodeSelector:  mongodb.Spec.NodeSelector,
					Affinity:      mongodb.Spec.Affinity,
					SchedulerName: mongodb.Spec.SchedulerName,
					Tolerations:   mongodb.Spec.Tolerations,
				},
			},
		},
	}

	if mongodb.Spec.Monitor != nil &&
		mongodb.Spec.Monitor.Agent == api.AgentCoreosPrometheus &&
		mongodb.Spec.Monitor.Prometheus != nil {
		exporter := core.Container{
			Name: "exporter",
			Args: []string{
				"export",
				fmt.Sprintf("--address=:%d", mongodb.Spec.Monitor.Prometheus.Port),
				"--v=3",
			},
			Image:           docker.ImageOperator + ":" + c.opt.ExporterTag,
			ImagePullPolicy: core.PullIfNotPresent,
			Ports: []core.ContainerPort{
				{
					Name:          api.PrometheusExporterPortName,
					Protocol:      core.ProtocolTCP,
					ContainerPort: mongodb.Spec.Monitor.Prometheus.Port,
				},
			},
		}
		statefulSet.Spec.Template.Spec.Containers = append(statefulSet.Spec.Template.Spec.Containers, exporter)
	}

	if mongodb.Spec.DatabaseSecret == nil {
		secretVolumeSource, err := c.createDatabaseSecret(mongodb)
		if err != nil {
			return nil, err
		}

		_mongodb, err := util.TryPatchMongoDB(c.ExtClient, mongodb.ObjectMeta, func(in *api.MongoDB) *api.MongoDB {
			in.Spec.DatabaseSecret = secretVolumeSource
			return in
		})
		if err != nil {
			c.recorder.Eventf(mongodb.ObjectReference(), core.EventTypeWarning, eventer.EventReasonFailedToUpdate, err.Error())
			return nil, err
		}
		mongodb.Spec.DatabaseSecret = _mongodb.Spec.DatabaseSecret
	}

	//Set root user password from Secret
	setEnvFromSecret(statefulSet, mongodb.Spec.DatabaseSecret)

	// Add Data volume for StatefulSet
	addDataVolume(statefulSet, mongodb.Spec.Storage)

	// Add InitialScript to run at startup
	if mongodb.Spec.Init != nil && mongodb.Spec.Init.ScriptSource != nil {
		addInitialScript(statefulSet, mongodb.Spec.Init.ScriptSource)
	}

	if c.opt.EnableRbac {
		// Ensure ClusterRoles for database statefulsets
		if err := c.createRBACStuff(mongodb); err != nil {
			return nil, err
		}

		statefulSet.Spec.Template.Spec.ServiceAccountName = mongodb.Name
	}

	if _, err := c.Client.AppsV1beta1().StatefulSets(statefulSet.Namespace).Create(statefulSet); err != nil {
		return nil, err
	}

	return statefulSet, nil
}

// Set root user password from Secret, Through Env.
func setEnvFromSecret(statefulSet *apps.StatefulSet, secSource *core.SecretVolumeSource) {
	statefulSet.Spec.Template.Spec.Containers[0].Env = append(statefulSet.Spec.Template.Spec.Containers[0].Env,
		core.EnvVar{
			Name: "MONGO_INITDB_ROOT_PASSWORD",
			ValueFrom: &core.EnvVarSource{
				SecretKeyRef: &core.SecretKeySelector{
					LocalObjectReference: core.LocalObjectReference{
						Name: secSource.SecretName,
					},
					Key: ".admin",
				},
			},
		},
	)
}

func (c *Controller) findSecret(secretName, namespace string) (bool, error) {
	secret, err := c.Client.CoreV1().Secrets(namespace).Get(secretName, metav1.GetOptions{})
	if err != nil {
		if kerr.IsNotFound(err) {
			return false, nil
		} else {
			return false, err
		}
	}
	if secret == nil {
		return false, nil
	}

	return true, nil
}

func (c *Controller) createDatabaseSecret(mongodb *api.MongoDB) (*core.SecretVolumeSource, error) {
	authSecretName := mongodb.Name + "-admin-auth"

	found, err := c.findSecret(authSecretName, mongodb.Namespace)
	if err != nil {
		return nil, err
	}

	if !found {
		MONGO_PASSWORD := fmt.Sprintf("%s", rand.GeneratePassword())
		data := map[string][]byte{
			".admin": []byte(MONGO_PASSWORD),
		}

		secret := &core.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name: authSecretName,
				Labels: map[string]string{
					api.LabelDatabaseKind: api.ResourceKindMongoDB,
				},
			},
			Type: core.SecretTypeOpaque,
			Data: data, // Add secret data
		}
		if _, err := c.Client.CoreV1().Secrets(mongodb.Namespace).Create(secret); err != nil {
			return nil, err
		}
	}

	return &core.SecretVolumeSource{
		SecretName: authSecretName,
	}, nil
}

func addDataVolume(statefulSet *apps.StatefulSet, pvcSpec *core.PersistentVolumeClaimSpec) {
	if pvcSpec != nil {
		if len(pvcSpec.AccessModes) == 0 {
			pvcSpec.AccessModes = []core.PersistentVolumeAccessMode{
				core.ReadWriteOnce,
			}
			log.Infof(`Using "%v" as AccessModes in "%v"`, core.ReadWriteOnce, *pvcSpec)
		}
		// volume claim templates
		// Dynamically attach volume
		statefulSet.Spec.VolumeClaimTemplates = []core.PersistentVolumeClaim{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "data",
					Annotations: map[string]string{
						"volume.beta.kubernetes.io/storage-class": *pvcSpec.StorageClassName,
					},
				},
				Spec: *pvcSpec,
			},
		}
	} else {
		// Attach Empty directory
		statefulSet.Spec.Template.Spec.Volumes = append(
			statefulSet.Spec.Template.Spec.Volumes,
			core.Volume{
				Name: "data",
				VolumeSource: core.VolumeSource{
					EmptyDir: &core.EmptyDirVolumeSource{},
				},
			},
		)
	}
}

func addInitialScript(statefulSet *apps.StatefulSet, script *api.ScriptSourceSpec) {
	statefulSet.Spec.Template.Spec.Containers[0].VolumeMounts = append(statefulSet.Spec.Template.Spec.Containers[0].VolumeMounts,
		core.VolumeMount{
			Name:      "initial-script",
			MountPath: "/docker-entrypoint-initdb.d",
		},
	)

	statefulSet.Spec.Template.Spec.Volumes = append(statefulSet.Spec.Template.Spec.Volumes,
		core.Volume{
			Name:         "initial-script",
			VolumeSource: script.VolumeSource,
		},
	)
}

func (c *Controller) createDormantDatabase(mongodb *api.MongoDB) (*api.DormantDatabase, error) {
	dormantDb := &api.DormantDatabase{
		ObjectMeta: metav1.ObjectMeta{
			Name:      mongodb.Name,
			Namespace: mongodb.Namespace,
			Labels: map[string]string{
				api.LabelDatabaseKind: api.ResourceKindMongoDB,
			},
		},
		Spec: api.DormantDatabaseSpec{
			Origin: api.Origin{
				ObjectMeta: metav1.ObjectMeta{
					Name:        mongodb.Name,
					Namespace:   mongodb.Namespace,
					Labels:      mongodb.Labels,
					Annotations: mongodb.Annotations,
				},
				Spec: api.OriginSpec{
					MongoDB: &mongodb.Spec,
				},
			},
		},
	}

	initSpec, _ := json.Marshal(mongodb.Spec.Init)
	if mongodb.Spec.Init != nil {
		dormantDb.Annotations = map[string]string{
			api.MongoDBInitSpec: string(initSpec),
		}
	}

	dormantDb.Spec.Origin.Spec.MongoDB.Init = nil

	return c.ExtClient.DormantDatabases(dormantDb.Namespace).Create(dormantDb)
}

func (c *Controller) reCreateMongoDB(mongodb *api.MongoDB) error {
	_mongodb := &api.MongoDB{
		ObjectMeta: metav1.ObjectMeta{
			Name:        mongodb.Name,
			Namespace:   mongodb.Namespace,
			Labels:      mongodb.Labels,
			Annotations: mongodb.Annotations,
		},
		Spec:   mongodb.Spec,
		Status: mongodb.Status,
	}

	if _, err := c.ExtClient.MongoDBs(_mongodb.Namespace).Create(_mongodb); err != nil {
		return err
	}

	return nil
}

const (
	SnapshotProcess_Restore  = "restore"
	snapshotType_DumpRestore = "dump-restore"
)

func (c *Controller) createRestoreJob(mongodb *api.MongoDB, snapshot *api.Snapshot) (*batch.Job, error) {
	databaseName := mongodb.Name
	jobName := snapshot.OffshootName()
	jobLabel := map[string]string{
		api.LabelDatabaseName: databaseName,
		api.LabelJobType:      SnapshotProcess_Restore,
	}
	backupSpec := snapshot.Spec.SnapshotStorageSpec
	bucket, err := backupSpec.Container()
	if err != nil {
		return nil, err
	}

	// Get PersistentVolume object for Backup Util pod.
	persistentVolume, err := c.getVolumeForSnapshot(mongodb.Spec.Storage, jobName, mongodb.Namespace)
	if err != nil {
		return nil, err
	}

	// Folder name inside Cloud bucket where backup will be uploaded
	folderName, _ := snapshot.Location()

	job := &batch.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:   jobName,
			Labels: jobLabel,
		},
		Spec: batch.JobSpec{
			Template: core.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: jobLabel,
				},
				Spec: core.PodSpec{
					Containers: []core.Container{
						{
							Name: SnapshotProcess_Restore,
							//Image: fmt.Sprintf("%s:%s-util", docker.ImageMongoDB, mongodb.Spec.Version), //todo
							Image: fmt.Sprintf("kubedb/mongodb:3.4-util"), //todo
							Args: []string{
								fmt.Sprintf(`--process=%s`, SnapshotProcess_Restore),
								fmt.Sprintf(`--host=%s`, databaseName),
								fmt.Sprintf(`--bucket=%s`, bucket),
								fmt.Sprintf(`--folder=%s`, folderName),
								fmt.Sprintf(`--snapshot=%s`, snapshot.Name),
							},
							Resources: snapshot.Spec.Resources,
							VolumeMounts: []core.VolumeMount{
								{
									Name:      "secret",
									MountPath: "/srv/" + api.ResourceNameMongoDB + "/secrets",
								},
								{
									Name:      persistentVolume.Name,
									MountPath: "/var/" + snapshotType_DumpRestore + "/",
								},
								{
									Name:      "osmconfig",
									MountPath: storage.SecretMountPath,
									ReadOnly:  true,
								},
							},
						},
					},
					Volumes: []core.Volume{
						{
							Name: "secret",
							VolumeSource: core.VolumeSource{
								Secret: &core.SecretVolumeSource{
									SecretName: mongodb.Spec.DatabaseSecret.SecretName,
								},
							},
						},
						{
							Name:         persistentVolume.Name,
							VolumeSource: persistentVolume.VolumeSource,
						},
						{
							Name: "osmconfig",
							VolumeSource: core.VolumeSource{
								Secret: &core.SecretVolumeSource{
									SecretName: snapshot.OSMSecretName(),
								},
							},
						},
					},
					RestartPolicy: core.RestartPolicyNever,
				},
			},
		},
	}
	if snapshot.Spec.SnapshotStorageSpec.Local != nil {
		job.Spec.Template.Spec.Containers[0].VolumeMounts = append(job.Spec.Template.Spec.Containers[0].VolumeMounts, core.VolumeMount{
			Name:      "local",
			MountPath: snapshot.Spec.SnapshotStorageSpec.Local.Path,
		})
		volume := core.Volume{
			Name:         "local",
			VolumeSource: snapshot.Spec.SnapshotStorageSpec.Local.VolumeSource,
		}
		job.Spec.Template.Spec.Volumes = append(job.Spec.Template.Spec.Volumes, volume)
	}
	return c.Client.BatchV1().Jobs(mongodb.Namespace).Create(job)
}
