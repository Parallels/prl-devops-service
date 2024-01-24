package restapi

import (
	"html/template"
	"net/http"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"

	swaggerFiles "github.com/swaggo/files/v2"
	"github.com/swaggo/swag"
)

// SwaggerConfig stores httpSwagger configuration variables.
type SwaggerConfig struct {
	// The url pointing to API definition (normally swagger.json or swagger.yaml). Default is `doc.json`.
	URL                      string
	DocExpansion             string
	DomID                    string
	InstanceName             string
	BeforeScript             template.JS
	AfterScript              template.JS
	Plugins                  []template.JS
	UIConfig                 map[template.JS]template.JS
	DeepLinking              bool
	PersistAuthorization     bool
	Layout                   SwaggerLayout
	DefaultModelsExpandDepth ModelsExpandDepthType
}

// SwaggerURL presents the url pointing to API definition (normally swagger.json or swagger.yaml).
func SwaggerURL(url string) func(*SwaggerConfig) {
	return func(c *SwaggerConfig) {
		c.URL = url
	}
}

// SwaggerDeepLinking true, false.
func SwaggerDeepLinking(deepLinking bool) func(*SwaggerConfig) {
	return func(c *SwaggerConfig) {
		c.DeepLinking = deepLinking
	}
}

// SwaggerDocExpansion list, full, none.
func SwaggerDocExpansion(docExpansion string) func(*SwaggerConfig) {
	return func(c *SwaggerConfig) {
		c.DocExpansion = docExpansion
	}
}

// SwaggerDomID #swagger-ui.
func SwaggerDomID(domID string) func(*SwaggerConfig) {
	return func(c *SwaggerConfig) {
		c.DomID = domID
	}
}

// SwaggerInstanceName set the instance name that was used to generate the swagger documents
// Defaults to swag.Name ("swagger").
func SwaggerInstanceName(name string) func(*SwaggerConfig) {
	return func(c *SwaggerConfig) {
		c.InstanceName = name
	}
}

// SwaggerPersistAuthorization Persist authorization information over browser close/refresh.
// Defaults to false.
func SwaggerPersistAuthorization(persistAuthorization bool) func(*SwaggerConfig) {
	return func(c *SwaggerConfig) {
		c.PersistAuthorization = persistAuthorization
	}
}

// SwaggerPlugins specifies additional plugins to load into Swagger UI.
func SwaggerPlugins(plugins []string) func(*SwaggerConfig) {
	return func(c *SwaggerConfig) {
		vs := make([]template.JS, len(plugins))
		for i, v := range plugins {
			vs[i] = template.JS(v) // #nosec G203 thi
		}
		c.Plugins = vs
	}
}

// SwaggerUIConfig specifies additional SwaggerUIBundle config object properties.
func SwaggerUIConfig(props map[string]string) func(*SwaggerConfig) {
	return func(c *SwaggerConfig) {
		vs := make(map[template.JS]template.JS, len(props))
		for k, v := range props {
			vs[template.JS(k)] = template.JS(v) // #nosec G203 the user does not have access to this
		}
		c.UIConfig = vs
	}
}

// SwaggerBeforeScript holds JavaScript to be run right before the Swagger UI object is created.
func SwaggerBeforeScript(js string) func(*SwaggerConfig) {
	return func(c *SwaggerConfig) {
		c.BeforeScript = template.JS(js) // #nosec G203 the user does not have access to this
	}
}

// SwaggerAfterScript holds JavaScript to be run right after the Swagger UI object is created
// and set on the window.
func SwaggerAfterScript(js string) func(*SwaggerConfig) {
	return func(c *SwaggerConfig) {
		c.AfterScript = template.JS(js) // #nosec G203 the user does not have access to this
	}
}

type SwaggerLayout string

const (
	BaseLayout       SwaggerLayout = "BaseLayout"
	StandaloneLayout SwaggerLayout = "StandaloneLayout"
)

// Define SwaggerLayoutConfig options are BaseLayout or StandaloneLayout
func SwaggerLayoutConfig(layout SwaggerLayout) func(*SwaggerConfig) {
	return func(c *SwaggerConfig) {
		c.Layout = layout
	}
}

type ModelsExpandDepthType int

const (
	ShowModel ModelsExpandDepthType = 1
	HideModel ModelsExpandDepthType = -1
)

// SwaggerDefaultModelsExpandDepth presents the model of response and request.
// set the default expansion depth for models
func SwaggerDefaultModelsExpandDepth(defaultModelsExpandDepth ModelsExpandDepthType) func(*SwaggerConfig) {
	return func(c *SwaggerConfig) {
		c.DefaultModelsExpandDepth = defaultModelsExpandDepth
	}
}

func swaggerNewConfig(configFns ...func(*SwaggerConfig)) *SwaggerConfig {
	config := SwaggerConfig{
		URL:                      "doc.json",
		DocExpansion:             "list",
		DomID:                    "swagger-ui",
		InstanceName:             "swagger",
		DeepLinking:              true,
		PersistAuthorization:     false,
		Layout:                   StandaloneLayout,
		DefaultModelsExpandDepth: ShowModel,
	}

	for _, fn := range configFns {
		fn(&config)
	}

	if config.InstanceName == "" {
		config.InstanceName = swag.Name
	}

	return &config
}

// SwaggerHandler wraps `http.SwaggerHandler` into `http.HandlerFunc`.
func SwaggerHandler(configFns ...func(*SwaggerConfig)) http.HandlerFunc {
	config := swaggerNewConfig(configFns...)

	// create a template with name
	index, _ := template.New("swagger_index.html").Parse(swaggerIndexTempl)

	re := regexp.MustCompile(`^(.*/)([^?].*)?[?|.]*$`)

	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)

			return
		}
		requestUri := r.RequestURI
		if !strings.HasSuffix(requestUri, "/") {
			filePathExt := filepath.Ext(requestUri)
			if filePathExt == "" {
				requestUri = requestUri + "/"
			}
		}

		matches := re.FindStringSubmatch(requestUri)

		path := matches[2]

		switch filepath.Ext(path) {
		case ".html":
			w.Header().Set("Content-Type", "text/html; charset=utf-8")

		case ".css":
			w.Header().Set("Content-Type", "text/css; charset=utf-8")
		case ".js":
			w.Header().Set("Content-Type", "application/javascript")
		case ".png":
			w.Header().Set("Content-Type", "image/png")
		case ".json":
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
		}

		switch path {
		case "index.html":
			_ = index.Execute(w, config)
		case "doc.json":
			doc, err := swag.ReadDoc(config.InstanceName)
			if err != nil {
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)

				return
			}

			_, _ = w.Write([]byte(doc))
		case "":
			http.Redirect(w, r, matches[1]+"/"+"index.html", http.StatusMovedPermanently)
		default:
			var err error
			r.URL, err = url.Parse(matches[2])
			if err != nil {
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)

				return
			}
			http.FileServer(http.FS(swaggerFiles.FS)).ServeHTTP(w, r)
		}
	}
}

const swaggerIndexTempl = `<!-- HTML for static distribution bundle build -->
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>Swagger UI</title>
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
  {{- if .BeforeScript}}
  {{.BeforeScript}}
  {{- end}}
  // Build a system
  const ui = SwaggerUIBundle({
    url: "{{.URL}}",
    deepLinking: {{.DeepLinking}},
    docExpansion: "{{.DocExpansion}}",
    dom_id: "#{{.DomID}}",
    persistAuthorization: {{.PersistAuthorization}},
    validatorUrl: null,
    presets: [
      SwaggerUIBundle.presets.apis,
      SwaggerUIStandalonePreset
    ],
    plugins: [
      SwaggerUIBundle.plugins.DownloadUrl
      {{- range $plugin := .Plugins }},
      {{$plugin}}
      {{- end}}
    ],
    {{- range $k, $v := .UIConfig}}
    {{$k}}: {{$v}},
    {{- end}}
    layout: "{{$.Layout}}",
    defaultModelsExpandDepth: {{.DefaultModelsExpandDepth}}
  })

  window.ui = ui
  {{- if .AfterScript}}
  {{.AfterScript}}
  {{- end}}
}
</script>
</body>

</html>
`
