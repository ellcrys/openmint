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
	"github.com/ellcrys/ellcrys/www"
	"github.com/ellcrys/ellcrys/extend"
	"github.com/ellcrys/util"
	"github.com/ellcrys/crypto"
	"github.com/garyburd/redigo/redis"
	"github.com/labstack/echo"
	"github.com/labstack/echo/engine/standard"
)

var MongoSes *mgo.Session
var RedisPool *redis.Pool

// Initialize package by setting
// up the application and a test mongo database
func InitTestPackage() {
	_, MongoSes, RedisPool = www.App(true, true)
}

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

// Get test database session
func ConnectToTestMongoSession() (*mgo.Session, error) {
	return mgo.DialWithInfo(&mgo.DialInfo{
		Addrs:    []string{util.Env("MONGO_DB_HOST", "localhost:27017")},
		Timeout:  5 * time.Second,
		Database: util.Env("MONGO_TEST_DB_NAME", "db_name"),
		Username: util.Env("MONGO_USERNAME", ""),
		Password: util.Env("MONGO_PASSWORD", ""),
	})
}

// Get mongo session
func GetMongoSession() *mgo.Session {
	return MongoSes
}


