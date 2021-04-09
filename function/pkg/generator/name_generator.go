package generator

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

var (
	left = [...]string{
		"admiring",
		"adoring",
		"affectionate",
		"agitated",
		"amazing",
		"angry",
		"awesome",
		"beautiful",
		"blissful",
		"bold",
		"boring",
		"brave",
		"busy",
		"charming",
		"clever",
		"cool",
		"compassionate",
		"competent",
		"condescending",
		"confident",
		"cranky",
		"crazy",
		"dazzling",
		"determined",
		"distracted",
		"dreamy",
		"eager",
		"ecstatic",
		"elastic",
		"elated",
		"elegant",
		"eloquent",
		"epic",
		"exciting",
		"fervent",
		"festive",
		"flamboyant",
		"focused",
		"friendly",
		"frosty",
		"funny",
		"gallant",
		"gifted",
		"goofy",
		"gracious",
		"great",
		"happy",
		"hardcore",
		"heuristic",
		"hopeful",
		"hungry",
		"infallible",
		"inspiring",
		"interesting",
		"intelligent",
		"jolly",
		"jovial",
		"keen",
		"kind",
		"laughing",
		"loving",
		"lucid",
		"magical",
		"mystifying",
		"modest",
		"musing",
		"naughty",
		"nervous",
		"nice",
		"nifty",
		"nostalgic",
		"objective",
		"optimistic",
		"peaceful",
		"pedantic",
		"pensive",
		"practical",
		"priceless",
		"quirky",
		"quizzical",
		"recursing",
		"relaxed",
		"reverent",
		"romantic",
		"sad",
		"serene",
		"sharp",
		"silly",
		"sleepy",
		"stoic",
		"strange",
		"stupefied",
		"suspicious",
		"sweet",
		"tender",
		"thirsty",
		"trusting",
		"unruffled",
		"upbeat",
		"vibrant",
		"vigilant",
		"vigorous",
		"wizardly",
		"wonderful",
		"xenodochial",
		"youthful",
		"zealous",
		"zen",
	}

	right = [...]string{
		"karolina",
		"rafal",
		"krzysztof",
		"michal",
		"mateusz",
		"tomasz",
		"marcin",
		"damian",
		"filip",
		"artur",
		"karol",
		"maciej",
	}
)

// GenerateName generates a random name from the list of adjectives and names of some kyma creators
// formatted as "adjective-name". For example 'quizzical_rafal'. If retry is true, a random
// integer between 0 and 10 will be added to the end of the name, e.g `focused_filip3`
func GenerateName(isSuffix bool) (string, error) {
	adjIndex, err := rand.Int(rand.Reader, big.NewInt(int64(len(left))))
	if err != nil {
		return "", err
	}

	nameIndex, err := rand.Int(rand.Reader, big.NewInt(int64(len(right))))
	if err != nil {
		return "", err
	}

	name := fmt.Sprintf("%s-%s", left[adjIndex.Int64()], right[nameIndex.Int64()])
	if isSuffix {
		index, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			return "", err
		}
		name = fmt.Sprintf("%s%d", name, index.Int64())
	}
	return name, nil
}
