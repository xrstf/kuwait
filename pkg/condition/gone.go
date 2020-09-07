package condition

import (
	"context"
	"fmt"

	"go.xrstf.de/kuwait/pkg/util"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type goneCondition struct {
	baseCondition
}

func (c *goneCondition) Satisfied(ctx context.Context) (bool, error) {
	if c.target.Name == "*" {
		list, err := c.dynamicInterface.List(ctx, v1.ListOptions{Limit: 1})
		if err != nil {
			if errors.IsNotFound(err) {
				return true, nil
			}

			return false, err
		}

		return len(list.Items) == 0, nil
	}

	_, err := c.dynamicInterface.Get(ctx, c.target.Name, v1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return true, nil
		}

		return false, err
	}

	return false, nil
}

func (c *goneCondition) String() string {
	ns := c.target.Namespace
	name := c.target.Name
	gvk := util.PrettyGVK(c.gvk)

	if ns == "" {
		if name == "*" {
			return fmt.Sprintf("No %s resource must exist anymore in the cluster.", gvk)
		}

		return fmt.Sprintf("The %s %q must be be gone.", gvk, name)
	}

	if name == "*" {
		return fmt.Sprintf("No %s resource must exist in the %q namespace.", gvk, ns)
	}

	return fmt.Sprintf("The %s \"%s/%s\" must be gone.", gvk, ns, name)
}
