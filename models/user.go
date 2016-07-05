package models

import (
	"github.com/ellcrys/openmint/config"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"time"
)

type UserModel struct {

	// Collection attributes
	Id        bson.ObjectId `json:"id" bson:"_id"`
	Fullname  string        `json:"full_name" bson:"full_name" valid:"required"`
	Email     string        `json:"email" bson:"email" valid:"required,email"`
	Password  string        `json:"password,omitempty" bson:"password" valid:"required,isValidPassword"`
	CreatedAt time.Time     `json:"created_at" bson:"created_at"`

	// Hidden fields
	MinPasswordLength int    `json:"-" bson:"-"`
	TokenString       string `json:"token,omitempty" bson:"-"`
}

var (
	User = UserModel{}
)

func (m *UserModel) EnsureIndex(ses *mgo.Session) {
	ses.SetMode(mgo.Monotonic, true)
	colName := config.C.GetString("mongo_cloudmint_user_col")
	c := ses.DB(config.C.GetString("mongo_database")).C(colName)
	if c.EnsureIndexKey("email") != nil {
		panic("failed to ensure index in " + colName + " collection")
	}
}

// find by a field name
func (m *UserModel) FindByField(ses *mgo.Session, field, value string) (*UserModel, error) {
	ses.SetMode(mgo.Monotonic, true)
	c := ses.DB(config.C.GetString("mongo_database")).C(config.C.GetString("mongo_cloudmint_user_col"))
	asset := UserModel{}
	err := c.Find(bson.M{field: value}).One(&asset)
	return &asset, err
}

// find by id
func (m *UserModel) FindById(ses *mgo.Session, id string) (*UserModel, error) {
	ses.SetMode(mgo.Monotonic, true)
	c := ses.DB(config.C.GetString("mongo_database")).C(config.C.GetString("mongo_cloudmint_user_col"))
	asset := UserModel{}
	err := c.FindId(bson.ObjectIdHex(id)).One(&asset)
	return &asset, err
}

// add new app entry
func (m *UserModel) Create(ses *mgo.Session, data *UserModel) error {
	data.CreatedAt = time.Now().UTC()
	ses.SetMode(mgo.Monotonic, true)
	c := ses.DB(config.C.GetString("mongo_database")).C(config.C.GetString("mongo_cloudmint_user_col"))
	return c.Insert(data)
}

// delete app
func (m *UserModel) Delete(ses *mgo.Session, id string) error {
	ses.SetMode(mgo.Monotonic, true)
	c := ses.DB(config.C.GetString("mongo_database")).C(config.C.GetString("mongo_cloudmint_user_col"))
	return c.RemoveId(bson.ObjectIdHex(id))
}

func (m *UserModel) UpdateField(ses *mgo.Session, id, field, newValue string) error {
	ses.SetMode(mgo.Monotonic, true)
	c := ses.DB(config.C.GetString("mongo_database")).C(config.C.GetString("mongo_cloudmint_user_col"))
	return c.UpdateId(bson.ObjectIdHex(id), bson.M{"$set": bson.M{field: newValue}})
}
