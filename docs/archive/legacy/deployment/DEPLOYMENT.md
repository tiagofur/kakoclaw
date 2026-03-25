# MakoClaw - Guía de Despliegue en VPS Ubuntu

Esta guía te ayudará a desplegar MakoClaw en un VPS con Ubuntu usando Docker.

## Requisitos Previos

- VPS con Ubuntu 20.04 o superior
- Mínimo 2GB RAM, 2 vCPUs recomendados
- 10GB de espacio en disco
- Acceso SSH al servidor
- (Opcional) Dominio apuntando a la IP del VPS

## Despliegue Rápido

### 1. Clonar el Repositorio

```bash
# En tu VPS
git clone https://github.com/sipeed/makoclaw.git
cd makoclaw
```

### 2. Instalación Automática (Recomendado)

El script `deploy.sh` automatiza todo el proceso:

```bash
# Hacer el script ejecutable
chmod +x deploy.sh

# Instalación completa con Nginx y SSL
./deploy.sh --install-prereqs \
            --setup-nginx \
            --setup-ssl \
            --domain tu-dominio.com \
            --email tu@email.com \
            --build \
            --start
```

**Nota:** Durante la ejecución, el script te pedirá que edites el archivo `.env`. Ver sección de configuración abajo.

### 3. Configuración Manual (Alternativa)

Si prefieres hacerlo paso a paso:

#### a. Instalar Docker y Docker Compose

```bash
# Instalar Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh
sudo usermod -aG docker $USER

# Instalar Docker Compose
sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" \
     -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose

# Verificar instalación
docker --version
docker-compose --version

# Cerrar sesión y volver a entrar para aplicar el grupo docker
```

#### b. Configurar Variables de Entorno

```bash
# Copiar el archivo de ejemplo
cp .env.example .env

# Editar con tu editor favorito
nano .env
```

**Configuración mínima requerida en `.env`:**

```bash
# Contraseña de acceso web (¡CAMBIAR!)
MAKOCLAW_WEB_PASSWORD=tu_contraseña_segura_aqui

# Proveedor LLM (elegir uno)
MAKOCLAW_AGENTS_DEFAULTS_PROVIDER=openrouter
MAKOCLAW_AGENTS_DEFAULTS_MODEL=anthropic/claude-3.5-sonnet

# API Key del proveedor elegido
MAKOCLAW_PROVIDERS_OPENROUTER_API_KEY=tu_api_key_aqui
```

#### c. Construir y Ejecutar

```bash
# Construir la imagen
docker-compose -f docker-compose.prod.yml build

# Iniciar los servicios
docker-compose -f docker-compose.prod.yml up -d

# Verificar que esté corriendo
docker-compose -f docker-compose.prod.yml ps
docker-compose -f docker-compose.prod.yml logs -f
```

#### d. (Opcional) Configurar Nginx como Reverse Proxy

```bash
# Instalar Nginx
sudo apt-get install -y nginx

# Crear configuración
sudo nano /etc/nginx/sites-available/makoclaw
```

Contenido del archivo:

```nginx
server {
    listen 80;
    listen [::]:80;
    server_name tu-dominio.com;

    client_max_body_size 100M;

    location / {
        proxy_pass http://127.0.0.1:18880;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_cache_bypass $http_upgrade;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;

        # WebSocket support
        proxy_read_timeout 86400;
    }
}
```

```bash
# Habilitar el sitio
sudo ln -s /etc/nginx/sites-available/makoclaw /etc/nginx/sites-enabled/

# Verificar configuración
sudo nginx -t

# Recargar Nginx
sudo systemctl reload nginx
```

#### e. (Opcional) Configurar SSL con Let's Encrypt

```bash
# Instalar Certbot
sudo apt-get install -y certbot python3-certbot-nginx

# Obtener certificado
sudo certbot --nginx -d tu-dominio.com

# Habilitar renovación automática
sudo systemctl enable certbot.timer
```

## Configuración Detallada

### Proveedores LLM Soportados

#### OpenRouter (Recomendado - Acceso a múltiples modelos)

```bash
MAKOCLAW_AGENTS_DEFAULTS_PROVIDER=openrouter
MAKOCLAW_AGENTS_DEFAULTS_MODEL=anthropic/claude-3.5-sonnet
MAKOCLAW_PROVIDERS_OPENROUTER_API_KEY=sk-or-v1-...
```

Obtener API key en: https://openrouter.ai/keys

#### Anthropic Claude

```bash
MAKOCLAW_AGENTS_DEFAULTS_PROVIDER=anthropic
MAKOCLAW_AGENTS_DEFAULTS_MODEL=claude-3-sonnet-20240229
MAKOCLAW_PROVIDERS_ANTHROPIC_API_KEY=sk-ant-...
```

#### OpenAI

```bash
MAKOCLAW_AGENTS_DEFAULTS_PROVIDER=openai
MAKOCLAW_AGENTS_DEFAULTS_MODEL=gpt-4-turbo-preview
MAKOCLAW_PROVIDERS_OPENAI_API_KEY=sk-...
```

#### Groq (Rápido y con tier gratuito)

```bash
MAKOCLAW_AGENTS_DEFAULTS_PROVIDER=groq
MAKOCLAW_AGENTS_DEFAULTS_MODEL=llama2-70b-4096
MAKOCLAW_PROVIDERS_GROQ_API_KEY=gsk_...
```

#### Ollama (Auto-hospedado, gratis)

```bash
MAKOCLAW_AGENTS_DEFAULTS_PROVIDER=ollama
MAKOCLAW_AGENTS_DEFAULTS_MODEL=llama3
MAKOCLAW_PROVIDERS_OLLAMA_API_BASE=http://host.docker.internal:11434
```

**Nota:** Para Ollama, necesitas instalarlo en el servidor:
```bash
curl -fsSL https://ollama.com/install.sh | sh
ollama pull llama3
```

### Configuración de Canales (Opcional)

#### Telegram Bot

```bash
MAKOCLAW_CHANNELS_TELEGRAM_ENABLED=true
MAKOCLAW_CHANNELS_TELEGRAM_TOKEN=123456:ABC-DEF...
MAKOCLAW_CHANNELS_TELEGRAM_ALLOW_FROM=123456789,987654321
```

#### Discord Bot

```bash
MAKOCLAW_CHANNELS_DISCORD_ENABLED=true
MAKOCLAW_CHANNELS_DISCORD_TOKEN=MTk...
MAKOCLAW_CHANNELS_DISCORD_ALLOW_FROM=123456789,987654321
```

## Comandos Útiles

### Gestión del Contenedor

```bash
# Ver logs en tiempo real
docker-compose -f docker-compose.prod.yml logs -f

# Ver logs de las últimas 100 líneas
docker-compose -f docker-compose.prod.yml logs --tail=100

# Reiniciar servicios
docker-compose -f docker-compose.prod.yml restart

# Detener servicios
docker-compose -f docker-compose.prod.yml down

# Detener y eliminar volúmenes (¡CUIDADO! Borra datos)
docker-compose -f docker-compose.prod.yml down -v

# Ver estado de los servicios
docker-compose -f docker-compose.prod.yml ps

# Acceder a la shell del contenedor
docker-compose -f docker-compose.prod.yml exec makoclaw bash
```

### Actualización

```bash
# Detener servicios
docker-compose -f docker-compose.prod.yml down

# Actualizar código
git pull

# Reconstruir imagen
docker-compose -f docker-compose.prod.yml build --no-cache

# Iniciar servicios
docker-compose -f docker-compose.prod.yml up -d
```

### Backup y Restauración

```bash
# Crear backup de datos
tar -czf makoclaw-backup-$(date +%Y%m%d).tar.gz MakoClaw-data/

# Restaurar desde backup
tar -xzf makoclaw-backup-YYYYMMDD.tar.gz
```

### Monitoreo

```bash
# Ver uso de recursos
docker stats

# Ver salud del contenedor
docker inspect makoclaw_prod | grep -A 10 Health

# Verificar conectividad
curl http://localhost:18880/health
```

## Seguridad

### Recomendaciones

1. **Firewall**: Configura UFW para permitir solo los puertos necesarios:

```bash
sudo ufw allow 22/tcp    # SSH
sudo ufw allow 80/tcp    # HTTP
sudo ufw allow 443/tcp   # HTTPS
sudo ufw enable
```

2. **Contraseña Fuerte**: Genera una contraseña segura:

```bash
openssl rand -base64 32
```

3. **Actualizaciones**: Mantén el sistema actualizado:

```bash
sudo apt-get update && sudo apt-get upgrade -y
```

4. **SSL/TLS**: Usa siempre HTTPS en producción con Let's Encrypt.

5. **Backup Regular**: Configura backups automáticos:

```bash
# Crear cron job para backup diario
crontab -e

# Agregar línea:
0 2 * * * cd /ruta/a/makoclaw && tar -czf backup-$(date +\%Y\%m\%d).tar.gz MakoClaw-data/
```

## Troubleshooting

### El contenedor no inicia

```bash
# Ver logs detallados
docker-compose -f docker-compose.prod.yml logs

# Verificar configuración
docker-compose -f docker-compose.prod.yml config

# Verificar permisos de archivos
ls -la MakoClaw-data/
```

### Error de conexión con el proveedor LLM

```bash
# Verificar variables de entorno
docker-compose -f docker-compose.prod.yml exec makoclaw env | grep MAKOCLAW

# Probar conectividad desde el contenedor
docker-compose -f docker-compose.prod.yml exec makoclaw curl -I https://api.openai.com
```

### Problemas de memoria

Ajusta los límites en `docker-compose.prod.yml`:

```yaml
deploy:
  resources:
    limits:
      memory: 4G  # Aumentar si es necesario
```

### Puerto ya en uso

```bash
# Ver qué proceso usa el puerto 18880
sudo lsof -i :18880

# Cambiar puerto en .env
MAKOCLAW_PORT=8080:18880
```

## Soporte

- **Issues**: https://github.com/sipeed/makoclaw/issues
- **Documentación**: https://github.com/sipeed/makoclaw
- **Discusiones**: https://github.com/sipeed/makoclaw/discussions

## Licencia

Ver archivo LICENSE en el repositorio.
