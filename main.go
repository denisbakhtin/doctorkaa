package main

import (
	"flag"
	"fmt"
	"net/http"
	"path"

	"github.com/Sirupsen/logrus"
	"github.com/denisbakhtin/doctorkaa/config"
	"github.com/denisbakhtin/doctorkaa/controllers"
	"github.com/denisbakhtin/doctorkaa/models"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/memstore"
	"github.com/gin-gonic/gin"
	"github.com/utrack/gin-csrf"
)

func main() {
	mode := flag.String("mode", "debug", "Application mode: debug, release, test")
	flag.Parse()

	gin.SetMode(*mode)

	initLogger()
	config.LoadConfig()
	models.SetDB(config.GetConnectionString())
	models.AutoMigrate()

	// Creates a gin router with default middleware:
	// logger and recovery (crash-free) middleware
	router := gin.Default()
	//better use nginx to serve assets (Cache-Control, Etag, fast gzip, etc)
	router.StaticFS("/images", http.Dir(path.Join(config.PublicPath(), "images")))
	router.StaticFS("/assets", http.Dir(path.Join(config.PublicPath(), "assets")))
	router.StaticFS("/uploads", http.Dir(path.Join(config.PublicPath(), "uploads")))
	router.StaticFS("/vendor", http.Dir(path.Join(config.PublicPath(), "vendor")))
	router.StaticFile("/robots.txt", path.Join(config.PublicPath(), "robots.txt"))
	controllers.LoadTemplates(router)

	//setup sessions
	conf := config.GetConfig()
	store := memstore.NewStore([]byte(conf.SessionSecret))
	store.Options(sessions.Options{HttpOnly: true, MaxAge: 7 * 86400}) //Also set Secure: true if using SSL, you should though
	router.Use(sessions.Sessions("gin-session", store))
	router.Use(controllers.ContextData())

	//setup csrf protection
	router.Use(csrf.Middleware(csrf.Options{
		Secret: conf.SessionSecret,
		ErrorFunc: func(c *gin.Context) {
			logrus.Error("CSRF token mismatch")
			controllers.ShowErrorPage(c, 400, fmt.Errorf("CSRF token mismatch"))
			c.Abort()
		},
	}))

	router.GET("/", controllers.HomeGet)
	router.NoMethod(controllers.MethodNotAllowed)

	if config.GetConfig().SignupEnabled {
		router.GET("/signup", controllers.SignUpGet)
		router.POST("/signup", controllers.SignUpPost)
	}
	router.GET("/signin", controllers.SignInGet)
	router.POST("/signin", controllers.SignInPost)
	router.GET("/signout", controllers.SignoutGet)
	router.GET("/forgot", controllers.ForgotGet)
	router.POST("/forgot", controllers.ForgotPost)

	router.POST("/feedback", controllers.FeedbackPost)
	router.GET("/pr/:hash", controllers.PasswordResetGet)
	router.POST("/pr", controllers.PasswordResetPost)

	//router.GET("/:slug", controllers.PageGet)
	router.GET("/blog", controllers.PostsGet)
	router.GET("/blog/:slug", controllers.PostGet)

	authorized := router.Group("/admin")
	authorized.Use(controllers.AuthRequired())
	{
		authorized.GET("/", controllers.AdminGet)

		authorized.POST("/upload", controllers.UploadPost) //image upload

		authorized.GET("/users", controllers.UserIndex)
		authorized.GET("/new_user", controllers.UserNew)
		authorized.POST("/new_user", controllers.UserCreate)
		authorized.GET("/users/:id/edit", controllers.UserEdit)
		authorized.POST("/users/:id/edit", controllers.UserUpdate)
		authorized.POST("/users/:id/delete", controllers.UserDelete)

		authorized.GET("/pages", controllers.PageIndex)
		authorized.GET("/new_page", controllers.PageNew)
		authorized.POST("/new_page", controllers.PageCreate)
		authorized.GET("/pages/:id/edit", controllers.PageEdit)
		authorized.POST("/pages/:id/edit", controllers.PageUpdate)
		authorized.POST("/pages/:id/delete", controllers.PageDelete)

		authorized.GET("/posts", controllers.PostIndex)
		authorized.GET("/new_post", controllers.PostNew)
		authorized.POST("/new_post", controllers.PostCreate)
		authorized.GET("/posts/:id/edit", controllers.PostEdit)
		authorized.POST("/posts/:id/edit", controllers.PostUpdate)
		authorized.POST("/posts/:id/delete", controllers.PostDelete)

		authorized.GET("/menus", controllers.MenuIndex)
		authorized.GET("/new_menu", controllers.MenuNew)
		authorized.POST("/new_menu", controllers.MenuCreate)
		authorized.GET("/menus/:id/edit", controllers.MenuEdit)
		authorized.POST("/menus/:id/edit", controllers.MenuUpdate)
		authorized.POST("/menus/:id/delete", controllers.MenuDelete)

		authorized.GET("/settings", controllers.SettingIndex)
		authorized.GET("/new_setting", controllers.SettingNew)
		authorized.POST("/new_setting", controllers.SettingCreate)
		authorized.GET("/settings/:id/edit", controllers.SettingEdit)
		authorized.POST("/settings/:id/edit", controllers.SettingUpdate)
		authorized.POST("/settings/:id/delete", controllers.SettingDelete)

		authorized.GET("/slides", controllers.SlideIndex)
		authorized.GET("/new_slide", controllers.SlideNew)
		authorized.POST("/new_slide", controllers.SlideCreate)
		authorized.GET("/slides/:id/edit", controllers.SlideEdit)
		authorized.POST("/slides/:id/edit", controllers.SlideUpdate)
		authorized.POST("/slides/:id/delete", controllers.SlideDelete)

	}

	//catch all unmatched routes with page handler
	router.NoRoute(controllers.PageGet)

	// Listen and server on 0.0.0.0:8081
	router.Run(":8089")
}

//initLogger initializes logrus logger with some defaults
func initLogger() {
	logrus.SetFormatter(&logrus.TextFormatter{})
	if gin.Mode() == gin.DebugMode {
		logrus.SetLevel(logrus.DebugLevel)
	}
}
