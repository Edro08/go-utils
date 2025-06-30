## 💻 go-utils

**Descripción:** Una librería de utilidades generales para proyectos en Go. Incluye herramientas reutilizables para lectura de configuraciones, logs.

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

- Crear una nueva configuración con `NewConfig`
```go
opts := config.Opts{
	File: "app.yaml",
}

cfg, err := config.NewConfig(opts)
if err != nil {
	log.Fatalf("Error cargando configuración: %v", err)
}

```

- ***Opcional:*** Acceder a valores de configuración
```go
valor := cfg.GetString("database.host")
````

### 🧊 Logs
- Importar el paquete
```go
import (
    "github.com/edro08/go-utils/logger"
)
```
- Crear un nuevo logger con `NewLogger`
```go
opts := logger.Opts{
	MinLevel: logger.DEBUG, 
	Format:   logger.FormatText,
}

newLogger, err := logger.NewLogger(opts)
if err != nil {
	log.Fatalf("Error creando logger: %v", err)
}
```
- ***Opcional:*** Agregar un nuevo log
```go
newLogger.Info("Titulo", "key", "value")
```

