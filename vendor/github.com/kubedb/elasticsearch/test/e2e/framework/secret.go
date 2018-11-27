package framework

import (
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/appscode/go/crypto/rand"
	"github.com/appscode/go/log"
	api "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	. "github.com/onsi/gomega"
	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	store "kmodules.xyz/objectstore-api/api/v1"
)

func (i *Invocation) SecretForLocalBackend() *core.Secret {
	return &core.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      rand.WithUniqSuffix(i.app + "-local"),
			Namespace: i.namespace,
		},
		Data: map[string][]byte{},
	}
}

func (i *Invocation) SecretForS3Backend() *core.Secret {
	if os.Getenv(store.AWS_ACCESS_KEY_ID) == "" ||
		os.Getenv(store.AWS_SECRET_ACCESS_KEY) == "" {
		return &core.Secret{}
	}

	return &core.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      rand.WithUniqSuffix(i.app + "-s3"),
			Namespace: i.namespace,
		},
		Data: map[string][]byte{
			store.AWS_ACCESS_KEY_ID:     []byte(os.Getenv(store.AWS_ACCESS_KEY_ID)),
			store.AWS_SECRET_ACCESS_KEY: []byte(os.Getenv(store.AWS_SECRET_ACCESS_KEY)),
		},
	}
}

func (i *Invocation) SecretForGCSBackend() *core.Secret {
	if os.Getenv(store.GOOGLE_PROJECT_ID) == "" ||
		(os.Getenv("GOOGLE_APPLICATION_CREDENTIALS") == "" && os.Getenv(store.GOOGLE_SERVICE_ACCOUNT_JSON_KEY) == "") {
		return &core.Secret{}
	}

	jsonKey := os.Getenv(store.GOOGLE_SERVICE_ACCOUNT_JSON_KEY)
	if jsonKey == "" {
		if keyBytes, err := ioutil.ReadFile(os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")); err == nil {
			jsonKey = string(keyBytes)
		}
	}

	return &core.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      rand.WithUniqSuffix(i.app + "-gcs"),
			Namespace: i.namespace,
		},
		Data: map[string][]byte{
			store.GOOGLE_PROJECT_ID:               []byte(os.Getenv(store.GOOGLE_PROJECT_ID)),
			store.GOOGLE_SERVICE_ACCOUNT_JSON_KEY: []byte(jsonKey),
		},
	}
}

func (i *Invocation) SecretForAzureBackend() *core.Secret {
	if os.Getenv(store.AZURE_ACCOUNT_NAME) == "" ||
		os.Getenv(store.AZURE_ACCOUNT_KEY) == "" {
		return &core.Secret{}
	}

	return &core.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      rand.WithUniqSuffix(i.app + "-azure"),
			Namespace: i.namespace,
		},
		Data: map[string][]byte{
			store.AZURE_ACCOUNT_NAME: []byte(os.Getenv(store.AZURE_ACCOUNT_NAME)),
			store.AZURE_ACCOUNT_KEY:  []byte(os.Getenv(store.AZURE_ACCOUNT_KEY)),
		},
	}
}

func (i *Invocation) SecretForSwiftBackend() *core.Secret {
	if os.Getenv(store.OS_AUTH_URL) == "" ||
		(os.Getenv(store.OS_TENANT_ID) == "" && os.Getenv(store.OS_TENANT_NAME) == "") ||
		os.Getenv(store.OS_USERNAME) == "" ||
		os.Getenv(store.OS_PASSWORD) == "" {
		return &core.Secret{}
	}

	return &core.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      rand.WithUniqSuffix(i.app + "-swift"),
			Namespace: i.namespace,
		},
		Data: map[string][]byte{
			store.OS_AUTH_URL:    []byte(os.Getenv(store.OS_AUTH_URL)),
			store.OS_TENANT_ID:   []byte(os.Getenv(store.OS_TENANT_ID)),
			store.OS_TENANT_NAME: []byte(os.Getenv(store.OS_TENANT_NAME)),
			store.OS_USERNAME:    []byte(os.Getenv(store.OS_USERNAME)),
			store.OS_PASSWORD:    []byte(os.Getenv(store.OS_PASSWORD)),
			store.OS_REGION_NAME: []byte(os.Getenv(store.OS_REGION_NAME)),
		},
	}
}

// TODO: Add more methods for Swift, Backblaze B2, Rest server backend.
func (f *Framework) CreateSecret(obj *core.Secret) error {
	_, err := f.kubeClient.CoreV1().Secrets(obj.Namespace).Create(obj)
	return err
}

func (f *Framework) UpdateSecret(meta metav1.ObjectMeta, transformer func(core.Secret) core.Secret) error {
	attempt := 0
	for ; attempt < maxAttempts; attempt = attempt + 1 {
		cur, err := f.kubeClient.CoreV1().Secrets(meta.Namespace).Get(meta.Name, metav1.GetOptions{})
		if kerr.IsNotFound(err) {
			return nil
		} else if err == nil {
			modified := transformer(*cur)
			_, err = f.kubeClient.CoreV1().Secrets(cur.Namespace).Update(&modified)
			if err == nil {
				return nil
			}
		}
		log.Errorf("Attempt %d failed to update Secret %s@%s due to %s.", attempt, cur.Name, cur.Namespace, err)
		time.Sleep(updateRetryInterval)
	}
	return fmt.Errorf("failed to update Secret %s@%s after %d attempts", meta.Name, meta.Namespace, attempt)
}

func (f *Framework) DeleteSecret(meta metav1.ObjectMeta) error {
	err := f.kubeClient.CoreV1().Secrets(meta.Namespace).Delete(meta.Name, deleteInForeground())
	if !kerr.IsNotFound(err) {
		return err
	}
	return nil
}

func (f *Framework) EventuallyDBSecretCount(meta metav1.ObjectMeta) GomegaAsyncAssertion {
	labelMap := map[string]string{
		api.LabelDatabaseKind: api.ResourceKindElasticsearch,
		api.LabelDatabaseName: meta.Name,
	}
	labelSelector := labels.SelectorFromSet(labelMap)

	return Eventually(
		func() int {
			secretList, err := f.kubeClient.CoreV1().Secrets(meta.Namespace).List(
				metav1.ListOptions{
					LabelSelector: labelSelector.String(),
				},
			)
			Expect(err).NotTo(HaveOccurred())

			return len(secretList.Items)
		},
		time.Minute*5,
		time.Second*5,
	)
}
