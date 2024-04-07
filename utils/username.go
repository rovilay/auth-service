package utils

import (
	"math/rand"
	"strings"
	"time"
)

var adjectives = []string{"happy", "blue", "swift", "clever", "smart", "quick", "red"}
var nouns = []string{"panda", "tree", "rocket", "coffee", "ninja", "star", "unicorn"}

func GenerateUsername(firstName string) string {
	// Seed for randomness
	rand.New(rand.NewSource(time.Now().UnixNano()))
	adjective := adjectives[rand.Intn(len(adjectives))]
	noun := nouns[rand.Intn(len(nouns))]

	return strings.ToLower(firstName + adjective + noun)
}
