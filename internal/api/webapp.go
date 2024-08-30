package api

import (
	"context"
	"pipe/internal/services"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"gopkg.in/telebot.v3"
)

// //go:embed static
// var embededFiles embed.FS

type WebApp struct {
	addr string
	App  *services.App
	e    *echo.Echo
	bot  *telebot.Bot
}

func NewWebApp(
	addr string,
	app *services.App,
	bot *telebot.Bot,
) *WebApp {
	e := echo.New()
	wa := &WebApp{
		App:  app,
		e:    e,
		addr: addr,
		bot:  bot,
	}
	wa.routes()
	// wa.static()
	return wa
}

func (w *WebApp) Start() error {
	w.e.Use(middleware.Recover())
	return w.e.Start(w.addr)
}

func (w *WebApp) Shutdown(ctx context.Context) error {
	return w.e.Shutdown(ctx)
}

// func (w *WebApp) static() {
// 	assetHandler := http.FileServer(getFileSystem())
// 	w.e.GET("/static/*",
// 		echo.WrapHandler(http.StripPrefix("/static/", assetHandler)),
// 		func(next echo.HandlerFunc) echo.HandlerFunc {
// 			return func(c echo.Context) error {
// 				// c.Response().Header().Set(
// 				// 	"Cache-Control",
// 				// 	fmt.Sprintf("public,max-age=%d",
// 				// 		int((time.Hour*24*7).Seconds())),
// 				// )
// 				err := next(c)
// 				if err != nil {
// 					return err
// 				}
// 				return nil
// 			}
// 		},
// 	)

// }

// func getFileSystem() http.FileSystem {
// 	fSys, err := fs.Sub(embededFiles, "static")
// 	if err != nil {
// 		log.Fatal("couldn't init static embedding")
// 	}
// 	return http.FS(fSys)
// }
