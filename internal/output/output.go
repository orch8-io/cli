package output

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"text/tabwriter"
)

// Out is the default writer for all output functions.
// It defaults to os.Stdout and can be overridden in tests.
var Out io.Writer = os.Stdout

// JSON prints v as indented JSON to Out.
func JSON(v any) {
	enc := json.NewEncoder(Out)
	enc.SetIndent("", "  ")
	enc.Encode(v)
}

// Table prints rows as an aligned table to Out.
func Table(headers []string, rows [][]string) {
	w := tabwriter.NewWriter(Out, 0, 0, 2, ' ', 0)
	for i, h := range headers {
		if i > 0 {
			fmt.Fprint(w, "\t")
		}
		fmt.Fprint(w, h)
	}
	fmt.Fprintln(w)

	for _, row := range rows {
		for i, col := range row {
			if i > 0 {
				fmt.Fprint(w, "\t")
			}
			fmt.Fprint(w, col)
		}
		fmt.Fprintln(w)
	}
	w.Flush()
}

// Errorf formats an error message and returns it as an error.
// Use with RunE: return output.Errorf("...", err)
func Errorf(msg string, args ...any) error {
	return fmt.Errorf(msg, args...)
}

// Truncate shortens s to maxLen with ellipsis.
func Truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
