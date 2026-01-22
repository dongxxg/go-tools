package util

import "github.com/pkg/errors"

func NewError(errs []error) error {
	if len(errs) == 0 {
		return nil
	}
	err := errs[0]

	for i := 1; i < len(errs); i++ {
		err = errors.Wrap(err, errs[i].Error())
	}

	return err
}
