package schema

import "fmt"

func prefixErrors(prefix string, errs []error) []error {
	var res []error
	for _, e := range errs {
		res = append(res, fmt.Errorf("%s: %v", prefix, e))
	}
	return res
}
