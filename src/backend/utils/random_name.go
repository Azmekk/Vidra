package utils

import (
	"fmt"
	"math/rand/v2"
)

var (
	adjectives = []string{
		"amber", "bold", "brave", "bright", "calm", "clever", "cosmic", "crisp", "daring", "dusty",
		"eager", "electric", "fancy", "fierce", "gentle", "golden", "happy", "hidden", "icy", "jolly",
		"keen", "lively", "lucky", "lunar", "mellow", "misty", "neon", "noble", "odd", "polar",
		"quick", "quiet", "rapid", "rustic", "shiny", "silent", "sleek", "sly", "snowy", "solar",
		"steady", "stormy", "sunny", "swift", "tidy", "urban", "velvet", "vivid", "wild", "witty",
	}
	nouns = []string{
		"badger", "bison", "comet", "coral", "crane", "falcon", "ferret", "finch", "fox", "gecko",
		"heron", "ibis", "jaguar", "koala", "lemur", "lynx", "marten", "meteor", "moose", "nebula",
		"newt", "ocelot", "orca", "otter", "owl", "panda", "pelican", "puffin", "quasar", "raven",
		"robin", "salmon", "seal", "shark", "sparrow", "squid", "stork", "tapir", "tiger", "toucan",
		"trout", "turtle", "viper", "walrus", "wasp", "whale", "wolf", "wombat", "yak", "zebra",
	}
)

// RandomName returns a readable name such as "swift-otter-4821".
func RandomName() string {
	return fmt.Sprintf("%s-%s-%04d", adjectives[rand.IntN(len(adjectives))], nouns[rand.IntN(len(nouns))], rand.IntN(10000))
}
