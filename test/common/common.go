package common

import (
	"gopkg.in/mgo.v2"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"time"

	"github.com/ellcrys/openmint/extend"
	"github.com/ellcrys/openmint/models"
	"github.com/ellcrys/openmint/www"
	"github.com/ellcrys/util"
	"github.com/labstack/echo"
	"github.com/labstack/echo/engine/standard"
	"golang.org/x/crypto/bcrypt"
)

var MongoSes *mgo.Session

// Initialize package by setting
// up the application and a test mongo database
func InitTestPackage() {
	_, MongoSes = www.App(true, false)
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

	if util.InStringSlice([]string{"post", "put"}, strings.ToLower(method)) {
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

// directly create a user in the test database
func CreateTestUser() (*models.UserModel, error) {

	newUser := &models.UserModel{
		Id:       models.NewId(),
		Fullname: util.RandString(10),
		Email:    util.RandString(10) + "@example.com",
		Password: util.RandString(10),
	}

	unHashedPassword := newUser.Password
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(newUser.Password), bcrypt.DefaultCost)
	newUser.Password = string(hashedPassword)

	if err := models.User.Create(MongoSes, newUser); err != nil {
		return nil, err
	}

	newUser.Password = unHashedPassword
	return newUser, nil
}
