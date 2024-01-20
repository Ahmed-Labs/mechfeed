package errors

import "fmt"

type FetchError struct {
    Code    int
    Message string
}

func (e FetchError) Error() string {
    return fmt.Sprintf("Error %d: %s\n", e.Code, e.Message)
}