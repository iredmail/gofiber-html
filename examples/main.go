package main

import (
	"embed"
	"io/fs"

	"github.com/gofiber/fiber/v2"
	html "github.com/iredmail/gofiber-html"
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
	engine.AddLayouts("layout")
	app := fiber.New(fiber.Config{
		Views: engine,
	})

	app.Get("/", func(c *fiber.Ctx) error {
		return c.Render("index", fiber.Map{
			"Var1": "value1",
		})
	})

	app.Get("/dashboard", func(c *fiber.Ctx) error {
		return c.Render("dashboard", fiber.Map{
			"Var2": "value2",
		})
	})

	app.Listen(":3000")
}
