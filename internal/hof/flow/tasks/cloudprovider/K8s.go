package cloudprovider

import (
	"context"
	"fmt"
	"path"

	"cuelang.org/go/cue"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"

	hofcontext "github.com/opentofu/opentofu/internal/hof/flow/context"
	"github.com/opentofu/opentofu/internal/hof/flow/tasks/kubernetes"
	"github.com/opentofu/opentofu/internal/hof/lib/mantis"
)

type K8sTask struct{}

func NewK8sTask(val cue.Value) (hofcontext.Runner, error) {
	return &K8sTask{}, nil
}

func (t *K8sTask) Run(ctx *hofcontext.Context) (any, error) {
	v := ctx.Value

	configValue := v.LookupPath(cue.ParsePath("config"))

	client, err := kubernetes.NewClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create Kubernetes client: %v", err)
	}

	resources, err := queryK8sResources(client, configValue)
	if err != nil {
		return nil, fmt.Errorf("failed to query Kubernetes resources: %v", err)
	}

	newV := v.FillPath(cue.ParsePath(mantis.MantisTaskOuts), resources)
	return newV, nil
}

func queryK8sResources(client *kubernetes.Client, config cue.Value) (map[string]interface{}, error) {
	resourceTypeStr, err := config.LookupPath(cue.ParsePath("resourceType")).String()
	if err != nil {
		return nil, fmt.Errorf("failed to get resource type: %v", err)
	}

	filters := struct {
		Namespace string `json:"namespace"`
		Name      string `json:"name"`
		Label     string `json:"label"`
	}{}

	if filtersValue := config.LookupPath(cue.ParsePath("filters")); filtersValue.Exists() {
		if err := filtersValue.Decode(&filters); err != nil {
			return nil, fmt.Errorf("failed to decode filters: %v", err)
		}
	}

	if filters.Namespace == "" {
		filters.Namespace = "default"
	}

	labelSelector := filters.Label

	gvr, err := getGVRForResource(resourceTypeStr)
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	var resourceList *unstructured.UnstructuredList

	dynamicClient, err := dynamic.NewForConfig(client.GetConfig())
	if err != nil {
		return nil, fmt.Errorf("failed to create dynamic client: %v", err)
	}

	if gvr.Resource == "namespaces" {
		resourceList, err = dynamicClient.Resource(gvr).List(ctx, metav1.ListOptions{
			LabelSelector: labelSelector,
		})
	} else {
		resourceList, err = dynamicClient.Resource(gvr).Namespace(filters.Namespace).List(ctx, metav1.ListOptions{
			LabelSelector: labelSelector,
		})
	}

	if err != nil {
		return nil, fmt.Errorf("failed to list resources: %v", err)
	}

	var filteredItems []unstructured.Unstructured
	for _, item := range resourceList.Items {
		if filters.Name != "" {
			matched, _ := path.Match(filters.Name, item.GetName())
			if !matched {
				continue
			}
		}
		filteredItems = append(filteredItems, item)
	}

	return map[string]interface{}{resourceTypeStr: filteredItems}, nil
}

func getGVRForResource(resourceType string) (schema.GroupVersionResource, error) {
	switch resourceType {
	case "pods":
		return schema.GroupVersionResource{Version: "v1", Resource: "pods"}, nil
	case "deployments":
		return schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"}, nil
	case "services":
		return schema.GroupVersionResource{Version: "v1", Resource: "services"}, nil
	case "configmaps":
		return schema.GroupVersionResource{Version: "v1", Resource: "configmaps"}, nil
	case "secrets":
		return schema.GroupVersionResource{Version: "v1", Resource: "secrets"}, nil
	case "namespaces":
		return schema.GroupVersionResource{Version: "v1", Resource: "namespaces"}, nil
	default:
		return schema.GroupVersionResource{}, fmt.Errorf("unsupported resource type: %s", resourceType)
	}
}
