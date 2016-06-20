package models

import (
	"gopkg.in/mgo.v2/bson"
)

func NewId() bson.ObjectId {
	return bson.NewObjectId()
}

func IsId(v string) bool {
	return bson.IsObjectIdHex(v)
}