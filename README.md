## 💻 go-utils

**Descripción:** Una librería de utilidades generales para proyectos en Go. Incluye herramientas reutilizables para configuraciones, logs, CORS, tokens JWT y cifrado.

### 🔌 Instalación

```bash
go get github.com/edro08/go-utils
```

### 🧊 Configuración

- Importar el paquete

```go
import (
    "github.com/edro08/go-utils/config"
)
```

- Crear una nueva configuración con `config.New`

```go
cfg := config.New(config.Options{})

// Cargar desde archivo YAML
cfg.LoadFile("app.yaml")

// Cargar desde struct
cfg.LoadStruct(&myConfig{})
```

- ***Opcional:*** Acceder a valores de configuración

```go
valor  := cfg.GetString("database.host")
puerto := cfg.GetInt("server.port")
debug  := cfg.GetBool("debug")
```

- Establecer o actualizar valores

```go
cfg.Set("app.version", "2.0")
```

### 🧊 Logs

- Importar el paquete

```go
import (
    "github.com/edro08/go-utils/logger"
)
```

- Crear un nuevo logger con `logger.New`

```go
opts := logger.Options{
	MinLevel: logger.DEBUG,
	Format:   logger.FormatText,
}

log, _ := logger.New(opts)
```

- ***Opcional:*** Agregar un nuevo log

```go
log.Info("Titulo", "key", "value")
log.Debug("Mensaje debug", "modulo", "auth")
log.Error("Error capturado", "err", err)
```

### 🧊 CORS

- Importar el paquete

```go
import (
    "github.com/edro08/go-utils/cors"
)
```

- Crear un handler CORS con `cors.NewCors`

```go
opts := cors.Options{
	AllowedOrigins:   []string{"https://example.com"},
	AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
	AllowedHeaders:   []string{"Content-Type", "Authorization"},
	AllowCredentials: true,
}

handler := cors.NewCors(opts)
```

- ***Opcional:*** Envolver tu mux o handler principal

```go
http.ListenAndServe(":8080", handler.Middleware(mux))
```

### 🧊 JWT

- Importar el paquete

```go
import (
    "github.com/edro08/go-utils/jwt"
)
```

- Crear y firmar un token con `jwt.New`

```go
token := jwt.New(jwt.NewTokenOpts{
	Algorithm:      jwt.RS256,
	Issuer:         "api",
	Subject:        "user-123",
	ExpiredEnabled: true,
	ExpiredTime:    1 * time.Hour,
})

key, _ := jwt.ParseRSAPrivateKeyFromPEMFile("private.pem")
signed, _ := token.Sign(key)
```

- Verificar un token

```go
err := jwt.Verify(signed, &key.PublicKey)
```

- Parsear sin verificar

```go
claims, _ := jwt.Parse(signed)
```

### 🧊 Cifrado RSA

- Importar el paquete

```go
import (
    "github.com/edro08/go-utils/cipher/rsa"
)
```

- Cifrar y descifrar con RSA

```go
privKey, _ := rsa.ParsePrivateKeyFromPEMFile("private.pem")
pubKey, _  := rsa.ParsePublicKeyFromPEMFile("public.pem")

ciphertext, _ := rsa.Encrypt(pubKey, []byte("mensaje secreto"))
plaintext, _  := rsa.Decrypt(privKey, ciphertext)
```

### 🧊 Cifrado AES-GCM

- Importar el paquete

```go
import (
    "github.com/edro08/go-utils/cipher/aesgcm"
)
```

- Cifrar y descifrar con AES-GCM

```go
key := []byte("32-byte-key-here-1234567890abcd")

ciphertext, _ := aesgcm.Encrypt(key, []byte("mensaje secreto"))
plaintext, _  := aesgcm.Decrypt(key, ciphertext)
```

### 🧊 Cifrado AES-CBC

- Importar el paquete

```go
import (
    "github.com/edro08/go-utils/cipher/aescbc"
)
```

- Cifrar y descifrar con AES-CBC

```go
key := []byte("32-byte-key-here-1234567890abcd")
iv  := []byte("16-byte-iv-here!")

ciphertext, _ := aescbc.Encrypt(key, iv, []byte("mensaje secreto"))
plaintext, _  := aescbc.Decrypt(key, iv, ciphertext)
```
