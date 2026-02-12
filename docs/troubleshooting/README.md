# Solución de Problemas

Ayuda para resolver problemas comunes con PicoClaw.

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
picoclaw agent --debug

# Logs a archivo
picoclaw gateway --debug 2>&1 | tee debug.log
```

### Comandos Útiles
```bash
# Ver estado
picoclaw status

# Ver configuración
cat ~/.picoclaw/config.json

# Ver workspace
tree ~/.picoclaw/workspace/
```

## Soporte

Si no encuentras tu problema aquí:

1. Busca en la [FAQ](./faq.md)
2. Revisa [GitHub Issues](https://github.com/sipeed/picoclaw/issues)
3. Únete a [Discord](https://discord.gg/V4sAZ9XWpN)
4. Crea un nuevo issue

---

Para documentación completa, visita la [documentación principal](../README.md).
