package dto

type APIResponse[T any] struct {
	Success bool   `json:"success"`
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    T      `json:"data"`
}

// SwaggerDocResponse is a non-generic response type used solely for Swagger/OpenAPI documentation.
// swag cannot introspect generic types like APIResponse[T], so this concrete type is defined.
type SwaggerDocResponse struct {
	Success bool   `json:"success" example:"true"`
	Code    int    `json:"code"    example:"200"`
	Message string `json:"message" example:"success"`
	Data    any    `json:"data"`
}
