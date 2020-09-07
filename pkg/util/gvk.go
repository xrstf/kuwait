package util

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func PrettyGVK(gvk schema.GroupVersionKind) string {
	// this usually is nice and obvious enough
	return gvk.Kind

	// if gvk.Group == "" {
	// 	return gvk.Kind
	// }

	// return fmt.Sprintf("%s.%s/%s", gvk.Version, gvk.Group, gvk.Kind)
}
