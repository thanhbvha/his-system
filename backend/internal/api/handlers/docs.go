package handlers

import (
	"github.com/gofiber/fiber/v2"
)

type DocsHandler struct{}

func NewDocsHandler() *DocsHandler {
	return &DocsHandler{}
}

// ServeReDoc hiển thị giao diện ReDoc sử dụng các tệp tin lưu trữ nội bộ
func (h *DocsHandler) ServeReDoc(c *fiber.Ctx) error {
	c.Set("Content-Type", "text/html; charset=utf-8")
	html := `<!DOCTYPE html>
<html>
  <head>
    <title>ReDoc - HIS System API</title>
    <!-- Cấu hình Font trực tiếp -->
    <link href="https://fonts.googleapis.com/css?family=Montserrat:300,400,700|Roboto:300,400,700" rel="stylesheet">
    <style>
      body {
        margin: 0;
        padding: 0;
      }
    </style>
  </head>
  <body>
    <!-- ReDoc Container -->
    <redoc spec-url="/docs/swagger.json"></redoc>
    <!-- Local ReDoc Bundle -->
    <script src="/static/docs/redoc.standalone.js"></script>
  </body>
</html>`
	return c.SendString(html)
}

// ServeSwaggerUI hiển thị giao diện Swagger UI sử dụng các tệp tin lưu trữ nội bộ
func (h *DocsHandler) ServeSwaggerUI(c *fiber.Ctx) error {
	c.Set("Content-Type", "text/html; charset=utf-8")
	html := `<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="UTF-8">
    <title>Swagger UI - HIS System API</title>
    <!-- Local Swagger CSS -->
    <link rel="stylesheet" type="text/css" href="/static/docs/swagger-ui.css" />
    <style>
      html { box-sizing: border-box; overflow: -moz-scrollbars-vertical; overflow-y: scroll; }
      *, *:before, *:after { box-sizing: inherit; }
      body { margin:0; background: #fafafa; }
    </style>
  </head>
  <body>
    <div id="swagger-ui"></div>
    <!-- Local Swagger JS -->
    <script src="/static/docs/swagger-ui-bundle.js"></script>
    <script src="/static/docs/swagger-ui-standalone-preset.js"></script>
    <script>
      window.onload = function() {
        const ui = SwaggerUIBundle({
          url: "/docs/swagger.json",
          dom_id: '#swagger-ui',
          deepLinking: true,
          presets: [
            SwaggerUIBundle.presets.apis,
            SwaggerUIStandalonePreset
          ],
          plugins: [
            SwaggerUIBundle.plugins.DownloadUrl
          ],
          layout: "StandaloneLayout"
        });
        window.ui = ui;
      };
    </script>
  </body>
</html>`
	return c.SendString(html)
}
