package condition

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
)

type baseCondition struct {
	gvk              schema.GroupVersionKind
	clientset        *kubernetes.Clientset
	dynamicInterface dynamic.ResourceInterface
	target           types.NamespacedName
}
