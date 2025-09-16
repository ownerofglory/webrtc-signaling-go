package services

import (
	"fmt"
	"math/rand/v2"
	"sync"
)

type nicknameGenerator struct {
	mu    sync.Mutex
	inUse map[string]struct{}
}

var adjectives = []string{
	"adventurous", "affectionate", "agreeable", "amazing", "ambitious",
	"amiable", "artistic", "attentive", "balanced", "beautiful",
	"blissful", "bold", "brilliant", "bubbly", "calm",
	"carefree", "charming", "cheerful", "clever", "colorful",
	"compassionate", "confident", "cool", "courageous", "creative",
	"curious", "daring", "delightful", "eager", "easygoing",
	"ecstatic", "efficient", "energetic", "enthusiastic", "excellent",
	"fabulous", "faithful", "fantastic", "fearless", "festive",
	"friendly", "fun", "funny", "generous", "gentle",
	"glad", "gleeful", "graceful", "gracious", "happy",
	"harmonious", "helpful", "honest", "hopeful", "humble",
	"imaginative", "incredible", "independent", "ingenious", "inspiring",
	"jolly", "joyful", "jubilant", "keen", "kind",
	"lively", "lovely", "lucky", "magnificent", "marvelous",
	"mellow", "mindful", "mysterious", "nice", "noble",
	"optimistic", "outgoing", "peaceful", "playful", "pleasant",
	"polite", "positive", "powerful", "precious", "radiant",
	"reliable", "resourceful", "respectful", "shiny", "sincere",
	"smart", "smiling", "spirited", "splendid", "stellar",
	"strong", "sunny", "sweet", "talented", "thoughtful",
	"tranquil", "trustworthy", "unique", "upbeat", "vibrant",
	"victorious", "vivacious", "warm", "whimsical", "wise",
	"witty", "wonderful", "zealous", "zesty",
}

var animals = []string{
	// real animals
	"ant", "antelope", "armadillo", "badger", "bat",
	"bear", "beaver", "bee", "bison", "buffalo",
	"butterfly", "camel", "canary", "capybara", "caribou",
	"cat", "caterpillar", "cheetah", "chicken", "chimpanzee",
	"chipmunk", "cow", "coyote", "crab", "crane",
	"crocodile", "crow", "deer", "dingo", "dolphin",
	"donkey", "dove", "dragonfly", "duck", "eagle",
	"eel", "elephant", "elk", "falcon", "ferret",
	"finch", "firefly", "fish", "flamingo", "fox",
	"frog", "gazelle", "gecko", "giraffe", "goat",
	"goose", "gorilla", "grasshopper", "hamster", "hare",
	"hedgehog", "heron", "hippo", "horse", "hummingbird",
	"ibis", "iguana", "jackal", "jaguar", "jellyfish",
	"kangaroo", "kingfisher", "koala", "lemur", "leopard",
	"lion", "lizard", "llama", "lobster", "lynx",
	"magpie", "manatee", "meerkat", "mole", "monkey",
	"moose", "mouse", "narwhal", "newt", "nightingale",
	"octopus", "opossum", "orangutan", "oriole", "ostrich",
	"otter", "owl", "ox", "panda", "panther",
	"parrot", "peacock", "pelican", "penguin", "pigeon",
	"platypus", "pony", "porcupine", "possum", "puffin",
	"quail", "rabbit", "raccoon", "ram", "rat",
	"raven", "reindeer", "robin", "salamander", "seal",
	"seahorse", "shark", "sheep", "skunk", "sloth",
	"snail", "snake", "sparrow", "squid", "squirrel",
	"starfish", "stork", "swan", "tapir", "tiger",
	"toad", "tortoise", "toucan", "trout", "turkey",
	"turtle", "vulture", "walrus", "weasel", "whale",
	"wolf", "wombat", "woodpecker", "yak", "zebra",

	// mythical creatures
	"centaur", "chimera", "cyclops", "dragon", "fairy",
	"griffin", "hydra", "kraken", "mermaid", "minotaur",
	"pegasus", "phoenix", "pixie", "satyr", "sphinx",
	"troll", "unicorn", "wyvern", "yeti", "zephyr",
}

// NewNicknameGenerator creates a new nickname generator.
func NewNicknameGenerator() *nicknameGenerator {
	return &nicknameGenerator{
		inUse: make(map[string]struct{}),
	}
}

// Generate returns a unique nickname
func (g *nicknameGenerator) Generate() string {
	g.mu.Lock()
	defer g.mu.Unlock()

	maxAttempts := len(adjectives) * len(animals)
	for i := 0; i < maxAttempts; i++ {
		adj := adjectives[rand.IntN(len(adjectives))]
		animal := animals[rand.IntN(len(animals))]
		name := fmt.Sprintf("%s_%s", adj, animal)
		if _, exists := g.inUse[name]; !exists {
			g.inUse[name] = struct{}{}
			return name
		}
	}

	// fallback: append random number if exhausted
	for {
		adj := adjectives[rand.IntN(len(adjectives))]
		animal := animals[rand.IntN(len(animals))]
		name := fmt.Sprintf("%s_%s_%d", adj, animal, rand.IntN(1000))
		if _, exists := g.inUse[name]; !exists {
			g.inUse[name] = struct{}{}
			return name
		}
	}
}

// Release frees a nickname for reuse.
func (g *nicknameGenerator) Release(name string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	delete(g.inUse, name)
}
