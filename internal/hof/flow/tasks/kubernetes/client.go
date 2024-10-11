package kubernetes

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"

	"gopkg.in/yaml.v3" // YAML v3 library for decoding documents
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/kylelemons/godebug/diff"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"
)

type Client struct {
	clientset *kubernetes.Clientset
	dynamic   dynamic.Interface
	mapper    *restmapper.DeferredDiscoveryRESTMapper
}

func getConfig() (*rest.Config, error) {
	// Try in-cluster config first
	config, err := rest.InClusterConfig()
	if err == nil {
		return config, nil
	}

	// If that fails, try the default kubeconfig
	kubeconfigPath := os.Getenv("KUBECONFIG")
	if kubeconfigPath == "" {
		kubeconfigPath = filepath.Join(os.Getenv("HOME"), ".kube", "config")
	}

	config, err = clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load kubeconfig: %v", err)
	}

	return config, nil
}

func NewClient() (*Client, error) {
	config, err := getConfig()
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	mapper := restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(clientset.Discovery()))

	return &Client{
		clientset: clientset,
		dynamic:   dynamicClient,
		mapper:    mapper,
	}, nil
}

func (c *Client) Apply(manifest string) error {
	return c.applyOrDelete(manifest, false, false, false)
}

func (c *Client) Delete(manifest string) error {
	return c.applyOrDelete(manifest, true, false, false)
}

func (c *Client) Plan(manifest string) error {
	return c.applyOrDelete(manifest, false, false, true)
}

func (c *Client) applyOrDelete(manifestYAML string, delete bool, dryRun bool, plan bool) error {
	var obj map[string]interface{}
	err := yaml.Unmarshal([]byte(manifestYAML), &obj)
	if err != nil {
		return fmt.Errorf("failed to unmarshal manifest: %w", err)
	}

	u := &unstructured.Unstructured{Object: obj}
	gvk := u.GroupVersionKind()

	// Get the GVR
	mapping, err := c.mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return fmt.Errorf("failed to connect to kubernetes cluster: %w", err)
	}

	gvr := mapping.Resource

	var resourceClient dynamic.ResourceInterface
	namespace := u.GetNamespace()
	if mapping.Scope.Name() == meta.RESTScopeNameRoot {
		// Cluster-scoped resource
		resourceClient = c.dynamic.Resource(gvr)
	} else {
		// Namespace-scoped resource
		if namespace == "" {
			namespace = "default" // or any other default namespace you want to use
		}
		resourceClient = c.dynamic.Resource(gvr).Namespace(namespace)
	}

	ctx := context.Background()

	if delete {
		if dryRun {
			fmt.Printf("Would delete %s/%s (dry run)\n", gvk.Kind, u.GetName())
		} else {
			err = resourceClient.Delete(ctx, u.GetName(), metav1.DeleteOptions{})
			if err != nil && !k8serrors.IsNotFound(err) {
				return fmt.Errorf("failed to delete %s/%s: %w", gvk.Kind, u.GetName(), err)
			}
		}
	} else {
		if dryRun || plan {
			if plan {
				err = c.showDiff(resourceClient, u)
				if err != nil {
					return fmt.Errorf("failed to show diff for %s/%s: %w", gvk.Kind, u.GetName(), err)
				}
			}
		} else {
			_, err = resourceClient.Apply(ctx, u.GetName(), u, metav1.ApplyOptions{FieldManager: "client"})
			if err != nil {
				return fmt.Errorf("failed to apply %s/%s: %w", gvk.Kind, u.GetName(), err)
			}
		}
	}

	return nil
}

func (c *Client) showDiff(resourceClient dynamic.ResourceInterface, expected *unstructured.Unstructured) error {
	actual, err := resourceClient.Get(context.TODO(), expected.GetName(), metav1.GetOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			fmt.Printf("Resource does not exist. It will be created.\n")
			return nil
		}
		return err
	}

	// Remove fields that are not relevant for comparison
	removeFields(actual)
	removeFields(expected)

	if equality.Semantic.DeepEqual(actual, expected) {
		fmt.Printf("No changes detected.\n")
		return nil
	}

	actualYAML, err := runtime.Encode(unstructured.UnstructuredJSONScheme, actual)
	if err != nil {
		return err
	}

	expectedYAML, err := runtime.Encode(unstructured.UnstructuredJSONScheme, expected)
	if err != nil {
		return err
	}

	diffString := diff.Diff(string(actualYAML), string(expectedYAML))

	if diffString == "" {
		fmt.Printf("No differences detected.\n")
		return nil // No differences found
	}
	lines := strings.Split(diffString, "\n")
	for _, line := range lines {
		switch {
		case strings.HasPrefix(line, "+"):
			fmt.Printf("\033[32m%s\033[0m\n", line) // Green for additions
		case strings.HasPrefix(line, "-"):
			fmt.Printf("\033[31m%s\033[0m\n", line) // Red for deletions
		default:
			fmt.Println(line)
		}
	}
	return nil
}

func removeFields(obj *unstructured.Unstructured) {
	unstructured.RemoveNestedField(obj.Object, "metadata", "creationTimestamp")
	unstructured.RemoveNestedField(obj.Object, "metadata", "resourceVersion")
	unstructured.RemoveNestedField(obj.Object, "metadata", "uid")
	unstructured.RemoveNestedField(obj.Object, "metadata", "generation")
	unstructured.RemoveNestedField(obj.Object, "status")
}
