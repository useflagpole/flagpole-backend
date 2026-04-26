package response

type APIResponse interface {
	isAPIResponse()
}

type DataResponse struct {
	Data any `json:"data"`
}

func (DataResponse) isAPIResponse() {}

type ErrorResponse struct {
	Error string `json:"error"`
}

func (ErrorResponse) isAPIResponse() {}

type ConflictResponse struct {
	Fields []string `json:"fields"`
}

func (ConflictResponse) isAPIResponse() {}

var (
	Error500              = ErrorResponse{Error: "Internal server error"}
	ErrInvalidCredentials = ErrorResponse{Error: "Invalid credentials"}
	ErrEmailTaken         = ErrorResponse{Error: "Email is already taken"}
	ErrOrgLimitReached    = ErrorResponse{Error: "Organization limit reached"}
)
