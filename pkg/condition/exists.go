package condition

import (
	"context"
	"fmt"

	"go.xrstf.de/kuwait/pkg/util"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type existsCondition struct {
	baseCondition
}

func (c *existsCondition) Satisfied(ctx context.Context) (bool, error) {
	if c.target.Name == "*" {
		// this only ensures that at least one item exists
		list, err := c.dynamicInterface.List(ctx, v1.ListOptions{Limit: 1})
		if err != nil {
			if errors.IsNotFound(err) {
				return false, nil
			}

			return false, err
		}

		return len(list.Items) > 0, nil
	}

	_, err := c.dynamicInterface.Get(ctx, c.target.Name, v1.GetOptions{})
	if err != nil && !errors.IsNotFound(err) {
		return false, err
	}

	return err == nil, nil
}

func (c *existsCondition) String() string {
	ns := c.target.Namespace
	name := c.target.Name
	gvk := util.PrettyGVK(c.gvk)

	if ns == "" {
		if name == "*" {
			return fmt.Sprintf("Any %s resource must exist in the cluster.", gvk)
		}

		return fmt.Sprintf("The %s %q must exist in the cluster.", gvk, name)
	}

	if name == "*" {
		return fmt.Sprintf("Any %s resource must exist in the %q namespace.", gvk, ns)
	}

	return fmt.Sprintf("The %s \"%s/%s\" must exist.", gvk, ns, name)
}
