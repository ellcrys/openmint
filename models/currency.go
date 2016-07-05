package models

import (
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"time"
	"github.com/ellcrys/openmint/config"
)

type CurrencyModel struct {
    Id     				bson.ObjectId 	`json:"id" bson:"_id"`
    Address 			string 			`json:"name" bson:"name"`
    ImageID 	   		string 			`json:"image_id" bson:"image_id"`
    Code 		 		string 			`json:"code" bson:"code"`
    SuggestedDenom		int 			`json:"suggested_denom" bson:"suggested_denom"`
    Status 				string 			`json:"status" bson:"status"`
    CreatedAt 			time.Time 	  	`json:"created_at" bson:"created_at"`	
}

var (
	Currency = CurrencyModel{}
)

func (m *CurrencyModel) EnsureIndex(ses *mgo.Session) {
	ses.SetMode(mgo.Monotonic, true)
	colName := config.C.GetString("mongo_currency_collection")
	c := ses.DB(config.C.GetString("mongo_database")).C(colName)
	if c.EnsureIndexKey("code", "suggested_denom") != nil {
		panic("failed to ensure index in "+colName+" collection")
	}
}

// find by a field name 
func (m *CurrencyModel) FindByField(ses *mgo.Session, field, value string) (*CurrencyModel, error) {
	ses.SetMode(mgo.Monotonic, true)
	c := ses.DB(config.C.GetString("mongo_database")).C(config.C.GetString("mongo_currency_collection"))
	asset := CurrencyModel{}
	err := c.Find(bson.M{ field: value }).One(&asset)
	return &asset, err
}

// find by id
func (m *CurrencyModel) FindById(ses *mgo.Session, id string) (*CurrencyModel, error) {
	ses.SetMode(mgo.Monotonic, true)
	c := ses.DB(config.C.GetString("mongo_database")).C(config.C.GetString("mongo_currency_collection"))
	asset := CurrencyModel{}
	err := c.FindId(bson.ObjectIdHex(id)).One(&asset)
	return &asset, err
}

// add new app entry
func (m *CurrencyModel) Create(ses *mgo.Session, data *CurrencyModel) error {
	data.CreatedAt = time.Now().UTC()
	ses.SetMode(mgo.Monotonic, true)
	c := ses.DB(config.C.GetString("mongo_database")).C(config.C.GetString("mongo_currency_collection"))
	return c.Insert(data)
}

// delete app
func (m *CurrencyModel) Delete(ses *mgo.Session, id string) error {
	ses.SetMode(mgo.Monotonic, true)
	c := ses.DB(config.C.GetString("mongo_database")).C(config.C.GetString("mongo_currency_collection"))
	return c.RemoveId(bson.ObjectIdHex(id))
}

func (m *CurrencyModel) UpdateField(ses *mgo.Session, id, field, newValue string) error {
	ses.SetMode(mgo.Monotonic, true)
	c := ses.DB(config.C.GetString("mongo_database")).C(config.C.GetString("mongo_currency_collection"))
	return c.UpdateId(bson.ObjectIdHex(id), bson.M{ "$set": bson.M{ field: newValue }})
}

func (m *CurrencyModel) UpdateStatus(ses *mgo.Session, id, newStatus string) error {
	return m.UpdateField(ses, id, "status", newStatus)
}
