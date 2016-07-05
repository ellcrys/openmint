package integration

import (
	"testing"

	"github.com/ellcrys/openmint/config"
	"github.com/ellcrys/openmint/test/common"
	"gopkg.in/mgo.v2"
)

func ClearDBCollection(colname string) {
	common.MongoSes.SetMode(mgo.Monotonic, true)
	c := common.MongoSes.DB(config.C.GetString("mongo_database")).C(config.C.GetString(colname))
	c.DropCollection()
}

func TestShutdown(t *testing.T) {
	ClearDBCollection("mongo_cloudmint_user_col")
	ClearDBCollection("mongo_currency_collection")
}
