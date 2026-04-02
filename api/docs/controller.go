package docs

import (
	"html/template"
	"net/http"
	"sync-backend/arch/network"

	"github.com/gin-gonic/gin"
)

type docsController struct {
	network.BaseController
}

func NewDocsController() network.Controller {
	return &docsController{
		BaseController: network.NewBaseController("/docs", nil),
	}
}

func (c *docsController) MountRoutes(router *gin.RouterGroup) {
	// Serve Swagger UI HTML page
	router.GET("/", c.serveSwaggerUI)
	router.GET("", c.serveSwaggerUI)

	// Serve OpenAPI spec
	router.GET("/openapi.yaml", c.serveOpenAPISpec)
	router.GET("/swagger.yaml", c.serveOpenAPISpec)
}

func (c *docsController) serveSwaggerUI(ctx *gin.Context) {
	htmlTemplate := `
<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<title>Sync API Documentation</title>
	<link rel="icon" type="image/png" href="data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg==">
	<link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5.11.0/swagger-ui.css" />
	<style>
		body {
			margin: 0;
			padding: 0;
			background-color: #fafafa;
		}
		.swagger-ui .topbar {
			background-color: #1a1a1a;
			padding: 15px 0;
		}
		.swagger-ui .topbar .download-url-wrapper {
			display: none;
		}
		.swagger-ui .info {
			margin: 50px 0;
		}
		.swagger-ui .info .title {
			color: #1a1a1a;
			font-size: 36px;
		}
		.swagger-ui .scheme-container {
			background-color: #fff;
			box-shadow: 0 1px 2px 0 rgba(0,0,0,.15);
		}
		.topbar-wrapper img {
			content: url("data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iNDAiIGhlaWdodD0iNDAiIHZpZXdCb3g9IjAgMCA0MCA0MCIgZmlsbD0ibm9uZSIgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIj4KPHJlY3Qgd2lkdGg9IjQwIiBoZWlnaHQ9IjQwIiByeD0iOCIgZmlsbD0iIzNiODJmNiIvPgo8dGV4dCB4PSI1MCUiIHk9IjUwJSIgZm9udC1mYW1pbHk9IkFyaWFsIiBmb250LXNpemU9IjI0IiBmb250LXdlaWdodD0iYm9sZCIgZmlsbD0id2hpdGUiIHRleHQtYW5jaG9yPSJtaWRkbGUiIGR5PSIuMzVlbSI+UzwvdGV4dD4KPC9zdmc+");
		}
	</style>
</head>
<body>
	<div id="swagger-ui"></div>
	<script src="https://unpkg.com/swagger-ui-dist@5.11.0/swagger-ui-bundle.js"></script>
	<script src="https://unpkg.com/swagger-ui-dist@5.11.0/swagger-ui-standalone-preset.js"></script>
	<script>
		window.onload = function() {
			window.ui = SwaggerUIBundle({
				url: "/api/v1/docs/openapi.yaml",
				dom_id: '#swagger-ui',
				deepLinking: true,
				presets: [
					SwaggerUIBundle.presets.apis,
					SwaggerUIStandalonePreset
				],
				plugins: [
					SwaggerUIBundle.plugins.DownloadUrl
				],
				layout: "StandaloneLayout",
				defaultModelsExpandDepth: 1,
				defaultModelExpandDepth: 1,
				docExpansion: "list",
				filter: true,
				showRequestHeaders: true,
				showExtensions: true,
				showCommonExtensions: true,
				tryItOutEnabled: true,
				persistAuthorization: true,
				displayOperationId: false,
				displayRequestDuration: true
			});
		};
	</script>
</body>
</html>
	`

	tmpl, err := template.New("swagger").Parse(htmlTemplate)
	if err != nil {
		ctx.String(http.StatusInternalServerError, "Failed to load Swagger UI")
		return
	}

	ctx.Header("Content-Type", "text/html; charset=utf-8")
	err = tmpl.Execute(ctx.Writer, nil)
	if err != nil {
		ctx.String(http.StatusInternalServerError, "Failed to render Swagger UI")
	}
}

func (c *docsController) serveOpenAPISpec(ctx *gin.Context) {
	// Serve the swagger.yaml file from the docs directory
	ctx.File("docs/swagger.yaml")
}
