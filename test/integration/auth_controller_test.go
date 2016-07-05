package integration

import (
	"bufio"
	"bytes"
	"testing"

	"github.com/ellcrys/openmint/config"
	"github.com/ellcrys/openmint/lib"
	"github.com/ellcrys/openmint/models"
	"github.com/ellcrys/openmint/test/common"
	"github.com/ellcrys/util"
	. "github.com/franela/goblin"
	. "github.com/onsi/gomega"
)

var authCntrl *lib.AuthController
var testUser *models.UserModel

func init() {
	var err error
	authCntrl = lib.NewAuthController(common.MongoSes)
	testUser, err = common.CreateTestUser()
	if err != nil {
		panic("init: could not create test user")
	}
}

func TestUserAuth(t *testing.T) {
	g := Goblin(t)
	RegisterFailHandler(func(m string, _ ...int) { g.Fail(m) })
	g.Describe("UserAuth()", func() {

		g.It("should fail when invalid request body is received", func() {

			var bodies = [][]interface{}{
				{`{}`, 400, "email and password are required"},
				{`{ "email": "ex@example.com" }`, 400, "email and password are required"},
				{`{ "password": "somepassword" }`, 400, "email and password are required"},
				{`{ "email": "` + testUser.Email + `", "password": "wrong" }`, 400, "user email or password are invalid"},
				{`{ "email": "wrong@example.com", "password": "` + testUser.Password + `" }`, 400, "user email or password are invalid"},
			}

			for _, body := range bodies {
				ctx := common.NewContext("POST", "/v1/auth/login", nil, body[0].(string), nil)
				err := authCntrl.UserAuth(ctx)

				if err != nil {
					err := err.(*config.HTTPError)
					Expect(err.StatusCode).To(Equal(body[1].(int)))
					Expect(err.Error()).To(Equal(body[2].(string)))
				}
			}
		})

		g.It("should succeed", func() {
			body := `{ "email": "` + testUser.Email + `", "password": "` + testUser.Password + `" }`
			ctx := common.NewContext("POST", "/v1/auth/login", nil, body, nil)
			var buffer bytes.Buffer
			writer := bufio.NewWriter(&buffer)
			ctx.Response().SetWriter(writer)

			err := authCntrl.UserAuth(ctx)

			writer.Flush()
			body = buffer.String()
			m, _ := util.DecodeJSONToMap(body)

			Expect(err).To(BeNil())
			Expect(util.Map(m)["token"]).ToNot(BeNil())
		})
	})
}
