package helpers

// run: go test -v ./pkg/helpers

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/Flaque/filet"
	"github.com/stretchr/testify/assert"
)

var o *Origin

func TestMain(m *testing.M) {
	// Setup git clone of repo
	setup()
	os.Exit(m.Run())
}

func setup() {
	o = NewOrigin("https://github.com/snipem/monako.git", "master", ".", "test/docs")
	o.CloneDir()
}

func TestIsMarkdown(t *testing.T) {
	assert.True(t, isMarkdown("markdown.md"), "Check should be true")
	assert.True(t, isMarkdown("markdown.MD"), "Check should be true")
	assert.False(t, isMarkdown("somefolderwith.md-init/somefile.tmp"), "Asciidoc not detected correctly")
}

func TestIsAsciidoc(t *testing.T) {
	assert.True(t, isAsciidoc("asciidoc.adoc"), "Check should be true")
	assert.True(t, isAsciidoc("asciidoc.ADOC"), "Check should be true")
	assert.False(t, isAsciidoc("somefolderwith.adoc-init/somefile.tmp"), "Asciidoc not detected correctly")
}

func TestGitCommiter(t *testing.T) {
	fileName := "README.md"

	ci, err := GetCommitInfo(o.repo, fileName)

	assert.NoError(t, err, "Could not retrieve commit info")
	assert.Contains(t, ci.Committer.Email, "@")

}

func TestGitCommiterFileNotFound(t *testing.T) {
	fileName := "Not existing file...."
	_, err := GetCommitInfo(o.repo, fileName)

	assert.Error(t, err, "Expect error for non existing file")
}

func TestGitCommiterSubfolder(t *testing.T) {
	fileName := "test/config.menu.local.md"
	ci, err := GetCommitInfo(o.repo, fileName)

	assert.NoError(t, err, "Could not retrieve commit info")
	assert.Contains(t, ci.Committer.Email, "@")
}

// TestCopyDir is a test for testing the copying capability of a single directory
func TestCopyDir(t *testing.T) {

	// TODO Get own temporary test folder
	// targetDir := filepath.Join(os.TempDir(), "tmp/testrun/", t.Name())
	// defer os.RemoveAll(targetDir)

	var whitelist = []string{".md", ".png"}

	t.Run("start in single directory 'test'", func(t *testing.T) {
		o.SourceDir = "test"
		o.FileWhitelist = whitelist
		tempDir := filet.TmpDir(t, "")
		o.ComposeDir(tempDir)
		expectedTargetFile := filepath.Join(tempDir, "compose", "test_docs/test_doc_markdown.md")
		b, err := ioutil.ReadFile(expectedTargetFile)

		assert.NoError(t, err, "File not found")
		assert.Contains(t, string(b), "# Markdown Doc 1")
	})

	t.Run("start in deeper directory 'test/test_docs/'", func(t *testing.T) {
		o.SourceDir = "test/test_docs/"
		o.FileWhitelist = whitelist
		tempDir := filet.TmpDir(t, "")
		o.ComposeDir(tempDir)
		expectedTargetFile := filepath.Join(tempDir, "compose", "/test_doc_markdown.md")
		b, err := ioutil.ReadFile(expectedTargetFile)

		assert.NoError(t, err, "File not found")
		assert.Contains(t, string(b), "# Markdown Doc 1")
	})

}
