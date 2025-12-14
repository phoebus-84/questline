package engine

import "strings"

// ParseAttribute parses user input (str|int|wis|art) to an Attribute.
// If input is empty, returns DefaultAttribute.
func ParseAttribute(input string) Attribute {
	s := strings.TrimSpace(strings.ToLower(input))
	switch s {
	case "":
		return DefaultAttribute
	case "str":
		return AttributeSTR
	case "int":
		return AttributeINT
	case "wis":
		return AttributeWIS
	case "art":
		return AttributeART
	default:
		return DefaultAttribute
	}
}
