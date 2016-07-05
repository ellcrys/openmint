// This controller contains address related actions
package lib

import (
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/ellcrys/openmint/config"
	"github.com/ellcrys/openmint/extend"
	"github.com/ellcrys/openmint/models"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/mgo.v2"
	"time"
)

type AuthController struct {
	mongoSession *mgo.Session
}

func NewAuthController(mongoSession *mgo.Session) *AuthController {
	return &AuthController{mongoSession}
}

// API: 			POST /v1/auth/login
// Description: 	Authenticate a cloud mint user and return a session token
// Content-Type: 	application/json
// Body Params: 	email {string}, password {string}
// Response 200: 	id {string}, full_name {string}, email {string}, created_at {Date}, token {string}
func (self *AuthController) UserAuth(c *extend.Context) error {

	var body models.UserModel
	if c.BindJSON(&body) != nil {
		return config.NewHTTPError(c.Lang(), 400, "e001")
	}

	if body.Email == "" || body.Password == "" {
		return config.NewHTTPError(c.Lang(), 400, "e007")
	}

	// find user with matching email
	user, err := models.User.FindByField(self.mongoSession, "email", body.Email)
	if err != nil && err == mgo.ErrNotFound {
		return config.NewHTTPError(c.Lang(), 400, "e008")
	} else if err != nil {
		return config.NewHTTPError(c.Lang(), 500, "e500")
	}

	if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(body.Password)) != nil {
		return config.NewHTTPError(c.Lang(), 400, "e008")
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":  body.Id.Hex(),
		"iat": time.Now().Unix(),
	})

	user.TokenString, err = token.SignedString([]byte(config.C.GetString("hmac_key")))
	if err != nil {
		return config.NewHTTPError(c.Lang(), 500, "e500")
	}

	return c.JSON(200, user)
}
