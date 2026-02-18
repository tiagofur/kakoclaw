# Configuracion de Email

Esta guia explica como habilitar y configurar el envio de emails en PicoClaw para que el agente pueda enviarte reportes, resumenes y notificaciones por correo electronico.

## Tabla de Contenidos

- [Como Funciona](#como-funciona)
- [Requisitos](#requisitos)
- [Paso 1: Crear un App Password de Gmail](#paso-1-crear-un-app-password-de-gmail)
- [Paso 2: Configurar las Variables de Entorno](#paso-2-configurar-las-variables-de-entorno)
- [Paso 3: Habilitar el Email Tool](#paso-3-habilitar-el-email-tool)
- [Paso 4: Reiniciar y Verificar](#paso-4-reiniciar-y-verificar)
- [Configuracion con config.json](#configuracion-con-configjson)
- [Uso de Otros Proveedores SMTP](#uso-de-otros-proveedores-smtp)
- [Referencia de Variables](#referencia-de-variables)
- [Solucion de Problemas](#solucion-de-problemas)

## Como Funciona

PicoClaw incluye una herramienta integrada llamada `send_email_report` que permite al agente enviar correos electronicos directamente usando SMTP. No necesitas instalar software adicional como `sendmail` o `msmtp`; el envio se realiza desde el propio binario de Go usando el paquete `net/smtp`.

El agente puede usar esta herramienta para:

- Enviar resumenes semanales de tareas completadas
- Notificar resultados de tareas programadas (cron)
- Enviar alertas o informacion importante
- Generar y enviar reportes bajo demanda

## Requisitos

- PicoClaw instalado y funcionando
- Una cuenta de correo con acceso SMTP (Gmail, Outlook, etc.)
- Para Gmail: verificacion en 2 pasos activada y un App Password generado

## Paso 1: Crear un App Password de Gmail

Gmail no permite autenticacion SMTP con tu contrasena normal. Necesitas generar una **App Password** (contrasena de aplicacion).

### 1.1 Activar verificacion en 2 pasos

1. Ve a [https://myaccount.google.com/security](https://myaccount.google.com/security)
2. En la seccion "Como inicias sesion en Google", haz clic en **Verificacion en 2 pasos**
3. Sigue las instrucciones para activarla (necesitaras tu telefono)

### 1.2 Generar el App Password

1. Ve a [https://myaccount.google.com/apppasswords](https://myaccount.google.com/apppasswords)
2. En "Seleccionar aplicacion", escribe un nombre descriptivo (ej: `PicoClaw`)
3. Haz clic en **Crear**
4. Google te mostrara un codigo de 16 caracteres, por ejemplo: `abcd efgh ijkl mnop`
5. **Copia este codigo** (sin espacios): `abcdefghijklmnop`
6. Guardalo en un lugar seguro; no podras verlo de nuevo

> **Importante:** Este App Password es el unico metodo soportado para Gmail. Las contrasenas normales seran rechazadas con un error de autenticacion.

## Paso 2: Configurar las Variables de Entorno

### Despliegue con Docker (recomendado)

Edita tu archivo `.env` en la raiz del proyecto:

```bash
# Email Tools
PICOCLAW_TOOLS_EMAIL_USERNAME=tu_correo@gmail.com
PICOCLAW_TOOLS_EMAIL_PASSWORD=abcdefghijklmnop
PICOCLAW_TOOLS_EMAIL_FROM=PicoClaw <tu_correo@gmail.com>
PICOCLAW_TOOLS_EMAIL_TO=destinatario@ejemplo.com
```

| Variable | Descripcion | Ejemplo |
|----------|-------------|---------|
| `EMAIL_USERNAME` | Correo usado para autenticacion SMTP | `mi_correo@gmail.com` |
| `EMAIL_PASSWORD` | App Password generado en el paso anterior | `abcdefghijklmnop` |
| `EMAIL_FROM` | Nombre y correo que aparece como remitente | `PicoClaw <mi_correo@gmail.com>` |
| `EMAIL_TO` | Destinatario por defecto de los emails | `yo@ejemplo.com` |

> **Nota:** El archivo `.env` ya esta incluido en `.gitignore` para evitar que las credenciales se suban al repositorio.

### Despliegue sin Docker

Exporta las variables en tu shell o agregalas a `~/.bashrc` / `~/.zshrc`:

```bash
export PICOCLAW_TOOLS_EMAIL_ENABLED=true
export PICOCLAW_TOOLS_EMAIL_HOST=smtp.gmail.com
export PICOCLAW_TOOLS_EMAIL_PORT=587
export PICOCLAW_TOOLS_EMAIL_USERNAME=tu_correo@gmail.com
export PICOCLAW_TOOLS_EMAIL_PASSWORD=abcdefghijklmnop
export PICOCLAW_TOOLS_EMAIL_FROM="PicoClaw <tu_correo@gmail.com>"
export PICOCLAW_TOOLS_EMAIL_TO=destinatario@ejemplo.com
```

## Paso 3: Habilitar el Email Tool

### Con Docker Compose

En `docker-compose.yml`, asegurate de que la variable `PICOCLAW_TOOLS_EMAIL_ENABLED` este en `"true"`:

```yaml
services:
  picoclaw:
    environment:
      # ... otras variables ...
      PICOCLAW_TOOLS_EMAIL_ENABLED: "true"
      PICOCLAW_TOOLS_EMAIL_HOST: "smtp.gmail.com"
      PICOCLAW_TOOLS_EMAIL_PORT: "587"
      PICOCLAW_TOOLS_EMAIL_USERNAME: "${PICOCLAW_TOOLS_EMAIL_USERNAME}"
      PICOCLAW_TOOLS_EMAIL_PASSWORD: "${PICOCLAW_TOOLS_EMAIL_PASSWORD}"
      PICOCLAW_TOOLS_EMAIL_FROM: "${PICOCLAW_TOOLS_EMAIL_FROM}"
      PICOCLAW_TOOLS_EMAIL_TO: "${PICOCLAW_TOOLS_EMAIL_TO}"
```

### Con config.json

Alternativamente, agrega la seccion `email` dentro de `tools` en tu `config.json`:

```json
{
  "tools": {
    "email": {
      "enabled": true,
      "host": "smtp.gmail.com",
      "port": 587,
      "username": "tu_correo@gmail.com",
      "password": "abcdefghijklmnop",
      "from": "PicoClaw <tu_correo@gmail.com>",
      "to": "destinatario@ejemplo.com"
    }
  }
}
```

## Paso 4: Reiniciar y Verificar

### Con Docker

```bash
# Reconstruir y reiniciar el contenedor
docker compose down && docker compose up -d --build
```

### Sin Docker

```bash
# Reiniciar el agente o gateway
picoclaw gateway
```

### Probar el envio

Una vez reiniciado, puedes pedirle al agente que envie un email de prueba:

```
picoclaw agent -m "Enviale un email de prueba a mi correo con el asunto 'Test PicoClaw'"
```

O desde cualquier canal conectado (Telegram, Discord, web), simplemente escribe:

```
Enviame un email de prueba para verificar que funciona el correo
```

El agente usara la herramienta `send_email_report` automaticamente.

## Configuracion con config.json

Si prefieres no usar variables de entorno, puedes agregar toda la configuracion directamente en `~/.picoclaw/config.json`. Aqui un ejemplo completo:

```json
{
  "agents": {
    "defaults": {
      "provider": "zhipu",
      "model": "glm-4.7"
    }
  },
  "tools": {
    "web": {
      "search": {
        "api_key": "TU_BRAVE_API_KEY"
      }
    },
    "email": {
      "enabled": true,
      "host": "smtp.gmail.com",
      "port": 587,
      "username": "tu_correo@gmail.com",
      "password": "tu_app_password",
      "from": "PicoClaw <tu_correo@gmail.com>",
      "to": "destinatario@ejemplo.com"
    }
  }
}
```

> **Nota:** Las variables de entorno tienen prioridad sobre los valores en `config.json`.

## Uso de Otros Proveedores SMTP

Si no usas Gmail, puedes configurar cualquier servidor SMTP que soporte autenticacion `PLAIN` sobre TLS (puerto 587).

### Outlook / Hotmail

```bash
PICOCLAW_TOOLS_EMAIL_HOST=smtp.office365.com
PICOCLAW_TOOLS_EMAIL_PORT=587
PICOCLAW_TOOLS_EMAIL_USERNAME=tu_correo@outlook.com
PICOCLAW_TOOLS_EMAIL_PASSWORD=tu_contrasena
```

### Yahoo Mail

```bash
PICOCLAW_TOOLS_EMAIL_HOST=smtp.mail.yahoo.com
PICOCLAW_TOOLS_EMAIL_PORT=587
PICOCLAW_TOOLS_EMAIL_USERNAME=tu_correo@yahoo.com
PICOCLAW_TOOLS_EMAIL_PASSWORD=tu_app_password
```

### SendGrid

```bash
PICOCLAW_TOOLS_EMAIL_HOST=smtp.sendgrid.net
PICOCLAW_TOOLS_EMAIL_PORT=587
PICOCLAW_TOOLS_EMAIL_USERNAME=apikey
PICOCLAW_TOOLS_EMAIL_PASSWORD=SG.tu_api_key
```

### Servidor SMTP propio

```bash
PICOCLAW_TOOLS_EMAIL_HOST=mail.tudominio.com
PICOCLAW_TOOLS_EMAIL_PORT=587
PICOCLAW_TOOLS_EMAIL_USERNAME=usuario@tudominio.com
PICOCLAW_TOOLS_EMAIL_PASSWORD=tu_contrasena
```

## Referencia de Variables

| Variable de Entorno | Campo JSON | Tipo | Default | Descripcion |
|---------------------|------------|------|---------|-------------|
| `PICOCLAW_TOOLS_EMAIL_ENABLED` | `tools.email.enabled` | bool | `false` | Activa o desactiva la herramienta de email |
| `PICOCLAW_TOOLS_EMAIL_HOST` | `tools.email.host` | string | `smtp.gmail.com` | Servidor SMTP |
| `PICOCLAW_TOOLS_EMAIL_PORT` | `tools.email.port` | int | `587` | Puerto SMTP (587 para STARTTLS) |
| `PICOCLAW_TOOLS_EMAIL_USERNAME` | `tools.email.username` | string | `""` | Usuario para autenticacion SMTP |
| `PICOCLAW_TOOLS_EMAIL_PASSWORD` | `tools.email.password` | string | `""` | Contrasena o App Password |
| `PICOCLAW_TOOLS_EMAIL_FROM` | `tools.email.from` | string | `""` | Remitente (formato: `Nombre <correo>`) |
| `PICOCLAW_TOOLS_EMAIL_TO` | `tools.email.to` | string | `""` | Destinatario por defecto |

## Solucion de Problemas

### Error: "email tool is disabled in configuration"

La herramienta no esta habilitada. Verifica que:

```bash
# Variable de entorno
PICOCLAW_TOOLS_EMAIL_ENABLED=true

# O en config.json
"tools": { "email": { "enabled": true } }
```

### Error: "535 Authentication failed" o "Username and Password not accepted"

- **Gmail:** Asegurate de usar un App Password, no tu contrasena normal
- Verifica que la verificacion en 2 pasos este activa en tu cuenta de Google
- Confirma que el App Password no tiene espacios
- Si regeneraste el App Password, actualiza el valor en `.env`

### Error: "534 Application-specific password required"

Estas usando tu contrasena normal de Gmail. Genera un App Password siguiendo el [Paso 1](#paso-1-crear-un-app-password-de-gmail).

### Error: "connection refused" o "dial tcp: lookup smtp.gmail.com"

- Verifica que el contenedor tenga acceso a internet
- Revisa que el host y puerto SMTP sean correctos
- Si usas un proxy o firewall, asegurate de que el puerto 587 este abierto

### Error: "no recipient specified"

No se especifico destinatario. Configura `PICOCLAW_TOOLS_EMAIL_TO` o pasa el parametro `to` al usar la herramienta.

### Los emails llegan a Spam

- Configura el campo `FROM` con el formato correcto: `PicoClaw <tu_correo@gmail.com>`
- Evita palabras como "test" o "prueba" en el asunto durante las primeras pruebas
- Marca los emails como "No es spam" en tu bandeja para entrenar el filtro

---

Para mas informacion sobre herramientas disponibles, consulta la [Tools API](../api-reference/tools.md).
Para configurar tareas programadas que envien emails automaticamente, ver [Tareas Programadas](./cron-jobs.md).
