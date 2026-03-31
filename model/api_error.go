package model

type APIErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type APIErrorResponse struct {
	Error APIErrorBody `json:"error"`
}
