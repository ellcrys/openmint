// Special Response Errors
package response 

import (
	"strings"
	"github.com/ellcrys/openmint/config"
	"github.com/ellcrys/openmint/extend"
	"github.com/ellcrys/util"
)

type Err struct {
	Message string 		`json:"message"`
	Code 	string 		`json:"code,omitempty"`
	Hint 	interface{}	`json:"hint,omitempty"`
	Param   interface{} `json:"param,omitempty"`
}

type ErrorResp struct {
	Error Err 			`json:"error"`
}

type CustomResponse struct {
}

var customResponse *CustomResponse

func init() {
	customResponse = &CustomResponse{}
}

// Given a martini context, it will write a JSON response
func (cr *CustomResponse) JSONError(c *extend.Context) {
	lang := "en"
	c.JSON(400, extend.H{
		"error": extend.H{
			"message": config.GetError(lang, "e001"),
			"code": "e001",
		},
	})
}

// It will write a response describing a validation failure
func (cr *CustomResponse) ValidationError(c *extend.Context, err error) {
	lang := "en"
	errTexts := strings.Split(err.Error(), ";")
	msg := config.GetError(lang, errTexts[0])
	msgPart := strings.Split(msg, ":")
	c.JSON(400, extend.H{
		"error": extend.H{
			"message": msgPart[1],
			"code": "invalid_parameter",
			"param": msgPart[0],
		},
	})
}

// It will write a response describing an error. The error messages
// is located in the /config/errors.go and accessible via their associated keys
func (cr *CustomResponse) Error(c *extend.Context, statusCode int, code string) {
	c.JSON(statusCode, ErrorResp{
		Error: Err {
			Message: config.GetError(c.Lang(), code),
			Code: code,
		},
	})
}

// It will write a response describing an error with an explicit error message provided. 
func (cr *CustomResponse) ErrorMsg(c *extend.Context, statusCode int, msg, code string) {
	c.JSON(statusCode, ErrorResp{
		Error: Err {
			Message: msg,
			Code: code,
		},
	})
}

// Respond with an error object that has a hint field specified
func (cr *CustomResponse) ErrorMsgWithHint(c *extend.Context, statusCode int, msg, code string, hint interface{}) {
	c.JSON(statusCode, ErrorResp{
		Error: Err {
			Message: msg,
			Code: code,
			Hint: hint,		
		},
	})
}

// Respond with an error object that has a param field specified
func (cr *CustomResponse) ErrorMsgWithParam(c *extend.Context, statusCode int, msg, code string, param interface{}) {
	c.JSON(statusCode, ErrorResp{
		Error: Err {
			Message: msg,
			Code: code,
			Param: param,
		},
	})
}

// for when JSON request body could not be parsed
func JSONError(c *extend.Context) {
	customResponse.JSONError(c)
}

// validation error
func ValidationError(c *extend.Context, err error) {
	customResponse.ValidationError(c, err)
}

// regular error
func Error(c *extend.Context, statusCode int, code string) {
	customResponse.Error(c, statusCode, code)
}

// regular error with message passed directly
func ErrorMsg(c *extend.Context, statusCode int, msg, code string) {
	customResponse.ErrorMsg(c, statusCode, msg, code)
}

// regular error with message passed directly with a hint included
func ErrorMsgWithHint(c *extend.Context, statusCode int, msg, code string, hint interface{}) {
	customResponse.ErrorMsgWithHint(c, statusCode, msg, code, hint)
}

// regular error with message passed directly with a param included
func ErrorMsgWithParam(c *extend.Context, statusCode int, msg, code string, param interface{}) {
	customResponse.ErrorMsgWithParam(c, statusCode, msg, code, param)
}

// Log message (err, info).
// Todo: Send to an error management service?
func Log(v interface{}) *CustomResponse {
	util.Println(v)
	return customResponse
}

