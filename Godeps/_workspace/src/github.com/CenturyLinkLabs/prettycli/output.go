package prettycli

import (
	"bytes"
	"fmt"
	"sort"
	"strings"
	"text/tabwriter"
)

type Output interface {
	ToPrettyOutput() string
}

type ListOutput struct {
	Labels []string
	Rows   []map[string]string
}

func (o *ListOutput) AddRow(m map[string]string) {
	o.Rows = append(o.Rows, m)
}

func (o ListOutput) ToPrettyOutput() string {
	b := bytes.NewBuffer([]byte{})
	w := tabwriter.NewWriter(b, 0, 8, 2, '\t', 0)

	// Print Heading Row
	for _, l := range o.Labels {
		fmt.Fprintf(w, "%v\t", strings.ToUpper(l))
	}
	fmt.Fprintln(w)

	// Print Rows
	for _, r := range o.Rows {
		row := make([]string, len(o.Labels))
		for i, l := range o.Labels {
			row[i] = r[l]
		}

		fmt.Fprintf(w, "%s\n", strings.Join(row, "\t"))
	}

	w.Flush()
	return strings.TrimSpace(b.String())
}

type PlainOutput struct {
	Output string
}

func (o PlainOutput) ToPrettyOutput() string {
	return o.Output
}

type DetailOutput struct {
	Details map[string]string
	Order   []string
}

func (o DetailOutput) ToPrettyOutput() string {
	b := bytes.NewBuffer([]byte{})
	w := tabwriter.NewWriter(b, 0, 8, 2, '\t', 0)

	var keys []string
	for k := range o.Details {
		shouldBeOrdered := true
		for _, ok := range o.Order {
			if k == ok {
				shouldBeOrdered = false
			}
		}

		if shouldBeOrdered {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)

	for _, k := range append(o.Order, keys...) {
		fmt.Fprintf(w, "%s\t%v\n", k, o.Details[k])
	}

	w.Flush()
	return strings.TrimSpace(b.String())
}

type CombinedOutput struct {
	Outputs []sectionedOutput
}

type sectionedOutput struct {
	Heading string
	Output  Output
}

func (o *CombinedOutput) AddOutput(heading string, output Output) {
	o.Outputs = append(o.Outputs, sectionedOutput{heading, output})
}

func (o *CombinedOutput) ToPrettyOutput() string {
	outputStrs := make([]string, len(o.Outputs))
	for i, out := range o.Outputs {
		var s string
		if out.Heading == "" {
			s = out.Output.ToPrettyOutput()
		} else {
			s = fmt.Sprintf("%s\n%s", strings.ToUpper(out.Heading), out.Output.ToPrettyOutput())
		}
		outputStrs[i] = s
	}
	return strings.Join(outputStrs, "\n\n")
}
