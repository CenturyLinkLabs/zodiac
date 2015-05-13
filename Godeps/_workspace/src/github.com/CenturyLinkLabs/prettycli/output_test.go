package prettycli

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPlainOutput(t *testing.T) {
	assert.Equal(t, "Test", PlainOutput{"Test"}.ToPrettyOutput())
}

func TestListOutput(t *testing.T) {
	lo := ListOutput{Labels: []string{"ID"}}
	lo.AddRow(map[string]string{"ID": "10"})
	s := lo.ToPrettyOutput()

	assert.Contains(t, s, "ID")
	assert.Contains(t, s, "10")
	assert.False(t, strings.HasSuffix(s, "\n"))
}

func TestDetailOutput(t *testing.T) {
	do := DetailOutput{
		Details: map[string]string{
			"Z":    "Other",
			"Name": "Test Name",
		},
	}
	s := do.ToPrettyOutput()

	assert.Regexp(t, `Name\s+Test Name\s+Z`, s)
	assert.False(t, strings.HasSuffix(s, "\n"))
}

func TestOrderedDetailOutput(t *testing.T) {
	do := DetailOutput{
		Details: map[string]string{
			"Order Me Z": "v",
			"Order Me X": "v",
			"Alpha Z":    "v",
			"Alpha X":    "v",
		},
		Order: []string{"Order Me Z", "Order Me X"},
	}
	s := do.ToPrettyOutput()

	assert.Regexp(t, `Order Me Z\s+v\s+Order Me X\s+v\s+Alpha X\s+v\s+Alpha Z`, s)
	assert.False(t, strings.HasSuffix(s, "\n"))
}

func TestCombinedOutput(t *testing.T) {
	first := PlainOutput{"First"}
	second := PlainOutput{"Second"}
	o := CombinedOutput{}
	o.AddOutput("", first)
	o.AddOutput("Heading 2", second)
	assert.Equal(t, "First\n\nHEADING 2\nSecond", o.ToPrettyOutput())
}
