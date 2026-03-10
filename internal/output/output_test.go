package output

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		seconds float64
		want    string
	}{
		{0, "0s"},
		{30, "30s"},
		{59, "59s"},
		{60, "1m"},
		{90, "1m"},
		{120, "2m"},
		{3600, "1h 0m"},
		{3661, "1h 1m"},
		{7200, "2h 0m"},
		{5400, "1h 30m"},
		{86400, "24h 0m"},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.want, FormatDuration(tt.seconds), "FormatDuration(%v)", tt.seconds)
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		s    string
		max  int
		want string
	}{
		{"hello", 10, "hello"},
		{"hello world", 8, "hello..."},
		{"abc", 3, "abc"},
		{"abcd", 4, "abcd"},
		{"abcde", 4, "a..."},
		{"", 5, ""},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.want, Truncate(tt.s, tt.max), "Truncate(%q, %d)", tt.s, tt.max)
	}
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w

	fn()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	return buf.String()
}

func TestPrintJSON(t *testing.T) {
	out := captureStdout(t, func() {
		PrintJSON(map[string]string{"key": "value"})
	})

	var result map[string]string
	require.NoError(t, json.Unmarshal([]byte(out), &result))
	assert.Equal(t, "value", result["key"])
}

func TestPrintJSON_Array(t *testing.T) {
	out := captureStdout(t, func() {
		PrintJSON([]int{1, 2, 3})
	})

	var result []int
	require.NoError(t, json.Unmarshal([]byte(out), &result))
	assert.Equal(t, []int{1, 2, 3}, result)
}

func TestPrintJSON_Indented(t *testing.T) {
	out := captureStdout(t, func() {
		PrintJSON(map[string]string{"a": "b"})
	})
	assert.Contains(t, out, "  ")
}

func TestPrintTable(t *testing.T) {
	out := captureStdout(t, func() {
		headers := []string{"ID", "NAME"}
		rows := [][]string{{"1", "Alice"}, {"2", "Bob"}}
		PrintTable(headers, rows)
	})

	assert.Contains(t, out, "ID")
	assert.Contains(t, out, "NAME")
	assert.Contains(t, out, "Alice")
	assert.Contains(t, out, "Bob")
}

func TestPrintTable_Empty(t *testing.T) {
	out := captureStdout(t, func() {
		PrintTable([]string{"ID"}, nil)
	})

	assert.Contains(t, out, "No results.")
}

func TestPrintTable_EmptyRows(t *testing.T) {
	out := captureStdout(t, func() {
		PrintTable([]string{"ID"}, [][]string{})
	})

	assert.Contains(t, out, "No results.")
}

func TestJSONOutputFlag(t *testing.T) {
	assert.False(t, JSONOutput)
}
