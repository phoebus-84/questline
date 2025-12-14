package engine

type Attribute string

const (
	AttributeSTR Attribute = "STR"
	AttributeINT Attribute = "INT"
	AttributeWIS Attribute = "WIS"
	AttributeART Attribute = "ART"
)

func (a Attribute) IsValid() bool {
	switch a {
	case AttributeSTR, AttributeINT, AttributeWIS, AttributeART:
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
