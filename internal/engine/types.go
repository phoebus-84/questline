//go:build ignore
// +build ignore

package engine
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

























}	return d >= DifficultyTrivial && d <= DifficultyEpicfunc (d Difficulty) IsValid() bool {)	DifficultyEpic    Difficulty = 5	DifficultyHard    Difficulty = 4	DifficultyMedium  Difficulty = 3	DifficultyEasy    Difficulty = 2	DifficultyTrivial Difficulty = 1const (type Difficulty intconst DefaultAttribute Attribute = AttributeWIS// DefaultAttribute is used when user input is missing/invalid.}	}		return false	default:		return true	case AttributeSTR, AttributeINT, AttributeWIS, AttributeART:	switch a {func (a Attribute) IsValid() bool {