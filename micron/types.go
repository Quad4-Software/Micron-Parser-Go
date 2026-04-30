// Copyright Quad4 2026
// SPDX-License-Identifier: 0BSD

package micron

// Parser converts Micron markup to HTML.
type Parser struct {
	DarkTheme      bool
	ForceMonospace bool
}

// Formatting tracks inline text styling inside backtick formatting mode.
type Formatting struct {
	Bold      bool
	Underline bool
	Italic    bool
}

// Style is a resolved foreground, background, and font style.
type Style struct {
	FG        string
	BG        string
	Bold      bool
	Underline bool
	Italic    bool
}

// State holds parser state across lines.
type State struct {
	Literal      bool
	Depth        int
	FGColor      string
	BGColor      string
	Formatting   Formatting
	DefaultAlign string
	Align        string
	DefaultFG    string
	DefaultBG    string
}

// FieldKind selects the widget produced by a field span.
type FieldKind int

const (
	FieldText FieldKind = iota
	FieldCheckbox
	FieldRadio
)

// Field is a text field, checkbox, or radio control.
type Field struct {
	Kind       FieldKind
	Name       string
	Value      string
	Label      string
	Width      int
	Masked     bool
	Prechecked bool
	Style      Style
}

// Link is an anchor with optional field submission metadata.
type Link struct {
	URL    string
	Label  string
	Fields []string
	Style  Style
}

// Partial is an asynchronously loaded micron block with optional refresh interval.
type Partial struct {
	URL            string
	RefreshSeconds int
	Style          Style
}

// linePart is one segment on a line after makeOutput.
type linePart struct {
	style   Style
	text    string
	html    string
	field   *Field
	link    *Link
	partial *Partial
}
