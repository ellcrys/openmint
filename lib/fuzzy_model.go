package lib 

import (
	"github.com/sajari/fuzzy"
)

type FuzzyModel struct {
	model *fuzzy.Model
}

func NewFuzzyModel(words []string, depth int) *FuzzyModel {
	model := fuzzy.NewModel()
	model.SetThreshold(1)
	model.SetDepth(depth)
	model.Train(words)
	return &FuzzyModel{ model }
}

// Get suggestions for a given word
func (self *FuzzyModel) GetSuggestions(word string) []string {
	return self.model.Suggestions(word, false)
}

// Get single spelling suggestion
func (self *FuzzyModel) GetSpellingSuggestion(word string) []string {
	return []string{ self.model.SpellCheck(word) }
}