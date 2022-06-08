package main

import (
	"embed"
	"io/fs"

	"github.com/gofiber/fiber/v2"
	html "github.com/spiderd-io/gofiber-html"
)

var (
	//go:embed views
	embedStaticFiles embed.FS
)

func main() {
	viewsDir, err := fs.Sub(embedStaticFiles, "views")
	if err != nil {
		panic(err)
	}

	engine := html.NewFileSystem(viewsDir, ".gohtml")
	app := fiber.New(fiber.Config{
		Views: engine,
	})

	app.Get("/", func(c *fiber.Ctx) error {
		return c.Render("index", nil, "layout") // specify layout file
	})

	app.Get("/dashboard", func(c *fiber.Ctx) error {
		return c.Render("dashboard", nil, "layout") // specify layout file
	})

	app.Listen(":3000")
}
