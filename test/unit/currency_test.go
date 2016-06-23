package unit 

import (
    "testing"
    . "github.com/franela/goblin"
    . "github.com/onsi/gomega"
    "github.com/ellcrys/openmint/config"
)

func TestIsValidCode(t *testing.T) {
	g := Goblin(t)
	RegisterFailHandler(func(m string, _ ...int) { g.Fail(m) })
	g.Describe("IsValidCode()", func() {

		g.It("Should be true because USD is a valid currency code", func () {
			Expect(config.IsValidCode("USD")).To(Equal(true))
		})

		g.It("Should be false because XDP is not a vaild currency code", func () {
			Expect(config.IsValidCode("XDP")).To(Equal(false))
		})
	})
}



// func TestGetCurrencyLang(t *testing.T) {
// 	g := Goblin(t)
// 	RegisterFailHandler(func(m string, _ ...int) { g.Fail(m) })
// 	g.Describe("GetCurrencyLang()", func() {
// 		g.It("Should return `us`", func () {
// 			Expect(config.GetCurrencyLang("USD")).To(Equal("en"))
// 		})
// 	})
// }

// func TestGetDenomOccurrance(t *testing.T) {
// 	g := Goblin(t)
// 	RegisterFailHandler(func(m string, _ ...int) { g.Fail(m) })
// 	g.Describe("GetDenomOccurrance()", func() {
// 		g.It("Should return 2", func () {
// 			Expect(config.GetDenomOccurrance("RUB", 50)).To(Equal(2))
// 		})
// 		g.It("Should return 1", func () {
// 			Expect(config.GetDenomOccurrance("RUB", 5)).To(Equal(1))
// 		})
// 	})
// }

// func TestGetCurrencyDenoms(t *testing.T) {
// 	g := Goblin(t)
// 	RegisterFailHandler(func(m string, _ ...int) { g.Fail(m) })
// 	g.Describe("GetCurrencyDenoms()", func() {
// 		g.It("Should return 2", func () {
// 			var USDDenom = []string{ "1", "2", "5", "10", "20", "50", "100" }
// 			Expect(config.GetCurrencyDenoms("USD")).To(Equal(USDDenom))
// 		})
// 	})
// }