package controller

import (
	"fmt"
	"time"

	"github.com/appscode/go/crypto/rand"
	tapi "github.com/k8sdb/apimachinery/api"
	amc "github.com/k8sdb/apimachinery/pkg/controller"
	kapi "k8s.io/kubernetes/pkg/api"
	k8serr "k8s.io/kubernetes/pkg/api/errors"
	kapps "k8s.io/kubernetes/pkg/apis/apps"
	kbatch "k8s.io/kubernetes/pkg/apis/batch"
	"k8s.io/kubernetes/pkg/util/intstr"
)

const (
	annotationDatabaseVersion  = "elastic.kubedb.com/version"
	ImageElasticsearch         = "kubedb/elasticsearch"
	ImageOperatorElasticsearch = "kubedb/es-operator"
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

	if _, err := c.Client.Core().Services(namespace).Create(service); err != nil {
		return err
	}

	return nil
}

func (c *Controller) checkStatefulSet(elastic *tapi.Elastic) (*kapps.StatefulSet, error) {
	// SatatefulSet for Postgres database
	statefulSetName := getStatefulSetName(elastic.Name)
	statefulSet, err := c.Client.Apps().StatefulSets(elastic.Namespace).Get(statefulSetName)
	if err != nil {
		if k8serr.IsNotFound(err) {
			return nil, nil
		} else {
			return nil, err
		}
	}

	if statefulSet.Labels[amc.LabelDatabaseKind] != tapi.ResourceKindElastic {
		return nil, fmt.Errorf(`Intended statefulSet "%v" already exists`, statefulSetName)
	}

	return statefulSet, nil
}

func (c *Controller) createStatefulSet(elastic *tapi.Elastic) (*kapps.StatefulSet, error) {
	_statefulSet, err := c.checkStatefulSet(elastic)
	if err != nil {
		return nil, err
	}
	if _statefulSet != nil {
		return _statefulSet, nil
	}

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
	annotations[annotationDatabaseVersion] = elastic.Spec.Version

	podLabels := make(map[string]string)
	for key, val := range labels {
		podLabels[key] = val
	}
	podLabels[amc.LabelDatabaseName] = elastic.Name

	dockerImage := fmt.Sprintf("%v:%v", ImageElasticsearch, elastic.Spec.Version)
	initContainerImage := fmt.Sprintf("%v:%v", ImageOperatorElasticsearch, c.operatorTag)

	// SatatefulSet for Elastic database
	statefulSetName := getStatefulSetName(elastic.Name)
	statefulSet := &kapps.StatefulSet{
		ObjectMeta: kapi.ObjectMeta{
			Name:        statefulSetName,
			Namespace:   elastic.Namespace,
			Labels:      labels,
			Annotations: annotations,
		},
		Spec: kapps.StatefulSetSpec{
			Replicas:    elastic.Spec.Replicas,
			ServiceName: c.governingService,
			Template: kapi.PodTemplateSpec{
				ObjectMeta: kapi.ObjectMeta{
					Labels:      podLabels,
					Annotations: annotations,
				},
				Spec: kapi.PodSpec{
					Containers: []kapi.Container{
						{
							Name:            tapi.ResourceNameElastic,
							Image:           dockerImage,
							ImagePullPolicy: kapi.PullIfNotPresent,
							Ports: []kapi.ContainerPort{
								{
									Name:          "api",
									ContainerPort: 9200,
								},
								{
									Name:          "tcp",
									ContainerPort: 9300,
								},
							},
							VolumeMounts: []kapi.VolumeMount{
								{
									Name:      "discovery",
									MountPath: "/tmp/discovery",
								},
								{
									Name:      "data",
									MountPath: "/var/pv",
								},
							},
							Env: []kapi.EnvVar{
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
					InitContainers: []kapi.Container{
						{
							Name:            "discover",
							Image:           initContainerImage,
							ImagePullPolicy: kapi.PullIfNotPresent,
							Args: []string{
								"discover",
								fmt.Sprintf("--service=%v", elastic.Name),
								fmt.Sprintf("--namespace=%v", elastic.Namespace),
							},
							Env: []kapi.EnvVar{
								{
									Name: "POD_NAME",
									ValueFrom: &kapi.EnvVarSource{
										FieldRef: &kapi.ObjectFieldSelector{
											APIVersion: "v1",
											FieldPath:  "metadata.name",
										},
									},
								},
							},
							VolumeMounts: []kapi.VolumeMount{
								{
									Name:      "discovery",
									MountPath: "/tmp/discovery",
								},
							},
						},
					},
					NodeSelector: elastic.Spec.NodeSelector,
					Volumes: []kapi.Volume{
						{
							Name: "discovery",
							VolumeSource: kapi.VolumeSource{
								EmptyDir: &kapi.EmptyDirVolumeSource{},
							},
						},
					},
				},
			},
		},
	}

	// Add Data volume for StatefulSet
	addDataVolume(statefulSet, elastic.Spec.Storage)

	if _, err := c.Client.Apps().StatefulSets(statefulSet.Namespace).Create(statefulSet); err != nil {
		return nil, err
	}

	return statefulSet, nil
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

func (c *Controller) createDormantDatabase(elastic *tapi.Elastic) (*tapi.DormantDatabase, error) {
	dormantDb := &tapi.DormantDatabase{
		ObjectMeta: kapi.ObjectMeta{
			Name:      elastic.Name,
			Namespace: elastic.Namespace,
			Labels: map[string]string{
				amc.LabelDatabaseKind: tapi.ResourceKindElastic,
			},
		},
		Spec: tapi.DormantDatabaseSpec{
			Origin: tapi.Origin{
				ObjectMeta: kapi.ObjectMeta{
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
		ObjectMeta: kapi.ObjectMeta{
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

func (c *Controller) createRestoreJob(elastic *tapi.Elastic, snapshot *tapi.Snapshot) (*kbatch.Job, error) {

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
							Image: ImageElasticDump + ":" + c.elasticDumpTag,
							Args: []string{
								fmt.Sprintf(`--process=%s`, SnapshotProcess_Restore),
								fmt.Sprintf(`--host=%s`, databaseName),
								fmt.Sprintf(`--bucket=%s`, backupSpec.BucketName),
								fmt.Sprintf(`--folder=%s`, folderName),
								fmt.Sprintf(`--snapshot=%s`, snapshot.Name),
							},
							VolumeMounts: []kapi.VolumeMount{
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

	return c.Client.Batch().Jobs(elastic.Namespace).Create(job)
}

func getStatefulSetName(databaseName string) string {
	return fmt.Sprintf("%v-%v", databaseName, tapi.ResourceCodeElastic)
}
