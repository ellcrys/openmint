// This controller contains address related actions
package lib

import (
	"gopkg.in/mgo.v2"
	"github.com/ellcrys/openmint/models"
	"github.com/ellcrys/openmint/config"
	"github.com/asaskevich/govalidator"
	"github.com/ellcrys/openmint/extend"
	"golang.org/x/crypto/bcrypt"
)

type UserController struct {
	mongoSession 	*mgo.Session
}

func NewUserController(mongoSession *mgo.Session) *UserController {
	return &UserController{ mongoSession }
}


// API: 			POST /v1/users
// Description: 	Create a new cloud mint user.
// Content-Type: 	application/json
// Body Params: 	full_name {string}, email {string}, password {string}
// Response 200: 	id {string}, full_name {string}, email {string}, created_at {Date}
func (self *UserController) Create(c *extend.Context) error {

	// parse request body
	var body models.UserModel
	if c.BindJSON(&body) != nil {
		return config.NewHTTPError(c.Lang(), 400, "e001")
	}

	// validate request body
	body.MinPasswordLength = 6
	_, err := govalidator.ValidateStruct(body)
	if err != nil {
	    return config.ValidationError(c, err)
	}

	// find existing user with matching email
	_, err = models.User.FindByField(self.mongoSession, "email", body.Email)
	if err == nil {
		return config.NewHTTPError(c.Lang(), 400, "e006")
	} else if err != nil  && err != mgo.ErrNotFound {
		return config.NewHTTPError(c.Lang(), 500, "e500")
	}
	
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.DefaultCost)
	if err != nil {
		return config.NewHTTPError(c.Lang(), 400, "e500")
	}

	body.Id = models.NewId()
	body.Password = string(hashedPassword)
	if err = models.User.Create(self.mongoSession, &body); err != nil {
		return config.NewHTTPError(c.Lang(), 500, "e500")
	}

	body.Password = ""
	return c.JSON(201, body)
}
