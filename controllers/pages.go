package controllers

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"

	"html/template"

	"github.com/Sirupsen/logrus"
	"github.com/denisbakhtin/doctorkaa/config"
	"github.com/denisbakhtin/doctorkaa/models"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	gomail "gopkg.in/gomail.v2"
)

//PageGet handles GET /:slug route
func PageGet(c *gin.Context) {
	db := models.GetDB()
	page := models.Page{}
	slug := c.Request.URL.RequestURI()
	db.Where("slug = ?", strings.TrimLeft(strings.ToLower(slug), "/")).First(&page)
	if page.ID == 0 || (!page.Published && isReleaseMode()) {
		c.HTML(http.StatusNotFound, "errors/404", nil)
		return
	}
	h := DefaultH(c)
	session := sessions.Default(c)
	h["Flash"] = session.Flashes()
	h["Title"] = page.Title
	h["Page"] = page
	h["IsContactPage"] = isContactPage(slug)
	h["MetaKeywords"] = page.MetaKeywords
	h["MetaDescription"] = page.MetaDescription
	session.Save()
	c.HTML(http.StatusOK, "pages/show", h)
}

//PageIndex handles GET /admin/pages route
func PageIndex(c *gin.Context) {
	db := models.GetDB()
	var pages []models.Page
	db.Order("id").Find(&pages)
	h := DefaultH(c)
	h["Title"] = "Список страниц"
	h["Pages"] = pages
	c.HTML(http.StatusOK, "pages/index", h)
}

//PageNew handles GET /admin/new_page route
func PageNew(c *gin.Context) {
	h := DefaultH(c)
	h["Title"] = "Новая страница"
	session := sessions.Default(c)
	h["Flash"] = session.Flashes()
	h["Page"] = models.Page{Published: true}
	session.Save()

	c.HTML(http.StatusOK, "pages/form", h)
}

//PageCreate handles POST /admin/new_page route
func PageCreate(c *gin.Context) {
	db := models.GetDB()
	page := models.Page{}
	if err := c.ShouldBind(&page); err != nil {
		session := sessions.Default(c)
		session.AddFlash(err.Error())
		session.Save()
		c.Redirect(http.StatusSeeOther, "/admin/new_page")
		return
	}

	if err := db.Create(&page).Error; err != nil {
		c.HTML(http.StatusInternalServerError, "errors/500", nil)
		logrus.Error(err)
		return
	}
	c.Redirect(http.StatusFound, "/admin/pages")
}

//PageEdit handles GET /admin/pages/:id/edit route
func PageEdit(c *gin.Context) {
	db := models.GetDB()
	page := models.Page{}
	db.First(&page, c.Param("id"))
	if page.ID == 0 {
		c.HTML(http.StatusNotFound, "errors/404", nil)
		return
	}
	h := DefaultH(c)
	h["Title"] = "Редактирование страницы"
	h["Page"] = page
	session := sessions.Default(c)
	h["Flash"] = session.Flashes()
	session.Save()
	c.HTML(http.StatusOK, "pages/form", h)
}

//PageUpdate handles POST /admin/pages/:id/edit route
func PageUpdate(c *gin.Context) {
	page := models.Page{}
	db := models.GetDB()
	if err := c.ShouldBind(&page); err != nil {
		session := sessions.Default(c)
		session.AddFlash(err.Error())
		session.Save()
		c.Redirect(http.StatusSeeOther, "/admin/pages")
		return
	}
	if err := db.Save(&page).Error; err != nil {
		c.HTML(http.StatusInternalServerError, "errors/500", nil)
		logrus.Error(err)
		return
	}
	c.Redirect(http.StatusFound, "/admin/pages")
}

//PageDelete handles POST /admin/pages/:id/delete route
func PageDelete(c *gin.Context) {
	page := models.Page{}
	db := models.GetDB()
	db.First(&page, c.Param("id"))
	if page.ID == 0 {
		c.HTML(http.StatusNotFound, "errors/404", nil)
		return
	}
	if err := db.Delete(&page).Error; err != nil {
		c.HTML(http.StatusInternalServerError, "errors/500", nil)
		logrus.Error(err)
		return
	}
	c.Redirect(http.StatusFound, "/admin/pages")
}

//FeedbackPost handles POST /feedback route
func FeedbackPost(c *gin.Context) {
	session := sessions.Default(c)
	feedback := models.Feedback{}
	returnURL := getContactsPage().URL()
	if err := c.ShouldBind(&feedback); err != nil {
		session.AddFlash("Пожалуйста, внимательно заполните все поля.")
		session.Save()
		c.Redirect(http.StatusFound, returnURL)
		return
	}
	if err := notifyOwnerOfFeedback(c, &feedback); err != nil {
		logrus.Error(err)
		session.AddFlash("Ошибка отправки сообщения, внимательно заполните все поля.")
		session.Save()
		c.Redirect(http.StatusFound, returnURL)
		return
	}
	session.AddFlash("Спасибо! Ваше сообщение было успешно отправлено!")
	session.Save()
	c.Redirect(http.StatusFound, returnURL)
}

func notifyOwnerOfFeedback(c *gin.Context, feedback *models.Feedback) error {
	var b bytes.Buffer

	tmpl := template.New("").Funcs(getFuncMap())
	workingdir, _ := os.Getwd()
	tmpl, _ = tmpl.ParseFiles(path.Join(workingdir, "views", "emails", "feedback.gohtml"))
	if err := tmpl.Lookup("emails/feedback").Execute(&b, gin.H{"Feedback": feedback}); err != nil {
		return err
	}

	smtp := config.GetConfig().SMTP
	msg := gomail.NewMessage()
	msg.SetHeader("From", smtp.To)
	msg.SetHeader("To", smtp.To)
	msg.SetHeader("Subject", fmt.Sprintf("Новое сообщение на сайте doctorkaa.ru от: %s", feedback.Name))
	msg.SetBody(
		"text/html",
		b.String(),
	)

	port, _ := strconv.Atoi(smtp.Port)
	dialer := gomail.NewPlainDialer(smtp.SMTP, port, smtp.User, smtp.Password)
	sender, err := dialer.Dial()
	if err != nil {
		return err
	}
	if err := gomail.Send(sender, msg); err != nil {
		return err
	}
	return nil
}
