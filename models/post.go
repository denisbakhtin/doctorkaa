package models

import (
	"fmt"
	"regexp"
	"strings"
)

//Post type contains blog post info
type Post struct {
	ID              uint64 `gorm:"primary_key" form:"id"`
	Title           string `form:"title"`
	Content         string `form:"content"`
	Slug            string `form:"slug"`
	Published       bool   `form:"published"`
	MetaKeywords    string `form:"meta_keywords"`
	MetaDescription string `form:"meta_description"`
}

//URL returns article url
func (p *Post) URL() string {
	return fmt.Sprintf("/blog/%s", p.Slug)
}

//Excerpt is blog post summary
func (p *Post) Excerpt() string {
	re := regexp.MustCompile("<.*?>|&.*?;")
	s := re.ReplaceAllString(p.Content, "")
	return truncate(s, 300)
}

//BeforeCreate gorm hook
func (p *Post) BeforeCreate() (err error) {
	if strings.TrimSpace(p.Slug) == "" {
		p.Slug = createSlug(p.Title)
	}
	return
}

//BeforeSave gorm hook
func (p *Post) BeforeSave() (err error) {
	if strings.TrimSpace(p.Slug) == "" {
		p.Slug = createSlug(p.Title)
	}
	return
}
