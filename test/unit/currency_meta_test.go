package unit

import (
	"github.com/ellcrys/openmint/lib"
	"github.com/ellcrys/openmint/test/fixtures"
	"github.com/ellcrys/util"
	. "github.com/franela/goblin"
	. "github.com/onsi/gomega"
	"testing"
)

func TestIsValidCode(t *testing.T) {
	lib.SetCurrencyMeta(fixtures.TestCurrencyMeta)
	g := Goblin(t)
	RegisterFailHandler(func(m string, _ ...int) { g.Fail(m) })
	g.Describe("IsValidCode()", func() {

		g.It("Should be true because BSD is a valid currency code", func() {
			Expect(lib.IsValidCode("BSD")).To(Equal(true))
		})

		g.It("Should be false because XDP is not a vaild currency code", func() {
			Expect(lib.IsValidCode("XDP")).To(Equal(false))
		})
	})
}

func TestGetCurrencyLang(t *testing.T) {
	lib.SetCurrencyMeta(fixtures.TestCurrencyMeta)
	g := Goblin(t)
	RegisterFailHandler(func(m string, _ ...int) { g.Fail(m) })
	g.Describe("GetCurrencyLang()", func() {
		g.It("Should return `en`", func() {
			Expect(lib.GetCurrencyLang("BSD")).To(Equal("en"))
		})
	})
}

func TestGetCurrencyDenoms(t *testing.T) {
	lib.SetCurrencyMeta(fixtures.TestCurrencyMeta)
	g := Goblin(t)
	RegisterFailHandler(func(m string, _ ...int) { g.Fail(m) })
	g.Describe("GetCurrencyDenoms()", func() {
		g.It("Should return 1 and 5", func() {
			denoms := lib.GetCurrencyDenoms("BSD")
			Expect(util.InStringSlice(denoms, "1")).To(Equal(true))
			Expect(util.InStringSlice(denoms, "5")).To(Equal(true))
			Expect(util.InStringSlice(denoms, "10")).To(Equal(false))
		})
	})
}

func TestGetCurrencyTextMarks(t *testing.T) {
	lib.SetCurrencyMeta(fixtures.TestCurrencyMeta)
	g := Goblin(t)
	RegisterFailHandler(func(m string, _ ...int) { g.Fail(m) })
	g.Describe("GetCurrencyTextMarks()", func() {
		g.It("Should return [bahamas]", func() {
			Expect(lib.GetCurrencyTextMarks("BSD")).To(Equal([]string{"bahamas"}))
		})
	})
}

func TestGetCurrencySerialData(t *testing.T) {
	lib.SetCurrencyMeta(fixtures.TestCurrencyMeta)
	g := Goblin(t)
	RegisterFailHandler(func(m string, _ ...int) { g.Fail(m) })
	g.Describe("GetCurrencySerialData()", func() {
		g.It("Should return expected serial data", func() {
			expected := map[string]interface{}{
				"join_method": "space_delimited",
				"rx":          ".*([a-z]{1}[0-9 ]{6,})",
				"rx_group":    1,
				"filters":     []string{"remove-spaces"},
			}
			Expect(lib.GetCurrencySerialData("BSD")).To(Equal(expected))
		})
	})
}

func TestGetDenominationData(t *testing.T) {
	lib.SetCurrencyMeta(fixtures.TestCurrencyMeta)
	g := Goblin(t)
	RegisterFailHandler(func(m string, _ ...int) { g.Fail(m) })
	g.Describe("GetDenominationData()", func() {
		g.It("Should return expected denominations", func() {
			expected := map[string]interface{}{
				"1": map[string]interface{}{
					"join_method": "no",
					"rx":          []string{"one", "lynden", "pindling"},
				},
				"5": map[string]interface{}{
					"join_method": "no",
					"rx":          []string{"one", "lynden", "pindling"},
				},
			}
			Expect(lib.GetDenominationData("BSD")).To(Equal(expected))
		})
	})
}

func TestGetDefinedCurrencies(t *testing.T) {
	lib.SetCurrencyMeta(fixtures.TestCurrencyMeta2)
	g := Goblin(t)
	RegisterFailHandler(func(m string, _ ...int) { g.Fail(m) })
	g.Describe("GetDefinedCurrencies()", func() {
		g.It("Should return the list of currency codes that have their metadata defined", func() {
			expected := []string{"BSD", "USD"}
			Expect(lib.GetDefinedCurrencies()).To(ContainElement(expected[0]))
			Expect(lib.GetDefinedCurrencies()).To(ContainElement(expected[1]))
		})
	})
}
