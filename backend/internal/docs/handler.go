package docs

import (
	_ "embed"
	"html/template"
	"net/http"
)

//go:embed openapi.yaml
var openAPISpec []byte

//go:embed swagger.json
var swaggerJSON []byte

var swaggerTemplate = template.Must(template.New("swagger").Parse(`<!doctype html>
<html lang="en">
  <head>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
    <title>MoneyApp API Docs</title>
    <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css" />
    <style>
      html, body {
        margin: 0;
        background: #f7f8fa;
      }
      .topbar {
        display: none;
      }
    </style>
  </head>
  <body>
    <div id="swagger-ui"></div>
    <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
    <script>
      window.ui = SwaggerUIBundle({
        spec: {{ .SpecJSON }},
        dom_id: "#swagger-ui",
        deepLinking: true,
        displayRequestDuration: true,
        docExpansion: "list",
        persistAuthorization: true,
        tryItOutEnabled: true,
        filter: true,
        defaultModelsExpandDepth: 2,
      });
    </script>
  </body>
</html>
`))

func OpenAPI(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/yaml; charset=utf-8")
	_, _ = w.Write(openAPISpec)
}

func SwaggerJSON(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_, _ = w.Write(swaggerJSON)
}

func SwaggerUI(_ string) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_ = swaggerTemplate.Execute(w, map[string]any{
			"SpecJSON": template.JS(swaggerJSON),
		})
	}
}
