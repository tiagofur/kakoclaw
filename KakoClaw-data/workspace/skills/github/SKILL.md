# GitHub Skill

## Configuración

### 🔑 API Key de GitHub
**IMPORTANTE:** Debes agregar tu API key de GitHub Personal Access Token a continuación.

```
GITHUB_TOKEN="TU_API_KEY_AQUI"
```

**Cómo obtener tu API Key:**
1. Ve a https://github.com/settings/tokens
2. Clic en "Generate new token" → "Generate new token (classic)"
3. Selecciona los permisos necesarios:
   - `repo` - Para repositorios privados
   - `public_repo` - Para repositorios públicos
   - `issues` - Para issues y pull requests
   - `contents` - Para leer/escribir contenido de repositorios
4. Genera el token y cópialo (solo se muestra una vez)
5. Reemplaza "TU_API_KEY_AQUI" con el token

**⚠️ NUNCA compartas tu API Key públicamente ni la incluyas en commits.**

---

## Funciones Disponibles

### 📁 Repositorios

#### Listar Repositorios del Usuario
```bash
curl -s -H "Authorization: token $GITHUB_TOKEN" \
  https://api.github.com/user/repos
```

#### Crear un Nuevo Repositorio
```bash
curl -s -X POST -H "Authorization: token $GITHUB_TOKEN" \
  -H "Accept: application/vnd.github.v3+json" \
  -d '{"name":"nombre-repo","description":"Descripción","private":false}' \
  https://api.github.com/user/repos
```

#### Ver Información de un Repositorio
```bash
curl -s -H "Authorization: token $GITHUB_TOKEN" \
  https://api.github.com/repos/USERNAME/REPO_NAME
```

#### Eliminar un Repositorio
```bash
curl -s -X DELETE -H "Authorization: token $GITHUB_TOKEN" \
  https://api.github.com/repos/USERNAME/REPO_NAME
```

---

### 📋 Issues y Pull Requests

#### Crear un Issue
```bash
curl -s -X POST -H "Authorization: token $GITHUB_TOKEN" \
  -H "Accept: application/vnd.github.v3+json" \
  -d '{"title":"Título del issue","body":"Descripción detallada"}' \
  https://api.github.com/repos/USERNAME/REPO_NAME/issues
```

#### Listar Issues de un Repositorio
```bash
curl -s -H "Authorization: token $GITHUB_TOKEN" \
  https://api.github.com/repos/USERNAME/REPO_NAME/issues?state=open
```

#### Crear un Pull Request
```bash
curl -s -X POST -H "Authorization: token $GITHUB_TOKEN" \
  -H "Accept: application/vnd.github.v3+json" \
  -d '{"title":"Título del PR","head":"branch-feature","base":"main","body":"Descripción"}' \
  https://api.github.com/repos/USERNAME/REPO_NAME/pulls
```

---

### 📄 Contenido de Archivos

#### Leer el Contenido de un Archivo
```bash
curl -s -H "Authorization: token $GITHUB_TOKEN" \
  https://api.github.com/repos/USERNAME/REPO_NAME/contents/PATH/TO/FILE
```

#### Crear/Actualizar un Archivo
```bash
curl -s -X PUT -H "Authorization: token $GITHUB_TOKEN" \
  -H "Accept: application/vnd.github.v3+json" \
  -d '{"message":"Commit message","content":"BASE64_ENCODED_CONTENT","sha":"FILE_SHA"}' \
  https://api.github.com/repos/USERNAME/REPO_NAME/contents/PATH/TO/FILE
```

#### Eliminar un Archivo
```bash
curl -s -X DELETE -H "Authorization: token $GITHUB_TOKEN" \
  -H "Accept: application/vnd.github.v3+json" \
  -d '{"message":"Delete file","sha":"FILE_SHA"}' \
  https://api.github.com/repos/USERNAME/REPO_NAME/contents/PATH/TO/FILE
```

---

### 🔍 Búsqueda

#### Buscar Repositorios
```bash
curl -s -H "Authorization: token $GITHUB_TOKEN" \
  "https://api.github.com/search/repositories?q=language:python+stars:>100"
```

#### Buscar Código
```bash
curl -s -H "Authorization: token $GITHUB_TOKEN" \
  "https://api.github.com/search/code?q=filename:main.js+repo:USERNAME/REPO_NAME"
```

---

### 🌿 Branches y Commits

#### Listar Branches
```bash
curl -s -H "Authorization: token $GITHUB_TOKEN" \
  https://api.github.com/repos/USERNAME/REPO_NAME/branches
```

#### Crear un Branch
```bash
curl -s -X POST -H "Authorization: token $GITHUB_TOKEN" \
  -H "Accept: application/vnd.github.v3+json" \
  -d '{"ref":"refs/heads/nuevo-branch","sha":"COMMIT_SHA"}' \
  https://api.github.com/repos/USERNAME/REPO_NAME/git/refs
```

#### Ver Commits Recientes
```bash
curl -s -H "Authorization: token $GITHUB_TOKEN" \
  https://api.github.com/repos/USERNAME/REPO_NAME/commits
```

---

### 👤 Perfil de Usuario

#### Ver Perfil Propio
```bash
curl -s -H "Authorization: token $GITHUB_TOKEN" \
  https://api.github.com/user
```

#### Ver Perfil de Otro Usuario
```bash
curl -s -H "Authorization: token $GITHUB_TOKEN" \
  https://api.github.com/users/USERNAME
```

---

### 🏢 Organizaciones

#### Listar Organizaciones del Usuario
```bash
curl -s -H "Authorization: token $GITHUB_TOKEN" \
  https://api.github.com/user/orgs
```

#### Listar Repositorios de una Organización
```bash
curl -s -H "Authorization: token $GITHUB_TOKEN" \
  https://api.github.com/orgs/ORG_NAME/repos
```

---

### 📊 Estadísticas y Actividad

#### Ver Estadísticas del Repositorio
```bash
curl -s -H "Authorization: token $GITHUB_TOKEN" \
  https://api.github.com/repos/USERNAME/REPO_NAME/stats/contributors
```

#### Ver Activity Feed
```bash
curl -s -H "Authorization: token $GITHUB_TOKEN" \
  https://api.github.com/users/USERNAME/events
```

---

## Ejemplos de Uso

### Ejemplo 1: Listar mis repositorios
```bash
GITHUB_TOKEN="ghp_xxxxxxxxxxxx"
curl -s -H "Authorization: token $GITHUB_TOKEN" https://api.github.com/user/repos | jq '.[].name'
```

### Ejemplo 2: Crear un repositorio nuevo
```bash
GITHUB_TOKEN="ghp_xxxxxxxxxxxx"
curl -s -X POST -H "Authorization: token $GITHUB_TOKEN" \
  -H "Accept: application/vnd.github.v3+json" \
  -d '{"name":"mi-proyecto","description":"Mi nuevo proyecto","private":false}' \
  https://api.github.com/user/repos
```

### Ejemplo 3: Crear un issue
```bash
GITHUB_TOKEN="ghp_xxxxxxxxxxxx"
curl -s -X POST -H "Authorization: token $GITHUB_TOKEN" \
  -H "Accept: application/vnd.github.v3+json" \
  -d '{"title":"Bug encontrado","body":"El sistema no responde correctamente"}' \
  https://api.github.com/repos/usuario/mi-repo/issues
```

---

## Notas Importantes

1. **Rate Limiting:** GitHub API permite hasta 5000 requests/hora con autenticación.
2. **Permisos:** Asegúrate de que tu token tenga los permisos necesarios.
3. **Base64:** Para contenido de archivos, los datos deben estar codificados en Base64.
4. **SHA:** Para actualizar o eliminar archivos, necesitas el SHA actual del archivo.
5. **jq:** Usa `jq` para formatear el output JSON: `curl ... | jq '.'`

---

## Referencias

- [GitHub REST API Documentation](https://docs.github.com/en/rest)
- [Authentication](https://docs.github.com/en/rest/authentication)
- [Rate Limiting](https://docs.github.com/en/rest/overview/resources-in-the-rest-api#rate-limiting)
