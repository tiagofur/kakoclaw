# Sistema de Análisis de Páginas Web para Servicios de Rediseño

## 📋 Descripción

Este sistema automatiza el proceso de buscar pequeñas empresas en el área de C.P. 66035 y 66036 (Monterrey, Nuevo León), analizar sus sitios web e identificar oportunidades para ofrecer servicios de remodelación de páginas web.

## 🗂️ Estructura de Archivos

```
documentos/
├── empresas_66036.csv          # Base de datos de empresas analizadas
├── web_analyzer.py            # Script de análisis automatizado
├── demo_analisis.md           # Demo con ejemplos de análisis
└── README.md                  # Este archivo
```

## 🚀 Flujo de Trabajo

### Paso 1: Recolección de Empresas

**Opción A: Manual (recomendada actualmente)**
1. Visitar Google Maps
2. Buscar empresas en C.P. 66035-66036
3. Extraer datos manualmente:
   - Nombre del negocio
   - Dirección completa
   - Teléfono
   - Sitio web (si tiene)

**Opción B: Automatizada (en desarrollo)**
- Usar script para extraer datos de directorios de empresas
- Nota: Requiere configuración de API o servicio de scraping

### Paso 2: Análisis de Sitios Web

El script `web_analyzer.py` analiza automáticamente:

#### Tecnologías Detectadas:
- **Frameworks CSS**: Bootstrap, Tailwind, Bulma, Foundation
- **Frameworks JS**: React, Vue, Angular, jQuery
- **CMS**: WordPress, Drupal, Joomla
- **Plataformas**: Wix, Squarespace, Shopify

#### Indicadores de Calidad:
- ✅ SSL/HTTPS
- ✅ Responsividad (viewport)
- ✅ Framework moderno
- ✅ SEO básico (título, meta description)
- ✅ Formulario de contacto
- ✅ Integración con redes sociales
- ❌ Flash (obsoleto)
- ❌ HTML4 (obsoleto)
- ❌ Tablas para maquetación (obsoleto)
- ❌ Joomla versiones viejas

#### Sistema de Puntuación:
- **10/10**: Excelente - Sitio moderno, bien mantenido
- **8-9/10**: Muy bueno - Solo requiere ajustes menores
- **6-7/10**: Bueno - Puede tener mejoras opcionales
- **4-5/10**: Regular - Beneficia de rediseño
- **2-3/10**: Antiguo - Requiere rediseño
- **1/10**: Crítico - Urgente rediseño necesario

### Paso 3: Generación de Reportes

El sistema genera:

1. **Archivo CSV** con todos los datos de empresas analizadas
2. **Reporte resumido** con:
   - Total de empresas analizadas
   - Porcentaje con sitio web
   - Empresas con prioridad alta/media/baja
   - Calidad promedio

### Paso 4: Acción Comercial

Basado en el análisis, priorizar contactos:

| Prioridad | Score | Acción |
|-----------|-------|--------|
| 🔴 Alta | 1-3 | Contactar urgente, propuesta de rediseño completo |
| 🟡 Media | 4-5 | Contactar, propuesta de mejoras específicas |
| 🟢 Baja | 6-10 | Seguimiento opcional, mantenimiento |

## 💻 Uso del Script

### Instalación de Dependencias

```bash
pip3 install requests beautifulsoup4 lxml
```

### Ejecutar Análisis

```python
from web_analyzer import WebAnalyzer, analyze_companies_batch

# Crear lista de empresas
companies = [
    {
        'Nombre': 'Restaurante El Rincón',
        'Empresa': 'Restaurantes El Rincón S.A.',
        'Direccion': 'Av. Benito Juárez 123, 66036 Monterrey, NL',
        'Telefono': '(81) 1234-5678',
        'Sitio_Web': 'https://restaurante-elincon.com'
    },
    # ... más empresas
]

# Analizar
results = analyze_companies_batch(companies)

# Generar reporte
analyzer = WebAnalyzer()
report = analyzer.generate_report(results)
print(report)
```

### Ejecutar Directamente

```bash
python3 web_analyzer.py
```

## 📊 Formato del CSV

| Columna | Descripción |
|---------|-------------|
| Nombre | Nombre del negocio visible al cliente |
| Empresa | Razón social legal |
| Direccion | Dirección física completa |
| Telefono | Número de teléfono |
| Sitio_Web | URL del sitio web |
| Tecnologias | Tecnologías detectadas (separadas por coma) |
| Estado_Diseño | Crítico/Antiguo/Regular/Bueno/Excelente |
| Calidad_Diseño | Puntuación 1-10 |
| Observaciones | Detalles y oportunidades de mejora |
| Fecha_Analisis | Fecha en que se realizó el análisis |
| Añadido_CSV | Si/No - Confirmación de guardado |

## 🎯 Estrategia de Ventas

### Mensajes según Calidad del Sitio:

#### Para sitios Críticos (1-2/10):
```
"Hola [Nombre],

Noté que su sitio web usa tecnologías antiguas como Flash que ya no funcionan
en móviles ni navegadores modernos. Esto significa que están perdiendo el 70%
del tráfico móvil.

Podemos ayudarlo a crear un sitio moderno que funcione en todos los dispositivos
y aparezca en Google. ¿Le gustaría ver un ejemplo de cómo podría mejorar?
```

#### Para sitios Antiguos (2-3/10):
```
"Hola [Nombre],

Su sitio web podría tener un mejor diseño y funcionalidad. Actualmente no es
compatible con móviles y tiene varias mejoras que podrían aumentar sus ventas.

Ofrecemos rediseños modernos a partir de $X,XXX con entrega en 2 semanas.
¿Le gustaría saber más?
```

#### Para sitios Regulares (4-5/10):
```
"Hola [Nombre],

Su sitio web está bien pero podemos hacerlo aún mejor con:

- Mejor optimización para Google
- Integración con WhatsApp para contacto directo
- Mejoras visuales modernas
- Página más rápida

¿Le gustaría una evaluación gratuita de mejoras?
```

## 📈 Métricas de Éxito

### Objetivos Mensuales:
- [ ] Analizar 50 nuevas empresas
- [ ] Identificar 10 oportunidades de prioridad alta
- [ ] Contactar 5 empresas críticas
- [ ] Cerrar 2-3 contratos

### Indicadores Clave:
- **Tasa de conversión**: (Contratos / Contactos) × 100
- **Valor promedio de contrato**: $X,XXX
- **Tiempo de cierre promedio**: X días

## 🔧 Configuración Futura

Para automatizar completamente el proceso, se requiere:

1. **API de Google Maps Places** - Para buscar empresas automáticamente
   - Requiere cuenta de Google Cloud
   - Costo: ~$5 por 1,000 búsquedas

2. **Servicio de proxies rotativos** - Para evitar bloqueos
   - Ejemplo: ScraperAPI, ZenRows

3. **Base de datos PostgreSQL** - Para escalar más allá de CSV

4. **Integración con CRM** - Para seguimiento de clientes
   - Ejemplo: HubSpot, Pipedrive

## 📞 Soporte

Para preguntas o asistencia con el sistema, contactar a [tu-email].

---

**Última actualización**: 2026-02-18
**Versión**: 1.0.0
