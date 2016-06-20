package config

import (
	"fmt"
	"regexp"
	"strings"

	"bitbucket.org/ellcrys/keygenie/extend"
	"github.com/labstack/echo"
)

var errors = map[string]map[string]string {
	"en": map[string]string{
		"e500": "something bad has happened on our end",
		"e001": "could not parse request body",
		"e002": "currency image is required",
		"e003": "currency code is required",
		"e004": "currency code is invalid",
		"e005": "currency denomination is required",
		"e006": "currency denomination is unrecognized",
	},
}

func GetError(lang, code string) string {
	if messages, found := errors[lang]; found {
		for c, msg := range messages {
			r := fmt.Sprintf("(?m)^%s$", c)
			if match, _ := regexp.MatchString(r, code); match {
				return msg
			}
		}
		panic(fmt.Sprintf("message with code %s does not exist", code))
	}
	panic(fmt.Sprintf("messages for lang %s does not exist", lang))
}

type HTTPError struct {
	Msg 		string
	Code 		string
	StatusCode 	int
	Hint 		string
	Param 		string
}

type ErrorData struct {
	Message string 	`json:"message"`
	Code 	string  `json:"code"`
	Hint 	string  `json:"hint,omitempty"`
	Param 	string 	`json:"param,omitempty"`
}

type ErrorObj struct {
	Error ErrorData `json:"error"`
}

func NewHTTPError(lang string, statusCode int, code string) *HTTPError {
	var msg = ""
	if code != "" {
		msg = GetError(lang, code)
	}
	return &HTTPError{ Msg: msg, Code: code, StatusCode: statusCode }
}

func (e *HTTPError) Error() string {
	return e.Msg
}

func (e *HTTPError) SetCode(code string) *HTTPError {
	e.Code = code
	return e
}

func (e *HTTPError) SetMsg(msg string) *HTTPError {
	e.Msg = msg
	return e
}

func (e *HTTPError) SetHint(hint string) *HTTPError {
	e.Hint = hint
	return e
}

func (e *HTTPError) SetParam(param string) *HTTPError {
	e.Param = param
	return e
}

func ValidationError(c *extend.Context, err error) error {
	errTexts := strings.Split(err.Error(), ";")
	msg := GetError(c.Lang(), errTexts[0])
	msgPart := strings.Split(msg, ":")
	return NewHTTPError(c.Lang(), 400, "").SetMsg(msgPart[1]).SetCode("invalid_parameter").SetParam(msgPart[0])
}

func HandleError(e *echo.Echo) {
	e.SetHTTPErrorHandler(func(err error, c echo.Context){

		if err == echo.ErrNotFound {
			c.JSON(404, ErrorObj{
				Error: ErrorData{
					Message: "endpoint not found",
					Code: "e404",
				},
			})
			return
		}

		if err == echo.ErrMethodNotAllowed {
			c.JSON(404, ErrorObj{
				Error: ErrorData{
					Message: "request method not supported",
					Code: "e404",
				},
			})
			return
		}

		if httpErr, ok := err.(*HTTPError); ok {
			c.JSON(httpErr.StatusCode, ErrorObj{
				Error: ErrorData{
					Message: httpErr.Error(),
					Code: httpErr.Code,
					Hint: httpErr.Hint,
					Param: httpErr.Param,
				},
			})
			return 
		}

		// unhandled error
		fmt.Println(err)
		c.JSON(404, ErrorObj{
			Error: ErrorData{
				Message: "something bad has happened on our end",
				Code: "e500",
			},
		})
		return
	})
}