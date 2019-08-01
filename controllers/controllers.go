package controllers

import (
	"fmt"
	"html/template"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/denisbakhtin/doctorkaa/config"
	"github.com/denisbakhtin/doctorkaa/models"
	"github.com/gin-gonic/gin"
	csrf "github.com/utrack/gin-csrf"
)

const userIDKey = "UserID"

var tmpl *template.Template

//Option represents select option entry
type Option struct {
	Value string
	Text  string
}

//DefaultH returns common to all pages template data
func DefaultH(c *gin.Context) gin.H {
	return gin.H{
		"Title":           "", //page title:w
		"Context":         c,
		"Csrf":            csrf.GetToken(c),
		"MetaKeywords":    "",
		"MetaDescription": "",
	}
}

//LoadTemplates loads templates from views directory to gin engine
func LoadTemplates(router *gin.Engine) {
	router.SetFuncMap(getFuncMap())
	router.LoadHTMLGlob("views/**/*")
}

func getFuncMap() template.FuncMap {
	return template.FuncMap{
		"isActiveLink":        isActiveLink,
		"stringInSlice":       stringInSlice,
		"formatDateTime":      formatDateTime,
		"now":                 now,
		"activeUserEmail":     activeUserEmail,
		"activeUserID":        activeUserID,
		"isUserAuthenticated": isUserAuthenticated,
		"signUpEnabled":       SignUpEnabled,
		"noescape":            noescape,
		"navbarMenuItems":     navbarMenuItems,
		"refEqUint":           refEqUint,
		"getSetting":          getSetting,
		"isNotBlank":          isNotBlank,
		"tel":                 tel,
		"slides":              homeSlides,
		"fullResetURL":        fullResetURL,
		"adminMenuItems":      adminMenuItems,
		"mapAPIKey":           mapAPIKey,
	}
}

//atouint converts string to uint, returns 0 if error
func atouint(s string) uint {
	i, _ := strconv.ParseUint(s, 10, 32)
	return uint(i)
}

//atouint64 converts string to uint64, returns 0 if error
func atouint64(s string) uint64 {
	i, _ := strconv.ParseUint(s, 10, 64)
	return i
}

//isActiveLink checks uri against currently active (uri, or nil) and returns "active" if they are equal
func isActiveLink(c *gin.Context, uri string) string {
	if c != nil && c.Request.RequestURI == uri {
		return "active"
	}
	return ""
}

//formatDateTime prints timestamp in human format
func formatDateTime(t time.Time) string {
	return fmt.Sprintf("%d-%02d-%02d %02d:%02d:%02d", t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second())
}

//stringInSlice returns true if value is in list slice
func stringInSlice(value string, list []string) bool {
	for i := range list {
		if value == list[i] {
			return true
		}
	}
	return false
}

//now returns current timestamp
func now() time.Time {
	return time.Now()
}

//activeUserEmail returns currently authenticated user email
func activeUserEmail(c *gin.Context) string {
	if c != nil {
		u, _ := c.Get("User")
		if user, ok := u.(*models.User); ok {
			return user.Email
		}
	}
	return ""
}

//activeUserID returns currently authenticated user ID
func activeUserID(c *gin.Context) uint64 {
	if c != nil {
		u, _ := c.Get("User")
		if user, ok := u.(*models.User); ok {
			return user.ID
		}
	}
	return 0
}

//isUserAuthenticated returns true is user is authenticated
func isUserAuthenticated(c *gin.Context) bool {
	if c != nil {
		u, _ := c.Get("User")
		if _, ok := u.(*models.User); ok {
			return true
		}
	}
	return false
}

//SignUpEnabled returns true if sign up is enabled by config
func SignUpEnabled() bool {
	return config.GetConfig().SignupEnabled
}

//noescape unescapes html content
func noescape(content string) template.HTML {
	return template.HTML(content)
}

//top level menu items
func navbarMenuItems() []models.MenuItem {
	db := models.GetDB()
	var menus []models.MenuItem
	db.Order("ord asc").Find(&menus, "parent_id is null")
	return menus
}

//refEqUint checks if *uint64 equals uint64
func refEqUint(ref *uint64, val uint64) bool {
	if ref == nil {
		return false
	}
	return *ref == val
}

func getSetting(name string) template.HTML {
	return models.GetSetting(name)
}

func isNotBlank(content string) bool {
	return len(content) > 0 && content != "<p>&nbsp;</p>"
}

func tel(content string) string {
	reg, err := regexp.Compile("[^\\+0-9]+")
	if err != nil {
		log.Fatal(err)
	}
	processedString := reg.ReplaceAllString(content, "")
	return processedString
}

func homeSlides() []models.Slide {
	var slides []models.Slide
	models.GetDB().Order("ord").Find(&slides)
	return slides
}

func isReleaseMode() bool {
	return gin.Mode() == gin.ReleaseMode
}

func isContactPage(slug string) bool {
	return strings.Contains(strings.ToLower(slug), "contact")
}

func getContactsPage() *models.Page {
	page := models.Page{}
	models.GetDB().First(&page, "slug like ?", "%contact%")
	return &page
}

func fullResetURL(hash string) string {
	return fmt.Sprintf("%s/pr/%s", config.GetConfig().Domain, hash)
}

func adminMenuItems() []models.MenuItem {
	return []models.MenuItem{
		models.MenuItem{Title: "Страницы", URL: "/admin/pages"},
		models.MenuItem{Title: "Заметки", URL: "/admin/posts"},
		models.MenuItem{Title: "Пункты меню", URL: "/admin/menus"},
		models.MenuItem{Title: "Слайды", URL: "/admin/slides"},
		models.MenuItem{Title: "Настройки", URL: "/admin/settings"},
		models.MenuItem{Title: "Пользователи", URL: "/admin/users"},
	}
}

func mapAPIKey() string {
	return config.GetConfig().MapAPIKey
}
