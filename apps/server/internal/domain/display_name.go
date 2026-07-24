package domain

import (
	"errors"
	"strings"
	"unicode"

	"github.com/rivo/uniseg"
	"golang.org/x/text/cases"
	"golang.org/x/text/unicode/norm"
)

type DisplayName struct {
	display       string
	comparisonKey string
}

func NewDisplayName(value string) (DisplayName, error) {
	display := norm.NFC.String(strings.TrimSpace(value))
	graphemes := uniseg.GraphemeClusterCount(display)
	if graphemes < 2 || graphemes > 20 {
		return DisplayName{}, errors.New("display name must contain 2-20 grapheme clusters")
	}
	visible := false
	for _, char := range display {
		if unicode.IsControl(char) || prohibitedFormat(char) {
			return DisplayName{}, errors.New("display name contains prohibited formatting")
		}
		if !unicode.IsSpace(char) {
			visible = true
		}
	}
	if !visible {
		return DisplayName{}, errors.New("display name must contain visible content")
	}

	collapsed := strings.Join(strings.Fields(norm.NFKC.String(display)), " ")
	return DisplayName{
		display:       display,
		comparisonKey: cases.Fold().String(collapsed),
	}, nil
}

func prohibitedFormat(char rune) bool {
	if char == '\u200d' {
		return false
	}
	return unicode.In(char, unicode.Cf)
}

func (name DisplayName) String() string        { return name.display }
func (name DisplayName) ComparisonKey() string { return name.comparisonKey }
