package lib

import (
	validator "github.com/asaskevich/govalidator"
	"github.com/ellcrys/openmint/models"
)

func init() {

	// validate a password.
	// Password must have a least one character otherwise
	// it is considered valid. User `require` validator to check for
	// emptiness if you require it.
	validator.CustomTypeTagMap.Set("isValidPassword", validator.CustomTypeValidator(func(v interface{}, context interface{}) bool {
		switch obj := context.(type) {
		case models.UserModel:
			_ = obj
			// var strLen = len(v.(string))
			// if strLen > 0 && strLen < obj.MinPasswordLength {
			// 	return false
			// }
			return false
		default:
			panic("unsupported object type")
		}
		return true
	}))
}
