package handlers

type ApiResponse[T any] struct {
	Data    *T     `json:"data"`
	Message string `json:"message"`
}

func NewApiResponse[T any](data *T, message string) (*ApiResponse[T], error) {
	return &ApiResponse[T]{Data: data, Message: message}, nil
}

func NewErrorResponse[T any](err error) (*ApiResponse[T], error) {
	return nil, err
}
