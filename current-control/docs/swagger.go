package docs

import (
	_ "embed"
	"net/http"
)

//go:embed swagger.json
var swaggerJSON []byte

func ServeSpec(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write(swaggerJSON)
}

func ServeUI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(`
<!DOCTYPE html>
<html>
<head>
	<meta charset="utf-8">
	<title>Orders API</title>
	<link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5.17.14/swagger-ui.css">
</head>
<body>
	<div id="swagger-ui"></div>
	<script src="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5.17.14/swagger-ui-bundle.js"></script>
	<script>
		SwaggerUIBundle({
			url: '/swagger/doc.json',
			dom_id: '#swagger-ui'
		});
	</script>
</body>
</html>
`))
}
