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
cfg := config.New(config.Options{
	ExpandEnv:    true,  // Opcional: expande ${VAR} y ${VAR:-default} en YAML
	AutomaticEnv: true,  // Opcional: permite sobrescribir claves vía variables de entorno
	EnvPrefix:    "APP", // Opcional: ej. server.port -> APP_SERVER_PORT
})

// Cargar desde archivo YAML o bytes
_ = cfg.LoadFile("app.yaml")

// Cargar desde struct
_ = cfg.LoadStruct(&myConfig{})
```

- ***Opcional:*** Acceder a valores de configuración, defaults y `Unmarshal`

```go
valor   := cfg.GetString("database.host")
puerto  := cfg.GetIntOr("server.port", 8080)
timeout := cfg.GetDurationOr("server.readTimeout", 15*time.Second)
tlsOn   := cfg.GetBoolOr("server.tls", true)

// Decodificar una sección directamente en un struct tipado
var srvCfg ServerConfig
_ = cfg.Unmarshal("server", &srvCfg)
```

- Establecer o actualizar valores

```go
_ = cfg.Set("app.version", "2.0")
```

### 🧊 Logs

- Importar el paquete

```go
import (
    "github.com/edro08/go-utils/logger"
)
```

- Crear un nuevo logger con `logger.New` (`Options{}` usa `INFO`, `FormatJSON`, `Stdout` y `AddSource: false` por defecto)

```go
opts := logger.Options{
	MinLevel:  logger.INFO,       // DEBUG, INFO, WARN, ERROR, FATAL (o logger.ParseLevel("INFO"))
	Format:    logger.FormatJSON, // FormatJSON o FormatText
	Output:    logger.Both,       // Stdout, File o Both
	FilePath:  "./logs/app.log",  // Ruta de archivo .log o carpeta para timestamp automático
	AddSource: false,             // Opcional: si es true, registra paquete/archivo.go recortado
}

log, err := logger.New(opts)
if err != nil {
	panic(err)
}
defer func() { _ = log.Close() }()
```

- ***Opcional:*** Agregar registros, contexto (`*Ctx`) y grupos
```go
log.Info("Titulo", "key", "value")
log.Debug("Mensaje debug", "modulo", "auth")
log.Error("Error capturado", "err", err)
log.InfoCtx(ctx, "Operación con contexto", "id", 42)

// Logger contextual con atributos fijos o grupos (retorna logger.ILogger)
authLogger := log.With("modulo", "auth").WithGroup("http")
authLogger.Info("Petición procesada", "status", 200)

// Acceder directamente a la instancia *slog.Logger nativa
stdSlog := log.Slog()
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
	ExposedHeaders:   []string{"X-Total-Count"},
	AllowCredentials: true,
	MaxAge:           3600,
}

handler, err := cors.NewCors(opts)
if err != nil {
	panic(err)
}
```

- ***Opcional:*** Envolver tu mux o handler principal

```go
http.ListenAndServe(":8080", handler.Middleware(mux))
```

### 🧊 JWT

- Importar el paquete (soporta `HS256`/`HS384`/`HS512`, `RS256`/`RS384`/`RS512`, `PS256`/`PS384`/`PS512`, `ES256`/`ES384`/`ES512` y `EdDSA`)

```go
import (
    "github.com/edro08/go-utils/jwt"
)
```

- Firmar un token con `jwt.Sign`

```go
key, _ := jwt.ParseRSAPrivateKeyFromPEMFile("private.pem") // o jwt.ParseRSAPrivateKeyFromPEM(bytes)

signed, err := jwt.Sign(jwt.SignOptions{
	ID:            "token-123",
	Algorithm:     jwt.RS256,
	Key:           key,
	Issuer:        "api",
	Subject:       "user-123",
	Audience:      []string{"web-app"},
	Expiration:    1 * time.Hour,
	CustomHeaders: map[string]string{"kid": "key-2026-01"},
	CustomClaims:  map[string]any{"role": "admin"},
})
```

- Verificar un token y obtener sus claims con `jwt.Verify`

```go
claims, err := jwt.Verify(jwt.VerifyOptions{
	Token:     signed,
	Algorithm: jwt.RS256,
	Key:       &key.PublicKey,
	Issuer:    "api",           // Opcional
	Subject:   "user-123",      // Opcional
	Audience:  "web-app",       // Opcional
	Leeway:    5 * time.Second, // Opcional
})
```

- Inspeccionar claims y headers sin verificar firma (ej. para leer `kid` antes de llamar a `Verify`)

```go
rawClaims, headers, err := jwt.ParseUnverified(signed)
```

### 🧊 Cifrado RSA

- Importar el paquete

```go
import (
    "github.com/edro08/go-utils/cipher/rsa"
)
```

- Cifrar y descifrar con RSA-OAEP SHA-256 (soporta llaves privadas PKCS#1/PKCS#8 y públicas PKIX/PKCS#1)

```go
privKey, _ := rsa.ParsePrivateKeyFromPEMFile("private.pem") // o rsa.ParsePrivateKeyFromPEM(bytes)
pubKey, _  := rsa.ParsePublicKeyFromPEMFile("public.pem")   // o rsa.ParsePublicKeyFromPEM(bytes)

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
key := []byte("32-byte-key-here-1234567890abcde")

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
key := []byte("32-byte-key-here-1234567890abcde")
iv  := []byte("16-byte-iv-here!")

ciphertext, _ := aescbc.Encrypt(key, iv, []byte("mensaje secreto"))
plaintext, _  := aescbc.Decrypt(key, iv, ciphertext)
```
