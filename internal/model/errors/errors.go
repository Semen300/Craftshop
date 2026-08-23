package errors

type AccessRejectedError struct{}

func (AccessRejectedError) Error() string {
	return "Access rejected: attempt to access data, which not been assigned to user"
}
