package unit 

import (
	"testing"
	"github.com/ellcrys/openmint/lib"
	. "github.com/franela/goblin"
    . "github.com/onsi/gomega"
)

func TestRemoveSpaces(t *testing.T) {
	g := Goblin(t)
	RegisterFailHandler(func(m string, _ ...int) { g.Fail(m) })
	g.Describe("RemoveSpaces()", func() {

		g.It("remove all whitespaces", func () {
			actual := "the quick brown "
			expected := "thequickbrown"
			Expect(lib.RemoveSpaces(actual)).To(Equal(expected))
		})

	})
}

func TestFilter(t *testing.T) {
	g := Goblin(t)
	RegisterFailHandler(func(m string, _ ...int) { g.Fail(m) })
	g.Describe("Filter()", func() {

		g.It("call with no filter func should have no effect", func () {
			actual := "the quick brown "
			Expect(lib.Filter(actual, nil)).To(Equal(actual))
		})

		g.It("call with RemoveSpaces filter func should remove whitespaces", func () {
			actual := "the quick brown "
			expected := "thequickbrown"
			Expect(lib.Filter(actual, []string{"remove-spaces"})).To(Equal(expected))
		})

	})
}

