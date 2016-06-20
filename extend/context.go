package extend

import (
	"encoding/json"
	"io"
	"errors"
	
	"github.com/ellcrys/util"
	"github.com/labstack/echo"
	"github.com/labstack/echo/engine"
)

type (

	Context struct {
		context echo.Context
		resWritten 	bool
	}

	H map[string]interface{}
)

func NewContext(ctx echo.Context) *Context {
	return &Context{ context: ctx }
}

// Get request instance from underlying http engine.
func (c *Context) Request() engine.Request {
	return c.context.Request()
}

// Get response instance from underlying http engine
func (c *Context) Response() engine.Response {
	return c.context.Response()
}

// Decode a JSON string to a struct or a map
func (c *Context) BindJSON(container interface{}) error {
	decoder := json.NewDecoder(c.Request().Body())
	return decoder.Decode(container)
}

// Decode a JSON string but limit the size of the json string
func (c *Context) BindJSONLimited(container interface{}, byteLimit int64) error {
	decoder := json.NewDecoder(io.LimitReader(c.Request().Body(), byteLimit))
	return decoder.Decode(container)
}

// Decode a JSON string by reading the request body stream.
// Buffer size limits the size of the data read from the stream.
func (c *Context) BindJSONStream(container interface{}, buffSize int) error {
	var str = ""
	util.ReadReader(c.Request().Body(), buffSize, func(err error, bs []byte, done bool) bool {
		if !done {
			str += string(bs)
		}
		return true
	})
	return json.Unmarshal([]byte(str), container)
} 

// Decodes a JSON string by reading the reading the request body stream but 
// will stop reading after it has read a limited number of bytes
func (c *Context) BindJSONStreamLimited(container interface{}, buffSize int, byteLimit int64) error {
	var str = ""
	var err error
	util.ReadReader(c.Request().Body(), buffSize, func(err error, bs []byte, done bool) bool {
		if !done {
			if int64(len(str)) < byteLimit {
				str += string(bs)
			} else {
				err = errors.New("byte limit reached")
				return false
			}
		}
		return true
	})
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(str), container)
}

// Send json response. data can be a json string, a map or a struct
func (c *Context) JSON(statusCode int, data interface{}) error {
	if c.resWritten {
		return nil
	}
	c.resWritten = true
	return c.context.JSON(statusCode, data)
}

// Send string response
func (c *Context) String(statusCode int, data string) error {
	if c.resWritten {
		return nil
	}

	return c.context.String(statusCode, data)
}

// Get route params
func (c *Context) Param(key string) string {
	return c.context.Param(key)
}

// Get request object.
// Deprecated. Please use c.Request(). 
func (c *Context) Req() engine.Request {
	return c.Request()
}

// Get response writer object.
// Deprecated. Please use c.Response()
func (c *Context) Res() engine.Response {
	return c.Response()
}

// TODO: fetch from request header
func (c *Context) Lang() string {
	return "en"
}

// Get x-signature header value
func (c *Context) GetSignature() string {
	return c.Request().Header().Get("x-signature")
}

// Get 'Authorization' header value
func (c *Context)  GetAuthorization() string {
	return c.Request().Header().Get("authorization")
}

// Set all key/value pair in context parameter
func (c *Context) SetMany(params map[string]string) {
	for k, v := range params {
		c.context.Set(k, v)
	}
}

// Set new data in context
func (c *Context) Set(key, value string) {
	c.context.Set(key, value)
}

// Get data from context
func (c *Context) Get(key string) string {
	return c.context.Get(key).(string)
}

// Returns the echo context
func (c *Context) Echo() echo.Context {
	return c.context
}