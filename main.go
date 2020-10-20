package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"
)

// GOOS=linux GOARCH=amd64 CGO_ENABLED=0 ARG=v0.0.9 go generate main.go
//go:generate go build -o dron8s
//go:generate docker build -t bh90210/dron8s:latest -t bh90210/dron8s:$ARG .
//go:generate docker push bh90210/dron8s:$ARG
//go:generate docker push bh90210/dron8s:latest

func main() {
	var config *rest.Config
	// Lookup for env variable `PLUGIN_KUBECONFIG`.
	kubeconfig, exists := os.LookupEnv("PLUGIN_KUBECONFIG")
	switch exists {
	// If it does exists means user intents for out-of-cluster usage with provided kubeconfig
	case true:
		outOfCluster, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			panic(err.Error())
		}
		config = outOfCluster

	// If user didn't provide a kubeconfig dron8s defaults to create an in-cluster config
	case false:
		inCluster, err := rest.InClusterConfig()
		if err != nil {
			panic(err.Error())
		}
		config = inCluster
	}

	// run the server side apply function
	err := ssa(context.Background(), config)
	log.Println(err)
}

// https://ymmt2005.hatenablog.com/entry/2020/04/14/An_example_of_using_dynamic_client_of_k8s.io/client-go#Go-client-libraries
func ssa(ctx context.Context, cfg *rest.Config) error {
	var decUnstructured = yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)

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

	// 2.1. Read user's yaml
	yaml, err := ioutil.ReadFile(os.Getenv("PLUGIN_YAML"))
	if err != nil {
		log.Fatal(err)
	}

	// convert it to string
	text := string(yaml)
	// Parse each yaml from file
	configs := strings.Split(text, "---")
	// variable to hold and print how many yaml configs are present
	var sum int
	// Iterate over provided configs
	for i, v := range configs {
		// If a yaml starts with `---`
		// the first slice of `configs` will be empty
		// so we just skip (continue) to next iteration
		if len(v) == 0 {
			continue
		}

		fmt.Println("Applying yaml nu ", i)

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

		sum = i
	}

	fmt.Println("Dron8s finished applying ", sum+1, " resources.")

	return err
}
