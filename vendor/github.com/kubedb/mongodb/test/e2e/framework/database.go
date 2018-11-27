package framework

import (
	"fmt"
	"time"

	"github.com/appscode/kutil/tools/portforward"
	"github.com/globalsign/mgo/bson"
	"github.com/go-bongo/bongo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Todo: use official go-mongodb driver. https://github.com/mongodb/mongo-go-driver
// Currently in Alpha Release.
//
// Connect to each replica set instances to check data.
// Currently `Secondary Nodes` not supported in used drivers.

type KubedbTable struct {
	bongo.DocumentBase `bson:",inline"`
	FirstName          string
	LastName           string
}

func (f *Framework) ForwardPort(meta metav1.ObjectMeta, clientPodName string) (*portforward.Tunnel, error) {
	tunnel := portforward.NewTunnel(
		f.kubeClient.CoreV1().RESTClient(),
		f.restConfig,
		meta.Namespace,
		clientPodName,
		27017,
	)

	if err := tunnel.ForwardPort(); err != nil {
		return nil, err
	}
	return tunnel, nil
}

func (f *Framework) GetMongoDBClient(meta metav1.ObjectMeta, tunnel *portforward.Tunnel, dbName string) (*bongo.Connection, error) {
	mongodb, err := f.GetMongoDB(meta)
	if err != nil {
		return nil, err
	}

	user := "root"
	pass, err := f.GetMongoDBRootPassword(mongodb)

	config := &bongo.Config{
		ConnectionString: fmt.Sprintf("mongodb://%s:%s@127.0.0.1:%v", user, pass, tunnel.Local),
		Database:         dbName,
	}
	return bongo.Connect(config)
}

func (f *Framework) GetPrimaryInstance(meta metav1.ObjectMeta, dbName string) (string, error) {
	mongodb, err := f.GetMongoDB(meta)
	if err != nil {
		return "", err
	}

	if mongodb.Spec.ReplicaSet == nil {
		return fmt.Sprintf("%v-0", mongodb.Name), nil
	}

	// For MongoDB ReplicaSet, Find out the primary instance.
	// Current driver only connects to a primary instance.
	// So, try to connect to each instance, and once it is connected to onc,
	// that is our desired primary component!
	//
	// TODO: Extract information `IsMaster: true` from the component's status.
	// Keep track of official go-mongodb driver and introduce that once it is stable.

	for i := int32(0); i < *mongodb.Spec.Replicas; i++ {
		clientPodName := fmt.Sprintf("%v-%d", mongodb.Name, i)
		tunnel, err := f.ForwardPort(meta, clientPodName)
		if err != nil {
			return "", err
		}

		en, err := f.GetMongoDBClient(meta, tunnel, dbName)
		tunnel.Close()
		if err == nil {
			en.Session.Close()
			return clientPodName, nil
		}
		fmt.Println("GetMongoDB Client error", err)
	}
	return "", err
}

func (f *Framework) EventuallyInsertDocument(meta metav1.ObjectMeta, dbName string) GomegaAsyncAssertion {
	return Eventually(
		func() bool {
			podName, err := f.GetPrimaryInstance(meta, dbName)
			if err != nil {
				fmt.Println("GetPrimaryInstance error", err)
				return false
			}

			tunnel, err := f.ForwardPort(meta, podName)
			if err != nil {
				fmt.Println("Failed to forward port. Reason: ", err)
				return false
			}
			defer tunnel.Close()

			en, err := f.GetMongoDBClient(meta, tunnel, dbName)
			if err != nil {
				fmt.Println("GetMongoDB Client error", err)
				return false
			}

			defer en.Session.Close()

			if err := en.Session.Ping(); err != nil {
				fmt.Println("Ping error", err)
				return false
			}

			person := &KubedbTable{
				FirstName: "kubernetes",
				LastName:  "database",
			}

			if err := en.Collection("people").Save(person); err != nil {
				fmt.Println("creation error", err)
				return false
			}
			return true
		},
		time.Minute*5,
		time.Second*5,
	)
}

func (f *Framework) EventuallyDocumentExists(meta metav1.ObjectMeta, dbName string) GomegaAsyncAssertion {
	return Eventually(
		func() bool {
			podName, err := f.GetPrimaryInstance(meta, dbName)
			if err != nil {
				fmt.Println("GetPrimaryInstance error", err)
				return false
			}

			tunnel, err := f.ForwardPort(meta, podName)
			if err != nil {
				fmt.Println("Failed to forward port. Reason: ", err)
				return false
			}
			defer tunnel.Close()

			en, err := f.GetMongoDBClient(meta, tunnel, dbName)
			if err != nil {
				fmt.Println("GetMongoDB Client error", err)
				return false
			}
			defer en.Session.Close()

			if err := en.Session.Ping(); err != nil {
				fmt.Println("Ping error", err)
				return false
			}
			person := &KubedbTable{}

			if er := en.Collection("people").FindOne(bson.M{"firstname": "kubernetes"}, person); er == nil {
				return true
			} else {
				fmt.Println("checking error", er)
			}
			return false
		},
		time.Minute*5,
		time.Second*5,
	)
}
