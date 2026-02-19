# Issue #41 - Signal Channel Integration

## Estado: ✅ IMPLEMENTADO

## Descripción

Integración completa con Signal messenger para recibir y enviar mensajes a través de KakoClaw.

## Requisitos Previos

### Instalar signal-cli

Signal requiere `signal-cli` para funcionar:

```bash
# Linux (Ubuntu/Debian)
sudo apt install signal-cli

# macOS
brew install signal-cli

# O descargar manualmente desde:
# https://github.com/AsamK/signal-cli/releases
```

### Registrar Número de Teléfono

```bash
# Registrar número (se enviará código de verificación por SMS)
signal-cli -a +1234567890 register

# Verificar con código recibido
signal-cli -a +1234567890 verify 123456
```

## Configuración

Editar `~/.KakoClaw/config.json`:

```json
{
  "channels": {
    "signal": {
      "enabled": true,
      "phone_number": "+1234567890",
      "allow_from": ["+0987654321"]
    }
  }
}
```

### Opciones

- `enabled`: Habilitar/deshabilitar canal Signal
- `phone_number`: Número de teléfono registrado en Signal (formato internacional)
- `allow_from`: Lista de números permitidos (vacío = permitir todos)

## Uso

```bash
# Iniciar gateway
KakoClaw gateway
```

Envía un mensaje a tu número de Signal y KakoClaw responderá.

## Características

- ✅ Recibir mensajes de texto
- ✅ Enviar respuestas
- ✅ Lista de permitidos (allowlist)
- ✅ Soporta adjuntos (básico)
- ✅ Privacidad completa (mensajes E2E)

## Limitaciones

- ⚠️ Requiere `signal-cli` instalado y configurado
- ⚠️ Número debe estar registrado en Signal
- ⚠️ Polling cada 5 segundos (no es realtime)
- ⚠️ Solo mensajes directos 1:1 (no grupos)

## Troubleshooting

### "signal-cli not found"

Instala signal-cli siguiendo las instrucciones de arriba.

### "signal account not registered"

Registra el número primero:
```bash
signal-cli -a +1234567890 register
signal-cli -a +1234567890 verify CODIGO
```

### Mensajes no llegan

1. Verifica que el número esté en `allow_from`
2. Revisa logs: `KakoClaw gateway --debug`
3. Verifica signal-cli: `signal-cli -a +1234567890 receive`

## Privacidad

Signal es el canal más privado disponible:
- Mensajes E2E cifrados
- No requiere API keys de terceros
- Sin logs en servidores externos
- Ideal para comunicaciones sensibles

## Referencias

- signal-cli: https://github.com/AsamK/signal-cli
- Signal Protocol: https://signal.org/docs/
- Issue original: https://github.com/sipeed/KakoClaw/issues/41
