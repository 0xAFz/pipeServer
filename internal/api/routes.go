package api

import (
	"pipe/internal/config"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func (w *WebApp) routes() {
	w.e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{config.AppConfig.ClientURL},
		AllowMethods:     []string{echo.OPTIONS, echo.HEAD, echo.GET, echo.POST, echo.PUT, echo.DELETE, echo.PATCH, "*"},
		AllowHeaders:     []string{"*", echo.HeaderAuthorization},
		AllowCredentials: true,
	}))

	w.e.GET("/", w.index)
	w.e.GET("/getMe", w.getMe, w.withAuth)
	w.e.GET("/getUser/:privateID", w.getUser, w.withAuth)
	w.e.GET("/getMessages", w.getMessages, w.withAuth)
	w.e.POST("/sendMessage/:privateID", w.sendMessage, w.withAuth)
	w.e.DELETE("/deleteAccount", w.deleteAccount, w.withAuth)
	w.e.PATCH("/setPubKey", w.setPubKey, w.withAuth)
}
