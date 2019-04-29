// Package playstyle provides an enum for Akatsuki's playstyles.
package playstyle

import "strings"

// PlayStyle is a bitwise enum containing the instruments a Akatsuki user likes
// to play with.
type PlayStyle int

// Various playstyles on Akatsuki.
const (
	Mouse int = 1 << iota
	Tablet
	TouchscreenKB
	Touchscreen
	MouseOnly
	TapX
	Vive
	Oculus
)

// Styles are string representations of the various playstyles someone can have.
var Styles = [...]string{
	"Mouse & Keyboard",
	"Tablet & Keyboard",
	"Touchscreen & Keyboard",
	"Touchscreen Only",
	"Mouse Only",
	"Tap-X",
	"HTC Vive",
	"Oculus Rift",
}

// String is the string representation of a playstyle.
func (p PlayStyle) String() string {
	var parts []string

	i := int(p)
	for k, v := range Styles {
		if i&(1<<uint(k)) > 0 {
			parts = append(parts, v)
		}
	}

	return strings.Join(parts, ", ")
}
