package storage

import (
	"encoding/xml"
	"fmt"
	"io"
	"strings"
)

// ValidationError represents a Storage XHTML validation failure.
type ValidationError struct {
	Message string
	Tag     string
}

func (e *ValidationError) Error() string {
	if e.Tag != "" {
		return fmt.Sprintf("validation error: %s (tag: %s)", e.Message, e.Tag)
	}
	return fmt.Sprintf("validation error: %s", e.Message)
}

// ForbiddenTags are HTML tags not allowed in Confluence Storage Format.
var ForbiddenTags = map[string]bool{
	"thead":    true,
	"tfoot":    true,
	"colgroup": true,
	"col":      true,
	"div":      true,
	"span":     true,
	"script":   true,
	"style":    true,
	"iframe":   true,
	"form":     true,
	"input":    true,
	"button":   true,
}

// AllowedMacros is a configurable allowlist of permitted macro names.
// If empty, all macros are allowed.
var AllowedMacros = map[string]bool{}

// ValidatorOptions configures validation behavior.
type ValidatorOptions struct {
	// RequireTableTbody requires tables to have <tbody> elements.
	RequireTableTbody bool
	// AllowedMacros limits which macros are permitted. Empty means all allowed.
	AllowedMacros map[string]bool
	// ForbiddenTags specifies tags that are not allowed.
	ForbiddenTags map[string]bool
}

// DefaultValidatorOptions returns the default validation options.
func DefaultValidatorOptions() ValidatorOptions {
	return ValidatorOptions{
		RequireTableTbody: true,
		AllowedMacros:     nil,
		ForbiddenTags:     ForbiddenTags,
	}
}

// Validate checks if the given string is valid Confluence Storage XHTML.
func Validate(xhtml string) error {
	return ValidateWithOptions(xhtml, DefaultValidatorOptions())
}

// ValidateWithOptions checks if the given string is valid Confluence Storage XHTML
// using the provided options.
func ValidateWithOptions(xhtml string, opts ValidatorOptions) error {
	if xhtml == "" {
		return nil
	}

	// Wrap in root element for XML parsing
	wrapped := "<root xmlns:ac=\"http://atlassian.com/confluence\" xmlns:ri=\"http://atlassian.com/confluence\">" + xhtml + "</root>"
	decoder := xml.NewDecoder(strings.NewReader(wrapped))
	decoder.Entity = htmlEntities

	var tableDepth int
	var tbodyFound bool

	for {
		tok, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return &ValidationError{Message: fmt.Sprintf("invalid XML: %v", err)}
		}

		switch t := tok.(type) {
		case xml.StartElement:
			name := t.Name.Local

			// Check forbidden tags
			if opts.ForbiddenTags[name] {
				return &ValidationError{Message: "forbidden tag", Tag: name}
			}

			// Check allowed macros
			if name == "structured-macro" && len(opts.AllowedMacros) > 0 {
				macroName := getMacroName(t)
				if macroName != "" && !opts.AllowedMacros[macroName] {
					return &ValidationError{Message: "disallowed macro", Tag: macroName}
				}
			}

			// Track table structure
			if name == "table" {
				tableDepth++
				tbodyFound = false
			}
			if name == "tbody" && tableDepth > 0 {
				tbodyFound = true
			}
			if name == "tr" && tableDepth > 0 && !tbodyFound && opts.RequireTableTbody {
				return &ValidationError{Message: "<tr> must be inside <tbody>", Tag: "tr"}
			}

		case xml.EndElement:
			if t.Name.Local == "table" {
				if !tbodyFound && opts.RequireTableTbody {
					return &ValidationError{Message: "<table> must contain <tbody>", Tag: "table"}
				}
				tableDepth--
			}
		}
	}

	return nil
}

// getMacroName extracts the macro name from an ac:structured-macro element.
func getMacroName(el xml.StartElement) string {
	for _, attr := range el.Attr {
		if attr.Name.Local == "name" {
			return attr.Value
		}
	}
	return ""
}

// ValidateBlock validates a single block's rendered output.
func ValidateBlock(block Block) error {
	xhtml, err := RenderBlock(block)
	if err != nil {
		return err
	}
	return Validate(xhtml)
}

// ValidateBlockWithOptions validates a single block's rendered output with options.
func ValidateBlockWithOptions(block Block, opts ValidatorOptions) error {
	xhtml, err := RenderBlock(block)
	if err != nil {
		return err
	}
	return ValidateWithOptions(xhtml, opts)
}

// MustValidate panics if validation fails (useful for tests).
func MustValidate(xhtml string) {
	if err := Validate(xhtml); err != nil {
		panic(err)
	}
}

// IsValidXML checks if the string is well-formed XML.
func IsValidXML(xhtml string) bool {
	wrapped := "<root>" + xhtml + "</root>"
	decoder := xml.NewDecoder(strings.NewReader(wrapped))
	decoder.Entity = htmlEntities
	for {
		_, err := decoder.Token()
		if err == io.EOF {
			return true
		}
		if err != nil {
			return false
		}
	}
}
