# gofiber-html

`gofiber-html` uses the Go builtin `html/template` as [Fiber](https://gofiber.io) template engine. Here's the [original syntax of `html/template`](TEMPLATES_CHEATCHEET.md)

Check the sample app in `example/` directory to get start.

With `gofiber-html`, you must specify layout file name in `ctx.Render()` and
template files. For example:

```go
func Index(c *fiber.Ctx) error {
    // Specify the layout template file `layout` (omit the file extension)
    return c.Render("index", nil, "layout")
}
```

And in html template file, you must specify the layout template file too:
```html
{{template "layout.gohtml"}}

{{define "title"}}Index page{{end}}
```

## Notes

About the `html` module offered by [`gofiber/template`](https://github.com/gofiber/template/tree/master/html)
repo:

- it can not specify web page title in template files.

    If you define a template to hold page title like `{{template "title" .}}`
    in layout file, then define page title like `{{define "title"}}...{{end}}`
    in each template file, only the one in last parsed template file will be
    kept and used (earlies ones are overwrote),
    this is not ideal for
i18n.
