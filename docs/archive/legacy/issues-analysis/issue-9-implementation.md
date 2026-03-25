# Issue #9 - Rate Limiters

## Estado: ✅ IMPLEMENTADO

## Descripción

Sistema de rate limiting para prevenir abuso y controlar costos de API.

## Características

- ✅ **Por usuario**: Límites individuales por remitente
- ✅ **Por API**: Límites por proveedor de LLM
- ✅ **Por herramienta**: Límites para operaciones costosas
- ✅ **Configurable**: Fácil de ajustar límites
- ✅ **Información**: Mensajes claros cuando se excede el límite

## Límites por Defecto

### Usuarios

```go
"user:global"  // 100 solicitudes/hora por usuario
"user:burst"   // 10 solicitudes/minuto (burst)
```

### APIs

```go
"api:openai"     // 60 solicitudes/minuto
"api:anthropic"  // 40 solicitudes/minuto  
"api:openrouter" // 100 solicitudes/minuto
```

### Herramientas

```go
"tool:web_search" // 30 búsquedas/hora (Brave free tier)
"tool:shell"      // 20 ejecuciones/hora
```

## Uso

### Mensaje al Usuario

Cuando se excede el límite:

```
Usuario: [envía muchos mensajes rápidamente]
Bot: "Rate limit exceeded. Please wait a moment before sending more messages."
```

### En Código

```go
import "github.com/sipeed/MakoClaw/pkg/ratelimit"

// Verificar si permitido
limiter := ratelimit.GetGlobalLimiter()
if !limiter.Allow("user:123") {
    return "Rate limit exceeded", nil
}

// Con contexto y reintentos
err := ratelimit.WithRateLimitContext(ctx, "api:openai", func() error {
    // Llamada a API
    return callOpenAI()
})
```

### Configurar Límites Personalizados

```go
limiter := ratelimit.GetGlobalLimiter()

// 50 solicitudes por hora
limiter.SetLimit("custom:key", 50, time.Hour)

// Verificar estado
remaining, resetTime := limiter.GetRemaining("user:123")
fmt.Printf("Remaining: %d, Reset at: %v\n", remaining, resetTime)
```

## API del Rate Limiter

### Métodos Principales

```go
// Crear nuevo limiter
rl := ratelimit.NewRateLimiter()

// Configurar límite
rl.SetLimit(key string, requests int, window time.Duration)

// Verificar si permitido
allowed := rl.Allow(key)

// Verificar con tiempo de espera
allowed, waitTime := rl.AllowWithWait(key)

// Obtener información
remaining, resetTime := rl.GetRemaining(key)

// Limpiar entradas antiguas
rl.Cleanup()

// Resetear contador
rl.Reset(key)
```

### Funciones Helper

```go
// Wrapper simple
err := ratelimit.WithRateLimit(key, func() error {
    return doSomething()
})

// Wrapper con contexto
err := ratelimit.WithRateLimitContext(ctx, key, func() error {
    return doSomething()
})

// Singleton global
limiter := ratelimit.GetGlobalLimiter()
```

## Ventajas

- 🛡️ **Previene abuso**: Limita usuarios que envían spam
- 💰 **Controla costos**: Evita gastos inesperados en APIs
- ⚖️ **Fair use**: Distribuye recursos equitativamente
- 📊 **Monitoreo**: Logs de cuándo se aplican límites

## Casos de Uso

### 1. Protección contra Spam

```
Usuario envía 50 mensajes en 1 minuto → Rate limit activado
Usuario debe esperar antes de continuar
```

### 2. Control de Costos API

```
Cada llamada a GPT-4 cuesta $0.03
Límite: 100/hora = máximo $3/hora por usuario
```

### 3. Fair Use en Grupos

```
Grupo con 100 usuarios
Sin rate limit: 1 usuario monopoliza
Con rate limit: todos tienen oportunidad
```

## Troubleshooting

### "Rate limit exceeded"

- Espera unos minutos
- Reduce la frecuencia de mensajes
- Contacta admin si necesitas más límites

### Ajustar Límites

Editar código en `pkg/ratelimit/ratelimit.go`:

```go
// En GetGlobalLimiter()
globalLimiter.SetLimit("user:global", 200, time.Hour) // Aumentar a 200/hora
```

## Referencias

- Issue original: https://github.com/sipeed/MakoClaw/issues/9
- Token bucket algorithm: https://en.wikipedia.org/wiki/Token_bucket
