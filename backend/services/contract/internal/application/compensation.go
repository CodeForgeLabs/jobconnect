package application

import "fmt"

func wrapCompensationError(primaryErr error, compensationErr error, action string) error {
	if primaryErr == nil {
		return nil
	}
	if compensationErr == nil {
		return primaryErr
	}
	return fmt.Errorf("%v; compensation failed (%s): %v", primaryErr, action, compensationErr)
}
