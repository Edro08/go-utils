module github.com/edro08/go-utils

go 1.27.1

require (
	github.com/golang-jwt/jwt/v5 v5.3.1
	gopkg.in/yaml.v3 v3.0.1
)

retract (
	v1.0.1 // Broken/removed tag superseded by v1.0.2
	v1.0.0 // Experimental release superseded by v1.0.2
)
