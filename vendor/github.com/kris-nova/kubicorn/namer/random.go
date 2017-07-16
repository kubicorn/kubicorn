package namer

import (
	"fmt"
	"math/rand"
	"time"
)

func RandomName() string {
	adjl := len(adjectives)
	nounl := len(nouns)
	return fmt.Sprintf("%s-%s", adjectives[ran(1, adjl)], nouns[ran(1, nounl)])

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

var nouns = []string{
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
