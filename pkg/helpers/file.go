package helpers

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/snipem/monako/internal/workarounds"
	"gopkg.in/src-d/go-billy.v4"
	"gopkg.in/src-d/go-billy.v4/memfs"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/http"
	"gopkg.in/src-d/go-git.v4/storage/memory"
)

var filemode = os.FileMode(0700)

const Asciidoc = "ASCIIDOC"
const Markdown = "MARKDOWN"

func CloneDir(url string, branch string, username string, password string) (*git.Repository, billy.Filesystem) {

	log.Printf("Cloning in to %s with branch %s", url, branch)

	fs := memfs.New()

	basicauth := http.BasicAuth{}

	if username != "" && password != "" {
		log.Printf("Using username and password")
		basicauth = http.BasicAuth{
			Username: username,
			Password: password,
		}
	}

	repo, err := git.Clone(memory.NewStorage(), fs, &git.CloneOptions{
		URL:           url,
		Depth:         1,
		ReferenceName: plumbing.ReferenceName(fmt.Sprintf("refs/heads/%s", branch)),
		SingleBranch:  true,
		Auth:          &basicauth,
	})

	if err != nil {
		log.Fatal(err)
	}

	return repo, fs
}

func shouldIgnoreFile(filename string, whitelist []string) bool {
	for _, whitelisted := range whitelist {
		if strings.HasSuffix(strings.ToLower(filename), strings.ToLower(whitelisted)) {
			return false
		}
	}
	return true
}

func isMarkdown(filename string) bool {
	return strings.HasSuffix(strings.ToLower(filename), strings.ToLower(".md"))
}

func isAsciidoc(filename string) bool {
	return strings.HasSuffix(strings.ToLower(filename), strings.ToLower(".adoc")) ||
		strings.HasSuffix(strings.ToLower(filename), strings.ToLower(".asciidoc")) ||
		strings.HasSuffix(strings.ToLower(filename), strings.ToLower(".asc"))
}

func DetermineFormat(filename string) string {
	if isMarkdown(filename) {
		return Markdown
	} else if isAsciidoc(filename) {
		return Asciidoc
	} else {
		return ""
	}
}

func CopyDir(fs billy.Filesystem, subdir string, target string, whitelist []string) {

	log.Printf("Copying subdir '%s' to target dir %s", subdir, target)
	var files []os.FileInfo

	_, err := fs.Stat(subdir)
	if err != nil {
		log.Fatalf("Error while reading subdir %s: %s", subdir, err)
	}

	fs, err = fs.Chroot(subdir)
	if err != nil {
		log.Fatal(err)
	}

	files, err = fs.ReadDir(".")
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range files {

		if file.IsDir() {
			// TODO is this memory consuming or is fsSubdir freed after recursion?
			// fsSubdir := fs
			CopyDir(fs, file.Name(), target+"/"+file.Name(), whitelist)
			continue
		} else if shouldIgnoreFile(file.Name(), whitelist) {
			continue
		}

		f, err := fs.Open(file.Name())
		if err != nil {
			log.Fatal(err)
		}

		err = os.MkdirAll(target, filemode)
		if err != nil {
			log.Fatal(err)
		}

		var targetFilename = target + "/" + file.Name()
		contentFormat := DetermineFormat(file.Name())

		switch contentFormat {
		case Asciidoc:
		case Markdown:

			var dirty, _ = ioutil.ReadAll(f)
			var content []byte

			if contentFormat == Markdown {

				content = workarounds.MarkdownPostprocessing(dirty)
				// content = file.ExpandFrontmatter(content)

			} else if contentFormat == Asciidoc {

				content = workarounds.AsciidocPostprocessing(dirty)
				// content = file.ExpandFrontmatter(content)

			}
			ioutil.WriteFile(targetFilename, content, filemode)

		default:
			copyFile(targetFilename, f)
		}

		log.Printf("%s -> %s\n", file.Name(), targetFilename)

	}

}

func copyFile(targetFilename string, from io.Reader) {
	t, err := os.Create(targetFilename)
	if err != nil {
		log.Fatal(err)
	}

	if _, err = io.Copy(t, from); err != nil {
		log.Fatal(err)
	}

}

func GetCommitInfo(r *git.Repository, filename string) *object.Commit {

	cIter, err := r.Log(&git.LogOptions{
		FileName: &filename,
		All:      true,
	})

	if err != nil {
		log.Fatal(err)
	}

	commit, err := cIter.Next()
	if err != nil {
		log.Fatal(err)
	}

	return commit
}
