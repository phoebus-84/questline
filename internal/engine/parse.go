package engine

import "strings"

// ParseAttribute parses user input to an Attribute.
// Supported: str, int, wis, art, home, out, read, cinema, career
// If input is empty or unrecognized, returns DefaultAttribute.
func ParseAttribute(input string) Attribute {
	s := strings.TrimSpace(strings.ToLower(input))
	switch s {
	case "":
		return DefaultAttribute
	// Original 4
	case "str":
		return AttributeSTR
	case "int":
		return AttributeINT
	case "wis":
		return AttributeWIS
	case "art":
		return AttributeART
	// New 5
	case "home":
		return AttributeHOME
	case "out", "outdoors":
		return AttributeOUT
	case "read", "reading":
		return AttributeREAD
	case "cinema", "movies", "film":
		return AttributeCINEMA
	case "career", "finance", "work":
		return AttributeCAREER
	default:
		return DefaultAttribute
	}
}
