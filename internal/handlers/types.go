package handlers

import (
	"github.com/go-fuego/fuego"
	"github.com/unshade/citadelle/internal/pagination"
)

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

func NewPaginatedApiResponse[T any](data []T, p pagination.Result, message string) *PaginatedApiResponse[T] {
	return &PaginatedApiResponse[T]{
		Data:    data,
		Message: message,
		Page:    p.Page,
		PerPage: p.PerPage,
		Total:   p.Total,
	}
}

// PaginationParamsFromRequest extracts page/perPage query params and normalizes them.
func PaginationParamsFromRequest[T any](c fuego.Context[T, any]) pagination.Params {
	return pagination.Params{
		Page:    uint64(c.QueryParamInt("page")),
		PerPage: uint64(c.QueryParamInt("perPage")),
	}.Normalize()
}
