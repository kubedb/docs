package validator

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/appscode/go/arrays"
	"github.com/pkg/errors"
	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	mona "kmodules.xyz/monitoring-agent-api/api/v1"
	store "kmodules.xyz/objectstore-api/api/v1"
	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha1"
)

func ValidateStorage(client kubernetes.Interface, storageType api.StorageType, spec *core.PersistentVolumeClaimSpec, storageSpecPath ...string) error {
	if storageType == api.StorageTypeEphemeral {
		return nil
	}

	storagePath := "spec.storage"
	if len(storageSpecPath) != 0 {
		storagePath = strings.Join(storageSpecPath, ".")
	}

	if spec == nil {
		return fmt.Errorf(`%v is missing for durable storage type`, storagePath)
	}

	if spec.StorageClassName != nil {
		if _, err := client.StorageV1beta1().StorageClasses().Get(*spec.StorageClassName, metav1.GetOptions{}); err != nil {
			if kerr.IsNotFound(err) {
				return fmt.Errorf(`%v.storageClassName "%v" not found`, storagePath, *spec.StorageClassName)
			}
			return err
		}
	}

	if val, found := spec.Resources.Requests[core.ResourceStorage]; found {
		if val.Value() <= 0 {
			return errors.New("invalid ResourceStorage request")
		}
	} else {
		return errors.New("missing ResourceStorage request")
	}

	return nil
}

func ValidateBackupSchedule(client kubernetes.Interface, spec *api.BackupScheduleSpec, namespace string) error {
	if spec == nil {
		return nil
	}
	// CronExpression can't be empty
	if spec.CronExpression == "" {
		return errors.New("invalid cron expression")
	}

	return ValidateSnapshotSpec(spec.Backend)
}

func ValidateSnapshotSpec(spec store.Backend) error {
	// BucketName can't be empty
	if spec.S3 == nil && spec.GCS == nil && spec.Azure == nil && spec.Swift == nil && spec.Local == nil {
		return errors.New("no storage provider is configured")
	}

	if spec.Local != nil {
		return nil
	}

	// Note: S3 & GCS bucket can be accessed with default IAM account credential. So do not require secret
	// Must provide Storage credentials for Azure & Swift
	if spec.Azure != nil || spec.Swift != nil {
		if spec.StorageSecretName == "" {
			return fmt.Errorf(`object 'SecretName' is missing in '%v'`, spec)
		}
	}

	return nil
}

// ValidateMonitorSpec validates the Monitoring spec after all the defaulting is done.
func ValidateMonitorSpec(monitorSpec *mona.AgentSpec) error {
	specData, err := json.Marshal(monitorSpec)
	if err != nil {
		return err
	}

	if monitorSpec.Agent == "" {
		return fmt.Errorf(`object 'Agent' is missing in '%v'`, string(specData))
	}

	if monitorSpec.Agent.Vendor() == mona.VendorPrometheus {
		if monitorSpec.Prometheus != nil &&
			monitorSpec.Prometheus.Port >= 1024 &&
			monitorSpec.Prometheus.Port <= 65535 {
			return nil
		}
		return fmt.Errorf(`invalid 'Monitor.Prometheus' in '%v'. Prometheus.Port value must be between 1024 and 65535, inclusive`, string(specData))
	}

	return fmt.Errorf(`invalid 'Agent' in '%v'`, string(specData))
}

func ValidateEnvVar(envs []core.EnvVar, forbiddenEnvs []string, resourceType string) error {
	for _, env := range envs {
		present, _ := arrays.Contains(forbiddenEnvs, env.Name)
		if present {
			return fmt.Errorf("environment variable %s is forbidden to use in %s spec", env.Name, resourceType)
		}
	}
	return nil
}
