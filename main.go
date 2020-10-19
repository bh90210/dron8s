package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/discovery"

	// memory "k8s.io/client-go/discovery/cached"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
)

// GOOS=linux GOARCH=amd64 CGO_ENABLED=0 ARG=v0.0.9 go generate main.go
//go:generate go build -o dron8s
//go:generate docker build -t bh90210/dron8s:latest -t bh90210/dron8s:$ARG .
//go:generate docker push bh90210/dron8s:$ARG
//go:generate docker push bh90210/dron8s:latest

func main() {
	// body := strings.NewReader(
	// 	os.Getenv("PLUGIN_BODY"),
	// )

	// req, err := http.NewRequest(
	// 	os.Getenv("PLUGIN_METHOD"),
	// 	os.Getenv("PLUGIN_URL"),
	// 	body,
	// )

	// fmt.Println(req)

	fmt.Println("k8s starting")
	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	// get pods in all the namespaces by omitting namespace
	// Or specify namespace to get pods in particular namespace
	pods, err := clientset.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}
	fmt.Printf("There are %d pods in the cluster\n", len(pods.Items))

	// Examples for error handling:
	// - Use helper functions e.g. errors.IsNotFound()
	// - And/or cast to StatusError and use its properties like e.g. ErrStatus.Message
	_, err = clientset.CoreV1().Pods("default").Get(context.TODO(), "example-xxxxx", metav1.GetOptions{})
	if errors.IsNotFound(err) {
		fmt.Printf("Pod example-xxxxx not found in default namespace\n")
	} else if statusError, isStatus := err.(*errors.StatusError); isStatus {
		fmt.Printf("Error getting pod %v\n", statusError.ErrStatus.Message)
	} else if err != nil {
		panic(err.Error())
	} else {
		fmt.Printf("Found example-xxxxx pod in default namespace\n")
	}

	fmt.Println("k8s ended")
	// if err != nil {
	// 	os.Exit(1)
	// }

	//
	//
	//

	er := doSSA(context.Background(), config)
	log.Println(er)
}

const deploymentYAML = `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
  namespace: default
spec:
  selector:
    matchLabels:
      app: nginx
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
      - name: nginx
        image: nginx:latest`

var decUnstructured = yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)

func doSSA(ctx context.Context, cfg *rest.Config) error {

	// 1. Prepare a RESTMapper to find GVR
	dc, err := discovery.NewDiscoveryClientForConfig(cfg)
	if err != nil {
		return err
	}
	mapper := restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(dc))

	// 2. Prepare the dynamic client
	dyn, err := dynamic.NewForConfig(cfg)
	if err != nil {
		return err
	}

	// 	// 2.1.
	// 	yamlfile, err := os.Open(os.Getenv("PLUGIN_YAML"))
	// 	defer yamlfile.Close()

	// 	// Splits on newlines by default.
	// 	scanner := bufio.NewScanner(yamlfile)

	// 	line := 1
	// 	// https://golang.org/pkg/bufio/#Scanner.Scan
	// 	for scanner.Scan() {
	// 		if strings.Contains(scanner.Text(), `
	// ---
	// `) {
	// 			fmt.Println(line)
	// 		}

	// 		line++
	// 	}

	// 	if err := scanner.Err(); err != nil {
	// 		fmt.Println(err)
	// 	}

	yamlfile2, err := ioutil.ReadFile(os.Getenv("PLUGIN_YAML"))
	if err != nil {
		log.Fatal(err)
	}
	text := string(yamlfile2)

	resources := strings.Split(text, "---")

	for _, v := range resources {
		// 3. Decode YAML manifest into unstructured.Unstructured
		obj := &unstructured.Unstructured{}
		_, gvk, err := decUnstructured.Decode([]byte(v), nil, obj)
		if err != nil {
			return err
		}

		// 4. Find GVR
		mapping, err := mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
		if err != nil {
			return err
		}

		// 5. Obtain REST interface for the GVR
		var dr dynamic.ResourceInterface
		if mapping.Scope.Name() == meta.RESTScopeNameNamespace {
			// namespaced resources should specify the namespace
			dr = dyn.Resource(mapping.Resource).Namespace(obj.GetNamespace())
		} else {
			// for cluster-wide resources
			dr = dyn.Resource(mapping.Resource)
		}

		// 6. Marshal object into JSON
		data, err := json.Marshal(obj)
		if err != nil {
			return err
		}

		// 7. Create or Update the object with SSA
		//     types.ApplyPatchType indicates SSA.
		//     FieldManager specifies the field owner ID.
		_, err = dr.Patch(ctx, obj.GetName(), types.ApplyPatchType, data, metav1.PatchOptions{
			FieldManager: "dron8s-plugin",
		})
	}

	return err

}
