package condition

import (
	"context"
)

type Condition interface {
	String() string
	Satisfied(ctx context.Context) (bool, error)
}
