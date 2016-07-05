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

func TestCreate(t *testing.T) {
	g := Goblin(t)
	RegisterFailHandler(func(m string, _ ...int) { g.Fail(m) })
	g.Describe("Create()", func() {

		g.It("should fail when invalid request body is received", func() {

			var bodies = [][]interface{}{
				{`{}`, 400, "fullname is required"},
				{`{ "full_name": "john doe" }`, 400, "email is required"},
				{`{ "full_name": "john doe", "email": "invalid" }`, 400, "email is not valid"},
				{`{ "full_name": "john doe", "email": "me@example.com" }`, 400, "password is required"},
				{`{ "full_name": "john doe", "email": "me@example.com", "password": "abc" }`, 400, "password must have atleast 6 characters"},
			}

			for _, body := range bodies {
				ctx := common.NewContext("POST", "/v1/users", nil, body[0].(string), nil)
				err := userCntrl.Create(ctx)

				if err != nil {
					err := err.(*config.HTTPError)
					Expect(err.StatusCode).To(Equal(body[1].(int)))
					Expect(err.Error()).To(Equal(body[2].(string)))
				}
			}
		})

		g.It("should succeed", func() {
			body := `{ "full_name": "john doe", "email": "me2@example.com", "password": "abcdef" }`
			ctx := common.NewContext("POST", "/v1/users", nil, body, nil)
			err := userCntrl.Create(ctx)
			Expect(err).To(BeNil())
		})
	})
}
