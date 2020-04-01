package compose

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/snipem/monako/internal/workarounds"
	"github.com/snipem/monako/pkg/helpers"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
)

const standardFilemode = os.FileMode(0700)

// OriginFile represents a single file of an origin
type OriginFile struct {

	// Commit is the commit info about this file
	Commit *object.Commit
	// RemotePath is the path in the origin repository
	RemotePath string
	// LocalPath is the absolute path on the local disk
	LocalPath string

	// parentOrigin of this file
	parentOrigin *Origin
}

// ComposeDir copies a subdir of a virtual filesystem to a target in the local relative filesystem.
// The copied files can be limited by a whitelist. The Git repository is used to obtain Git commit
// information
func (origin Origin) ComposeDir() {
	origin.Files = origin.getWhitelistedFiles(origin.SourceDir)

	if len(origin.Files) == 0 {
		fmt.Printf("Found no matching files in '%s' with branch '%s' in folder '%s'\n", origin.URL, origin.Branch, origin.SourceDir)
	}

	for _, file := range origin.Files {
		file.composeFile()
	}
}

// NewOrigin returns a new origin with all needed fields
func NewOrigin(url string, branch string, sourceDir string, targetDir string) *Origin {
	o := new(Origin)
	o.URL = url
	o.Branch = branch
	o.SourceDir = sourceDir
	o.TargetDir = targetDir
	return o
}

func (origin Origin) getWhitelistedFiles(startdir string) []OriginFile {

	var originFiles []OriginFile

	files, _ := origin.filesystem.ReadDir(startdir)
	for _, file := range files {

		// This is the path as stored in the remote repo
		// This can only be gathered here, because of recursing through
		// the file system
		remotePath := filepath.Join(startdir, file.Name())

		if file.IsDir() {
			// Recurse over file and add their files to originFiles
			originFiles = append(originFiles,
				origin.getWhitelistedFiles(
					remotePath,
				)...)
		} else if helpers.FileIsWhitelisted(file.Name(), origin.FileWhitelist) {

			localPath := getLocalFilePath(origin.config.ContentWorkingDir, origin.SourceDir, origin.TargetDir, remotePath)

			originFile := OriginFile{
				RemotePath: remotePath,
				LocalPath:  localPath,

				parentOrigin: &origin,
			}

			// Add the current file to the list of files returned
			originFiles = append(originFiles, originFile)
		}

	}
	return originFiles
}

func (file OriginFile) composeFile() {

	file.createParentDir()
	contentFormat := file.GetFormat()

	switch contentFormat {
	case Asciidoc, Markdown:
		file.copyMarkupFile()
	default:
		file.copyRegularFile()
	}
	fmt.Printf("%s -> %s\n", file.RemotePath, file.LocalPath)

}

// createParentDir creates the parent directories for the file in the local filesystem
func (file OriginFile) createParentDir() {
	log.Debugf("Creating local folder '%s'", filepath.Dir(file.LocalPath))
	err := os.MkdirAll(filepath.Dir(file.LocalPath), standardFilemode)
	if err != nil {
		log.Fatalf("Error when creating '%s': %s", filepath.Dir(file.LocalPath), err)
	}
}

func (file OriginFile) copyRegularFile() {

	origin := file.parentOrigin
	f, err := origin.filesystem.Open(file.RemotePath)

	if err != nil {
		log.Fatal(err)
	}
	t, err := os.Create(file.LocalPath)
	if err != nil {
		log.Fatal(err)
	}

	if _, err = io.Copy(t, f); err != nil {
		log.Fatal(err)
	}

}

func (file OriginFile) copyMarkupFile() {

	// TODO: Only use strings not []byte

	// TODO: Add GetCommitInfo function
	// commitinfo, err := GetCommitInfo(g, gitFilepath)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	bf, err := file.parentOrigin.filesystem.Open(file.RemotePath)
	if err != nil {
		log.Fatalf("Error copying markup file %s", err)
	}

	var dirty, _ = ioutil.ReadAll(bf)
	var content []byte
	contentFormat := file.GetFormat()

	if contentFormat == Markdown {
		content = workarounds.MarkdownPostprocessing(dirty)
	} else if contentFormat == Asciidoc {
		content = workarounds.AsciidocPostprocessing(dirty)
	}

	// TODO: Add ExpandFrontmatter function
	// content = []byte(ExpandFrontmatter(string(content), gitFilepath, file.Commit))

	err = ioutil.WriteFile(file.LocalPath, content, standardFilemode)
	if err != nil {
		log.Fatalf("Error writing file %s", err)
	}
}

// getLocalFilePath returns the desired local file path for a remote file in the local filesystem.
// It is based on the local absolute composeDir, the remoteDocDir to strip it's path from the local file,
// the target dir to generate the local path and the file name itself
func getLocalFilePath(composeDir, remoteDocDir string, targetDir string, remoteFile string) string {
	// Since a remoteDocDir is defined, this should not be created in the local filesystem
	relativeFilePath := strings.TrimPrefix(remoteFile, remoteDocDir)
	return filepath.Join(composeDir, targetDir, relativeFilePath)
}
