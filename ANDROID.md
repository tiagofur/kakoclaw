# 📱 PicoClaw en Android (Termux)

¡PicoClaw funciona perfectamente en Android a través de Termux!

## ⚡ Instalación Rápida

Copia y pega esto en Termux:

```bash
curl -fsSL https://raw.githubusercontent.com/sipeed/picoclaw/main/scripts/install-termux.sh | bash
```

## 📋 Requisitos

- **Android**: 7.0+ (API 24+)
- **Termux**: Desde F-Droid (NO Google Play)
- **RAM**: 2GB+ recomendado
- **Espacio**: 100MB libres

## 🚀 Uso Básico

```bash
# Verificar instalación
picoclaw version

# Modo interactivo
picoclaw agent

# Comando directo
picoclaw agent -m "Hola desde Android"

# Ver estado
picoclaw status
```

## 🔧 Configuración Rápida

### Opción 1: Ollama (Sin API keys, offline)

```bash
# En proot-distro
proot-distro install alpine
proot-distro login alpine
apk add ollama
ollama serve &
ollama pull llama3.2

# Configurar PicoClaw
# ~/.picoclaw/config.json:
{
  "agents": {
    "defaults": {
      "model": "llama3.2"
    }
  }
}
```

### Opción 2: Con API Keys

```bash
nano ~/.picoclaw/config.json
```

```json
{
  "agents": {
    "defaults": {
      "model": "openai/gpt-4"
    }
  },
  "providers": {
    "openai": {
      "api_key": "sk-..."
    }
  }
}
```

## 💡 Características en Android

- ✅ **Asistente personal** completo
- ✅ **Modo offline** con Ollama
- ✅ **Bot de Telegram** desde tu teléfono
- ✅ **Automatización** de tareas
- ✅ **Batería optimizada** (<10MB RAM)

## 📚 Documentación Completa

Ver: [docs/deployment/termux-android.md](docs/deployment/termux-android.md)

## 🐛 Soporte

- **Issues**: https://github.com/sipeed/picoclaw/issues
- **Discord**: https://discord.gg/V4sAZ9XWpN

---

**¡Tu asistente de IA en el bolsillo! 🦞📱**
