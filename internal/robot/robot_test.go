package robot

import (
	"github.com/stretchr/testify/assert"
	"math/rand"
	"testing"
)

func TestSplitSecret(t *testing.T) {
	ass := assert.New(t)
	input := "Hidden beneath the old oak tree, golden coins patiently await discovery."
	expected := []string{"Hidden", "beneath", "the", "old", "oak", "tree,", "golden", "coins", "patiently", "await", "discovery."}

	result := SplitSecret(input)
	ass.Equal(expected, result)
}

func TestCreateRobots(t *testing.T) {
	ass := assert.New(t)
	words := []string{"a", "b", "c", "d"}
	robots := CreateRobots(words, 3, 10)

	ass.Equal(3, len(robots))
	total := 0
	for _, r := range robots {
		ass.Equal(10, cap(r.Inbox))
		total += len(r.Words)
	}
	ass.Equal(total, len(words))
}

func TestHasReconstructedSecret(t *testing.T) {
	ass := assert.New(t)
	expected := []string{"a", "b", "c"}
	tests := []struct {
		name   string
		words  []string
		result bool
	}{
		{"Exact match", []string{"a", "b", "c"}, true},
		{"Missing one", []string{"a", "b"}, false},
		{"Extra word", []string{"a", "b", "c", "d"}, true},
		{"Duplicated words", []string{"a", "a", "b", "c"}, true},
	}

	for _, tt := range tests {
		ass.Equal(tt.result, HasReconstructedSecret(expected, tt.words))
	}
}

func TestExchangeAtLeastOneMessage(t *testing.T) {
	ass := assert.New(t)
	rand.Seed(42)
	robots := CreateRobots([]string{"alpha", "beta", "gamma"}, 3, 10)
	sent := ExchangeMessage(robots, 0, 0, 0, 50)
	ass.GreaterOrEqual(sent, 1)
	ass.LessOrEqual(sent, 50)
}
