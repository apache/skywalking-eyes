package header

import (
	"fmt"
	"strings"
)

type Result struct {
	Success []string
	Failure []string
	Ignored []string
	Fixed   []string
}

func (result *Result) Fail(file string) {
	result.Failure = append(result.Failure, file)
}

func (result *Result) Succeed(file string) {
	result.Success = append(result.Success, file)
}

func (result *Result) Ignore(file string) {
	result.Ignored = append(result.Ignored, file)
}

func (result *Result) Fix(file string) {
	result.Fixed = append(result.Fixed, file)
}

func (result *Result) HasFailure() bool {
	return len(result.Failure) > 0
}

func (result *Result) Error() error {
	return fmt.Errorf(
		"the following files don't have a valid license header: \n%v",
		strings.Join(result.Failure, "\n"),
	)
}

func (result *Result) String() string {
	return fmt.Sprintf(
		"Totally checked %d files, valid: %d, invalid: %d, ignored: %d, fixed: %d",
		len(result.Success)+len(result.Failure)+len(result.Ignored),
		len(result.Success),
		len(result.Failure),
		len(result.Ignored),
		len(result.Fixed),
	)
}
