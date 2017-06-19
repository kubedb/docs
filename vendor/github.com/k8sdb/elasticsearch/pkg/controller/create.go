package controller

import (
	"fmt"
	"time"

	"github.com/appscode/go/crypto/rand"
	tapi "github.com/k8sdb/apimachinery/api"
	amc "github.com/k8sdb/apimachinery/pkg/controller"
	"github.com/k8sdb/apimachinery/pkg/docker"
	"github.com/k8sdb/apimachinery/pkg/monitor"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	apiv1 "k8s.io/client-go/pkg/api/v1"
	apps "k8s.io/client-go/pkg/apis/apps/v1beta1"
	batch "k8s.io/client-go/pkg/apis/batch/v1"
)

const (
	annotationDatabaseVersion = "elastic.kubedb.com/version"
	// Duration in Minute
	// Check whether pod under StatefulSet is running or not
	// Continue checking for this duration until failure
	durationCheckStatefulSet = time.Minute * 30
)

func (c *Controller) findService(name, namespace string) (bool, error) {
	service, err := c.Client.CoreV1().Services(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		if kerr.IsNotFound(err) {
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
	label := map[string]string{
		amc.LabelDatabaseName: name,
	}
	service := &apiv1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: label,
		},
		Spec: apiv1.ServiceSpec{
			Ports: []apiv1.ServicePort{
				{
					Name:       "api",
					Port:       9200,
					TargetPort: intstr.FromString("api"),
				},
				{
					Name:       "tcp",
					Port:       9300,
					TargetPort: intstr.FromString("tcp"),
				},
			},
			Selector: label,
		},
	}

	if _, err := c.Client.CoreV1().Services(namespace).Create(service); err != nil {
		return err
	}

	return nil
}

func (c *Controller) findStatefulSet(elastic *tapi.Elastic) (bool, error) {
	// SatatefulSet for Postgres database
	statefulSetName := getStatefulSetName(elastic.Name)
	statefulSet, err := c.Client.AppsV1beta1().StatefulSets(elastic.Namespace).Get(statefulSetName, metav1.GetOptions{})
	if err != nil {
		if kerr.IsNotFound(err) {
			return false, nil
		} else {
			return false, err
		}
	}

	if statefulSet.Labels[amc.LabelDatabaseKind] != tapi.ResourceKindElastic {
		return false, fmt.Errorf(`Intended statefulSet "%v" already exists`, statefulSetName)
	}

	return true, nil
}

func (c *Controller) createStatefulSet(elastic *tapi.Elastic) (*apps.StatefulSet, error) {
	// Set labels
	labels := make(map[string]string)
	for key, val := range elastic.Labels {
		labels[key] = val
	}
	labels[amc.LabelDatabaseKind] = tapi.ResourceKindElastic

	// Set Annotations
	annotations := make(map[string]string)
	for key, val := range elastic.Annotations {
		annotations[key] = val
	}
	annotations[annotationDatabaseVersion] = string(elastic.Spec.Version)

	podLabels := make(map[string]string)
	for key, val := range labels {
		podLabels[key] = val
	}
	podLabels[amc.LabelDatabaseName] = elastic.Name

	dockerImage := fmt.Sprintf("%v:%v", docker.ImageElasticsearch, elastic.Spec.Version)
	initContainerImage := fmt.Sprintf("%v:%v", docker.ImageElasticOperator, c.opt.DiscoveryTag)

	// SatatefulSet for Elastic database
	statefulSetName := getStatefulSetName(elastic.Name)
	statefulSet := &apps.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:        statefulSetName,
			Namespace:   elastic.Namespace,
			Labels:      labels,
			Annotations: annotations,
		},
		Spec: apps.StatefulSetSpec{
			Replicas:    &elastic.Spec.Replicas,
			ServiceName: c.opt.GoverningService,
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      podLabels,
					Annotations: annotations,
				},
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{
							Name:            tapi.ResourceNameElastic,
							Image:           dockerImage,
							ImagePullPolicy: apiv1.PullIfNotPresent,
							Ports: []apiv1.ContainerPort{
								{
									Name:          "api",
									ContainerPort: 9200,
								},
								{
									Name:          "tcp",
									ContainerPort: 9300,
								},
							},
							VolumeMounts: []apiv1.VolumeMount{
								{
									Name:      "discovery",
									MountPath: "/tmp/discovery",
								},
								{
									Name:      "data",
									MountPath: "/var/pv",
								},
							},
							Env: []apiv1.EnvVar{
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
					InitContainers: []apiv1.Container{
						{
							Name:            "discover",
							Image:           initContainerImage,
							ImagePullPolicy: apiv1.PullIfNotPresent,
							Args: []string{
								"discover",
								fmt.Sprintf("--service=%v", elastic.Name),
								fmt.Sprintf("--namespace=%v", elastic.Namespace),
							},
							Env: []apiv1.EnvVar{
								{
									Name: "POD_NAME",
									ValueFrom: &apiv1.EnvVarSource{
										FieldRef: &apiv1.ObjectFieldSelector{
											APIVersion: "v1",
											FieldPath:  "metadata.name",
										},
									},
								},
							},
							VolumeMounts: []apiv1.VolumeMount{
								{
									Name:      "discovery",
									MountPath: "/tmp/discovery",
								},
							},
						},
					},
					NodeSelector: elastic.Spec.NodeSelector,
					Volumes: []apiv1.Volume{
						{
							Name: "discovery",
							VolumeSource: apiv1.VolumeSource{
								EmptyDir: &apiv1.EmptyDirVolumeSource{},
							},
						},
					},
				},
			},
		},
	}

	if elastic.Spec.Monitor != nil &&
		elastic.Spec.Monitor.Agent == monitor.AgentCoreosPrometheus &&
		elastic.Spec.Monitor.Prometheus != nil {
		exporter := apiv1.Container{
			Name: "exporter",
			Args: []string{
				"exporter",
				fmt.Sprintf("--address=:%d", elastic.Spec.Monitor.Prometheus.TargetPort.IntVal),
				"--v=3",
			},
			Image:           docker.ImageOperator + ":" + c.opt.ExporterTag,
			ImagePullPolicy: apiv1.PullIfNotPresent,
			Ports: []apiv1.ContainerPort{
				{
					Name:          "http",
					Protocol:      apiv1.ProtocolTCP,
					ContainerPort: elastic.Spec.Monitor.Prometheus.TargetPort.IntVal,
				},
			},
		}
		statefulSet.Spec.Template.Spec.Containers = append(statefulSet.Spec.Template.Spec.Containers, exporter)
	}

	// Add Data volume for StatefulSet
	addDataVolume(statefulSet, elastic.Spec.Storage)

	if _, err := c.Client.AppsV1beta1().StatefulSets(statefulSet.Namespace).Create(statefulSet); err != nil {
		return nil, err
	}

	return statefulSet, nil
}

func addDataVolume(statefulSet *apps.StatefulSet, storage *tapi.StorageSpec) {
	if storage != nil {
		// volume claim templates
		// Dynamically attach volume
		storageClassName := storage.Class
		statefulSet.Spec.VolumeClaimTemplates = []apiv1.PersistentVolumeClaim{
			{
				ObjectMeta: metav1.ObjectMeta{
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
			apiv1.Volume{
				Name: "data",
				VolumeSource: apiv1.VolumeSource{
					EmptyDir: &apiv1.EmptyDirVolumeSource{},
				},
			},
		)
	}
}

func (c *Controller) createDormantDatabase(elastic *tapi.Elastic) (*tapi.DormantDatabase, error) {
	dormantDb := &tapi.DormantDatabase{
		ObjectMeta: metav1.ObjectMeta{
			Name:      elastic.Name,
			Namespace: elastic.Namespace,
			Labels: map[string]string{
				amc.LabelDatabaseKind: tapi.ResourceKindElastic,
			},
		},
		Spec: tapi.DormantDatabaseSpec{
			Origin: tapi.Origin{
				ObjectMeta: metav1.ObjectMeta{
					Name:        elastic.Name,
					Namespace:   elastic.Namespace,
					Labels:      elastic.Labels,
					Annotations: elastic.Annotations,
				},
				Spec: tapi.OriginSpec{
					Elastic: &elastic.Spec,
				},
			},
		},
	}
	return c.ExtClient.DormantDatabases(dormantDb.Namespace).Create(dormantDb)
}

func (c *Controller) reCreateElastic(elastic *tapi.Elastic) error {
	_elastic := &tapi.Elastic{
		ObjectMeta: metav1.ObjectMeta{
			Name:        elastic.Name,
			Namespace:   elastic.Namespace,
			Labels:      elastic.Labels,
			Annotations: elastic.Annotations,
		},
		Spec:   elastic.Spec,
		Status: elastic.Status,
	}

	if _, err := c.ExtClient.Elastics(_elastic.Namespace).Create(_elastic); err != nil {
		return err
	}

	return nil
}

const (
	SnapshotProcess_Restore  = "restore"
	snapshotType_DumpRestore = "dump-restore"
)

func (c *Controller) createRestoreJob(elastic *tapi.Elastic, snapshot *tapi.Snapshot) (*batch.Job, error) {
	databaseName := elastic.Name
	jobName := rand.WithUniqSuffix(databaseName)
	jobLabel := map[string]string{
		amc.LabelDatabaseName: databaseName,
		amc.LabelJobType:      SnapshotProcess_Restore,
	}
	backupSpec := snapshot.Spec.SnapshotStorageSpec

	// Get PersistentVolume object for Backup Util pod.
	persistentVolume, err := c.getVolumeForSnapshot(elastic.Spec.Storage, jobName, elastic.Namespace)
	if err != nil {
		return nil, err
	}

	// Folder name inside Cloud bucket where backup will be uploaded
	folderName := fmt.Sprintf("%v/%v/%v", amc.DatabaseNamePrefix, snapshot.Namespace, snapshot.Spec.DatabaseName)

	job := &batch.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:   jobName,
			Labels: jobLabel,
		},
		Spec: batch.JobSpec{
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: jobLabel,
				},
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{
							Name:  SnapshotProcess_Restore,
							Image: docker.ImageElasticdump + ":" + c.opt.ElasticDumpTag,
							Args: []string{
								fmt.Sprintf(`--process=%s`, SnapshotProcess_Restore),
								fmt.Sprintf(`--host=%s`, databaseName),
								fmt.Sprintf(`--bucket=%s`, backupSpec.BucketName),
								fmt.Sprintf(`--folder=%s`, folderName),
								fmt.Sprintf(`--snapshot=%s`, snapshot.Name),
							},
							VolumeMounts: []apiv1.VolumeMount{
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
					Volumes: []apiv1.Volume{
						{
							Name: "cloud",
							VolumeSource: apiv1.VolumeSource{
								Secret: backupSpec.StorageSecret,
							},
						},
						{
							Name:         persistentVolume.Name,
							VolumeSource: persistentVolume.VolumeSource,
						},
					},
					RestartPolicy: apiv1.RestartPolicyNever,
				},
			},
		},
	}

	return c.Client.BatchV1().Jobs(elastic.Namespace).Create(job)
}

func getStatefulSetName(databaseName string) string {
	return fmt.Sprintf("%v-%v", databaseName, tapi.ResourceCodeElastic)
}
