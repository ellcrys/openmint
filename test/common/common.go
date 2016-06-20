package common

import (
	"net/http/httptest"
	"net/http"
	"time"
	"strings"
	"net/url"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/ellcrys/util"
	"github.com/ellcrys/crypto"
	"github.com/garyburd/redigo/redis"
	"github.com/labstack/echo"
	"github.com/labstack/echo/engine/standard"
)

// create a context to use for testing with controller methods
func NewContext(method, urlPath string, params map[string]string, body string, headers map[string]string) *extend.Context {
	
	e := echo.New()
	r := new(http.Request)
	rec := httptest.NewRecorder()

	// set url
	u, _ := url.Parse(urlPath)
	r.URL = u

	// add header values
	var header = make(http.Header)
	for key, val := range headers {
		header.Set(key, val)
	}

	if util.InStringSlice([]string{"post","put"}, strings.ToLower(method)) {
		header.Set("Content-Type", "application/json")
	}

	r.Header = header

	// create echo context
	req := standard.NewRequest(r, e.Logger())
	echoContext := e.NewContext(req, standard.NewResponse(rec, e.Logger()))
	echoContext.Request().SetMethod(method)
	echoContext.Request().SetBody(strings.NewReader(body))
	c := extend.NewContext(echoContext)
	
	// set parameters
	keys := []string{}
	vals := []string{}
	for k, v := range params {
		keys = append(keys, k)
		vals = append(vals, v)
	}
	echoContext.SetParamNames(keys...)
	echoContext.SetParamValues(vals...)

	return c
}
