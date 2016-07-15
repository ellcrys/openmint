package integration

import (
	"testing"

	"github.com/ellcrys/openmint/config"
	"github.com/ellcrys/openmint/lib"
	"github.com/ellcrys/openmint/test/common"
	. "github.com/franela/goblin"
	. "github.com/onsi/gomega"
)

var userCntrl *lib.UserController

func init() {
	userCntrl = lib.NewUserController(common.MongoSes)
}
