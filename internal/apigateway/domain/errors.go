package domain

import "errors"

var (
	ErrServiceNotFound    = errors.New("service not found")
	ErrServiceUnavailable = errors.New("service unavailable")
	ErrRouteNotFound      = errors.New("route not found")
)
