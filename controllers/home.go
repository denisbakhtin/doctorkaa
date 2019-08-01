package controllers

import (
	"net/http"

	"github.com/denisbakhtin/doctorkaa/models"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

//HomeGet handles GET / route
func HomeGet(c *gin.Context) {
	db := models.GetDB()
	page := models.Page{}

	db.First(&page, "slug = ?", "/")
	if page.ID == 0 {
		c.HTML(http.StatusNotFound, "errors/404", nil)
		return
	}
	h := DefaultH(c)
	session := sessions.Default(c)
	h["Page"] = page
	h["Flash"] = session.Flashes()
	session.Save()

	c.HTML(http.StatusOK, "home/show", h)
}
