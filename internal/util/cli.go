package util

import "strings"

const (
	ModeStdOut     = "stdout"
	ModeFileWriter = "file"
	ModeTest       = "test"

	FileExtJSON = "json"
	FileExtYAML = "yaml"
	FileExtYML  = "yml"

	FormatJSON = "json"
	FormatYAML = "yaml"
)

type CLIOutput string

func (c CLIOutput) GetMode() string {
	output := string(c)
	if output != "" {
		if strings.Contains(output, FileExtJSON) ||
			strings.Contains(output, FileExtYAML) ||
			strings.Contains(output, FileExtYML) {
			return ModeFileWriter
		}
	}
	return ModeStdOut
}

func (c CLIOutput) GetFormat() string {
	output := string(c)
	format := FormatJSON
	if strings.Contains(output, FileExtYAML) || strings.Contains(output, FileExtYML) {
		return FormatYAML
	}
	return format
}
