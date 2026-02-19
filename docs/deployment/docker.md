# Despliegue con Docker Desktop

## Objetivo
Esta guía explica cómo levantar KakoClaw dentro de Docker Desktop aprovechando el Dockerfile multi-stage y el docker-compose.yml ya incluidos en el repositorio.

## Requisitos
- Docker Desktop (Windows) o Docker Engine 24+ con Compose (Docker Compose v2) instalado.
- Al menos 1 GB de RAM y red activa para descargar dependencias la primera vez.
- La CLI de `docker compose` en el PATH.

## Cómo se construye la imagen
El Dockerfile compila el frontend en `/pkg/web/frontend`, lo incrusta vía go:embed y luego compila el binario `KakoClaw` con CGO disabled. La imagen final es Debian slim con el binario copiado a `/usr/local/bin/KakoClaw` y expone el puerto 18880.

## Preparar datos persistentes
1. Crea la carpeta de datos que montará el contenedor:
   ```powershell
   mkdir KakoClaw-data\.KakoClaw
   ```
2. Copia el archivo `config.example.json` (ya versionado y sin claves) dentro de ese volumen:
   ```powershell
   cp config.example.json KakoClaw-data\.KakoClaw\config.json
   ```
3. Esa copia será la configuración base: puedes guardarla en otro repositorio o zip para reutilizarla en otra máquina, porque el volumen `KakoClaw-data` guarda configuración y workspace y está ignorado por Git.

## Configuración base sin claves
- Edita `KakoClaw-data/.KakoClaw/config.json` para ajustar modelos, workspace o puertos si quieres, pero deja las claves en blanco.
- El archivo `config.example.json` ya contiene la misma estructura y es seguro subirlo a GitHub; úsalo como punto de partida en cada máquina nueva.

## Variables de entorno y secretos
1. Copia `docs/deployment/docker.env.example` al `.env` del repositorio raíz antes de ejecutar `docker compose`:
   ```powershell
   cp docs/deployment/docker.env.example .env
   ```
2. Rellena las variables sensibles, por ejemplo:
   - `KakoClaw_WEB_PASSWORD`: contraseña para la UI web (el usuario es `admin`).
   - `KakoClaw_PROVIDERS_ZHIPU_API_KEY`: tu API key de Zhipu u otro proveedor.
3. `docker-compose.yml` ya referencia estas variables con `${...}`, así que no es necesario editar el YAML; la carpeta `.env` y `KakoClaw-data` están en `.gitignore`.

## Iniciar en Docker Desktop
1. Construye la imagen y descarga dependencias nuevas:
   ```bash
   docker compose build --pull
   ```
2. Arranca el servicio en segundo plano:
   ```bash
   docker compose up -d
   ```
3. Sigue los logs para confirmar el arranque:
   ```bash
   docker compose logs -f
   ```

## Acceso y comprobación
- La UI web queda disponible en `http://localhost:18880` con usuario `admin` y la contraseña definida en `.env`.
- También puedes verificar el estado con `docker compose ps` o `curl http://localhost:18880/health`.

## Actualizar y detener
- Para detener y limpiar recursos conservando datos:
  ```bash
  docker compose down
  ```
- Para aplicar cambios de código o dependencias:
  ```bash
  docker compose up -d --build
  ```
- Si necesitas resetear el workspace, borra `KakoClaw-data/.KakoClaw/workspace` antes de arrancar.

## Compartir plantillas seguras
- `config.example.json` y `docs/deployment/docker.env.example` no contienen claves y se pueden commitear para que cada computadora copie el mismo punto de partida.
- Cuando cambies de equipo, clona el repositorio, copia el `config.example` al volumen `KakoClaw-data/.KakoClaw/config.json` y crea un `.env` a partir de `docs/deployment/docker.env.example` con los secretos locales.
- Nunca subas un `.env` real ni la carpeta `KakoClaw-data`, pero puedes mantener una versión zip de la carpeta sin claves para agilizar el setup.
