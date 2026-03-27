package words

import (
	"bytes"
	"maps"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewStemmer(t *testing.T) {
	if NewStemmer(nil) == nil {
		t.Error("NewStemmer should constructed")
	}

	if NewStemmer(map[string]struct{}{"word": {}, "someWord": {}}) == nil {
		t.Error("NewStemmer should constructed")
	}
}

func TestParseStopWords(t *testing.T) {
	buf := bytes.NewBufferString("an apple a day")
	stopWords := map[string]struct{}{"an": {}, "apple": {}, "a": {}, "day": {}}
	words := ParseStopWords(buf)

	if !maps.Equal(stopWords, words) {
		t.Error("some words don't match")
	}
}

func TestParsePhrase(t *testing.T) {
	type words struct {
		phrase string
		words  []string
	}

	tests := []words{
		{"So close, no matter how far", []string{"so", "close", "no", "matter", "how", "far"}},
		{"Couldn't^be	much&more%from the heart", []string{"couldn", "t", "be", "much", "more", "from", "the", "heart"}},
		{"12 34 56 78", []string{"12", "34", "56", "78"}},
	}
	for _, test := range tests {
		res := ParsePhrase(test.phrase)
		expected := test.words
		sort.Strings(expected)
		sort.Strings(res)
		assert.Equal(t, expected, res)
	}
}

func TestStemmer(t *testing.T) {
	stemmer := NewStemmer(nil)
	stemmer.Stem(ParsePhrase("Forever trusting who we are"))
}

func TestStemmer_ShortWordsFiltered(t *testing.T) {
	stemmer := NewStemmer(nil)
	// Words strictly < 4 chars: "a"(1), "be"(2), "cat"(3) → all filtered
	// "dogs"(4) passes the length filter but may still be filtered by stop words
	result := stemmer.Stem(ParsePhrase("a be cat"))
	assert.Empty(t, result, "all words with len < 4 should be filtered")
}

func TestStemmer_StopWordFiltered(t *testing.T) {
	stemmer := NewStemmer(nil)
	result := stemmer.Stem(ParsePhrase("about above after"))
	assert.Empty(t, result)
}

func TestStemmer_EmptyInput(t *testing.T) {
	stemmer := NewStemmer(nil)
	result := stemmer.Stem(ParsePhrase(""))
	assert.Empty(t, result)
}

func TestStemmer_StemsSameRoot(t *testing.T) {
	stemmer := NewStemmer(nil)
	r1 := stemmer.Stem(ParsePhrase("connections"))
	r2 := stemmer.Stem(ParsePhrase("connected"))
	// Both "connections" and "connected" should stem to "connect"
	assert.NotEmpty(t, r1)
	assert.NotEmpty(t, r2)
	for k := range r1 {
		_, ok := r2[k]
		assert.True(t, ok, "stemmed forms should share root: %s", k)
	}
}

func TestParsePhrase_EmptyString(t *testing.T) {
	words := ParsePhrase("")
	assert.Empty(t, words)
}

func TestParsePhrase_OnlySpecialChars(t *testing.T) {
	words := ParsePhrase("!@#$%^&*()")
	assert.Empty(t, words)
}

func TestParseStopWords_Empty(t *testing.T) {
	buf := strings.NewReader("")
	result := ParseStopWords(buf)
	assert.Empty(t, result)
}
