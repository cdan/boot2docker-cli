package errors

import (
	original "errors"
	"fmt"
)

func New(message string) error {
	return original.New(message)
}

func NewWithFmt(message string, args ...interface{}) error {
	return original.New(fmt.Sprintf(message, args...))
}

func NewWithError(message string, err error) error {
	return NewWithFmt("%s: %s", message, err.Error())
}
