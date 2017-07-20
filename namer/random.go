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
	"extravagant",
	"ridiculous",
	"exciting",
	"amazing",
	"adorable",
	"ferocious",
	"hilarious",
	"scrumptious",
	"floral",
}

var nouns = []string{
	"unicorn",
	"rainbow",
	"painting",
	"badger",
	"mongoose",
	"sculpture",
	"creature",
	"gopher",
	"wardrobe",
	"mushroom",
	"hideout",
	"party",
	"monster",
	"sheepdog",
	"outfit",
}
