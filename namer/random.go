package namer

import (
	"fmt"
	"math/rand"
	"time"
)

func RandomName() string {
	adjl := len(adjectives)
	wordl := len(words)
	return fmt.Sprintf("%s-%s", adjectives[ran(1, adjl)], words[ran(1, wordl)])

}

func ran(min, max int) int {
	rand.Seed(time.Now().Unix())
	return rand.Intn(max-min) + min
}

var adjectives = []string{
	"unique",
	"beautiful",
	"attractive",
	"wonderful",
	"fabulous",
	"extravagent",
	"exciting",
	"amazing",
	"adorable",
	"ferocious",
}

var words = []string{
	"unicorn",
	"rainbow",
	"painting",
	"badger",
	"mongoose",
	"sculpture",
	"creature",
	"mushroom",
	"hideout",
	"party",
	"monster",
	"sheepdog",
}
