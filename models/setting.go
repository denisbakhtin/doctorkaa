package models

import (
	"html/template"
	"strings"
)

//Setting type contains settings info
type Setting struct {
	ID          uint64 `gorm:"primary_key" form:"id"`
	Name        string `binding:"required" form:"name"`
	Description string `form:"description"`
	Content     string `form:"content"`
	ContentType string `form:"content_type"`
}

//GetSetting returns site setting by its name
func GetSetting(name string) template.HTML {
	db := GetDB()
	setting := Setting{}
	db.Where("name = ?", strings.ToLower(name)).First(&setting)
	return template.HTML(setting.Content)
}
