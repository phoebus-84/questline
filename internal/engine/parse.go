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

// ParseAttributes parses multi-attribute input like "str:50,int:50" or "str" (100%).
// Returns primary attribute and weight map.
// Format: "attr1:weight1,attr2:weight2,..." or just "attr" for single attribute.
func ParseAttributes(input string) (primary Attribute, weights map[Attribute]int) {
	input = strings.TrimSpace(input)
	if input == "" {
		return DefaultAttribute, nil
	}

	// Check if it's a single attribute (no colon, no comma)
	if !strings.Contains(input, ":") && !strings.Contains(input, ",") {
		return ParseAttribute(input), nil
	}

	weights = make(map[Attribute]int)
	parts := strings.Split(input, ",")
	var firstAttr Attribute

	for i, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		// Parse "attr:weight" or just "attr" (defaults to equal weight)
		var attrStr string
		var weight int = 100

		if idx := strings.Index(part, ":"); idx != -1 {
			attrStr = strings.TrimSpace(part[:idx])
			weightStr := strings.TrimSpace(part[idx+1:])
			if w := parseWeight(weightStr); w > 0 {
				weight = w
			}
		} else {
			attrStr = part
		}

		attr := ParseAttribute(attrStr)
		if i == 0 {
			firstAttr = attr
		}
		weights[attr] = weight
	}

	// If only one attribute specified, no need for weights map
	if len(weights) == 1 {
		return firstAttr, nil
	}

	return firstAttr, weights
}

func parseWeight(s string) int {
	var w int
	for _, c := range s {
		if c >= '0' && c <= '9' {
			w = w*10 + int(c-'0')
		}
	}
	return w
}
