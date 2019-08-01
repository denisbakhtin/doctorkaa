package models

import (
	"fmt"
	"strings"
)

//Page type contains page info
type Page struct {
	ID              uint64 `gorm:"primary_key" form:"id"`
	Title           string `form:"title"`
	Content         string `form:"content"`
	Slug            string `form:"slug"`
	Published       bool   `form:"published"`
	MetaKeywords    string `form:"meta_keywords"`
	MetaDescription string `form:"meta_description"`
}

//Feedback view model
type Feedback struct {
	Name    string `form:"name" binding:"required"`
	Email   string `form:"email" binding:"required"`
	Phone   string `form:"phone"`
	Message string `form:"message" binding:"required"`
}

//URL returns article url
func (p *Page) URL() string {
	return fmt.Sprintf("/%s", p.Slug)
}

//BeforeCreate gorm hook
func (p *Page) BeforeCreate() (err error) {
	if strings.TrimSpace(p.Slug) == "" {
		p.Slug = createSlug(p.Title)
	}
	return
}

//BeforeSave gorm hook
func (p *Page) BeforeSave() (err error) {
	if strings.TrimSpace(p.Slug) == "" {
		p.Slug = createSlug(p.Title)
	}
	return
}
