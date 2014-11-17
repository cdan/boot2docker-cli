package errors

import "fmt"

type VmError struct {
	vm        string
	operation string
	reason    string
}

func NewVmError(vm, operation, reason string) error {
	err := VmError{
		vm:        vm,
		operation: operation,
		reason:    reason,
	}
	return &err
}

func (err *VmError) Error() string {
	return fmt.Sprintf("Unable to %s virtual machine %s due to %s", err.operation, err.vm, err.reason)
}
