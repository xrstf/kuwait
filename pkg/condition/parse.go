package condition

import (
	"errors"
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
)

func Parse(condition string, clientset *kubernetes.Clientset, dynamicClient dynamic.Interface, mapper meta.RESTMapper) (Condition, error) {
	parts := strings.Split(condition, "/")
	if len(parts) < 2 {
		return nil, errors.New("must be in `kind/name[/condition=ready]` (for non-namespaced resouces) or `kind/namespace/name[/condition=ready]` format")
	}

	kind := strings.ToLower(parts[0])

	// parts[1] is only used for the namespace if the resource is actually namespaced
	gvk, namespaced, dynamicInterface, err := getDynamicInterface(kind, parts[1], dynamicClient, mapper)
	if err != nil {
		return nil, fmt.Errorf("invalid kind %q: %v", kind, err)
	}

	cond := "ready"
	target := types.NamespacedName{}

	if namespaced {
		if len(parts) < 3 {
			return nil, fmt.Errorf("must be in `%s/namespace/name[/condition=ready]` format", kind)
		}

		target.Namespace = strings.TrimSpace(parts[1])
		target.Name = strings.TrimSpace(parts[2])

		if len(target.Namespace) == 0 {
			return nil, fmt.Errorf("%q is a namespaced resource, so a namespace must be given in the condition", kind)
		}

		if len(parts) >= 4 {
			cond = parts[3]
		}
	} else {
		if len(parts) < 2 {
			return nil, fmt.Errorf("must be in `%s/name[/condition=ready]` format", kind)
		}

		target.Name = parts[1]

		if len(parts) >= 3 {
			cond = parts[2]
		}
	}

	base := baseCondition{
		gvk:              gvk,
		clientset:        clientset,
		dynamicInterface: dynamicInterface,
		target:           target,
	}

	switch strings.ToLower(cond) {
	case "exist":
		fallthrough
	case "exists":
		return &existsCondition{
			baseCondition: base,
		}, nil

	case "removed":
		fallthrough
	case "gone":
		return &goneCondition{
			baseCondition: base,
		}, nil

	case "ready":
		return &readyCondition{
			baseCondition: base,
		}, nil
	}

	return nil, fmt.Errorf("unknown condition %q given", cond)
}

func getDynamicInterface(resource string, namespace string, dynamicClient dynamic.Interface, mapper meta.RESTMapper) (schema.GroupVersionKind, bool, dynamic.ResourceInterface, error) {
	gvk, err := mapper.KindFor(schema.GroupVersionResource{
		Resource: resource,
	})
	if err != nil {
		return schema.GroupVersionKind{}, false, nil, fmt.Errorf("failed to resolve %q: %v", resource, err)
	}

	mapping, err := mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return gvk, false, nil, fmt.Errorf("failed to determine mapping for %q: %v", resource, err)
	}

	namespaced := mapping.Scope.Name() == meta.RESTScopeNameNamespace

	var dr dynamic.ResourceInterface
	if namespaced {
		// namespaced resources should specify the namespace
		dr = dynamicClient.Resource(mapping.Resource).Namespace(namespace)
	} else {
		// for cluster-wide resources
		dr = dynamicClient.Resource(mapping.Resource)
	}

	return gvk, namespaced, dr, nil
}
