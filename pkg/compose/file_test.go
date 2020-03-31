package compose

// run: go test -v ./pkg/compose

import (
	"path/filepath"
	"testing"

	"github.com/alecthomas/assert"
)

func TestLocalPath(t *testing.T) {

	equalPath(t,
		"/tmp/compose/filename.md",
		getLocalFilePath("/tmp/compose", ".", ".", "filename.md"),
		"Simple setup, always first level")

	equalPath(t,
		"/tmp/compose/filename.md",
		getLocalFilePath("/tmp/compose", "docs", ".", "docs/filename.md"),
		"With remote 'docs' folder")

	equalPath(t,
		"/tmp/compose/docs/filename.md",
		getLocalFilePath("/tmp/compose", ".", ".", "docs/filename.md"),
		"With remote 'docs' folder, but keep structure")

	equalPath(t,
		"compose/filename.md",
		getLocalFilePath("./compose", ".", ".", "filename.md"),
		"Path is relative")

	equalPath(t,
		"/tmp/compose/localTarget/filename.md",
		getLocalFilePath("/tmp/compose", ".", "localTarget", "filename.md"),
		"Local Target folder")

	equalPath(t,
		"/tmp/compose/filename.md",
		getLocalFilePath("/tmp/compose", ".", "", "filename.md"),
		"Empty local target folder")
}

func equalPath(t *testing.T, expected string, actual string, msg string) {

	assert.Equal(t,
		filepath.ToSlash(expected),
		filepath.ToSlash(actual),
		msg,
	)
}
