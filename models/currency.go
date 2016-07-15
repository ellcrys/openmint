package models

import (
	"github.com/ellcrys/openmint/config"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"time"
)

type Vote struct {
	Label  string        `json:"label" bson:"label"`
	UserId bson.ObjectId `json:"user_id" bson:"user_id"`
}

type CurrencyModel struct {
	Id               bson.ObjectId `json:"id" bson:"_id"`
	UserId           bson.ObjectId `json:"user_id" bson:"user_id"`
	ImageURL         string        `json:"image_url" bson:"image_url"`
	OriginalImageURL string        `json:"original_image_url" bson:"original_image_url"`
	CurrencyCode     string        `json:"currency_code" bson:"currency_code"`
	Denomination     string        `json:"denomination" bson:"denomination"`
	Serial           string        `json:"serial" bson:"serial"`
	Status           string        `json:"status" bson:"status"`
	Votes            []Vote        `json:"votes" bson:"votes"`
	Multiplier       float64       `json:"multiplier" bson:"multiplier"`
	CreatedAt        time.Time     `json:"created_at" bson:"created_at"`
}

var (
	Currency = CurrencyModel{}
)

func (m *CurrencyModel) EnsureIndex(ses *mgo.Session) {
	ses.SetMode(mgo.Monotonic, true)
	colName := config.C.GetString("mongo_currency_collection")
	c := ses.DB(config.C.GetString("mongo_database")).C(colName)

	index := mgo.Index{
		Key: []string{"currency_code", "denomination", "serial"},
	}

	err := c.EnsureIndex(index)
	if err != nil {
		panic("failed to ensure compound index in " + colName + " collection")
	}

	if c.EnsureIndexKey("user_id") != nil {
		panic("failed to ensure index in " + colName + " collection")
	}
}

func (m *CurrencyModel) FindCurrency(ses *mgo.Session, curCode, denomination, serial string) (*CurrencyModel, error) {
	ses.SetMode(mgo.Monotonic, true)
	c := ses.DB(config.C.GetString("mongo_database")).C(config.C.GetString("mongo_currency_collection"))
	result := CurrencyModel{}
	err := c.Find(bson.M{"currency_code": curCode, "denomination": denomination, "serial": serial}).One(&result)
	return &result, err
}

// find by a field name
func (m *CurrencyModel) FindByField(ses *mgo.Session, field, value string) (*CurrencyModel, error) {
	ses.SetMode(mgo.Monotonic, true)
	c := ses.DB(config.C.GetString("mongo_database")).C(config.C.GetString("mongo_currency_collection"))
	result := CurrencyModel{}
	err := c.Find(bson.M{field: value}).One(&result)
	return &result, err
}

// find by id
func (m *CurrencyModel) FindById(ses *mgo.Session, id string) (*CurrencyModel, error) {
	ses.SetMode(mgo.Monotonic, true)
	c := ses.DB(config.C.GetString("mongo_database")).C(config.C.GetString("mongo_currency_collection"))
	result := CurrencyModel{}
	err := c.FindId(bson.ObjectIdHex(id)).One(&result)
	return &result, err
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
	return c.UpdateId(bson.ObjectIdHex(id), bson.M{"$set": bson.M{field: newValue}})
}

func (m *CurrencyModel) UpdateStatus(ses *mgo.Session, id, newStatus string) error {
	return m.UpdateField(ses, id, "status", newStatus)
}

func (m *CurrencyModel) FindWithDateSortAndSkip(ses *mgo.Session, userId, sort string, limit, skip int) ([]CurrencyModel, error) {
	ses.SetMode(mgo.Monotonic, true)
	c := ses.DB(config.C.GetString("mongo_database")).C(config.C.GetString("mongo_currency_collection"))
	results := []CurrencyModel{}
	err := c.Find(bson.M{"user_id": bson.ObjectIdHex(userId)}).Limit(limit).Sort(sort).Skip(skip).All(&results)
	return results, err
}
