package unit

import (
	"github.com/ellcrys/openmint/lib"
	"testing"
	// "github.com/ellcrys/util"
	"github.com/ellcrys/openmint/test/fixtures"
	. "github.com/franela/goblin"
	. "github.com/onsi/gomega"
)

func TestHasTextMarks(t *testing.T) {
	g := Goblin(t)
	RegisterFailHandler(func(m string, _ ...int) { g.Fail(m) })
	g.Describe("HasTextMarks()", func() {
		g.It("should return true when all text marks exists and false if otherwise", func() {
			var tokens = []string{"one", "dollar", "bahamas"}
			var tokens2 = []string{"one"}
			Expect(lib.HasTextMarks("BSD", tokens)).To(Equal(true))
			Expect(lib.HasTextMarks("BSD", tokens2)).To(Equal(false))
		})
	})
}

func TestJoinToken(t *testing.T) {
	g := Goblin(t)
	RegisterFailHandler(func(m string, _ ...int) { g.Fail(m) })
	g.Describe("JoinToken()", func() {

		g.It("should successfully join tokens with space between them", func() {
			var tokens = []string{"one", "dollar", "bahamas"}
			Expect(lib.JoinToken(tokens, "space_delimited")).To(Equal("one dollar bahamas"))
		})

		g.It("should successfully join tokens with no space between them", func() {
			var tokens = []string{"one", "dollar", "bahamas"}
			Expect(lib.JoinToken(tokens, "no_delimiter")).To(Equal("onedollarbahamas"))
		})
	})
}

func TestExtractSerial(t *testing.T) {
	g := Goblin(t)
	RegisterFailHandler(func(m string, _ ...int) { g.Fail(m) })
	g.Describe("ExtractSerial()", func() {

		g.It("should use default regex instruction when denomination is not passed", func() {

			fixtures.TestCurrencyMeta["BSD"]["serial"] = map[string]interface{}{
				"join_token_method": "space_delimited",
				"rx":                "(my_serial)",
				"rx_group":          1,
				"filters":           []string{"remove-spaces"},
			}

			var tokens = []string{"word", "term", "my_serial"}
			serial, err := lib.ExtractSerial("", "BSD", tokens)
			Expect(err).To(BeNil())
			Expect(serial).To(Equal("my_serial"))
		})

		g.It("should use denomination regex instruction when denomination is passed", func() {

			fixtures.TestCurrencyMeta["BSD"]["serial"] = map[string]interface{}{
				"join_token_method": "space_delimited",
				"rx":                "(my_serial)",
				"rx_group":          1,
				"filters":           []string{"remove-spaces"},
				"rx_100": map[string]interface{}{
					"rx":                `(my_100_serial)`,
					"rx_group":          1,
					"join_token_method": "space_delimited",
					"filters":           []string{},
				},
			}

			var tokens = []string{"word", "term", "my_serial", "my_100_serial"}
			serial, err := lib.ExtractSerial("100", "BSD", tokens)
			Expect(err).To(BeNil())
			Expect(serial).To(Equal("my_100_serial"))
		})

		g.It("should be able to use referenced denomination instructions", func() {

			fixtures.TestCurrencyMeta["BSD"]["serial"] = map[string]interface{}{
				"join_token_method": "space_delimited",
				"rx":                "(my_serial)",
				"rx_group":          1,
				"filters":           []string{"remove-spaces"},
				"rx_50": map[string]interface{}{
					"rx":                `(my_100_serial)`,
					"rx_group":          1,
					"join_token_method": "space_delimited",
					"filters":           []string{},
				},
				"rx_100": "rx_50",
			}

			var tokens = []string{"word", "term", "my_serial", "my_100_serial"}
			serial, err := lib.ExtractSerial("100", "BSD", tokens)
			Expect(err).To(BeNil())
			Expect(serial).To(Equal("my_100_serial"))
		})

		g.It("should be able to use filters", func() {

			fixtures.TestCurrencyMeta["BSD"]["serial"] = map[string]interface{}{
				"join_token_method": "space_delimited",
				"rx":                "(my_serial)",
				"rx_group":          1,
				"filters":           []string{"remove-spaces"},
				"rx_50": map[string]interface{}{
					"rx":                `(my 100 serial)`,
					"rx_group":          1,
					"join_token_method": "space_delimited",
					"filters":           []string{"remove-spaces"},
				},
				"rx_100": "rx_50",
			}

			var tokens = []string{"word", "term", "my_serial", "my 100 serial"}
			serial, err := lib.ExtractSerial("100", "BSD", tokens)
			Expect(err).To(BeNil())
			Expect(serial).To(Equal("my100serial"))
		})

		g.It("should fail when referenced denomination instructions does not exists", func() {

			fixtures.TestCurrencyMeta["BSD"]["serial"] = map[string]interface{}{
				"join_token_method": "space_delimited",
				"rx":                "(my_serial)",
				"rx_group":          1,
				"filters":           []string{"remove-spaces"},
				"rx_50": map[string]interface{}{
					"rx":                `(my_100_serial)`,
					"rx_group":          1,
					"join_token_method": "space_delimited",
					"filters":           []string{},
				},
				"rx_100": "rx_10",
			}

			var tokens = []string{"word", "term", "my_serial", "my_100_serial"}
			serial, err := lib.ExtractSerial("100", "BSD", tokens)
			Expect(err).NotTo(BeNil())
			Expect(err.Error()).To(Equal("regex instruction 'rx_10' not defined"))
			Expect(serial).To(Equal(""))
		})

		g.It("should fail when referenced denomination instructions type is unsupported", func() {

			fixtures.TestCurrencyMeta["BSD"]["serial"] = map[string]interface{}{
				"join_token_method": "space_delimited",
				"rx":                "(my_serial)",
				"rx_group":          1,
				"filters":           []string{"remove-spaces"},
				"rx_50":             "unsupported type",
				"rx_100":            "rx_50",
			}

			var tokens = []string{"word", "term", "my_serial", "my_100_serial"}
			serial, err := lib.ExtractSerial("100", "BSD", tokens)
			Expect(err).NotTo(BeNil())
			Expect(err.Error()).To(Equal("unsupported reference regex directive type"))
			Expect(serial).To(Equal(""))
		})

		g.It("should fail when reference denomination instruction is not defined", func() {

			fixtures.TestCurrencyMeta["BSD"]["serial"] = map[string]interface{}{
				"join_token_method": "space_delimited",
				"rx":                "(my_serial)",
				"rx_group":          1,
				"filters":           []string{"remove-spaces"},
				"rx_50":             123,
			}

			var tokens = []string{"word", "term", "my_serial", "my_100_serial"}
			serial, err := lib.ExtractSerial("50", "BSD", tokens)
			Expect(err).NotTo(BeNil())
			Expect(err.Error()).To(Equal("unsupported regex directive type"))
			Expect(serial).To(Equal(""))
		})
	})
}

func TestDetermineDenomination(t *testing.T) {
	g := Goblin(t)
	RegisterFailHandler(func(m string, _ ...int) { g.Fail(m) })
	g.Describe("DetermineDenomination()", func() {

		g.It("should successfully determine denomination to be 5", func() {

			fixtures.TestCurrencyMeta["BSD"]["denominations"] = map[string]interface{}{
				"5": map[string]interface{}{
					"join_token_method": "no",
					"rx":                []string{"(?i)five", "(?i)naira", "(?i)central"},
				},
				"100": map[string]interface{}{
					"join_token_method": "no",
					"rx":                []string{"(?i)1909", "1987", "(?i)currency"},
				},
			}

			var tokens = []string{"five", "naira", "CENTRAL"}
			denom := lib.DetermineDenomination("BSD", tokens)
			Expect(denom).To(Equal("5"))
		})

		g.It("should successfully determine denomination to be 100", func() {

			fixtures.TestCurrencyMeta["BSD"]["denominations"] = map[string]interface{}{
				"5": map[string]interface{}{
					"join_token_method": "no",
					"rx":                []string{"(?i)five", "(?i)naira", "(?i)central"},
				},
				"100": map[string]interface{}{
					"join_token_method": "no",
					"rx":                []string{"(?i)1909", "1987", "(?i)currency"},
				},
			}

			var tokens = []string{"1909", "1987", "currENcy"}
			denom := lib.DetermineDenomination("BSD", tokens)
			Expect(denom).To(Equal("100"))
		})

		g.It("should fail to determine denomination since not all rx patterns matched", func() {

			fixtures.TestCurrencyMeta["BSD"]["denominations"] = map[string]interface{}{
				"5": map[string]interface{}{
					"join_token_method": "no",
					"rx":                []string{"(?i)five", "(?i)naira", "(?i)central"},
				},
				"100": map[string]interface{}{
					"join_token_method": "no",
					"rx":                []string{"(?i)1909", "1987", "(?i)currency"},
				},
			}

			var tokens = []string{"fi", "nai", "CENTRAL"}
			denom := lib.DetermineDenomination("BSD", tokens)
			Expect(denom).To(Equal(""))
		})
	})
}
