# 🐸 KakoClaw - Sistema de Leads para Rediseño Web

**Estado: Parcialmente funcional**

---

## ✅ Lo que KakoClaw PUEDE hacer AHORA

| Función | Estado | Descripción |
|---------|--------|-------------|
| **web_fetch** | ✅ Funcional | Extrae contenido HTML de cualquier URL |
| **Análisis web** | ✅ Funcional | Detecta tecnologías, problemas, califica 1-10 |
| **Generar propuestas** | ✅ Funcional | Crea propuestas personalizadas para cada negocio |
| **task_manager** | ✅ Funcional | Crea tareas de seguimiento en el sistema |
| **send_email_report** | ✅ Funcional | Envía reportes por email (kakoclaw@gmail.com → tiagofur@gmail.com) |
| **Ejecutar scripts** | ✅ Funcional | Puede ejecutar Python y generar archivos |
| **spawn** | ✅ Funcional | Ejecuta tareas en background |
| **query_knowledge** | ✅ Funcional | Busca en documentos subidos |

---

## ❌ Limitaciones Actuales

| Función | Estado | Razón | Solución |
|---------|--------|-------|----------|
| **web_search** | ⚠️ API Key | `BRAVE_API_KEY not configured` | Configurar API key o usar fuente alternativa |
| **Google Maps** | ❌ Bloqueado | Protección anti-scraping, requiere API | Búsqueda manual por el usuario |
| **Cron jobs** | ⏳ No disponible | No hay cron en este entorno | Usar `spawn` para background tasks |

---

## 🎯 Workflow Funcional (Recomendado)

```
USUARIO                              KakoClaw
─────────────────────────────────────────────────────────────────────
1. Buscar en Google Maps         →
2. Extraer: nombre, web, teléfono→
3. Pasar lista a KakoClaw        →  →  Recibir datos
                                        ↓
                                 →  web_fetch(url) x N
                                        ↓
                                 →  Analizar sitios
                                        ↓
                                 →  Calificar 1-10
                                        ↓
                                 →  Generar propuestas
                                        ↓
                                 →  Crear tasks
                                        ↓
                                 →  Enviar email
                                        ↓
                                 →  ← CSV + Reporte
```

---

## 📋 Cómo Usar el Sistema

### Paso 1: Buscar Negocios (Manual)

1. Ve a Google Maps
2. Busca: "restaurantes 66035 Monterrey" o "negocios 66036"
3. Extrae de cada resultado:
   - Nombre
   - Dirección
   - Teléfono
   - Sitio web (si tiene)

### Paso 2: Enviar Datos a KakoClaw

**Opción A - Formato simple (chat):**
```
Analiza estos sitios web:
- Restaurante Los Abuelos: www.losabuelos.com
- Despacho García: www.abogadosgarcia.net
- TechStore: www.techstore.mx
```

**Opción B - Formato CSV:**
```csv
nombre,tipo,direccion,telefono,email,web
Restaurante Los Abuelos,Restaurante,Av. Lincoln 123,81-1234-5678,,www.losabuelos.com
Despacho García,Legal,Calle 5 de Mayo 456,81-8765-4321,,www.abogadosgarcia.net
```

### Paso 3: KakoClaw Procesa

KakoClaw automáticamente hará:
1. ✅ Fetch HTML de cada sitio
2. ✅ Analizar tecnologías y problemas
3. ✅ Calificar (1-10)
4. ✅ Determinar prioridad
5. ✅ Generar propuestas
6. ✅ Crear tasks
7. ✅ Enviar email con reporte

---

## 📁 Archivos Disponibles

| Archivo | Descripción |
|---------|-------------|
| `sistema_automatizado.py` | Motor de análisis web |
| `KakoClaw_workflow.py` | Demo del flujo completo |
| `web_analyzer.py` | Analizador de calidad web |
| `empresas_66036.csv` | Base de datos de ejemplo |
| `demo_analisis.md` | Demo de 10 empresas |

---

## 🚀 Ejemplo de Uso Rápido

### Ejecutar Demo:
```bash
cd documentos
python3 sistema_automatizado.py
```

### Ejecutar Workflow Demo:
```bash
cd documentos
python3 KakoClaw_workflow.py
```

---

## 💡 Casos de Uso

### Caso 1: Análisis de 1 sitio web
```
Usuario: "Analiza www.ejemplo.com"

KakoClaw:
→ web_fetch("www.ejemplo.com")
→ Analizar HTML
→ Reportar calificación y problemas
```

### Caso 2: Análisis de múltiples sitios
```
Usuario: "Analiza estos sitios:
- www.restaurante.com
- www.abogados.com
- www.tienda.com"

KakoClaw:
→ web_fetch() x3
→ Analizar cada uno
→ Generar CSV con resultados
→ Enviar email con resumen
```

### Caso 3: Flujo completo con tasks
```
Usuario: "Analiza y crea tareas de seguimiento"

KakoClaw:
→ web_fetch() → Analizar → Calificar
→ task_manager(create) para leads urgentes
→ send_email_report con reporte completo
```

---

## 📊 Resultados que Obtendrás

### Por cada negocio:
- 📊 Calificación (1-10)
- ⚠️ Problemas detectados
- 🚀 Oportunidades de mejora
- 🎯 Prioridad (URGENTE / ALTA / MEDIA / BAJA)
- 📝 Propuesta personalizada

### Archivos generados:
- `leads_generados.csv` - Base de datos de leads
- `reporte.md` - Reporte en Markdown
- Propuestas individuales por negocio

### Tareas:
- Tasks creadas en el sistema para seguimiento

### Notificaciones:
- Email con resumen al destinatario configurado

---

## 🔧 Configuración de Email

**Email de envío:** kakoclaw@gmail.com
**Email destino:** tiagofur@gmail.com

*Se puede cambiar configurando el sistema.*

---

## 📞 Ejemplo de Interacción

```
Usuario: KakoClaw, analiza estos 3 sitios:
- www.restaurante-losabuelos.com
- www.abogados-garcia.net
- www.techstore.mx
- Crea tareas de seguimiento para los urgentes
- Envíame un reporte por email

KakoClaw:
✅ Analizando www.restaurante-losabuelos.com...
   Calificación: 3/10 - URGENTE
   Problemas: Flash, no responsive, HTTP

✅ Analizando www.abogados-garcia.net...
   Calificación: 4/10 - ALTA
   Problemas: HTML4, frameset, no responsive

✅ Analizando www.techstore.mx...
   Calificación: 7/10 - MEDIA
   Problemas: Falta meta description

✅ Creando tareas de seguimiento...
   → Task creada: Contactar Restaurante Los Abuelos (URGENTE)
   → Task creada: Contactar Despacho García (ALTA)

✅ Enviando reporte por email...
   → Email enviado a tiagofur@gmail.com

📄 Archivos generados:
   - documentos/leads.csv
   - documentos/reporte.md
   - documentos/propuestas/
```

---

## 🎯 Próximos Pasos

1. **Para empezar YA:**
   - Busca 3-5 sitios web reales
   - Pásamelos para analizar
   - Revisa los resultados

2. **Para automatizar completamente:**
   - Configurar BRAVE_API_KEY para web_search
   - O usar un servicio externo de scraping
   - Implementar programación de tareas (spawn)

3. **Para escalar:**
   - Crear script que procese listas más grandes
   - Integrar con Google Places API (requiere API key)
   - Automatizar envío de propuestas

---

## ❓ Preguntas Frecuentes

**Q: ¿Por qué no puedes scrapear Google Maps directamente?**
A: Google tiene protecciones anti-scraping. Requiere API de pago y autenticación.

**Q: ¿Cuántos sitios puedes analizar a la vez?**
A: No hay límite técnico, pero para empezar recomiendo 5-10 para validar el proceso.

**Q: ¿Puedo cambiar el email de destino?**
A: Sí, solo dímelo y configuro el sistema.

**Q: ¿Qué tan precisa es la calificación?**
A: Basada en análisis técnico de HTML/CSS. No perfecto, pero muy útil para priorizar.

---

## 📞 ¿Listo para Empezar?

Envíame una lista de sitios web para analizar y generar leads.

Ejemplo:
```
Analiza estos sitios:
- www.tienda-local.com
- www.restaurante-casual.com
- www.dentista-tradicion.net
```
