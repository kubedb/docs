package controller

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/appscode/go/log"
	"github.com/appscode/go/types"
	"github.com/appscode/kutil"
	"github.com/golang/glog"
	_ "github.com/graymeta/stow/azure"
	_ "github.com/graymeta/stow/google"
	_ "github.com/graymeta/stow/s3"
	apps "k8s.io/api/apps/v1beta1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	typ "k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
)

const (
	sleepDuration = time.Second * 10
)

func (c *Controller) deleteDeployment(name, namespace string) error {
	deployment, err := c.Client.AppsV1beta1().Deployments(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		if kerr.IsNotFound(err) {
			return nil
		} else {
			return err
		}
	}

	// Update Deployments
	_, err = tryPatchDeployment(c.Client, deployment.ObjectMeta, func(in *apps.Deployment) *apps.Deployment {
		in.Spec.Replicas = types.Int32P(0)
		return in
	})
	if err != nil {
		return err
	}

	var checkSuccess bool = false
	then := time.Now()
	now := time.Now()
	for now.Sub(then) < time.Minute*10 {
		podList, err := c.Client.CoreV1().Pods(metav1.NamespaceAll).List(metav1.ListOptions{
			LabelSelector: labels.Set(deployment.Spec.Selector.MatchLabels).AsSelector().String(),
		})
		if err != nil {
			return err
		}
		if len(podList.Items) == 0 {
			checkSuccess = true
			break
		}

		time.Sleep(sleepDuration)
		now = time.Now()
	}

	if !checkSuccess {
		return errors.New("Fail to delete Deployments Pods")
	}
	// Delete Deployments
	return c.Client.AppsV1beta1().Deployments(deployment.Namespace).Delete(deployment.Name, nil)
}

func tryPatchDeployment(c kubernetes.Interface, meta metav1.ObjectMeta, transform func(*apps.Deployment) *apps.Deployment) (result *apps.Deployment, err error) {
	attempt := 0
	err = wait.PollImmediate(kutil.RetryInterval, kutil.RetryTimeout, func() (bool, error) {
		attempt++
		cur, e2 := c.AppsV1beta1().Deployments(meta.Namespace).Get(meta.Name, metav1.GetOptions{})
		if kerr.IsNotFound(e2) {
			return false, e2
		} else if e2 == nil {
			result, e2 = patchDeployments(c, cur, transform)
			return e2 == nil, nil
		}
		glog.Errorf("Attempt %d failed to patch Deployments %s/%s due to %v.", attempt, cur.Namespace, cur.Name, e2)
		return false, nil
	})

	if err != nil {
		err = fmt.Errorf("failed to patch Deployments %s/%s after %d attempts due to %v", meta.Namespace, meta.Name, attempt, err)
	}
	return
}

func patchDeployments(c kubernetes.Interface, cur *apps.Deployment, transform func(*apps.Deployment) *apps.Deployment) (*apps.Deployment, error) {
	curJson, err := json.Marshal(cur)
	if err != nil {
		return nil, err
	}

	modJson, err := json.Marshal(transform(cur.DeepCopy()))
	if err != nil {
		return nil, err
	}

	patch, err := strategicpatch.CreateTwoWayMergePatch(curJson, modJson, apps.Deployment{})
	if err != nil {
		return nil, err
	}
	if len(patch) == 0 || string(patch) == "{}" {
		return cur, nil
	}
	glog.V(3).Infof("Patching Deployment %s/%s with %s.", cur.Namespace, cur.Name, string(patch))
	return c.AppsV1beta1().Deployments(cur.Namespace).Patch(cur.Name, typ.StrategicMergePatchType, patch)
}

func (c *Controller) checkDeploymentPodStatus(deployment *apps.Deployment, checkDuration time.Duration) error {
	podReady := false
	then := time.Now()
	now := time.Now()
	for now.Sub(then) < checkDuration {
		dep, err := c.Client.AppsV1beta1().Deployments(deployment.Namespace).Get(deployment.Name, metav1.GetOptions{})
		if err != nil {
			if kerr.IsNotFound(err) {
				if dep.Status.Replicas != 0 && dep.Status.Replicas == dep.Status.ReadyReplicas {
					break
				}
				time.Sleep(sleepDuration)
				now = time.Now()
				continue
			} else {
				return err
			}
		}
		log.Debugf("Available replicas: %v", dep.Status.ReadyReplicas)

		// If job is success
		if dep.Status.Replicas != 0 && dep.Status.Replicas == dep.Status.ReadyReplicas {
			podReady = true
			break
		}

		time.Sleep(sleepDuration)
		now = time.Now()
	}
	if !podReady {
		return errors.New("Database fails to be Ready")
	}
	return nil
}
