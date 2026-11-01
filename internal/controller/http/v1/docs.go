package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const scalarHTML = `<!doctype html>
<html>
<head>
    <title>Araya Business Hub API</title>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
</head>
<body>
    <div id="app"></div>

    <script type="module">
        import { createApiReference } from
        'https://cdn.jsdelivr.net/npm/@scalar/api-reference/esm.js'

        createApiReference('#app', {
            url: '/openapi.yaml',
            theme: 'purple',
            layout: 'modern'
        })
    </script>
</body>
</html>`

func docsHandler(c *gin.Context) {
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(scalarHTML))
}
