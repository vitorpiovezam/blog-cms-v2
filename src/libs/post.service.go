package libs

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"blog-cms-v2/src/definitions"
)

var titleOverrides = map[string]string{
	"angular-15-standalone-components-not-the-silver-bullet": "Angular 15 standalone components",
}

type PostService struct {
	postsPath string
	s3Bucket  string
	offline   bool
}

func NewPostService() *PostService {
	return &PostService{
		postsPath: envOrDefault("POSTS_PATH", "src/posts"),
		s3Bucket:  envOrDefault("POSTS_BUCKET", "vitorpiovezam.dev-posts"),
		offline:   os.Getenv("IS_OFFLINE") == "true" || os.Getenv("SERVE_LOCAL") == "true",
	}
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func (s *PostService) GetAllPosts(ctx context.Context) ([]definitions.Post, error) {
	readPath := s.postsPath

	if !s.offline {
		synced, err := s.refreshPostsFromBucket(ctx)
		if err != nil {
			log.Printf("s3 sync skipped (non-fatal): %v", err)
		} else {
			readPath = synced
		}
	}

	return s.readPosts(readPath)
}

func (s *PostService) GetPostBySlug(ctx context.Context, slug string) (*definitions.Post, error) {
	posts, err := s.GetAllPosts(ctx)
	if err != nil {
		return nil, err
	}
	for i := range posts {
		if posts[i].Slug == slug {
			return &posts[i], nil
		}
	}
	return nil, nil
}

func (s *PostService) refreshPostsFromBucket(ctx context.Context) (string, error) {
	dest := "/tmp/posts"
	if err := os.MkdirAll(dest, 0o755); err != nil {
		return "", fmt.Errorf("mkdir %s: %w", dest, err)
	}
	cmd := exec.CommandContext(ctx, "aws", "s3", "sync", "s3://"+s.s3Bucket, dest)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("aws s3 sync: %w", err)
	}
	return dest, nil
}

func (s *PostService) readPosts(dir string) ([]definitions.Post, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", dir, err)
	}

	posts := make([]definitions.Post, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		p, err := parsePost(filepath.Join(dir, e.Name()), e.Name())
		if err != nil {
			log.Printf("skipping %s: %v", e.Name(), err)
			continue
		}
		posts = append(posts, p)
	}

	sort.Slice(posts, func(i, j int) bool {
		return posts[i].PostDate.After(posts[j].PostDate)
	})
	return posts, nil
}

// parsePost parses a filename with the convention MM-DD-YYYY#type#Slug-Title.md.
func parsePost(fullPath, filename string) (definitions.Post, error) {
	parts := strings.SplitN(filename, "#", 3)
	if len(parts) != 3 {
		return definitions.Post{}, fmt.Errorf("unexpected filename format: %s", filename)
	}

	postDate, err := time.Parse("01-02-2006", parts[0])
	if err != nil {
		return definitions.Post{}, fmt.Errorf("parsing date %q: %w", parts[0], err)
	}

	rawSlug := parts[2]
	slug := strings.ToLower(rawSlug)
	title := deriveTitle(rawSlug)

	content, err := os.ReadFile(fullPath)
	if err != nil {
		return definitions.Post{}, fmt.Errorf("reading %s: %w", filename, err)
	}

	return definitions.Post{
		Slug:        slug,
		Title:       title,
		Type:        " " + parts[1] + " ",
		Post:        string(content),
		TextPreview: makePreview(string(content), 150),
		PostDate:    postDate,
	}, nil
}

func deriveTitle(rawSlug string) string {
	nameNoExt := strings.TrimSuffix(strings.ToLower(rawSlug), ".md")
	if override, ok := titleOverrides[nameNoExt]; ok {
		return override
	}
	title := strings.ReplaceAll(nameNoExt, "-", " ")
	if title == "" {
		return title
	}
	return strings.ToUpper(title[:1]) + title[1:]
}

var (
	reFencedCode = regexp.MustCompile("(?s)```[^`]*```")
	reInlineCode = regexp.MustCompile("`[^`]+`")
	reHTMLTag    = regexp.MustCompile(`<[^>]+>`)
	reImage      = regexp.MustCompile(`!\[[^\]]*\]\([^)]*\)`)
	reLinkText   = regexp.MustCompile(`\[([^\]]*)\]\([^)]*\)`)
	reHeading    = regexp.MustCompile(`(?m)^#{1,6}\s+`)
	reBold       = regexp.MustCompile(`\*{1,3}([^*]+)\*{1,3}`)
	reItalic     = regexp.MustCompile(`_{1,2}([^_]+)_{1,2}`)
	reBlockquote = regexp.MustCompile(`(?m)^>\s+`)
	reHRule      = regexp.MustCompile(`(?m)^[-*_]{3,}\s*$`)
)

func makePreview(markdown string, limit int) string {
	s := reFencedCode.ReplaceAllString(markdown, "")
	s = reInlineCode.ReplaceAllString(s, "")
	s = reHTMLTag.ReplaceAllString(s, "")
	s = reImage.ReplaceAllString(s, "")
	s = reLinkText.ReplaceAllString(s, "$1")
	s = reHeading.ReplaceAllString(s, "")
	s = reBold.ReplaceAllString(s, "$1")
	s = reItalic.ReplaceAllString(s, "$1")
	s = reBlockquote.ReplaceAllString(s, "")
	s = reHRule.ReplaceAllString(s, "")
	s = strings.Join(strings.Fields(s), " ")

	runes := []rune(s)
	if len(runes) > limit {
		return string(runes[:limit]) + "..."
	}
	return s + "..."
}
