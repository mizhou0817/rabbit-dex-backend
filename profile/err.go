package profile

type TarantoolError struct {
	error
}

func (e TarantoolError) Error() string {
	return e.error.Error()
}
