package output

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"
)

// JSON prints v as indented JSON to stdout.
func JSON(v any) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.Encode(v)
}

// Table prints rows as an aligned table to stdout.
func Table(headers []string, rows [][]string) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
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

// Error prints an error message to stderr and exits with code 1.
func Error(msg string, args ...any) {
	fmt.Fprintf(os.Stderr, "Error: "+msg+"\n", args...)
	os.Exit(1)
}

// Truncate shortens s to maxLen with ellipsis.
func Truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
