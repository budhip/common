package error

import "google.golang.org/grpc/codes"

type ServiceError struct {
	Status     codes.Code
	Code       string
	Message    string
	Attributes map[string]string
}

func (e ServiceError) Error() string {
	return e.Message
}