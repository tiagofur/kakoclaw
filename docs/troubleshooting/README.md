# Solución de Problemas

Ayuda para resolver problemas comunes con KakoClaw.

## Archivos

- **[common-issues.md](./common-issues.md)** - Problemas comunes (pendiente)
- **[config-errors.md](./config-errors.md)** - Errores de configuración (pendiente)
- **[channel-issues.md](./channel-issues.md)** - Problemas de canales (pendiente)
- **[debugging.md](./debugging.md)** - Técnicas de debugging (pendiente)
- **[faq.md](./faq.md)** - Preguntas frecuentes

## Índice de Problemas

### Instalación
- "command not found"
- Errores de compilación
- Problemas de permisos

### Configuración
- API keys inválidas
- Configuración no encontrada
- Variables de entorno

### Uso
- Tools no funcionan
- Errores de LLM
- Problemas de memoria

### Canales
- Telegram no responde
- Discord connection failed
- WhatsApp no funciona

## Debugging

### Logs
```bash
# Modo debug
KakoClaw agent --debug

# Logs a archivo
KakoClaw gateway --debug 2>&1 | tee debug.log
```

### Comandos Útiles
```bash
# Ver estado
KakoClaw status

# Ver configuración
cat ~/.KakoClaw/config.json

# Ver workspace
tree ~/.KakoClaw/workspace/
```

## Soporte

Si no encuentras tu problema aquí:

1. Busca en la [FAQ](./faq.md)
2. Revisa [GitHub Issues](https://github.com/sipeed/KakoClaw/issues)
3. Únete a [Discord](https://discord.gg/V4sAZ9XWpN)
4. Crea un nuevo issue

---

Para documentación completa, visita la [documentación principal](../README.md).
