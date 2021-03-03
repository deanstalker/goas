package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCLIOutput(t *testing.T) {
	tests := map[string]struct {
		output     string
		wantMode   string
		wantFormat string
	}{
		"invalid file path defaults to stdout and json": {
			output:     "./test.csv",
			wantMode:   ModeStdOut,
			wantFormat: FormatJSON,
		},
		"yaml file switches mode to file writer and yaml": {
			output:     "./test.yaml",
			wantMode:   ModeFileWriter,
			wantFormat: FormatYAML,
		},
		"yml file switches mode to file writer and yaml": {
			output:     "./test.yml",
			wantMode:   ModeFileWriter,
			wantFormat: FormatYAML,
		},
		"json file switches mode to file writer and json": {
			output:     "./test.json",
			wantMode:   ModeFileWriter,
			wantFormat: FormatJSON,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			output := CLIOutput(tc.output)
			assert.Equal(t, tc.wantMode, output.GetMode())
			assert.Equal(t, tc.wantFormat, output.GetFormat())
		})
	}
}
