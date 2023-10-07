package api

//         ╓                                                          ╖
//         ║              Block Scanner Access Api routes              ║
//         ╙                                                          ╜

import (
	"github.com/gofiber/fiber/v2"
)

const (
	Get RouteMethods = iota
	Post
	Put
	Delete
)

type RouteMethods int

type Route struct {
	Path    string
	Method  RouteMethods
	Handler fiber.Handler
}

func RunApi(ListenAdd string, Routes []Route) {
	app := fiber.New()
	for _, route := range Routes {
		switch route.Method {
		case Get:
			app.Get(route.Path, route.Handler)
		case Post:
			app.Post(route.Path, route.Handler)
		case Put:
			app.Put(route.Path, route.Handler)
		case Delete:
			app.Delete(route.Path, route.Handler)
		}
	}
	app.Listen(ListenAdd)
}
