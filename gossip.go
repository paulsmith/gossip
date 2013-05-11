package gossip

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/russross/blackfriday"
)

// Format represents a source markup that can be converted to some output
// (namely HTML)
type Format int

const (
	MARKDOWN Format = iota
	HTML
)

var formats = map[string]Format{
	"txt":  MARKDOWN,
	"md":   MARKDOWN,
	"html": HTML,
}

// Convert converts a byte slice in a particular format to the format's output
// (namely HTML)
func (f Format) Convert(input []byte) []byte {
	switch f {
	case MARKDOWN:
		return blackfriday.MarkdownBasic(input)
	default:
		return input
	}
}

// Site is a generated static site, written to "Dest" from source files at
// "Source"
type Site struct {
	Source string
	Dest   string
}

func NewSite(source, dest string) *Site {
	if source == "" {
		source = "."
	}
	if dest == "" {
		dest = "./_site"
	}
	return &Site{source, dest}
}

func (s *Site) Generate() error {
	err := s.copyTree()
	if err != nil {
		return err
	}

	err = s.generatePosts()
	if err != nil {
		return err
	}

	return nil
}

func (s *Site) generatePosts() error {
	postsDir := filepath.Join(s.Source, "posts")
	tmplDir := filepath.Join(s.Source, "templates")
	if !exists(postsDir) || !exists(tmplDir) {
		return errors.New("gossip: posts and templates directories must exist")
	}
	tmpl := template.Must(template.ParseFiles(filepath.Join(tmplDir, "default.html")))
	entries, err := ioutil.ReadDir(postsDir)
	if err != nil {
		return err
	}
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), ".") {
			continue
		}

		path := filepath.Join(postsDir, e.Name())

		post, err := NewPostFromPath(path)
		if err != nil {
			return err
		}

		year, month := post.dateParts()
		dir := filepath.Join(s.Dest, year, month)
		os.MkdirAll(dir, 0755)

		f, err := os.Create(filepath.Join(dir, post.destFileName))
		if err != nil {
			return err
		}
		defer f.Close()

		post.Generate(f, tmpl)
	}
	return nil
}

// copyTree recursively copies files from the source dir to the dest, preserving
// directory structure, and skipping site-specific directories like "posts" and
// "templates"
func (s *Site) copyTree() error {
	ok := true
	filepath.Walk(s.Source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			ok = false
			return err
		}
		// Skip the posts and templates dirs
		if info.IsDir() && (info.Name() == "posts" || info.Name() == "templates") {
			return filepath.SkipDir
		}
		rel, err := filepath.Rel(s.Source, path)
		if err != nil {
			ok = false
			return err
		}
		// Filter
		if strings.HasPrefix(rel, ".") || strings.Contains(rel, "/.") {
			return nil
		}
		destName := filepath.Join(s.Dest, rel)
		if info.IsDir() {
			os.Mkdir(destName, info.Mode())
		} else {
			if _, err := copyFile(path, destName); err != nil {
				ok = false
				return err
			}
		}
		return nil
	})
	if !ok {
		return errors.New("error(s) copying source tree")
	}
	return nil
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func copyFile(srcName, destName string) (written int64, err error) {
	src, err := os.Open(srcName)
	if err != nil {
		return
	}
	defer src.Close()

	dest, err := os.Create(destName)
	if err != nil {
		return
	}
	defer dest.Close()

	fi, err := src.Stat()
	if err != nil {
		return
	}
	dest.Chmod(fi.Mode())

	return io.Copy(dest, src)
}

// Post is a blog post entry
type Post struct {
	content      []byte
	pubdate      time.Time
	destFileName string
	format       Format
	fileInfo     os.FileInfo
}

func NewPostFromPath(path string) (*Post, error) {
	fi, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	content, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	idx := strings.LastIndex(fi.Name(), ".")
	newName := fi.Name()[:idx] + ".html"

	ext := fi.Name()[idx+1:]
	fmt, ok := formats[ext]
	if !ok {
		return nil, errors.New("unknown format " + ext)
	}

	// TODO: get pubdate from contents
	pubdate := fi.ModTime()

	return &Post{
		content:      content,
		pubdate:      pubdate,
		fileInfo:     fi,
		destFileName: newName,
		format:       fmt,
	}, nil
}

func (p *Post) dateParts() (year, month string) {
	year = fmt.Sprintf("%d", p.pubdate.Year())
	month = fmt.Sprintf("%02d", p.pubdate.Month())
	return
}

func (p *Post) Generate(wr io.Writer, tmpl *template.Template) error {
	fmtContent := p.format.Convert(p.content)
	return tmpl.Execute(wr, struct{ Content string }{string(fmtContent)})
}

// Context is the object passed to the template for rendering
type Context struct {
	Content string
}
