package swagger

import (
	"bytes"
	"context"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sync"

	"github.com/oarkflow/frame"
	"github.com/oarkflow/frame/pkg/protocol/consts"
	swaggerFiles "github.com/swaggo/files"
	"github.com/swaggo/swag"
	"golang.org/x/net/webdav"
)

type swaggerConfig struct {
	URL                      string
	DocExpansion             string
	Title                    string
	Oauth2RedirectURL        template.JS
	DefaultModelsExpandDepth int
	DeepLinking              bool
	PersistAuthorization     bool
	Oauth2DefaultClientID    string
}

// Config stores hertzSwagger configuration variables.
type Config struct {
	// The url pointing to API definition (normally swagger.json or swagger.yaml). Default is `doc.json`.
	URL                      string
	DocExpansion             string
	InstanceName             string
	Title                    string
	DefaultModelsExpandDepth int
	DeepLinking              bool
	PersistAuthorization     bool
	Oauth2DefaultClientID    string
	Handler                  *webdav.Handler
}

func (config Config) toSwaggerConfig() swaggerConfig {
	return swaggerConfig{
		URL:                      config.URL,
		DeepLinking:              config.DeepLinking,
		DocExpansion:             config.DocExpansion,
		DefaultModelsExpandDepth: config.DefaultModelsExpandDepth,
		Oauth2RedirectURL: "`${window.location.protocol}//${window.location.host}$" +
			"{window.location.pathname.split('/').slice(0, window.location.pathname.split('/').length - 1).join('/')}" +
			"/oauth2-redirect.html`",
		Title:                 config.Title,
		PersistAuthorization:  config.PersistAuthorization,
		Oauth2DefaultClientID: config.Oauth2DefaultClientID,
	}
}

func defaultConfig() *Config {
	return &Config{
		URL:                      "doc.json",
		DocExpansion:             "list",
		InstanceName:             swag.Name,
		Title:                    "Swagger UI",
		DefaultModelsExpandDepth: 1,
		DeepLinking:              true,
	}
}

// New wraps `http.Handler` into `frame.HandlerFunc`.
func New(cfg ...*Config) frame.HandlerFunc {
	var config *Config
	if len(cfg) > 0 {
		config = cfg[0]
	} else {
		config = defaultConfig()
	}
	if config.URL == "" {
		config.URL = "doc.json"
	}
	if config.DocExpansion == "" {
		config.DocExpansion = "list"
	}
	if config.InstanceName == "" {
		config.InstanceName = swag.Name
	}
	if config.Title == "" {
		config.Title = "Swagger UI"
	}
	if config.DefaultModelsExpandDepth == 0 {
		config.DefaultModelsExpandDepth = 1
	}
	if config.Handler == nil {
		config.Handler = swaggerFiles.Handler
	}

	var once sync.Once

	// create a template with name
	index, _ := template.New("swagger_index.html").Parse(swaggerIndexTpl)

	matcher := regexp.MustCompile(`(.*)(index\.html|doc\.json|favicon-16x16\.png|favicon-32x32\.png|/oauth2-redirect\.html|swagger-ui\.css|swagger-ui\.css\.map|swagger-ui\.js|swagger-ui\.js\.map|swagger-ui-bundle\.js|swagger-ui-bundle\.js\.map|swagger-ui-standalone-preset\.js|swagger-ui-standalone-preset\.js\.map)[?|.]*`)

	return func(c context.Context, ctx *frame.Context) {
		if string(ctx.Request.Method()) != consts.MethodGet {
			ctx.AbortWithStatus(http.StatusMethodNotAllowed)
			return
		}

		matches := matcher.FindStringSubmatch(ctx.Request.URI().String())
		if len(matches) != 3 && ctx.Param("any") != "" {
			ctx.String(http.StatusNotFound, http.StatusText(http.StatusNotFound))

			return
		}
		path := ""
		if ctx.Param("any") != "" {
			path = matches[2]
		} else {
			path = "index.html"
		}
		prefix := ""
		if len(matches) != 3 {
			prefix = ctx.Request.URI().String()
		} else {
			prefix = matches[1]
		}

		once.Do(func() {
			config.Handler.Prefix = prefix
		})

		switch filepath.Ext(path) {
		case ".html":
			ctx.Header("Content-Type", "text/html; charset=utf-8")
		case ".css":
			ctx.Header("Content-Type", "text/css; charset=utf-8")
		case ".js":
			ctx.Header("Content-Type", "application/javascript")
		case ".png":
			ctx.Header("Content-Type", "image/png")
		case ".json":
			ctx.Header("Content-Type", "application/json; charset=utf-8")
		}

		switch path {
		case "index.html":
			_ = index.Execute(ctx, config.toSwaggerConfig())
		case "doc.json":
			doc, err := swag.ReadDoc(config.InstanceName)
			if err != nil {
				ctx.AbortWithStatus(http.StatusInternalServerError)
				return
			}
			if _, err = ctx.Write([]byte(doc)); err != nil {
				ctx.AbortWithStatus(http.StatusInternalServerError)
				return
			}

		default:
			f, err := config.Handler.FileSystem.OpenFile(c, path, os.O_RDONLY, 0)
			if err != nil {
				ctx.AbortWithStatus(http.StatusInternalServerError)
				return
			}
			buf := new(bytes.Buffer)
			if _, err = buf.ReadFrom(f); err != nil {
				ctx.AbortWithStatus(http.StatusInternalServerError)
				return
			}
			if _, err = ctx.Write(buf.Bytes()); err != nil {
				ctx.AbortWithStatus(http.StatusInternalServerError)
				return
			}
		}
	}
}

const swaggerIndexTpl = `<!-- HTML for static distribution bundle build -->
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>{{.Title}}</title>
  <link href="https://fonts.googleapis.com/css?family=Open+Sans:400,700|Source+Code+Pro:300,600|Titillium+Web:400,600,700" rel="stylesheet">
  <link rel="stylesheet" type="text/css" href="./swagger-ui.css" >
  <link rel="icon" type="image/png" href="./favicon-32x32.png" sizes="32x32" />
  <link rel="icon" type="image/png" href="./favicon-16x16.png" sizes="16x16" />
  <style>
    html
    {
        box-sizing: border-box;
        overflow: -moz-scrollbars-vertical;
        overflow-y: scroll;
    }
    *,
    *:before,
    *:after
    {
        box-sizing: inherit;
    }

    body {
      margin:0;
      background: #fafafa;
    }
  </style>
</head>

<body>

<svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" style="position:absolute;width:0;height:0">
  <defs>
    <symbol viewBox="0 0 20 20" id="unlocked">
          <path d="M15.8 8H14V5.6C14 2.703 12.665 1 10 1 7.334 1 6 2.703 6 5.6V6h2v-.801C8 3.754 8.797 3 10 3c1.203 0 2 .754 2 2.199V8H4c-.553 0-1 .646-1 1.199V17c0 .549.428 1.139.951 1.307l1.197.387C5.672 18.861 6.55 19 7.1 19h5.8c.549 0 1.428-.139 1.951-.307l1.196-.387c.524-.167.953-.757.953-1.306V9.199C17 8.646 16.352 8 15.8 8z"></path>
    </symbol>

    <symbol viewBox="0 0 20 20" id="locked">
      <path d="M15.8 8H14V5.6C14 2.703 12.665 1 10 1 7.334 1 6 2.703 6 5.6V8H4c-.553 0-1 .646-1 1.199V17c0 .549.428 1.139.951 1.307l1.197.387C5.672 18.861 6.55 19 7.1 19h5.8c.549 0 1.428-.139 1.951-.307l1.196-.387c.524-.167.953-.757.953-1.306V9.199C17 8.646 16.352 8 15.8 8zM12 8H8V5.199C8 3.754 8.797 3 10 3c1.203 0 2 .754 2 2.199V8z"/>
    </symbol>

    <symbol viewBox="0 0 20 20" id="close">
      <path d="M14.348 14.849c-.469.469-1.229.469-1.697 0L10 11.819l-2.651 3.029c-.469.469-1.229.469-1.697 0-.469-.469-.469-1.229 0-1.697l2.758-3.15-2.759-3.152c-.469-.469-.469-1.228 0-1.697.469-.469 1.228-.469 1.697 0L10 8.183l2.651-3.031c.469-.469 1.228-.469 1.697 0 .469.469.469 1.229 0 1.697l-2.758 3.152 2.758 3.15c.469.469.469 1.229 0 1.698z"/>
    </symbol>

    <symbol viewBox="0 0 20 20" id="large-arrow">
      <path d="M13.25 10L6.109 2.58c-.268-.27-.268-.707 0-.979.268-.27.701-.27.969 0l7.83 7.908c.268.271.268.709 0 .979l-7.83 7.908c-.268.271-.701.27-.969 0-.268-.269-.268-.707 0-.979L13.25 10z"/>
    </symbol>

    <symbol viewBox="0 0 20 20" id="large-arrow-down">
      <path d="M17.418 6.109c.272-.268.709-.268.979 0s.271.701 0 .969l-7.908 7.83c-.27.268-.707.268-.979 0l-7.908-7.83c-.27-.268-.27-.701 0-.969.271-.268.709-.268.979 0L10 13.25l7.418-7.141z"/>
    </symbol>


    <symbol viewBox="0 0 24 24" id="jump-to">
      <path d="M19 7v4H5.83l3.58-3.59L8 6l-6 6 6 6 1.41-1.41L5.83 13H21V7z"/>
    </symbol>

    <symbol viewBox="0 0 24 24" id="expand">
      <path d="M10 18h4v-2h-4v2zM3 6v2h18V6H3zm3 7h12v-2H6v2z"/>
    </symbol>

  </defs>
</svg>

<div id="swagger-ui"></div>

<script src="./swagger-ui-bundle.js"> </script>
<script src="./swagger-ui-standalone-preset.js"> </script>
<script>
window.onload = function() {
  // Build a system
  const ui = SwaggerUIBundle({
    url: "{{.URL}}",
    dom_id: '#swagger-ui',
    validatorUrl: null,
    oauth2RedirectUrl: {{.Oauth2RedirectURL}},
    persistAuthorization: {{.PersistAuthorization}},
    presets: [
      SwaggerUIBundle.presets.apis,
      SwaggerUIStandalonePreset
    ],
    plugins: [
      SwaggerUIBundle.plugins.DownloadUrl
    ],
	layout: "StandaloneLayout",
    docExpansion: "{{.DocExpansion}}",
	deepLinking: {{.DeepLinking}},
	defaultModelsExpandDepth: {{.DefaultModelsExpandDepth}}
  })

  const defaultClientId = "{{.Oauth2DefaultClientID}}";
  if (defaultClientId) {
    ui.initOAuth({
      clientId: defaultClientId
    })
  }

  window.ui = ui
}
</script>
</body>

</html>
`
