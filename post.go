package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"sort"
	"strings"

	"github.com/aymerick/raymond"
	"github.com/mitchellh/mapstructure"
	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting"
	meta "github.com/yuin/goldmark-meta"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
)

// Metadata stores the data about a post that needs to be visible
// at the home page.
type Metadata struct {
	Title   string
	Summary string
	Time    int64 // unix timestamp
}

// Post stores the contents of a blog post.
type Post struct {
	Slug     string
	Metadata Metadata
	Contents string
	Image    []byte
}

func newPost(slug string) (*Post, error) {
	data, err := os.ReadFile("posts/" + slug + ".md")
	if err != nil {
		return nil, fmt.Errorf("could not read file: %s", err)
	}

	md := goldmark.New(
		goldmark.WithExtensions(
			extension.Linkify,
			extension.Strikethrough,
			extension.Typographer,
			extension.Footnote,
			meta.Meta,
			highlighting.Highlighting,
		),
		goldmark.WithRendererOptions(
			html.WithUnsafe(),
		),
	)
	var converted bytes.Buffer
	ctx := parser.NewContext()
	err = md.Convert(data, &converted, parser.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("could not parse markdown: %s", err)
	}
	mdMap, err := meta.TryGet(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not parse metadata: %s", err)
	}
	var metadata Metadata
	err = mapstructure.Decode(mdMap, &metadata)
	if err != nil {
		return nil, fmt.Errorf("could not destructure metadata: %s", err)
	}

	post := &Post{
		Slug:     slug,
		Metadata: metadata,
		Contents: converted.String(),
	}

	url := blogURL + "/" + slug
	var buf bytes.Buffer
	err = createImage(post.Metadata.Title, post.Metadata.Summary, url, &buf)
	if err != nil {
		return nil, fmt.Errorf("could not create post image: %v", err)
	}
	post.Image, err = io.ReadAll(&buf)
	if err != nil {
		return nil, err
	}

	return post, nil
}

func (p *Post) render(tpl *raymond.Template) (string, error) {
	return tpl.Exec(p)
}

func (p *Post) String() string {
	return p.Slug
}

type postList []*Post

func newPostList() (postList, error) {
	files, err := os.ReadDir("posts/")
	if err != nil {
		return nil, err
	}

	pl := make(postList, 0, len(files))
	for _, f := range files {
		filename := f.Name()

		if strings.HasSuffix(filename, ".md") {
			post, err := newPost(strings.TrimSuffix(filename, ".md"))
			if err != nil {
				return nil, fmt.Errorf("could not render %s: %s", filename, err)
			}
			pl = append(pl, post)
			log.Printf("Loaded post %s", filename)
		}
	}
	sort.Sort(pl)

	return pl, nil
}

func insertOrUpdatePost(pl postList, p *Post) postList {
	for i, post := range pl {
		if post.Slug == p.Slug {
			pl[i] = p
			sort.Sort(pl)
			return pl
		}
	}
	pl = append(pl, p)
	sort.Sort(pl)
	return pl
}

func removePost(pl postList, slug string) postList {
	for i, post := range pl {
		if post.Slug == slug {
			pl = append(pl[:i], pl[i+1:]...)
			break
		}
	}
	fmt.Println(pl)
	return pl
}

// Len implements sort.Interface
func (pl postList) Len() int {
	return len(pl)
}

// Less implements sort.Interface
func (pl postList) Less(i, j int) bool {
	return pl[i].Metadata.Time > pl[j].Metadata.Time
}

// Swap implements sort.Interface
func (pl postList) Swap(i, j int) {
	temp := pl[i]
	pl[i] = pl[j]
	pl[j] = temp
}

func newPostListener(update func(func(postList) postList)) *listener {
	ln := &listener{
		folder: "posts/",
		update: func(file string) error {
			post, err := newPost(strings.TrimSuffix(file, ".md"))
			if err != nil {
				return err
			}
			update(func(oldList postList) postList {
				return insertOrUpdatePost(oldList, post)
			})
			return nil
		},
		clean: func(file string) error {
			update(func(oldList postList) postList {
				return removePost(oldList, strings.TrimSuffix(file, ".md"))
			})
			return nil
		},
	}
	return ln
}
