package types

import "fmt"

type RetryableExecutionError string

func (e RetryableExecutionError) Error() string {
	return fmt.Sprint(string(e))
}
