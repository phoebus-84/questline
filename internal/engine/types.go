package engine

type Attribute string

const (
	// Original 4 attributes
	AttributeSTR Attribute = "STR" // Strength - fitness, physical challenges
	AttributeINT Attribute = "INT" // Intelligence - learning, problem solving
	AttributeWIS Attribute = "WIS" // Wisdom - mindfulness, self-improvement
	AttributeART Attribute = "ART" // Art - creativity, culture

	// New 5 attributes
	AttributeHOME    Attribute = "HOME"    // Household & DIY
	AttributeOUT     Attribute = "OUT"     // Outdoors & nature
	AttributeREAD    Attribute = "READ"    // Reading
	AttributeCINEMA  Attribute = "CINEMA"  // Cinema - watching movies
	AttributeCAREER  Attribute = "CAREER"  // Career & Finance
)

// AllAttributes returns all valid attributes in display order.
var AllAttributes = []Attribute{
	AttributeSTR, AttributeINT, AttributeWIS, AttributeART,
	AttributeHOME, AttributeOUT, AttributeREAD, AttributeCINEMA, AttributeCAREER,
}

func (a Attribute) IsValid() bool {
	switch a {
	case AttributeSTR, AttributeINT, AttributeWIS, AttributeART,
		AttributeHOME, AttributeOUT, AttributeREAD, AttributeCINEMA, AttributeCAREER:
		return true
	default:
		return false
	}
}

// DefaultAttribute is used when user input is missing/invalid.
const DefaultAttribute Attribute = AttributeWIS

type Difficulty int

const (
	DifficultyTrivial Difficulty = 1
	DifficultyEasy    Difficulty = 2
	DifficultyMedium  Difficulty = 3
	DifficultyHard    Difficulty = 4
	DifficultyEpic    Difficulty = 5
)

func (d Difficulty) IsValid() bool {
	return d >= DifficultyTrivial && d <= DifficultyEpic
}
