package condition

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"go.xrstf.de/kuwait/pkg/util"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type readyCondition struct {
	baseCondition
}

func (c *readyCondition) Satisfied(ctx context.Context) (bool, error) {
	if c.target.Name == "*" {
		list, err := c.dynamicInterface.List(ctx, v1.ListOptions{})
		if err != nil {
			if errors.IsNotFound(err) {
				return false, nil
			}

			return false, err
		}

		// ensure all items are ready
		for _, item := range list.Items {
			if !c.isReady(item) {
				return false, nil
			}
		}

		return true, nil
	}

	obj, err := c.dynamicInterface.Get(ctx, c.target.Name, v1.GetOptions{})
	if err != nil && !errors.IsNotFound(err) {
		return false, err
	}

	return err == nil && c.isReady(*obj), nil
}

type objectWithConditions struct {
	Status struct {
		Conditions []struct {
			Status string `json:"status"`
			Type   string `json:"type"`
		} `json:"conditions"`
	} `json:"status"`
}

func (o *objectWithConditions) hasCondition(name string, status string) bool {
	for _, cond := range o.Status.Conditions {
		if strings.ToLower(cond.Type) == strings.ToLower(name) {
			return strings.ToLower(cond.Status) == strings.ToLower(status)
		}
	}

	return false
}

func (c *readyCondition) isReady(obj unstructured.Unstructured) bool {
	encoded, err := obj.MarshalJSON()
	if err != nil {
		return false
	}

	withConditions := objectWithConditions{}
	if err := json.Unmarshal(encoded, &withConditions); err != nil {
		return false
	}

	if withConditions.hasCondition("Ready", "true") {
		return true
	}

	// deployments
	if withConditions.hasCondition("Available", "true") {
		return true
	}

	return false
}

func (c *readyCondition) String() string {
	ns := c.target.Namespace
	name := c.target.Name
	gvk := util.PrettyGVK(c.gvk)

	if ns == "" {
		if name == "*" {
			return fmt.Sprintf("All %s resources in the cluster must be ready.", gvk)
		}

		return fmt.Sprintf("The %s %q must be ready.", gvk, name)
	}

	if name == "*" {
		return fmt.Sprintf("All %s resources in the %q namespace must be ready.", gvk, ns)
	}

	return fmt.Sprintf("The %s \"%s/%s\" must be ready.", gvk, ns, name)
}
