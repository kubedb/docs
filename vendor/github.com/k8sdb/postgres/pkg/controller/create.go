package controller

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/appscode/go/crypto/rand"
	"github.com/appscode/go/log"
	"github.com/appscode/go/types"
	tapi "github.com/k8sdb/apimachinery/apis/kubedb/v1alpha1"
	kutildb "github.com/k8sdb/apimachinery/client/typed/kubedb/v1alpha1/util"
	"github.com/k8sdb/apimachinery/pkg/docker"
	"github.com/k8sdb/apimachinery/pkg/eventer"
	"github.com/k8sdb/apimachinery/pkg/storage"
	apps "k8s.io/api/apps/v1beta1"
	batch "k8s.io/api/batch/v1"
	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	modeBasic = "basic"
	// Duration in Minute
	// Check whether pod under StatefulSet is running or not
	// Continue checking for this duration until failure
	durationCheckStatefulSet = time.Minute * 30
)

func (c *Controller) findService(postgres *tapi.Postgres) (bool, error) {
	name := postgres.OffshootName()
	service, err := c.Client.CoreV1().Services(postgres.Namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		if kerr.IsNotFound(err) {
			return false, nil
		} else {
			return false, err
		}
	}

	if service.Spec.Selector[tapi.LabelDatabaseName] != name {
		return false, fmt.Errorf(`intended service "%v" already exists`, name)
	}

	return true, nil
}

func (c *Controller) createService(postgres *tapi.Postgres) error {
	svc := &core.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:   postgres.OffshootName(),
			Labels: postgres.OffshootLabels(),
		},
		Spec: core.ServiceSpec{
			Ports: []core.ServicePort{
				{
					Name:       "db",
					Port:       5432,
					TargetPort: intstr.FromString("db"),
				},
			},
			Selector: postgres.OffshootLabels(),
		},
	}
	if postgres.Spec.Monitor != nil &&
		postgres.Spec.Monitor.Agent == tapi.AgentCoreosPrometheus &&
		postgres.Spec.Monitor.Prometheus != nil {
		svc.Spec.Ports = append(svc.Spec.Ports, core.ServicePort{
			Name:       tapi.PrometheusExporterPortName,
			Port:       tapi.PrometheusExporterPortNumber,
			TargetPort: intstr.FromString(tapi.PrometheusExporterPortName),
		})
	}

	if _, err := c.Client.CoreV1().Services(postgres.Namespace).Create(svc); err != nil {
		return err
	}

	return nil
}

func (c *Controller) findStatefulSet(postgres *tapi.Postgres) (bool, error) {
	// SatatefulSet for Postgres database
	statefulSet, err := c.Client.AppsV1beta1().StatefulSets(postgres.Namespace).Get(postgres.OffshootName(), metav1.GetOptions{})
	if err != nil {
		if kerr.IsNotFound(err) {
			return false, nil
		} else {
			return false, err
		}
	}

	if statefulSet.Labels[tapi.LabelDatabaseKind] != tapi.ResourceKindPostgres {
		return false, fmt.Errorf(`intended statefulSet "%v" already exists`, postgres.OffshootName())
	}

	return true, nil
}

func (c *Controller) createStatefulSet(postgres *tapi.Postgres) (*apps.StatefulSet, error) {
	// SatatefulSet for Postgres database
	statefulSet := &apps.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:        postgres.OffshootName(),
			Namespace:   postgres.Namespace,
			Labels:      postgres.StatefulSetLabels(),
			Annotations: postgres.StatefulSetAnnotations(),
		},
		Spec: apps.StatefulSetSpec{
			Replicas:    types.Int32P(1),
			ServiceName: c.opt.GoverningService,
			Template: core.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: postgres.OffshootLabels(),
				},
				Spec: core.PodSpec{
					Containers: []core.Container{
						{
							Name:            tapi.ResourceNamePostgres,
							Image:           fmt.Sprintf("%s:%s-db", docker.ImagePostgres, postgres.Spec.Version),
							ImagePullPolicy: core.PullIfNotPresent,
							Ports: []core.ContainerPort{
								{
									Name:          "db",
									ContainerPort: 5432,
								},
							},
							Resources: postgres.Spec.Resources,
							VolumeMounts: []core.VolumeMount{
								{
									Name:      "secret",
									MountPath: "/srv/" + tapi.ResourceNamePostgres + "/secrets",
								},
								{
									Name:      "data",
									MountPath: "/var/pv",
								},
							},
							Args: []string{modeBasic},
						},
					},
					NodeSelector:  postgres.Spec.NodeSelector,
					Affinity:      postgres.Spec.Affinity,
					SchedulerName: postgres.Spec.SchedulerName,
					Tolerations:   postgres.Spec.Tolerations,
				},
			},
		},
	}

	if postgres.Spec.Monitor != nil &&
		postgres.Spec.Monitor.Agent == tapi.AgentCoreosPrometheus &&
		postgres.Spec.Monitor.Prometheus != nil {
		exporter := core.Container{
			Name: "exporter",
			Args: []string{
				"export",
				fmt.Sprintf("--address=:%d", tapi.PrometheusExporterPortNumber),
				"--v=3",
			},
			Image:           docker.ImageOperator + ":" + c.opt.ExporterTag,
			ImagePullPolicy: core.PullIfNotPresent,
			Ports: []core.ContainerPort{
				{
					Name:          tapi.PrometheusExporterPortName,
					Protocol:      core.ProtocolTCP,
					ContainerPort: int32(tapi.PrometheusExporterPortNumber),
				},
			},
		}
		statefulSet.Spec.Template.Spec.Containers = append(statefulSet.Spec.Template.Spec.Containers, exporter)
	}

	if postgres.Spec.DatabaseSecret == nil {
		secretVolumeSource, err := c.createDatabaseSecret(postgres)
		if err != nil {
			return nil, err
		}

		_postgres, err := kutildb.TryPatchPostgres(c.ExtClient, postgres.ObjectMeta, func(in *tapi.Postgres) *tapi.Postgres {
			in.Spec.DatabaseSecret = secretVolumeSource
			return in
		})
		if err != nil {
			c.recorder.Eventf(postgres.ObjectReference(), core.EventTypeWarning, eventer.EventReasonFailedToUpdate, err.Error())
			return nil, err
		}
		postgres = _postgres
	}

	// Add secretVolume for authentication
	addSecretVolume(statefulSet, postgres.Spec.DatabaseSecret)

	// Add Data volume for StatefulSet
	addDataVolume(statefulSet, postgres.Spec.Storage)

	// Add InitialScript to run at startup
	if postgres.Spec.Init != nil && postgres.Spec.Init.ScriptSource != nil {
		addInitialScript(statefulSet, postgres.Spec.Init.ScriptSource)
	}

	if c.opt.EnableRbac {
		// Ensure ClusterRoles for database statefulsets
		if err := c.createRBACStuff(postgres); err != nil {
			return nil, err
		}

		statefulSet.Spec.Template.Spec.ServiceAccountName = postgres.Name
	}

	if _, err := c.Client.AppsV1beta1().StatefulSets(statefulSet.Namespace).Create(statefulSet); err != nil {
		return nil, err
	}

	return statefulSet, nil
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

func (c *Controller) createDatabaseSecret(postgres *tapi.Postgres) (*core.SecretVolumeSource, error) {
	authSecretName := postgres.Name + "-admin-auth"

	found, err := c.findSecret(authSecretName, postgres.Namespace)
	if err != nil {
		return nil, err
	}

	if !found {
		POSTGRES_PASSWORD := fmt.Sprintf("POSTGRES_PASSWORD=%s\n", rand.GeneratePassword())
		data := map[string][]byte{
			".admin": []byte(POSTGRES_PASSWORD),
		}
		secret := &core.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name: authSecretName,
				Labels: map[string]string{
					tapi.LabelDatabaseKind: tapi.ResourceKindPostgres,
				},
			},
			Type: core.SecretTypeOpaque,
			Data: data,
		}
		if _, err := c.Client.CoreV1().Secrets(postgres.Namespace).Create(secret); err != nil {
			return nil, err
		}
	}

	return &core.SecretVolumeSource{
		SecretName: authSecretName,
	}, nil
}

func addSecretVolume(statefulSet *apps.StatefulSet, secretVolume *core.SecretVolumeSource) error {
	statefulSet.Spec.Template.Spec.Volumes = append(statefulSet.Spec.Template.Spec.Volumes,
		core.Volume{
			Name: "secret",
			VolumeSource: core.VolumeSource{
				Secret: secretVolume,
			},
		},
	)
	return nil
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

func addInitialScript(statefulSet *apps.StatefulSet, script *tapi.ScriptSourceSpec) {
	statefulSet.Spec.Template.Spec.Containers[0].VolumeMounts = append(statefulSet.Spec.Template.Spec.Containers[0].VolumeMounts,
		core.VolumeMount{
			Name:      "initial-script",
			MountPath: "/var/db-script",
		},
	)
	statefulSet.Spec.Template.Spec.Containers[0].Args = []string{
		modeBasic,
		script.ScriptPath,
	}

	statefulSet.Spec.Template.Spec.Volumes = append(statefulSet.Spec.Template.Spec.Volumes,
		core.Volume{
			Name:         "initial-script",
			VolumeSource: script.VolumeSource,
		},
	)
}

func (c *Controller) createDormantDatabase(postgres *tapi.Postgres) (*tapi.DormantDatabase, error) {
	dormantDb := &tapi.DormantDatabase{
		ObjectMeta: metav1.ObjectMeta{
			Name:      postgres.Name,
			Namespace: postgres.Namespace,
			Labels: map[string]string{
				tapi.LabelDatabaseKind: tapi.ResourceKindPostgres,
			},
		},
		Spec: tapi.DormantDatabaseSpec{
			Origin: tapi.Origin{
				ObjectMeta: metav1.ObjectMeta{
					Name:        postgres.Name,
					Namespace:   postgres.Namespace,
					Labels:      postgres.Labels,
					Annotations: postgres.Annotations,
				},
				Spec: tapi.OriginSpec{
					Postgres: &postgres.Spec,
				},
			},
		},
	}

	initSpec, _ := json.Marshal(postgres.Spec.Init)
	if initSpec != nil {
		dormantDb.Annotations = map[string]string{
			tapi.PostgresInitSpec: string(initSpec),
		}
	}
	dormantDb.Spec.Origin.Spec.Postgres.Init = nil

	return c.ExtClient.DormantDatabases(dormantDb.Namespace).Create(dormantDb)
}

func (c *Controller) reCreatePostgres(postgres *tapi.Postgres) error {
	_postgres := &tapi.Postgres{
		ObjectMeta: metav1.ObjectMeta{
			Name:        postgres.Name,
			Namespace:   postgres.Namespace,
			Labels:      postgres.Labels,
			Annotations: postgres.Annotations,
		},
		Spec:   postgres.Spec,
		Status: postgres.Status,
	}

	if _, err := c.ExtClient.Postgreses(_postgres.Namespace).Create(_postgres); err != nil {
		return err
	}

	return nil
}

const (
	SnapshotProcess_Restore  = "restore"
	snapshotType_DumpRestore = "dump-restore"
)

func (c *Controller) createRestoreJob(postgres *tapi.Postgres, snapshot *tapi.Snapshot) (*batch.Job, error) {
	databaseName := postgres.Name
	jobName := snapshot.OffshootName()
	jobLabel := map[string]string{
		tapi.LabelDatabaseName: databaseName,
		tapi.LabelJobType:      SnapshotProcess_Restore,
	}
	backupSpec := snapshot.Spec.SnapshotStorageSpec
	bucket, err := backupSpec.Container()
	if err != nil {
		return nil, err
	}

	// Get PersistentVolume object for Backup Util pod.
	persistentVolume, err := c.getVolumeForSnapshot(postgres.Spec.Storage, jobName, postgres.Namespace)
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
							Name:  SnapshotProcess_Restore,
							Image: fmt.Sprintf("%s:%s-util", docker.ImagePostgres, postgres.Spec.Version),
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
									MountPath: "/srv/" + tapi.ResourceNamePostgres + "/secrets",
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
									SecretName: postgres.Spec.DatabaseSecret.SecretName,
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
									SecretName: snapshot.Name,
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
	return c.Client.BatchV1().Jobs(postgres.Namespace).Create(job)
}
