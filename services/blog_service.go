package services

import (
	"path/filepath"
	"io/ioutil"
	"gopkg.in/russross/blackfriday.v2"
	"encoding/json"
)

const BlogConfigFilename = "blog-config.json"

type BlogPost struct {
	// Slug is the post's slug
	Slug string

	// The Title of the blog post
	Title string

	// The Summary of the blog post
	Summary string

	// Image is the image to use for the post
	Image string

	// ImageAlt is the image alt to use for the post
	ImageAlt string

	// HTMLContent is the blog posts content
	HTMLContent string
}

type BlogService interface {
	// Posts returns all blog posts
	Posts() []*BlogPost

	// PostBySlug returns the blog post using its slug
	// Returns nil if the post does not exist
	PostBySlug(slug string) *BlogPost
}

// InMemoryBlogService is the in memory implementation of BlogService
type InMemoryBlogService struct {
	Logger          Logger
	BlogMarkdownDir string

	BlogPosts map[string]*BlogPost
}

type InMemoryBlogConfig struct {
	Posts []*InMemoryBlogPostConfig `json:"posts"`
}

func newBlankInMemoryBlogConfig() *InMemoryBlogConfig {
	return &InMemoryBlogConfig{
		Posts: make([]*InMemoryBlogPostConfig, 0),
	}
}

// TODO: Validate fields are not empty
type InMemoryBlogPostConfig struct {
	// Slug is the post's slug
	Slug string `json:"slug"`

	// The Title of the blog post
	Title string `json:"title"`

	// The Summary of the blog post
	Summary string `json:"summary"`

	// Image is the image to use for the post
	Image string `json:"image"`

	// ImageAlt is the image alt to use for the post
	ImageAlt string `json:"imageAlt"`

	// Filename
	Filename string `json:"filename"`
}

func NewInMemoryBlogService(logger Logger, blogMarkdownDirectory string) (*InMemoryBlogService, error) {
	blogConfigPath := filepath.Join(blogMarkdownDirectory, BlogConfigFilename)
	logger.Debug("Reading blog config file", Fields{"path": blogConfigPath})

	configFileData, err := ioutil.ReadFile(blogConfigPath)
	if err != nil {
		logger.Debug("Unable to find in memory blog post config file", Fields{"path": blogConfigPath})
		return nil, err
	}

	inMemoryBlogConfig := newBlankInMemoryBlogConfig()
	err = json.Unmarshal(configFileData, inMemoryBlogConfig)
	if err != nil {
		logger.Debug("In memory blog post file found, but invalid", Fields{"path": blogConfigPath})
		return nil, err
	}

	blogPosts := make(map[string]*BlogPost, 0)

	for _, blogPost := range inMemoryBlogConfig.Posts {
		filePath := filepath.Join(blogMarkdownDirectory, blogPost.Filename)
		logger.Debug("Loading blog post", Fields{"path": filePath, "slug": blogPost.Slug, "title": blogPost.Title})

		bytes, err := ioutil.ReadFile(filePath)
		if err != nil {
			logger.Debug("Unable to load blog post markdown from disk", Fields{"path": filePath, "slug": blogPost.Slug, "title": blogPost.Title})
			return nil, err
		}

		blogHtml := string(blackfriday.Run(bytes))

		loadedBlogPost := &BlogPost{
			Slug:        blogPost.Slug,
			Title:       blogPost.Title,
			Summary:     blogPost.Summary,
			Image:       blogPost.Image,
			ImageAlt:    blogPost.ImageAlt,
			HTMLContent: blogHtml,
		}

		blogPosts[loadedBlogPost.Slug] = loadedBlogPost
	}

	return &InMemoryBlogService{
		Logger:          logger,
		BlogMarkdownDir: blogMarkdownDirectory,
		BlogPosts:       blogPosts,
	}, nil
}

func (bs *InMemoryBlogService) Posts() []*BlogPost {
	posts := make([]*BlogPost, 0)
	for _, post := range bs.BlogPosts {
		posts = append(posts, post)
	}
	return posts
}

func (bs *InMemoryBlogService) PostBySlug(slug string) *BlogPost {
	return bs.BlogPosts[slug]
}
