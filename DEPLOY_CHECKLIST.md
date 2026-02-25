# Checklist de Despliegue MakoClaw

Usa esta lista para asegurar un despliegue exitoso en tu VPS Ubuntu.

## Pre-Despliegue

### En tu Máquina Local

- [ ] Todos los tests pasan: `go test ./...`
- [ ] El proyecto compila: `go build ./cmd/makoclaw`
- [ ] Tienes acceso SSH a tu VPS
- [ ] Has obtenido tu API key del proveedor LLM (OpenRouter, Anthropic, etc.)
- [ ] (Opcional) Tu dominio apunta a la IP del VPS

### Credenciales Necesarias

- [ ] API Key del proveedor LLM
- [ ] Contraseña segura para acceso web (genera con: `openssl rand -base64 32`)
- [ ] (Opcional) Token de bot de Telegram/Discord si usarás canales
- [ ] (Opcional) Email para Let's Encrypt si usarás SSL

## Despliegue en VPS

### 1. Preparación del Servidor

```bash
# Conectar al VPS
ssh usuario@tu-vps-ip

# Actualizar sistema
sudo apt-get update && sudo apt-get upgrade -y

# Clonar repositorio
git clone https://github.com/sipeed/makoclaw.git
cd makoclaw
```

- [ ] Conectado al VPS vía SSH
- [ ] Sistema actualizado
- [ ] Repositorio clonado

### 2. Instalación Automatizada

```bash
# Hacer ejecutable el script
chmod +x deploy.sh

# OPCIÓN A: Instalación completa (con dominio y SSL)
./deploy.sh --install-prereqs \
            --setup-nginx \
            --setup-ssl \
            --domain tu-dominio.com \
            --email tu@email.com \
            --build \
            --start

# OPCIÓN B: Instalación simple (sin dominio, solo localhost)
./deploy.sh --install-prereqs --build --start
```

- [ ] Script ejecutable
- [ ] Prerrequisitos instalados (Docker, Docker Compose)
- [ ] Archivo .env configurado con tus credenciales
- [ ] Imagen Docker construida
- [ ] Servicios iniciados

### 3. Configuración de .env

Durante el despliegue, el script pausará para que edites `.env`:

```bash
nano .env
```

**Configuración mínima obligatoria:**

```bash
# Seguridad
MAKOCLAW_WEB_PASSWORD=tu_contraseña_super_segura

# Proveedor LLM
MAKOCLAW_AGENTS_DEFAULTS_PROVIDER=openrouter
MAKOCLAW_AGENTS_DEFAULTS_MODEL=anthropic/claude-3.5-sonnet
MAKOCLAW_PROVIDERS_OPENROUTER_API_KEY=sk-or-v1-tu-api-key-aqui

# Puerto (si usas Nginx, dejar en 127.0.0.1)
MAKOCLAW_PORT=127.0.0.1:18880
```

- [ ] MAKOCLAW_WEB_PASSWORD configurado
- [ ] Proveedor LLM configurado
- [ ] API Key del proveedor configurada
- [ ] Puerto configurado correctamente

### 4. Verificación Post-Despliegue

```bash
# Verificar que el contenedor está corriendo
docker ps

# Ver logs
docker-compose -f docker-compose.prod.yml logs -f

# Probar conectividad local
curl http://localhost:18880/health

# Probar desde fuera (si configuraste dominio)
curl https://tu-dominio.com/health
```

- [ ] Contenedor corriendo (`docker ps` muestra makoclaw_prod)
- [ ] Logs sin errores críticos
- [ ] Health check responde OK
- [ ] Puedes acceder vía navegador

## Seguridad

### Firewall (UFW)

```bash
# Permitir SSH
sudo ufw allow 22/tcp

# Permitir HTTP/HTTPS si usas dominio
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp

# Habilitar firewall
sudo ufw enable

# Verificar estado
sudo ufw status
```

- [ ] Firewall configurado
- [ ] Solo puertos necesarios abiertos
- [ ] SSH accesible (¡importante!)

### SSL/TLS (si usas dominio)

```bash
# Verificar certificado
sudo certbot certificates

# Verificar renovación automática
sudo systemctl status certbot.timer
```

- [ ] Certificado SSL obtenido
- [ ] Renovación automática habilitada
- [ ] HTTPS funcionando

## Acceso Inicial

### Primera Conexión

1. Acceder a la interfaz web:
   - **Con dominio:** https://tu-dominio.com
   - **Sin dominio:** http://tu-vps-ip:18880

2. Credenciales de login:
   - **Usuario:** admin (o el que configuraste en MAKOCLAW_WEB_USERNAME)
   - **Contraseña:** La que configuraste en MAKOCLAW_WEB_PASSWORD

- [ ] Puedo acceder a la interfaz web
- [ ] Login exitoso
- [ ] Dashboard carga correctamente

### Prueba del Agente

1. Ir a la sección "Chat"
2. Enviar un mensaje de prueba: "Hola, ¿puedes ayudarme?"
3. Verificar que el agente responde

- [ ] Chat funciona
- [ ] El agente responde correctamente
- [ ] No hay errores de conexión con el proveedor LLM

## Mantenimiento

### Backups

```bash
# Crear directorio para backups
mkdir -p ~/backups

# Backup manual
tar -czf ~/backups/makoclaw-backup-$(date +%Y%m%d-%H%M%S).tar.gz MakoClaw-data/

# Configurar backup automático (cron)
crontab -e
# Agregar: 0 2 * * * cd ~/makoclaw && tar -czf ~/backups/makoclaw-$(date +\%Y\%m\%d).tar.gz MakoClaw-data/
```

- [ ] Directorio de backups creado
- [ ] Backup manual funciona
- [ ] Backup automático configurado (opcional)

### Monitoreo

```bash
# Ver logs en tiempo real
./deploy.sh --logs

# Ver uso de recursos
docker stats makoclaw_prod

# Ver salud del servicio
curl http://localhost:18880/health
```

- [ ] Logs accesibles
- [ ] Uso de recursos aceptable
- [ ] Health check responde

### Actualización

```bash
# Detener servicios
./deploy.sh --stop

# Actualizar código
git pull

# Reconstruir y reiniciar
./deploy.sh --build --start
```

- [ ] Procedimiento de actualización documentado
- [ ] Probado al menos una vez

## Troubleshooting Común

### El contenedor no inicia

```bash
# Ver logs
docker-compose -f docker-compose.prod.yml logs

# Verificar configuración
docker-compose -f docker-compose.prod.yml config

# Verificar .env
cat .env | grep -v "^#" | grep -v "^$"
```

### No puedo acceder desde fuera

```bash
# Verificar que el puerto está escuchando
sudo netstat -tlnp | grep 18880

# Verificar firewall
sudo ufw status

# Verificar Nginx (si aplica)
sudo nginx -t
sudo systemctl status nginx
```

### Errores con el proveedor LLM

```bash
# Verificar variables de entorno
docker-compose -f docker-compose.prod.yml exec makoclaw env | grep PROVIDER

# Probar conectividad
docker-compose -f docker-compose.prod.yml exec makoclaw curl -I https://api.openai.com
```

## Contacto y Soporte

- **GitHub Issues:** https://github.com/sipeed/makoclaw/issues
- **Documentación:** [DEPLOYMENT.md](DEPLOYMENT.md)

---

## Resumen de Comandos Útiles

```bash
# Ver estado
docker ps
docker-compose -f docker-compose.prod.yml ps

# Ver logs
./deploy.sh --logs
docker-compose -f docker-compose.prod.yml logs -f

# Reiniciar
./deploy.sh --restart

# Detener
./deploy.sh --stop

# Iniciar
./deploy.sh --start

# Backup
tar -czf backup.tar.gz MakoClaw-data/

# Actualizar
git pull && ./deploy.sh --build --start
```

---

**Fecha de despliegue:** _______________

**Notas adicionales:**

_______________________________________________

_______________________________________________

_______________________________________________
