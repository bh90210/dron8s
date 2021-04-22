package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
	"text/template"

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

func main() {
	// Lookup for env variable `PLUGIN_KUBECONFIG`.
	kubeconfig, exists := os.LookupEnv("PLUGIN_KUBECONFIG")
	switch exists {
	// If it does exists means user intents for out-of-cluster usage with provided kubeconfig
	case true:
		data := []byte(kubeconfig)
		// create a kubeconfig file
		err := ioutil.WriteFile("./kubeconfig", data, 0644)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		outOfCluster, err := clientcmd.BuildConfigFromFlags("", "./kubeconfig")
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		fmt.Println("Out-of-cluster SSA initiliazing")
		err = ssa(context.Background(), outOfCluster)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

	// If user didn't provide a kubeconfig dron8s defaults to create an in-cluster config
	case false:
		inCluster, err := rest.InClusterConfig()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		fmt.Println("In-cluster SSA initiliazing")
		err = ssa(context.Background(), inCluster)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}

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
		return err
	}

	// convert it to string
	text := string(yaml)
	// Parse variables
	t := template.Must(template.New("dron8s").Option("missingkey=zero").Parse(text))
	b := bytes.NewBuffer(make([]byte, 0, 512))
	err = t.Execute(b, getVariablesFromDrone())
	if err != nil {
		return err
	}
	text = b.String()
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

		// 3. Decode YAML manifest into unstructured.Unstructured
		obj := &unstructured.Unstructured{}
		_, gvk, err := decUnstructured.Decode([]byte(v), nil, obj)
		if "" == obj.GetNamespace() {
			obj.SetNamespace("default")
		}
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

		fmt.Println("Applying config #", i)
		// 7. Create or Update the object with SSA
		//     types.ApplyPatchType indicates SSA.
		//     FieldManager specifies the field owner ID.
		_, err = dr.Patch(ctx, obj.GetName(), types.ApplyPatchType, data, metav1.PatchOptions{
			FieldManager: "dron8s-plugin",
		})
		if err != nil {
			return err
		}

		sum = i
	}

	fmt.Println("Dron8s finished applying ", sum+1, " configs.")

	return nil
}

// getVariablesFromDrone Get variables from drone
func getVariablesFromDrone() map[string]string {
	ctx := make(map[string]string)
	pluginEnv := os.Environ()
	for _, value := range pluginEnv {
		re := regexp.MustCompile(`^PLUGIN_(.*)=(.*)`)
		if re.MatchString(value) {
			matches := re.FindStringSubmatch(value)
			key := strings.ToLower(matches[1])
			ctx[key] = matches[2]
		}

		re = regexp.MustCompile(`^DRONE_(.*)=(.*)`)
		if re.MatchString(value) {
			matches := re.FindStringSubmatch(value)
			key := strings.ToLower(matches[1])
			ctx[key] = matches[2]
		}
	}
	return ctx
}
