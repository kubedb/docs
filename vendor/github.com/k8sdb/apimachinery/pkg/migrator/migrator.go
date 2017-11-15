package migrator

import (
	"errors"
	"fmt"
	"time"

	"github.com/appscode/go/log"
	"github.com/hashicorp/go-version"
	api "github.com/k8sdb/apimachinery/apis/kubedb/v1alpha1"
	cs "github.com/k8sdb/apimachinery/client/typed/kubedb/v1alpha1"
	extensions "k8s.io/api/extensions/v1beta1"
	crd_api "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	crd_cs "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1beta1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
)

type migrationState struct {
	tprRegDeleted bool
	crdCreated    bool
}

type migrator struct {
	kubeClient       kubernetes.Interface
	apiExtKubeClient crd_cs.ApiextensionsV1beta1Interface
	extClient        cs.KubedbV1alpha1Interface

	migrationState *migrationState
}

func NewMigrator(kubeClient kubernetes.Interface, apiExtKubeClient crd_cs.ApiextensionsV1beta1Interface, extClient cs.KubedbV1alpha1Interface) *migrator {
	return &migrator{
		migrationState:   &migrationState{},
		kubeClient:       kubeClient,
		apiExtKubeClient: apiExtKubeClient,
		extClient:        extClient,
	}
}

func (m *migrator) isMigrationNeeded(runtimeObjs ...api.ResourceInfo) (bool, error) {
	v, err := m.kubeClient.Discovery().ServerVersion()
	if err != nil {
		return false, err
	}

	ver, err := version.NewVersion(v.String())
	if err != nil {
		return false, err
	}

	mv := ver.Segments()[1]

	if mv == 7 {
		for _, runtime := range runtimeObjs {
			_, err := m.kubeClient.ExtensionsV1beta1().ThirdPartyResources().Get(
				runtime.ResourceName()+"."+api.SchemeGroupVersion.Group,
				metav1.GetOptions{},
			)
			if err != nil {
				if !kerr.IsNotFound(err) {
					return false, err
				}
			} else {
				return true, nil
			}
		}
	}
	return false, nil
}

func (m *migrator) RunMigration(runtimeObjs ...api.ResourceInfo) error {
	needed, err := m.isMigrationNeeded(runtimeObjs...)
	if err != nil {
		return err
	}

	if needed {
		if err := m.migrateTPR2CRD(runtimeObjs...); err != nil {
			return m.rollback(runtimeObjs...)
		}
	}

	return nil
}

func (m *migrator) migrateTPR2CRD(runtimeObjs ...api.ResourceInfo) error {
	log.Debugln("Performing TPR to CRD migration.")

	log.Debugln("Deleting TPRs.")
	if err := m.deleteTPRs(runtimeObjs...); err != nil {
		return err
	}

	m.migrationState.tprRegDeleted = true

	log.Debugln("Creating CRDs.")
	if err := m.createCRDs(runtimeObjs...); err != nil {
		return err
	}

	m.migrationState.crdCreated = true

	log.Debugln("Waiting for CRDs to be ready.")
	if err := m.waitForCRDsReady(len(runtimeObjs)); err != nil {
		return err
	}

	return nil
}

func (m *migrator) deleteTPRs(runtimeObjs ...api.ResourceInfo) error {
	tprClient := m.kubeClient.ExtensionsV1beta1().ThirdPartyResources()

	deleteTPR := func(runtime api.ResourceInfo) error {
		name := runtime.ResourceName() + "." + api.SchemeGroupVersion.Group
		if err := tprClient.Delete(name, &metav1.DeleteOptions{}); err != nil && !kerr.IsNotFound(err) {
			return fmt.Errorf(`Failed to delete TPR "%s"`, name)
		}
		return nil
	}

	for _, runtime := range runtimeObjs {
		if err := deleteTPR(runtime); err != nil {
			return err
		}
	}
	return nil
}

func (m *migrator) createCRDs(runtimeObjs ...api.ResourceInfo) error {
	for _, runtime := range runtimeObjs {
		if err := m.createCRD(runtime); err != nil {
			return err
		}
	}
	return nil
}

func (m *migrator) createCRD(runtime api.ResourceInfo) error {
	crd := &crd_api.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: runtime.ResourceType() + "." + api.SchemeGroupVersion.Group,
			Labels: map[string]string{
				"app": "kubedb",
			},
		},
		Spec: crd_api.CustomResourceDefinitionSpec{
			Group:   api.SchemeGroupVersion.Group,
			Version: api.SchemeGroupVersion.Version,
			Scope:   crd_api.NamespaceScoped,
			Names: crd_api.CustomResourceDefinitionNames{
				Plural:     runtime.ResourceType(),
				ShortNames: []string{runtime.ResourceCode()},
				Kind:       runtime.ResourceKind(),
			},
		},
	}

	crdClient := m.apiExtKubeClient.CustomResourceDefinitions()

	if _, err := crdClient.Create(crd); err != nil && !kerr.IsAlreadyExists(err) {
		return fmt.Errorf(`Failed to create CRD "%v"`, crd.Spec.Names.Kind)
	}

	err := wait.Poll(500*time.Millisecond, 60*time.Second, func() (bool, error) {
		crdEst, err := crdClient.Get(crd.ObjectMeta.Name, metav1.GetOptions{})
		if err != nil {
			return false, err
		}
		for _, cond := range crdEst.Status.Conditions {
			switch cond.Type {
			case crd_api.Established:
				if cond.Status == crd_api.ConditionTrue {
					return true, nil
				}
			case crd_api.NamesAccepted:
				if cond.Status == crd_api.ConditionFalse {
					fmt.Printf("Name conflict. Reason: %v\n", cond.Reason)
				}
			}
		}
		return false, fmt.Errorf(`Failed to get CustomResourceDefinition "%v"`, runtime.ResourceKind())
	})

	return err
}

func (m *migrator) waitForCRDsReady(expectedCRD int) error {
	labelMap := map[string]string{
		"app": "kubedb",
	}

	return wait.Poll(3*time.Second, 10*time.Minute, func() (bool, error) {
		crdList, err := m.apiExtKubeClient.CustomResourceDefinitions().List(metav1.ListOptions{
			LabelSelector: labels.SelectorFromSet(labelMap).String(),
		})
		if err != nil {
			return false, err
		}

		if len(crdList.Items) == expectedCRD {
			return true, nil
		}

		return false, errors.New("Failed to get all CustomResourceDefinitions")
	})
}

func (m *migrator) rollback(runtimeObjs ...api.ResourceInfo) error {
	log.Debugln("Rolling back migration.")

	ms := m.migrationState

	if ms.crdCreated {
		log.Debugln("Deleting CRDs.")
		err := m.deleteCRDs()
		if err != nil {
			return err
		}
	}

	if ms.tprRegDeleted {
		log.Debugln("Creating TPRs.")
		err := m.CreateTPRs()
		if err != nil {
			return fmt.Errorf("Failed to recreate TPR. Error: %v", err.Error())
		}

		err = m.WaitForTPRsReady(len(runtimeObjs))
		if err != nil {
			return err
		}
	}

	return nil
}

func (m *migrator) deleteCRDs(runtimeObjs ...api.ResourceInfo) error {
	crdClient := m.apiExtKubeClient.CustomResourceDefinitions()

	deleteCRD := func(runtime api.ResourceInfo) error {
		name := runtime.ResourceType() + "." + api.SchemeGroupVersion.Group
		if err := crdClient.Delete(name, &metav1.DeleteOptions{}); err != nil && !kerr.IsNotFound(err) {
			return fmt.Errorf(`Failed to delete CRD "%s""`, name)
		}
		return nil
	}

	for _, runtime := range runtimeObjs {
		if err := deleteCRD(runtime); err != nil {
			return err
		}
	}
	return nil
}

func (m *migrator) CreateTPRs(runtimeObjs ...api.ResourceInfo) error {
	for _, runtime := range runtimeObjs {
		if err := m.createTPR(runtime); err != nil {
			return err
		}
	}
	return nil
}

func (m *migrator) createTPR(runtime api.ResourceInfo) error {
	name := runtime.ResourceName() + "." + api.SchemeGroupVersion.Group
	_, err := m.kubeClient.ExtensionsV1beta1().ThirdPartyResources().Get(name, metav1.GetOptions{})
	if !kerr.IsNotFound(err) {
		return err
	}

	thirdPartyResource := &extensions.ThirdPartyResource{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "extensions/v1beta1",
			Kind:       "ThirdPartyResource",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Labels: map[string]string{
				"app": "kubedb",
			},
		},
		Description: "Searchlight by AppsCode - Alerts for Kubernetes",
		Versions: []extensions.APIVersion{
			{
				Name: api.SchemeGroupVersion.Version,
			},
		},
	}

	_, err = m.kubeClient.ExtensionsV1beta1().ThirdPartyResources().Create(thirdPartyResource)
	if err != nil && !kerr.IsAlreadyExists(err) {
		return err
	}

	return nil
}

func (m *migrator) WaitForTPRsReady(expectedTPR int) error {
	labelMap := map[string]string{
		"app": "kubedb",
	}

	return wait.Poll(3*time.Second, 10*time.Minute, func() (bool, error) {
		crdList, err := m.kubeClient.ExtensionsV1beta1().ThirdPartyResources().List(metav1.ListOptions{
			LabelSelector: labels.SelectorFromSet(labelMap).String(),
		})
		if err != nil {
			return false, err
		}

		if len(crdList.Items) == expectedTPR {
			return true, nil
		}

		return false, errors.New("Failed to get all ThirdPartyResources")
	})
}
