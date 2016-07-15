package models

import (
	"github.com/ellcrys/openmint/config"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"time"
)

type TwitterAuthModel struct {
	Id               bson.ObjectId `json:"id" bson:"_id"`
	OauthTokenSecret string        `json:"oauth_token_secret" bson:"oauth_token_secret" valid:"required"`
	OauthToken       string        `json:"oauth_token" bson:"oauth_token" valid:"required,oauth_token"`
	CreatedAt        time.Time     `json:"created_at" bson:"created_at"`
}

var (
	TwitterAuth = TwitterAuthModel{}
)

func (m *TwitterAuthModel) EnsureIndex(ses *mgo.Session) {
	ses.SetMode(mgo.Monotonic, true)
	colName := config.C.GetString("mongo_twitter_auth_col")
	c := ses.DB(config.C.GetString("mongo_database")).C(colName)
	if c.EnsureIndexKey("oauth_token") != nil {
		panic("failed to ensure index in " + colName + " collection")
	}
}

// find by arbitrary query
func (m *TwitterAuthModel) Find(ses *mgo.Session, q bson.M) (*TwitterAuthModel, error) {
	ses.SetMode(mgo.Monotonic, true)
	c := ses.DB(config.C.GetString("mongo_database")).C(config.C.GetString("mongo_twitter_auth_col"))
	asset := TwitterAuthModel{}
	err := c.Find(q).One(&asset)
	return &asset, err
}

// find by a field name
func (m *TwitterAuthModel) FindByField(ses *mgo.Session, field, value string) (*TwitterAuthModel, error) {
	ses.SetMode(mgo.Monotonic, true)
	c := ses.DB(config.C.GetString("mongo_database")).C(config.C.GetString("mongo_twitter_auth_col"))
	asset := TwitterAuthModel{}
	err := c.Find(bson.M{field: value}).One(&asset)
	return &asset, err
}

// find by id
func (m *TwitterAuthModel) FindById(ses *mgo.Session, id string) (*TwitterAuthModel, error) {
	ses.SetMode(mgo.Monotonic, true)
	c := ses.DB(config.C.GetString("mongo_database")).C(config.C.GetString("mongo_twitter_auth_col"))
	asset := TwitterAuthModel{}
	err := c.FindId(bson.ObjectIdHex(id)).One(&asset)
	return &asset, err
}

// add new app entry
func (m *TwitterAuthModel) Create(ses *mgo.Session, data *TwitterAuthModel) error {
	data.CreatedAt = time.Now().UTC()
	ses.SetMode(mgo.Monotonic, true)
	c := ses.DB(config.C.GetString("mongo_database")).C(config.C.GetString("mongo_twitter_auth_col"))
	return c.Insert(data)
}

// delete app
func (m *TwitterAuthModel) Delete(ses *mgo.Session, id string) error {
	ses.SetMode(mgo.Monotonic, true)
	c := ses.DB(config.C.GetString("mongo_database")).C(config.C.GetString("mongo_twitter_auth_col"))
	return c.RemoveId(bson.ObjectIdHex(id))
}

func (m *TwitterAuthModel) UpdateField(ses *mgo.Session, id, field, newValue string) error {
	ses.SetMode(mgo.Monotonic, true)
	c := ses.DB(config.C.GetString("mongo_database")).C(config.C.GetString("mongo_twitter_auth_col"))
	return c.UpdateId(bson.ObjectIdHex(id), bson.M{"$set": bson.M{field: newValue}})
}
