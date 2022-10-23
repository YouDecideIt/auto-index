package operations

import (
	"context"
	"errors"
	"fmt"
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var Poll = time.Second * 10
var DefaultTimeout = time.Minute * 10

func WaitForCondition(
	ctx context.Context,
	c client.Client,
	obj client.Object,
	cond func() bool,
	timeout time.Duration,
) error {
	if err := wait.PollImmediate(Poll, timeout, func() (bool, error) {
		key := client.ObjectKeyFromObject(obj)
		if err := c.Get(ctx, key, obj); err != nil {
			if apierrors.IsNotFound(err) {
				return false, nil
			}

			return false, fmt.Errorf("can't get obj %s: %w", key, err)
		}

		if cond() {
			return true, nil
		}

		return false, nil
	}); err != nil {
		if errors.Is(err, wait.ErrWaitTimeout) {
			return fmt.Errorf("wait for object %v condition timeout: %w", obj, err)
		}

		return fmt.Errorf("can't wait for object %v condition, error : %w", obj, err)
	}

	return nil
}
