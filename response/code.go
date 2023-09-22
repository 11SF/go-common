package response

type ResponseCode string

var (
	SuccessCode     ResponseCode = "00000"
	BadRequestCode  ResponseCode = "E0400"
	UnAuthorizeCode ResponseCode = "E0401"
	NotFoundCode    ResponseCode = "E0404"

	GenericError ResponseCode = "E9999"
)
