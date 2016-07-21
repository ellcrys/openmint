package config

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/ellcrys/openmint/extend"
	"github.com/labstack/echo"
)

var errors = map[string]map[string]string{
	"en": map[string]string{
		"e500": "something bad has happened on our end",
		"e001": "could not parse request body",
		"e002": "currency image is required",
		"e003": "currency code is required",
		"e004": "currency code is invalid",
		"e005": "currency denomination is unrecognized",
		"e006": "email is not available",
		"e007": "email and password are required",
		"e008": "user email or password are invalid",
		"e009": "currency code is not currently supported",
		"e010": "invalid facebook user token",
		"e011": "user not found",
		"e012": "authorization header is required",
		"e013": "authorization header format is invalid",
		"e014": "bearer token is invalid",
		"e015": "image is not a currency",
		"e016": "Failed to read serial on currency",
		"e017": "currency has been indexed",
		"e018": "no currency in queue",
		"e019": "session is not active",
		"e020": "currency not found",
		"e021": "currency has enough votes",
		"e022": "vote session not active",
		"e023": "user already added a vote",

		"Fullname: non zero.*":    "full_name:fullname is required",
		"Email: non zero.*":       "email:email is required",
		"Email:.*as email":        "email:email is not valid",
		"PhotoURL: non zero.*":    "photo_url:photo_url is required",
		"Provider: non zero.*":    "provider:provider is required",
		"ProviderId: non zero.*":  "provider_id:provider_id is required",
		"AccessToken: non zero.*": "user_token:user_token is required",
		"Password: non zero.*":    "password:password is required",
		".*isValidPassword":       "password:password must have atleast 6 characters",
		"CurrencyId: non zero.*":  "currency_id:currency id is required",
		"VoteId: non zero.*":      "vote_id:vote id is required",
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
	Msg        string
	Code       string
	StatusCode int
	Hint       string
	Param      string
}

type ErrorData struct {
	Message string `json:"message"`
	Code    string `json:"code"`
	Hint    string `json:"hint,omitempty"`
	Param   string `json:"param,omitempty"`
}

type ErrorObj struct {
	Error ErrorData `json:"error"`
}

func NewHTTPError(lang string, statusCode int, code string) *HTTPError {
	var msg = ""
	if code != "" {
		msg = GetError(lang, code)
	}
	return &HTTPError{Msg: msg, Code: code, StatusCode: statusCode}
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
	e.SetHTTPErrorHandler(func(err error, c echo.Context) {

		if err == echo.ErrNotFound {
			c.JSON(404, ErrorObj{
				Error: ErrorData{
					Message: "endpoint not found",
					Code:    "e404",
				},
			})
			return
		}

		if err == echo.ErrMethodNotAllowed {
			c.JSON(404, ErrorObj{
				Error: ErrorData{
					Message: "request method not supported",
					Code:    "e404",
				},
			})
			return
		}

		if httpErr, ok := err.(*HTTPError); ok {
			c.JSON(httpErr.StatusCode, ErrorObj{
				Error: ErrorData{
					Message: httpErr.Error(),
					Code:    httpErr.Code,
					Hint:    httpErr.Hint,
					Param:   httpErr.Param,
				},
			})
			return
		}

		// unhandled error
		fmt.Println(err)
		c.JSON(404, ErrorObj{
			Error: ErrorData{
				Message: "something bad has happened on our end",
				Code:    "e500",
			},
		})
		return
	})
}
