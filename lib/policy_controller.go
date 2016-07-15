// This controller defines policies that govern
// how a request is treated before entering any main controller method
package lib

import (
	"fmt"
	"strings"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/ellcrys/openmint/config"
	"github.com/ellcrys/openmint/extend"
	"github.com/ellcrys/openmint/models"
	"github.com/ellcrys/util"
	"gopkg.in/mgo.v2"
)

type PolicyController struct {
	mongoSession *mgo.Session
}

// Create a new controller instance
func NewPolicyController(mgoSession *mgo.Session) *PolicyController {
	return &PolicyController{mgoSession}
}

// Authenticate policy.
// Ensures a valid bear/session token is included in the request.
// If token is valid, `auth_user` context data storage will hold
// the authenticated user id.
func (self *PolicyController) Authenticate(c *extend.Context) error {

	authorization := c.GetAuthorization()
	if authorization == "" {
		return config.NewHTTPError(c.Lang(), 401, "e012")
	}

	parts := strings.Split(authorization, " ")
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return config.NewHTTPError(c.Lang(), 401, "e013")
	}

	token, err := jwt.Parse(parts[1], func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(config.C.GetString("hmac_key")), nil
	})

	if err != nil {
		return config.NewHTTPError(c.Lang(), 401, "e014")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return config.NewHTTPError(c.Lang(), 401, "e014")
	}

	iat := claims["iat"].(float64)
	id := claims["id"].(string)

	// check if expired
	iatExpiryTime := util.UnixToTime(util.ToInt64(iat)).AddDate(0, 3, 0).UTC()
	if iatExpiryTime.Before(time.Now().UTC()) {
		return config.NewHTTPError(c.Lang(), 401, "e014")
	}

	// fetch user
	user, err := models.User.FindById(self.mongoSession, id)
	if err != nil {
		if err == mgo.ErrNotFound {
			util.Println("User associated with session token does not exist")
			return config.NewHTTPError(c.Lang(), 401, "e014")
		}
		return config.NewHTTPError(c.Lang(), 500, "e500")
	}

	c.Set("auth_user", user.Id.Hex())

	return nil
}
