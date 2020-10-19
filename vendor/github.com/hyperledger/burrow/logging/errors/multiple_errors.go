package errors

import "strings"

type MultipleErrors []error

func CombineErrors(errs []error) error {
	switch len(errs) {
	case 0:
		return nil
	case 1:
		return errs[0]
	default:
		return MultipleErrors(errs)
	}
}

func (errs MultipleErrors) Error() string {
	var errStrings []string
	for _, err := range errs {
		errStrings = append(errStrings, err.Error())
	}
	return strings.Join(errStrings, ";")
}
