package controller

import (
	"fmt"
	"time"

	"github.com/appscode/go/crypto/rand"
	tapi "github.com/k8sdb/apimachinery/api"
	amc "github.com/k8sdb/apimachinery/pkg/controller"
	"github.com/k8sdb/apimachinery/pkg/docker"
	kapi "k8s.io/kubernetes/pkg/api"
	k8serr "k8s.io/kubernetes/pkg/api/errors"
	kapps "k8s.io/kubernetes/pkg/apis/apps"
	kbatch "k8s.io/kubernetes/pkg/apis/batch"
	"k8s.io/kubernetes/pkg/util/intstr"
)

const (
	annotationDatabaseVersion = "postgres.kubedb.com/version"
	modeBasic                 = "basic"
	// Duration in Minute
	// Check whether pod under StatefulSet is running or not
	// Continue checking for this duration until failure
	durationCheckStatefulSet = time.Minute * 30
)

func (c *Controller) checkService(name, namespace string) (bool, error) {
	service, err := c.Client.Core().Services(namespace).Get(name)
	if err != nil {
		if k8serr.IsNotFound(err) {
			return false, nil
		} else {
			return false, err
		}
	}

	if service.Spec.Selector[amc.LabelDatabaseName] != name {
		return false, fmt.Errorf(`Intended service "%v" already exists`, name)
	}

	return true, nil
}

func (c *Controller) createService(name, namespace string) error {
	// Check if service name exists
	found, err := c.checkService(name, namespace)
	if err != nil {
		return err
	}
	if found {
		return nil
	}

	label := map[string]string{
		amc.LabelDatabaseName: name,
	}
	service := &kapi.Service{
		ObjectMeta: kapi.ObjectMeta{
			Name:   name,
			Labels: label,
		},
		Spec: kapi.ServiceSpec{
			Ports: []kapi.ServicePort{
				{
					Name:       "port",
					Port:       5432,
					TargetPort: intstr.FromString("port"),
				},
			},
			Selector: label,
		},
	}

	if _, err := c.Client.Core().Services(namespace).Create(service); err != nil {
		return err
	}

	return nil
}

func (c *Controller) checkStatefulSet(postgres *tapi.Postgres) (*kapps.StatefulSet, error) {
	// SatatefulSet for Postgres database
	statefulSetName := getStatefulSetName(postgres.Name)
	statefulSet, err := c.Client.Apps().StatefulSets(postgres.Namespace).Get(statefulSetName)
	if err != nil {
		if k8serr.IsNotFound(err) {
			return nil, nil
		} else {
			return nil, err
		}
	}

	if statefulSet.Labels[amc.LabelDatabaseKind] != tapi.ResourceKindPostgres {
		return nil, fmt.Errorf(`Intended statefulSet "%v" already exists`, statefulSetName)
	}

	return statefulSet, nil
}

func (c *Controller) createStatefulSet(postgres *tapi.Postgres) (*kapps.StatefulSet, error) {
	_statefulSet, err := c.checkStatefulSet(postgres)
	if err != nil {
		return nil, err
	}
	if _statefulSet != nil {
		return _statefulSet, nil
	}

	// Set labels
	labels := make(map[string]string)
	for key, val := range postgres.Labels {
		labels[key] = val
	}
	labels[amc.LabelDatabaseKind] = tapi.ResourceKindPostgres

	// Set Annotations
	annotations := make(map[string]string)
	for key, val := range postgres.Annotations {
		annotations[key] = val
	}
	annotations[annotationDatabaseVersion] = postgres.Spec.Version

	podLabels := make(map[string]string)
	for key, val := range labels {
		podLabels[key] = val
	}
	podLabels[amc.LabelDatabaseName] = postgres.Name

	// SatatefulSet for Postgres database
	statefulSetName := getStatefulSetName(postgres.Name)

	replicas := int32(1)
	statefulSet := &kapps.StatefulSet{
		ObjectMeta: kapi.ObjectMeta{
			Name:        statefulSetName,
			Namespace:   postgres.Namespace,
			Labels:      labels,
			Annotations: annotations,
		},
		Spec: kapps.StatefulSetSpec{
			Replicas:    replicas,
			ServiceName: c.opt.GoverningService,
			Template: kapi.PodTemplateSpec{
				ObjectMeta: kapi.ObjectMeta{
					Labels:      podLabels,
					Annotations: annotations,
				},
				Spec: kapi.PodSpec{
					Containers: []kapi.Container{
						{
							Name:            tapi.ResourceNamePostgres,
							Image:           fmt.Sprintf("%s:%s-db", docker.ImagePostgres, postgres.Spec.Version),
							ImagePullPolicy: kapi.PullIfNotPresent,
							Ports: []kapi.ContainerPort{
								{
									Name:          "port",
									ContainerPort: 5432,
								},
							},
							VolumeMounts: []kapi.VolumeMount{
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
					NodeSelector: postgres.Spec.NodeSelector,
				},
			},
		},
	}

	if postgres.Spec.DatabaseSecret == nil {
		secretVolumeSource, err := c.createDatabaseSecret(postgres)
		if err != nil {
			return nil, err
		}
		if postgres, err = c.ExtClient.Postgreses(postgres.Namespace).Get(postgres.Name); err != nil {
			return nil, err
		}

		postgres.Spec.DatabaseSecret = secretVolumeSource

		if _, err := c.ExtClient.Postgreses(postgres.Namespace).Update(postgres); err != nil {
			return nil, err
		}
	}

	// Add secretVolume for authentication
	addSecretVolume(statefulSet, postgres.Spec.DatabaseSecret)

	// Add Data volume for StatefulSet
	addDataVolume(statefulSet, postgres.Spec.Storage)

	// Add InitialScript to run at startup
	if postgres.Spec.Init != nil && postgres.Spec.Init.ScriptSource != nil {
		addInitialScript(statefulSet, postgres.Spec.Init.ScriptSource)
	}

	if _, err := c.Client.Apps().StatefulSets(statefulSet.Namespace).Create(statefulSet); err != nil {
		return nil, err
	}

	return statefulSet, nil
}

func (c *Controller) checkSecret(namespace, secretName string) (bool, error) {
	secret, err := c.Client.Core().Secrets(namespace).Get(secretName)
	if err != nil {
		if k8serr.IsNotFound(err) {
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

func (c *Controller) createDatabaseSecret(postgres *tapi.Postgres) (*kapi.SecretVolumeSource, error) {
	authSecretName := postgres.Name + "-admin-auth"

	found, err := c.checkSecret(postgres.Namespace, authSecretName)
	if err != nil {
		return nil, err
	}

	if !found {
		POSTGRES_PASSWORD := fmt.Sprintf("POSTGRES_PASSWORD=%s\n", rand.GeneratePassword())
		data := map[string][]byte{
			".admin": []byte(POSTGRES_PASSWORD),
		}
		secret := &kapi.Secret{
			ObjectMeta: kapi.ObjectMeta{
				Name: authSecretName,
				Labels: map[string]string{
					amc.LabelDatabaseKind: tapi.ResourceKindPostgres,
				},
			},
			Type: kapi.SecretTypeOpaque,
			Data: data,
		}
		if _, err := c.Client.Core().Secrets(postgres.Namespace).Create(secret); err != nil {
			return nil, err
		}
	}

	return &kapi.SecretVolumeSource{
		SecretName: authSecretName,
	}, nil
}

func addSecretVolume(statefulSet *kapps.StatefulSet, secretVolume *kapi.SecretVolumeSource) error {
	statefulSet.Spec.Template.Spec.Volumes = append(statefulSet.Spec.Template.Spec.Volumes,
		kapi.Volume{
			Name: "secret",
			VolumeSource: kapi.VolumeSource{
				Secret: secretVolume,
			},
		},
	)
	return nil
}

func addDataVolume(statefulSet *kapps.StatefulSet, storage *tapi.StorageSpec) {
	if storage != nil {
		// volume claim templates
		// Dynamically attach volume
		storageClassName := storage.Class
		statefulSet.Spec.VolumeClaimTemplates = []kapi.PersistentVolumeClaim{
			{
				ObjectMeta: kapi.ObjectMeta{
					Name: "data",
					Annotations: map[string]string{
						"volume.beta.kubernetes.io/storage-class": storageClassName,
					},
				},
				Spec: storage.PersistentVolumeClaimSpec,
			},
		}
	} else {
		// Attach Empty directory
		statefulSet.Spec.Template.Spec.Volumes = append(
			statefulSet.Spec.Template.Spec.Volumes,
			kapi.Volume{
				Name: "data",
				VolumeSource: kapi.VolumeSource{
					EmptyDir: &kapi.EmptyDirVolumeSource{},
				},
			},
		)
	}
}

func addInitialScript(statefulSet *kapps.StatefulSet, script *tapi.ScriptSourceSpec) {
	statefulSet.Spec.Template.Spec.Containers[0].VolumeMounts = append(statefulSet.Spec.Template.Spec.Containers[0].VolumeMounts,
		kapi.VolumeMount{
			Name:      "initial-script",
			MountPath: "/var/db-script",
		},
	)
	statefulSet.Spec.Template.Spec.Containers[0].Args = []string{
		modeBasic,
		script.ScriptPath,
	}

	statefulSet.Spec.Template.Spec.Volumes = append(statefulSet.Spec.Template.Spec.Volumes,
		kapi.Volume{
			Name:         "initial-script",
			VolumeSource: script.VolumeSource,
		},
	)
}

func (c *Controller) createDormantDatabase(postgres *tapi.Postgres) (*tapi.DormantDatabase, error) {
	dormantDb := &tapi.DormantDatabase{
		ObjectMeta: kapi.ObjectMeta{
			Name:      postgres.Name,
			Namespace: postgres.Namespace,
			Labels: map[string]string{
				amc.LabelDatabaseKind: tapi.ResourceKindPostgres,
			},
		},
		Spec: tapi.DormantDatabaseSpec{
			Origin: tapi.Origin{
				ObjectMeta: kapi.ObjectMeta{
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
	return c.ExtClient.DormantDatabases(dormantDb.Namespace).Create(dormantDb)
}

func (c *Controller) reCreatePostgres(postgres *tapi.Postgres) error {
	_postgres := &tapi.Postgres{
		ObjectMeta: kapi.ObjectMeta{
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

func (c *Controller) createRestoreJob(postgres *tapi.Postgres, snapshot *tapi.Snapshot) (*kbatch.Job, error) {

	databaseName := postgres.Name
	jobName := rand.WithUniqSuffix(databaseName)
	jobLabel := map[string]string{
		amc.LabelDatabaseName: databaseName,
		amc.LabelJobType:      SnapshotProcess_Restore,
	}
	backupSpec := snapshot.Spec.SnapshotStorageSpec

	// Get PersistentVolume object for Backup Util pod.
	persistentVolume, err := c.getVolumeForSnapshot(postgres.Spec.Storage, jobName, postgres.Namespace)
	if err != nil {
		return nil, err
	}

	// Folder name inside Cloud bucket where backup will be uploaded
	folderName := fmt.Sprintf("%v/%v/%v", amc.DatabaseNamePrefix, snapshot.Namespace, snapshot.Spec.DatabaseName)

	job := &kbatch.Job{
		ObjectMeta: kapi.ObjectMeta{
			Name:   jobName,
			Labels: jobLabel,
		},
		Spec: kbatch.JobSpec{
			Template: kapi.PodTemplateSpec{
				ObjectMeta: kapi.ObjectMeta{
					Labels: jobLabel,
				},
				Spec: kapi.PodSpec{
					Containers: []kapi.Container{
						{
							Name:  SnapshotProcess_Restore,
							Image: fmt.Sprintf("%s:%s-util", docker.ImagePostgres, postgres.Spec.Version),
							Args: []string{
								fmt.Sprintf(`--process=%s`, SnapshotProcess_Restore),
								fmt.Sprintf(`--host=%s`, databaseName),
								fmt.Sprintf(`--bucket=%s`, backupSpec.BucketName),
								fmt.Sprintf(`--folder=%s`, folderName),
								fmt.Sprintf(`--snapshot=%s`, snapshot.Name),
							},
							VolumeMounts: []kapi.VolumeMount{
								{
									Name:      "secret",
									MountPath: "/srv/" + tapi.ResourceNamePostgres + "/secrets",
								},
								{
									Name:      "cloud",
									MountPath: storageSecretMountPath,
								},
								{
									Name:      persistentVolume.Name,
									MountPath: "/var/" + snapshotType_DumpRestore + "/",
								},
							},
						},
					},
					Volumes: []kapi.Volume{
						{
							Name: "secret",
							VolumeSource: kapi.VolumeSource{
								Secret: &kapi.SecretVolumeSource{
									SecretName: postgres.Spec.DatabaseSecret.SecretName,
								},
							},
						},
						{
							Name: "cloud",
							VolumeSource: kapi.VolumeSource{
								Secret: backupSpec.StorageSecret,
							},
						},
						{
							Name:         persistentVolume.Name,
							VolumeSource: persistentVolume.VolumeSource,
						},
					},
					RestartPolicy: kapi.RestartPolicyNever,
				},
			},
		},
	}

	return c.Client.Batch().Jobs(postgres.Namespace).Create(job)
}

func getStatefulSetName(databaseName string) string {
	return fmt.Sprintf("%v-%v", databaseName, tapi.ResourceCodePostgres)
}
