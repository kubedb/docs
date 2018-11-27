package main

import (
	"github.com/appscode/go/log"
	gort "github.com/appscode/go/runtime"
	crdutils "github.com/appscode/kutil/apiextensions/v1beta1"
	"github.com/appscode/kutil/openapi"
	"github.com/go-openapi/spec"
	"github.com/golang/glog"
	"io/ioutil"
	crd_api "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/kube-openapi/pkg/common"
	"kmodules.xyz/custom-resources/apis"
	cataloginstall "kmodules.xyz/custom-resources/apis/appcatalog/install"
	catalog "kmodules.xyz/custom-resources/apis/appcatalog/v1alpha1"
	"os"
	"path/filepath"
)

func generateCRDDefinitions() {
	apis.EnableStatusSubresource = true

	err := os.MkdirAll(filepath.Join(gort.GOPath(), "/src/kmodules.xyz/custom-resources/api/crds"), 0755)
	if err != nil {
		log.Fatal(err)
	}

	crds := []*crd_api.CustomResourceDefinition{
		catalog.AppBinding{}.CustomResourceDefinition(),
	}
	for _, crd := range crds {
		filename := filepath.Join(gort.GOPath(), "/src/kmodules.xyz/custom-resources/api/crds", crd.Spec.Names.Singular+".yaml")
		f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			log.Fatal(err)
		}
		crdutils.MarshallCrd(f, crd, "yaml")
		f.Close()
	}
}

func generateSwaggerJson() {
	var (
		Scheme = runtime.NewScheme()
		Codecs = serializer.NewCodecFactory(Scheme)
	)

	cataloginstall.Install(Scheme)

	apispec, err := openapi.RenderOpenAPISpec(openapi.Config{
		Scheme: Scheme,
		Codecs: Codecs,
		Info: spec.InfoProps{
			Title:   "KubeDB",
			Version: "v0",
			Contact: &spec.ContactInfo{
				Name:  "AppsCode Inc.",
				URL:   "https://appscode.com",
				Email: "hello@appscode.com",
			},
			License: &spec.License{
				Name: "Apache 2.0",
				URL:  "https://www.apache.org/licenses/LICENSE-2.0.html",
			},
		},
		OpenAPIDefinitions: []common.GetOpenAPIDefinitions{
			catalog.GetOpenAPIDefinitions,
		},
		Resources: []openapi.TypeInfo{
			{catalog.SchemeGroupVersion, catalog.ResourceApps, catalog.ResourceKindApp, true},
		},
	})
	if err != nil {
		glog.Fatal(err)
	}

	filename := gort.GOPath() + "/src/kmodules.xyz/custom-resources/api/openapi-spec/swagger.json"
	err = os.MkdirAll(filepath.Dir(filename), 0755)
	if err != nil {
		glog.Fatal(err)
	}
	err = ioutil.WriteFile(filename, []byte(apispec), 0644)
	if err != nil {
		glog.Fatal(err)
	}
}

func main() {
	generateCRDDefinitions()
	generateSwaggerJson()
}
