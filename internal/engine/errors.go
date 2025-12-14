package engine

import "fmt"

// GateError indicates a feature is locked behind a required global level.
// This is returned by gate checks and should be shown to the user.
type GateError struct {
	Feature       string
	RequiredLevel int
}

func (e GateError) Error() string {
	if e.RequiredLevel <= 0 {
		return fmt.Sprintf("feature '%s' is locked", e.Feature)
	}
	return fmt.Sprintf("feature '%s' unlocks at level %d", e.Feature, e.RequiredLevel)
}
