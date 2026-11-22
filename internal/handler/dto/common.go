package dto

// SuccessResponse is returned by commands that have no payload.
type SuccessResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

var (
	Updated = SuccessResponse{Code: "0000", Message: "successfully updated"}
	Deleted = SuccessResponse{Code: "0000", Message: "successfully deleted"}
)

// ErrorResponse is the single error envelope for every endpoint.
type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
