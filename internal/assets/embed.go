package assets

import (
	_ "embed"
)

//go:embed openapi.yaml
var OpenAPISpec []byte
