// This controller contains address related actions
package lib

import (
	"strconv"

	"github.com/ellcrys/openmint/config"
	"github.com/ellcrys/openmint/extend"
	"github.com/ellcrys/openmint/models"
	"github.com/ellcrys/util"
	"gopkg.in/mgo.v2"
)

type UserController struct {
	mongoSession *mgo.Session
}

func NewUserController(mongoSession *mgo.Session) *UserController {
	return &UserController{mongoSession}
}

// @API: GET /v1/users/currencies
// @Description: Get currencies belonging to the authenticated user
func (self *UserController) GetCurrencies(c *extend.Context) error {

	var authUserId = c.Get("auth_user")
	var err error
	var skip = 0
	var limit = 10

	if _limit := c.Echo().QueryParam("limit"); _limit != "" {
		limit, err = strconv.Atoi(_limit)
		if err != nil {
			return config.NewHTTPError(c.Lang(), 500, "e500")
		}
	}

	if _skip := c.Echo().QueryParam("skip"); _skip != "" {
		skip, err = strconv.Atoi(_skip)
		if err != nil {
			return config.NewHTTPError(c.Lang(), 500, "e500")
		}
	}

	skip = skip * limit

	currencies, err := models.Currency.FindWithDateSortAndSkip(self.mongoSession, authUserId, "-created_at", limit, skip)
	if err != nil {
		util.Println("Failed to fetch currencies. ", err.Error())
		return config.NewHTTPError(c.Lang(), 500, "e500")
	}

	return c.JSON(200, currencies)
}
