package kubernetes

import (
	"context"
	"fmt"

	"gopkg.in/yaml.v3" // YAML v3 library for decoding documents
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"
)

type Client struct {
	clientset *kubernetes.Clientset
	dynamic   dynamic.Interface
	mapper    *restmapper.DeferredDiscoveryRESTMapper
}

func NewClient() (*Client, error) {
	config, err := clientcmd.BuildConfigFromFlags("", clientcmd.RecommendedHomeFile)
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

func (c *Client) Apply(manifests []byte) error {
	return c.applyOrDelete(manifests, false, false)
}

func (c *Client) Delete(manifests []byte) error {
	return c.applyOrDelete(manifests, true, false)
}

func (c *Client) Plan(manifests []byte) error {
	return c.applyOrDelete(manifests, false, true)
}

func (c *Client) applyOrDelete(manifests []byte, delete bool, dryRun bool) error {
	var objects []map[string]interface{}
	err := yaml.Unmarshal(manifests, &objects)
	if err != nil {
		return fmt.Errorf("failed to unmarshal manifests: %w", err)
	}

	for _, obj := range objects {
		u := &unstructured.Unstructured{Object: obj}
		gvk := u.GroupVersionKind()
		gvr, _ := schema.ParseResourceArg(gvk.GroupVersion().String() + "." + gvk.Kind)

		namespace := u.GetNamespace()
		if namespace == "" {
			namespace = "default"
		}

		resourceClient := c.dynamic.Resource(*gvr).Namespace(namespace)

		ctx := context.Background()

		if delete {
			if dryRun {
				fmt.Printf("Would delete %s/%s (dry run)\n", gvk.Kind, u.GetName())
			} else {
				err = resourceClient.Delete(ctx, u.GetName(), metav1.DeleteOptions{})
				if err != nil && !k8serrors.IsNotFound(err) {
					return fmt.Errorf("failed to delete %s/%s: %w", gvk.Kind, u.GetName(), err)
				}
				fmt.Printf("Deleted %s/%s\n", gvk.Kind, u.GetName())
			}
		} else {
			if dryRun {
				fmt.Printf("Would apply %s/%s (dry run)\n", gvk.Kind, u.GetName())
			} else {
				_, err = resourceClient.Apply(ctx, u.GetName(), u, metav1.ApplyOptions{FieldManager: "client"})
				if err != nil {
					return fmt.Errorf("failed to apply %s/%s: %w", gvk.Kind, u.GetName(), err)
				}
				fmt.Printf("Applied %s/%s\n", gvk.Kind, u.GetName())
			}
		}
	}

	return nil
}
