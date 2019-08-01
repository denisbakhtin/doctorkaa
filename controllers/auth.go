package controllers

import (
	"bytes"
	"html/template"
	"net/http"
	"os"
	"path"
	"strconv"

	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/denisbakhtin/doctorkaa/config"
	"github.com/denisbakhtin/doctorkaa/models"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	gomail "gopkg.in/gomail.v2"
)

//SignInGet handles GET /signin route
func SignInGet(c *gin.Context) {
	h := DefaultH(c)
	h["Title"] = "Вход в систему"
	session := sessions.Default(c)
	h["Flash"] = session.Flashes()
	session.Save()
	c.HTML(http.StatusOK, "auth/signin", h)
}

//SignInPost handles POST /signin route, authenticates user
func SignInPost(c *gin.Context) {
	session := sessions.Default(c)
	login := models.Login{}
	db := models.GetDB()
	if err := c.ShouldBind(&login); err != nil {
		session.AddFlash("Пожалуйста, укажите правильные данные.")
		session.Save()
		c.Redirect(http.StatusFound, "/signin")
		return
	}

	user := models.User{}
	db.Where("email = lower(?)", login.Email).First(&user)

	if user.ID == 0 || bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(login.Password)) != nil {
		logrus.Errorf("Login error, IP: %s, Email: %s", c.ClientIP(), login.Email)
		session.AddFlash("Электронная почта или пароль указаны неверно")
		session.Save()
		c.Redirect(http.StatusFound, "/signin")
		return
	}

	session.Set(userIDKey, user.ID)
	session.Save()
	c.Redirect(http.StatusFound, "/")
}

//SignUpGet handles GET /signup route
func SignUpGet(c *gin.Context) {
	h := DefaultH(c)
	h["Title"] = "Регистрация в системе"
	session := sessions.Default(c)
	h["Flash"] = session.Flashes()
	session.Save()
	c.HTML(http.StatusOK, "auth/signup", h)
}

//SignUpPost handles POST /signup route, creates new user
func SignUpPost(c *gin.Context) {
	session := sessions.Default(c)
	register := models.Register{}
	db := models.GetDB()
	if err := c.ShouldBind(&register); err != nil {
		session.AddFlash("Пожалуйста, заполните все обязательные поля.")
		session.Save()
		c.Redirect(http.StatusFound, "/signup")
		return
	}
	if register.Password != register.PasswordConfirm {
		session.AddFlash("Пароль и подтверждение пароля не совпадают.")
		session.Save()
		c.Redirect(http.StatusFound, "/signup")
		return
	}
	register.Email = strings.ToLower(register.Email)
	user := models.User{}
	db.Where("email = ?", register.Email).First(&user)
	if user.ID != 0 {
		session.AddFlash("Пользователь с такой электронной почтой уже существует")
		session.Save()
		c.Redirect(http.StatusFound, "/signup")
		return
	}
	//create user
	user.Email = register.Email
	user.Password = register.Password
	if err := db.Create(&user).Error; err != nil {
		session.AddFlash("Ошибка регистрации нового пользователя.")
		session.Save()
		logrus.Errorf("Error whilst registering user: %v", err)
		c.Redirect(http.StatusFound, "/signup")
		return
	}

	session.Set(userIDKey, user.ID)
	session.Save()
	c.Redirect(http.StatusFound, "/")
}

//SignoutGet handles GET /signout route
func SignoutGet(c *gin.Context) {
	session := sessions.Default(c)
	session.Delete(userIDKey)
	session.Save()
	c.Redirect(http.StatusSeeOther, "/")
}

//ForgotGet handles GET /forgot route
func ForgotGet(c *gin.Context) {
	h := DefaultH(c)
	h["Title"] = "Восстановление пароля"
	session := sessions.Default(c)
	h["Flash"] = session.Flashes()
	session.Save()
	c.HTML(http.StatusOK, "auth/forgot", h)
}

//ForgotPost handles POST /forgot route
func ForgotPost(c *gin.Context) {
	session := sessions.Default(c)
	forgot := models.Forgot{}
	db := models.GetDB()

	if err := c.ShouldBind(&forgot); err != nil {
		session.AddFlash("Пожалуйста, укажите правильные данные.")
		session.Save()
		c.Redirect(http.StatusFound, "/forgot")
		return
	}

	user := models.User{}
	db.Where("email = lower(?)", forgot.Email).First(&user)

	if user.ID == 0 {
		logrus.Errorf("Password reset error, IP: %s, Email: %s", c.ClientIP(), forgot.Email)
		session.AddFlash("Письмо с инструкциями по восстановлению пароля высланы на указанную почту!")
		session.Save()
		c.Redirect(http.StatusFound, "/forgot")
		return
	}
	if err := user.GenerateForgotHash(); err != nil {
		logrus.Errorf("Error generating password reset hash, IP: %s, Email: %s", c.ClientIP(), forgot.Email)
		session.AddFlash("Письмо с инструкциями по восстановлению пароля высланы на указанную почту!")
		session.Save()
		c.Redirect(http.StatusFound, "/forgot")
		return
	}
	if err := db.Save(&user).Error; err != nil {
		logrus.Errorf("Error updating user: %s", err.Error())
		session.AddFlash("Ошибка восстановления пароля, попробуйте операцию позже!")
		session.Save()
		c.Redirect(http.StatusFound, "/forgot")
		return
	}

	if err := notifyUserOfReset(c, &user); err != nil {
		logrus.Error(err)
		session.AddFlash("Ошибка отправки сообщения, повторите запрос позже.")
		session.Save()
		c.Redirect(http.StatusFound, "/forgot")
		return
	}
	session.AddFlash("Спасибо! На указанный почтовый адрес отправлено письмо с инструкцией по восстановлению пароля!")
	session.Save()
	c.Redirect(http.StatusFound, "/forgot")
}

func notifyUserOfReset(c *gin.Context, user *models.User) error {
	var b bytes.Buffer

	tmpl := template.New("").Funcs(getFuncMap())
	workingdir, _ := os.Getwd()
	tmpl, _ = tmpl.ParseFiles(path.Join(workingdir, "views", "emails", "forgot.gohtml"))
	if err := tmpl.Lookup("emails/forgot").Execute(&b, gin.H{"User": user}); err != nil {
		return err
	}

	smtp := config.GetConfig().SMTP
	msg := gomail.NewMessage()
	msg.SetHeader("From", smtp.To)
	msg.SetHeader("To", user.Email)
	msg.SetHeader("Subject", "Восстановление пароля на сайте doctorkaa.ru")
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

//PasswordResetGet handles GET /pr/:hash route
func PasswordResetGet(c *gin.Context) {
	session := sessions.Default(c)
	hash := c.Param("hash")
	db := models.GetDB()
	user := models.User{}

	db.Where("forgot_hash = ?", hash).First(&user)
	if user.ID == 0 || len(strings.TrimSpace(hash)) < 50 {
		session.AddFlash("Ссылка для восстановления пароля неверна, либо устарела.")
		session.Save()
		c.Redirect(http.StatusFound, "/")
		return
	}

	h := DefaultH(c)
	h["Hash"] = hash
	h["Title"] = "Укажите новый пароль"
	c.HTML(http.StatusOK, "auth/reset", h)
}

//PasswordResetPost handles POST /pr route
func PasswordResetPost(c *gin.Context) {
	session := sessions.Default(c)
	db := models.GetDB()
	user := models.User{}
	reset := models.Reset{}

	if err := c.ShouldBind(&reset); err != nil {
		session.AddFlash("Ошибка, неверно указаны данные.")
		session.Save()
		c.Redirect(http.StatusFound, "/")
		return
	}

	if len(strings.TrimSpace(reset.Password)) < 5 || reset.Password != reset.PasswordConfirm {
		session.AddFlash("Ошибка, неверно указаны данные, повторите попытку.")
		session.Save()
		c.Redirect(http.StatusFound, "/")
		return
	}

	db.Where("forgot_hash = ?", reset.Hash).First(&user)
	if user.ID == 0 || len(strings.TrimSpace(reset.Hash)) < 50 {
		session.AddFlash("Ссылка для восстановления пароля неверна, либо устарела.")
		session.Save()
		c.Redirect(http.StatusFound, "/")
		return
	}
	user.Password = reset.Password
	user.ForgotHash = ""
	if err := user.HashPassword(); err != nil {
		logrus.Error(err)
		session.AddFlash("Ошибка восстановления пароля, повторите попытку.")
		session.Save()
		c.Redirect(http.StatusFound, "/")
		return
	}
	if err := db.Save(&user).Error; err != nil {
		logrus.Error(err)
		session.AddFlash("Ошибка восстановления пароля, повторите попытку.")
		session.Save()
		c.Redirect(http.StatusFound, "/")
		return
	}

	session.AddFlash("Пароль успешно изменен!")
	session.Save()
	if err := notifyUserOfSuccessfulReset(c, &user); err != nil {
		logrus.Error(err)
	}
	c.Redirect(http.StatusFound, "/")
}

func notifyUserOfSuccessfulReset(c *gin.Context, user *models.User) error {
	var b bytes.Buffer
	tmpl := template.New("").Funcs(getFuncMap())
	workingdir, _ := os.Getwd()
	tmpl, _ = tmpl.ParseFiles(path.Join(workingdir, "views", "emails", "reset.gohtml"))
	if err := tmpl.Lookup("emails/reset").Execute(&b, gin.H{"User": user}); err != nil {
		return err
	}

	smtp := config.GetConfig().SMTP
	msg := gomail.NewMessage()
	msg.SetHeader("From", smtp.To)
	msg.SetHeader("To", user.Email)
	msg.SetHeader("Subject", "Новый пароль на сайте doctorkaa.ru успешно установлен")
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
