package controllers

import (
	"net/http"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/denisbakhtin/doctorkaa/models"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

//PostsGet handles GET /blog route
func PostsGet(c *gin.Context) {
	db := models.GetDB()
	var posts []models.Post
	db.Order("id desc").Where("published = ?", true).Find(&posts)
	h := DefaultH(c)
	h["Title"] = getSetting("blogs_title")
	h["Posts"] = posts
	c.HTML(http.StatusOK, "posts/public_index", h)
}

//PostGet handles GET /blog/:slug route
func PostGet(c *gin.Context) {
	db := models.GetDB()
	post := models.Post{}
	var posts []models.Post
	slug := c.Param("slug")
	db.Where("slug = ?", strings.ToLower(slug)).First(&post)
	if post.ID == 0 || (!post.Published && isReleaseMode()) {
		c.HTML(http.StatusNotFound, "errors/404", nil)
		return
	}
	db.Where("published = ? AND id != ?", true, post.ID).Order("id desc").Limit(5).Find(&posts)

	h := DefaultH(c)
	h["Title"] = post.Title
	h["Post"] = post
	h["Posts"] = posts
	h["MetaKeywords"] = post.MetaKeywords
	h["MetaDescription"] = post.MetaDescription
	c.HTML(http.StatusOK, "posts/show", h)
}

//PostIndex handles GET /admin/posts route
func PostIndex(c *gin.Context) {
	db := models.GetDB()
	var posts []models.Post
	db.Order("id").Find(&posts)
	h := DefaultH(c)
	h["Title"] = "Список публикаций"
	h["Posts"] = posts
	c.HTML(http.StatusOK, "posts/index", h)
}

//PostNew handles GET /admin/new_post route
func PostNew(c *gin.Context) {
	h := DefaultH(c)
	h["Title"] = "Новая публикация"
	session := sessions.Default(c)
	h["Flash"] = session.Flashes()
	h["Post"] = models.Post{Published: true}
	session.Save()

	c.HTML(http.StatusOK, "posts/form", h)
}

//PostCreate handles POST /admin/new_post route
func PostCreate(c *gin.Context) {
	db := models.GetDB()
	post := models.Post{}
	if err := c.ShouldBind(&post); err != nil {
		session := sessions.Default(c)
		session.AddFlash(err.Error())
		session.Save()
		c.Redirect(http.StatusSeeOther, "/admin/new_post")
		return
	}

	if err := db.Create(&post).Error; err != nil {
		c.HTML(http.StatusInternalServerError, "errors/500", nil)
		logrus.Error(err)
		return
	}
	c.Redirect(http.StatusFound, "/admin/posts")
}

//PostEdit handles GET /admin/posts/:id/edit route
func PostEdit(c *gin.Context) {
	db := models.GetDB()
	post := models.Post{}
	db.First(&post, c.Param("id"))
	if post.ID == 0 {
		c.HTML(http.StatusNotFound, "errors/404", nil)
		return
	}
	h := DefaultH(c)
	h["Title"] = "Редактирование публикации"
	h["Post"] = post
	session := sessions.Default(c)
	h["Flash"] = session.Flashes()
	session.Save()
	c.HTML(http.StatusOK, "posts/form", h)
}

//PostUpdate handles POST /admin/posts/:id/edit route
func PostUpdate(c *gin.Context) {
	post := models.Post{}
	db := models.GetDB()
	if err := c.ShouldBind(&post); err != nil {
		session := sessions.Default(c)
		session.AddFlash(err.Error())
		session.Save()
		c.Redirect(http.StatusSeeOther, "/admin/posts")
		return
	}
	if err := db.Save(&post).Error; err != nil {
		c.HTML(http.StatusInternalServerError, "errors/500", nil)
		logrus.Error(err)
		return
	}
	c.Redirect(http.StatusFound, "/admin/posts")
}

//PostDelete handles POST /admin/posts/:id/delete route
func PostDelete(c *gin.Context) {
	post := models.Post{}
	db := models.GetDB()
	db.First(&post, c.Param("id"))
	if post.ID == 0 {
		c.HTML(http.StatusNotFound, "errors/404", nil)
		return
	}
	if err := db.Delete(&post).Error; err != nil {
		c.HTML(http.StatusInternalServerError, "errors/500", nil)
		logrus.Error(err)
		return
	}
	c.Redirect(http.StatusFound, "/admin/posts")
}
