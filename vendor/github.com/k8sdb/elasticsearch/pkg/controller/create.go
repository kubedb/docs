package controller

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/appscode/go/log"
	api "github.com/k8sdb/apimachinery/apis/kubedb/v1alpha1"
	"github.com/k8sdb/apimachinery/pkg/docker"
	"github.com/k8sdb/apimachinery/pkg/storage"
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

func (c *Controller) findService(elastic *api.Elasticsearch) (bool, error) {
	name := elastic.OffshootName()
	service, err := c.Client.CoreV1().Services(elastic.Namespace).Get(name, metav1.GetOptions{})
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

func (c *Controller) createService(elastic *api.Elasticsearch) error {
	svc := &core.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:   elastic.Name,
			Labels: elastic.OffshootLabels(),
		},
		Spec: core.ServiceSpec{
			Ports: []core.ServicePort{
				{
					Name:       "db",
					Port:       9200,
					TargetPort: intstr.FromString("db"),
				},
				{
					Name:       "cluster",
					Port:       9300,
					TargetPort: intstr.FromString("cluster"),
				},
			},
			Selector: elastic.OffshootLabels(),
		},
	}
	if elastic.Spec.Monitor != nil &&
		elastic.Spec.Monitor.Agent == api.AgentCoreosPrometheus &&
		elastic.Spec.Monitor.Prometheus != nil {
		svc.Spec.Ports = append(svc.Spec.Ports, core.ServicePort{
			Name:       api.PrometheusExporterPortName,
			Port:       api.PrometheusExporterPortNumber,
			TargetPort: intstr.FromString(api.PrometheusExporterPortName),
		})
	}

	if _, err := c.Client.CoreV1().Services(elastic.Namespace).Create(svc); err != nil {
		return err
	}

	return nil
}

func (c *Controller) findStatefulSet(elastic *api.Elasticsearch) (bool, error) {
	// SatatefulSet for Elasticsearch database
	statefulSet, err := c.Client.AppsV1beta1().StatefulSets(elastic.Namespace).Get(elastic.OffshootName(), metav1.GetOptions{})
	if err != nil {
		if kerr.IsNotFound(err) {
			return false, nil
		} else {
			return false, err
		}
	}

	if statefulSet.Labels[api.LabelDatabaseKind] != api.ResourceKindElasticsearch {
		return false, fmt.Errorf(`Intended statefulSet "%v" already exists`, elastic.OffshootName())
	}

	return true, nil
}

func (c *Controller) createStatefulSet(elastic *api.Elasticsearch) (*apps.StatefulSet, error) {
	dockerImage := fmt.Sprintf("%v:%v", docker.ImageElasticsearch, elastic.Spec.Version)
	initContainerImage := fmt.Sprintf("%v:%v", docker.ImageElasticOperator, c.opt.DiscoveryTag)

	// SatatefulSet for Elasticsearch database
	statefulSet := &apps.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:        elastic.OffshootName(),
			Namespace:   elastic.Namespace,
			Labels:      elastic.StatefulSetLabels(),
			Annotations: elastic.StatefulSetAnnotations(),
		},
		Spec: apps.StatefulSetSpec{
			Replicas:    &elastic.Spec.Replicas,
			ServiceName: c.opt.GoverningService,
			Template: core.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: elastic.OffshootLabels(),
				},
				Spec: core.PodSpec{
					Containers: []core.Container{
						{
							Name:            api.ResourceNameElasticsearch,
							Image:           dockerImage,
							ImagePullPolicy: core.PullIfNotPresent,
							Ports: []core.ContainerPort{
								{
									Name:          "db",
									ContainerPort: 9200,
								},
								{
									Name:          "cluster",
									ContainerPort: 9300,
								},
							},
							Resources: elastic.Spec.Resources,
							VolumeMounts: []core.VolumeMount{
								{
									Name:      "discovery",
									MountPath: "/tmp/discovery",
								},
								{
									Name:      "data",
									MountPath: "/var/pv",
								},
							},
							Env: []core.EnvVar{
								{
									Name:  "CLUSTER_NAME",
									Value: elastic.Name,
								},
								{
									Name:  "KUBE_NAMESPACE",
									Value: elastic.Namespace,
								},
							},
						},
					},
					InitContainers: []core.Container{
						{
							Name:            "discover",
							Image:           initContainerImage,
							ImagePullPolicy: core.PullIfNotPresent,
							Args: []string{
								"discover",
								fmt.Sprintf("--service=%v", elastic.Name),
								fmt.Sprintf("--namespace=%v", elastic.Namespace),
							},
							Env: []core.EnvVar{
								{
									Name: "POD_NAME",
									ValueFrom: &core.EnvVarSource{
										FieldRef: &core.ObjectFieldSelector{
											APIVersion: "v1",
											FieldPath:  "metadata.name",
										},
									},
								},
							},
							VolumeMounts: []core.VolumeMount{
								{
									Name:      "discovery",
									MountPath: "/tmp/discovery",
								},
							},
						},
					},
					NodeSelector: elastic.Spec.NodeSelector,
					Volumes: []core.Volume{
						{
							Name: "discovery",
							VolumeSource: core.VolumeSource{
								EmptyDir: &core.EmptyDirVolumeSource{},
							},
						},
					},
					Affinity:      elastic.Spec.Affinity,
					SchedulerName: elastic.Spec.SchedulerName,
					Tolerations:   elastic.Spec.Tolerations,
				},
			},
		},
	}

	if elastic.Spec.Monitor != nil &&
		elastic.Spec.Monitor.Agent == api.AgentCoreosPrometheus &&
		elastic.Spec.Monitor.Prometheus != nil {
		exporter := core.Container{
			Name: "exporter",
			Args: []string{
				"export",
				fmt.Sprintf("--address=:%d", api.PrometheusExporterPortNumber),
				"--v=3",
			},
			Image:           docker.ImageOperator + ":" + c.opt.ExporterTag,
			ImagePullPolicy: core.PullIfNotPresent,
			Ports: []core.ContainerPort{
				{
					Name:          api.PrometheusExporterPortName,
					Protocol:      core.ProtocolTCP,
					ContainerPort: int32(api.PrometheusExporterPortNumber),
				},
			},
		}
		statefulSet.Spec.Template.Spec.Containers = append(statefulSet.Spec.Template.Spec.Containers, exporter)
	}

	// Add Data volume for StatefulSet
	addDataVolume(statefulSet, elastic.Spec.Storage)

	if c.opt.EnableRbac {
		// Ensure ClusterRoles for database statefulsets
		if err := c.createRBACStuff(elastic); err != nil {
			return nil, err
		}

		statefulSet.Spec.Template.Spec.ServiceAccountName = elastic.Name
	}

	if _, err := c.Client.AppsV1beta1().StatefulSets(statefulSet.Namespace).Create(statefulSet); err != nil {
		return nil, err
	}

	return statefulSet, nil
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

func (c *Controller) createDormantDatabase(elastic *api.Elasticsearch) (*api.DormantDatabase, error) {
	dormantDb := &api.DormantDatabase{
		ObjectMeta: metav1.ObjectMeta{
			Name:      elastic.Name,
			Namespace: elastic.Namespace,
			Labels: map[string]string{
				api.LabelDatabaseKind: api.ResourceKindElasticsearch,
			},
		},
		Spec: api.DormantDatabaseSpec{
			Origin: api.Origin{
				ObjectMeta: metav1.ObjectMeta{
					Name:        elastic.Name,
					Namespace:   elastic.Namespace,
					Labels:      elastic.Labels,
					Annotations: elastic.Annotations,
				},
				Spec: api.OriginSpec{
					Elasticsearch: &elastic.Spec,
				},
			},
		},
	}

	initSpec, _ := json.Marshal(elastic.Spec.Init)
	if initSpec != nil {
		dormantDb.Annotations = map[string]string{
			api.ElasticsearchInitSpec: string(initSpec),
		}
	}
	dormantDb.Spec.Origin.Spec.Elasticsearch.Init = nil

	return c.ExtClient.DormantDatabases(dormantDb.Namespace).Create(dormantDb)
}

func (c *Controller) reCreateElastic(elastic *api.Elasticsearch) error {
	_elastic := &api.Elasticsearch{
		ObjectMeta: metav1.ObjectMeta{
			Name:        elastic.Name,
			Namespace:   elastic.Namespace,
			Labels:      elastic.Labels,
			Annotations: elastic.Annotations,
		},
		Spec:   elastic.Spec,
		Status: elastic.Status,
	}

	if _, err := c.ExtClient.Elasticsearchs(_elastic.Namespace).Create(_elastic); err != nil {
		return err
	}

	return nil
}

const (
	SnapshotProcess_Restore  = "restore"
	snapshotType_DumpRestore = "dump-restore"
)

func (c *Controller) createRestoreJob(elastic *api.Elasticsearch, snapshot *api.Snapshot) (*batch.Job, error) {
	databaseName := elastic.Name
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
	persistentVolume, err := c.getVolumeForSnapshot(elastic.Spec.Storage, jobName, elastic.Namespace)
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
							Image: docker.ImageElasticdump + ":" + c.opt.ElasticDumpTag,
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
	return c.Client.BatchV1().Jobs(elastic.Namespace).Create(job)
}
