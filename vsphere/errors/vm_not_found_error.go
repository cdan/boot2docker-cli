package errors

type VmNotFoundError struct {
}

func NewVmNotFoundError() error {
	err := VmNotFoundError{}
	return &err
}

func (err *VmNotFoundError) Error() string {
	return "No docker host found"
}
