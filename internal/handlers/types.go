package handlers

import "github.com/go-fuego/fuego"

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

type PaginatedApiResponse[T any] struct {
	Data    []T    `json:"data"`
	Message string `json:"message"`
	Page    uint64 `json:"page"`
	PerPage uint64 `json:"perPage"`
	Total   uint64 `json:"total"`
}

func NewPaginatedApiResponse[T any](
	data []T,
	message string,
) *PaginatedApiResponse[T] {
	return &PaginatedApiResponse[T]{
		Data:    data,
		Message: message,
	}
}

func PaginationParamsFromRequest[T any](c fuego.Context[T, any]) PaginationParams {
	page := c.QueryParamInt("page")
	perPage := c.QueryParamInt("perPage")
	return PaginationParams{
		Page:    uint64(page),
		PerPage: uint64(perPage),
	}
}

type PaginationParams struct {
	Page    uint64
	PerPage uint64
}

type Pagination struct {
	PaginationParams
	Total uint64
}
