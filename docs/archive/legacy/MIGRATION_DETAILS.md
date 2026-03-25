# Migración de MakoClaw a MakoClaw

Este documento ayuda a migrar de **MakoClaw** (versión verde) a **MakoClaw** (versión azul con tiburón).

---

## 🎨 Cambios Visuales

### Colores

| Elemento | MakoClaw (Antes) | MakoClaw (Nuevo) |
|-----------|-------------------|------------------|
| Accent Color | Emerald (#10b981) | Blue (#3b82f6) |
| Accent Hover | Emerald 600 (#059669) | Blue 600 (#2563eb) |
| Theme Color | #10b981 | #3b82f6 |
| Mascota | 🦈 (Rana) | 🦈 (Tiburón) |

### Marca

- **Nombre**: MakoClaw → **MakoClaw**
- **Dominio**: makoclaw.com → **makoclaw.com**
- **Eslogan**: "Your AI Agent Platform" → "**The Apex AI Agent**"

---

## ⚙️ Cambios en Configuración

### Actualizar Configuración

No se requieren cambios en `config.json` para funcionar. Los valores de configuración son compatibles.

Sin embargo, puedes querer actualizar referencias:

```json
{
  "agents": {
    "defaults": {
      "model": "anthropic/claude-3.5-sonnet"
    }
  },
  // ... resto de configuración
}
```

### Actualizar Canales

Los canales no requieren cambios. Sigue funcionando igual.

**Nota**: Si tenías bots de Telegram/Discord configurados, seguirán funcionando sin cambios.

---

## 🔄 Actualización del Código

### Cambios en Colores CSS

Si personalizaste la UI con CSS personalizado, actualiza las referencias de color:

```css
/* Antes */
--pc-accent: 16 185 129;  /* Emerald 500 */
--pc-accent-hover: 5 150 105;  /* Emerald 600 */

/* Después */
--pc-accent: 59 130 246;  /* Blue 500 */
--pc-accent-hover: 37 99 235;  /* Blue 600 */
```

### Cambios en Logo

Si personalizaste el logo:

```html
<!-- Antes -->
<h1>MakoClaw</h1>
<img src="logo.png" alt="MakoClaw" />

<!-- Después -->
<h1>MakoClaw</h1>
<img src="assets/mascot.png" alt="MakoClaw" />
```

---

## 📚 Documentación

La documentación ha sido actualizada completamente:

- 📖 [Nueva documentación principal](../README.md)
- 🚀 [Guía de inicio rápido](../docs/guides/quickstart.md)
- 🏗️ [Arquitectura actualizada](../docs/architecture/overview.md)

---

## 🌐 Web UI

### Cambios en Landing Page

La Landing Page ha sido rediseñada completamente con:

- **Nueva identidad visual**: Colores azules y tiburón
- **Todas las funcionalidades**: Multi-Agent, Workflows, Kanban, Knowledge Base
- **Canales**: Sección actualizada con 9+ canales
- **Features**: 8 categorías de capacidades

### Panel Web

El panel web mantiene la misma funcionalidad con nueva identidad visual:

- Chat con historial
- Kanban Task Board
- Visual Workflows
- Multi-Agent System
- Gestión de archivos
- Base de conocimientos
- Cron jobs
- Métricas

---

## 🐛 Problemas Comunes

### El panel sigue mostrando colores antiguos

**Solución**:
1. Limpiar caché del navegador
2. Recargar la página (Ctrl+F5 o Cmd+Shift+R)
3. Verificar que el frontend haya sido recompilado

```bash
# Rebuild frontend
cd pkg/web/frontend
npm run build
```

### La Landing Page no muestra los cambios

**Solución**:
1. Verificar que `pkg/web/frontend/dist/index.html` tenga la nueva versión
2. Reiniciar el servidor web
3. Limpiar caché del navegador

### Error: "MakoClaw" en logs

**Solución**: Los logs internos pueden mantener el nombre anterior. Esto es normal y no afecta la funcionalidad. Se actualizará en futuras versiones.

---

## 🔄 Rolling Update

### Para Usuarios Auto-Hosted

```bash
# Actualizar código
git pull origin main

# Rebuild
make build

# Reiniciar servicio
sudo systemctl restart makoclaw
# o
MakoClaw web
```

### Para Docker Users

```bash
# Pull latest image
docker pull sipeed/makoclaw:latest

# Stop and remove old container
docker stop makoclaw
docker rm makoclaw

# Run new container
docker run -d --name makoclaw \
  -p 18880:18880 \
  -v ~/.makoclaw:/root/.makoclaw \
  sipeed/makoclaw:latest
```

---

## 📊 Nuevas Funcionalidades

MakoClaw incluye todas las funcionalidades de MakoClaw con mejoras:

### Mejoras Visuales
- Nueva paleta de colores azul
- Nueva identidad de marca
- Mejor contraste y legibilidad

### Mejoras de Documentación
- Documentación actualizada completa
- Ejemplos más claros
- Mejor organización

### Compatibilidad
- 100% backward compatible con configuración de MakoClaw
- Canales funcionan sin cambios
- Workflows sin modificaciones

---

## 🎓 Siguientes Pasos

1. **Actualizar el código**: `git pull && make build`
2. **Limpiar caché**: Browser cache y frontend build
3. **Probar funcionalidad**: Verificar que todo funcione
4. **Personalizar**: Ajustar branding si era necesario
5. **Actualizar documentación**: Si tienes documentación personalizada

---

## 💡 Tips

- **No hay cambios funcionales**: Solo cambios de marca y visuales
- **Configuración compatible**: No necesitas cambiar tu `config.json`
- **Canales iguales**: Bots de Telegram, Discord, etc. funcionan igual
- **Workflows idénticos**: Los workflows existentes funcionan sin cambios

---

## 🆘 Soporte

Si tienes problemas durante la migración:

- **GitHub Issues**: [Reportar problema](https://github.com/sipeed/MakoClaw/issues)
- **Discord**: [Comunidad](https://discord.gg/V4sAZ9XWpN)
- **Discussions**: [Preguntas](https://github.com/sipeed/MakoClaw/discussions)

---

<div align="center">

**🦈 ¡Bienvenido a MakoClaw!**

El tiburón más rápido del océano de la IA.

</div>
