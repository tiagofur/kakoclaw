#!/usr/bin/env python3
"""
🚀 PICOCLAW - FLUJO AUTOMATIZADO COMPLETO

Este script es la "plantilla" para que Picoclaw ejecute el flujo:

1. web_search -> Buscar negocios
2. web_fetch -> Extraer contenido de sitios
3. sistema_automatizado.py -> Analizar y generar propuestas
4. task_manager -> Crear tareas de seguimiento
5. send_email_report -> Enviar reportes por email

USO:
- Picoclaw puede ejecutar este script
- O usar las funciones directamente
"""

import sys
import os

# Añadir el directorio actual al path
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))

from sistema_automatizado import LeadGenerator


def flujo_completo_ejemplo():
    """
    Ejemplo del flujo completo que Picoclaw puede ejecutar

    Esto demuestra cómo integrar:
    - web_search (via Picoclaw)
    - web_fetch (via Picoclaw)
    - Análisis automatizado
    - Generación de reportes
    - Email (via Picoclaw)
    - Tasks (via Picoclaw)
    """
    
    print("""
╔════════════════════════════════════════════════════════════════╗
║          🦞 PICOCLAW - FLUJO AUTOMATIZADO COMPLETO            ║
╚════════════════════════════════════════════════════════════════╝

📋 PASO 1: BUSCAR NEGOCIOS
─────────────────────────────────────────────────────────────────
Comando: web_search(query="negocios restaurante Monterrey 66035")

Resultado esperado: Lista de URLs de páginas de negocios

📋 PASO 2: EXTRAER CONTENIDO WEB
─────────────────────────────────────────────────────────────────
Comando: web_fetch(url=[URL de cada negocio])

Resultado esperado: HTML de cada sitio para analizar

📋 PASO 3: ANALIZAR SITIOS WEB
─────────────────────────────────────────────────────────────────
Comando: sistema_automatizado.analizar_web(url, html_content)

Resultado esperado:
  - Calificación 1-10
  - Problemas detectados
  - Oportunidades
  - Prioridad del lead

📋 PASO 4: GENERAR PROPUESTAS
─────────────────────────────────────────────────────────────────
Comando: sistema_automatizado.generar_propuesta(negocio, analisis)

Resultado esperado:
  - Propuesta personalizada en Markdown
  - Lista de mejoras específicas
  - Estimación de precios

📋 PASO 5: CREAR TAREAS DE SEGUIMIENTO
─────────────────────────────────────────────────────────────────
Comando: task_manager(action="create", title="Contactar [Negocio]")

Resultado esperado: Tareas creadas en el sistema

📋 PASO 6: ENVIAR REPORTE POR EMAIL
─────────────────────────────────────────────────────────────────
Comando: send_email_report(subject="Reporte de Leads", body=...)

Resultado esperado: Email enviado con el reporte

╔════════════════════════════════════════════════════════════════╗
║                    🎯 EJEMPLO DE USO                          ║
╚════════════════════════════════════════════════════════════════╝
""")

    # Demo con datos de ejemplo
    generator = LeadGenerator(ciudad="Monterrey")
    
    # Simular HTML obtenido con web_fetch
    html_ejemplo = """
    <!DOCTYPE HTML PUBLIC "-//W3C//DTD HTML 4.01 Transitional//EN">
    <html>
    <head>
        <title>Mi Restaurante</title>
        <meta http-equiv="X-UA-Compatible" content="IE=8">
    </head>
    <body>
        <object width="300" height="200" data="banner.swf"></object>
        <p>Bienvenidos a nuestro restaurante</p>
    </body>
    </html>
    """
    
    negocio_ejemplo = {
        "nombre": "Restaurante Ejemplo",
        "tipo": "Restaurante",
        "direccion": "Av. Ejemplo 123, Monterrey 66035",
        "telefono": "+52 81 1234-5678",
        "email": "contacto@ejemplo.com",
        "web": "www.mirestaurante.com"
    }
    
    print("📊 Analizando sitio web...")
    analisis = generator.analizar_web(negocio_ejemplo["web"], html_ejemplo)
    
    print(f"""
✅ Análisis completado:
   Calificación: {analisis['calificacion']}/10
   Prioridad: {analisis['prioridad']}
   Problemas: {len(analisis['problemas'])}
   Oportunidades: {len(analisis['oportunidades'])}
""")
    
    print("\n📝 Generando propuesta...")
    propuesta = generator.generar_propuesta(negocio_ejemplo, analisis)
    
    print("\n" + "="*60)
    print(propuesta)
    print("="*60)
    
    print(f"""
╔════════════════════════════════════════════════════════════════╗
║              ✅ COMANDOS PARA PICOCLAW                          ║
╚════════════════════════════════════════════════════════════════╝

Para ejecutar el flujo real, Picoclaw usaría estos comandos:

# 1. Crear tarea de seguimiento
task_manager(
    action="create",
    title="Contactar Restaurante Ejemplo - Lead URGENTE",
    description=f"Calificación: {analisis['calificacion']}/10 - Prioridad: {analisis['prioridad']}"
)

# 2. Enviar reporte por email
send_email_report(
    subject="🎯 Nuevo Lead Urgente: Restaurante Ejemplo",
    body=f"""
Lead detectado para rediseño web:

🏢 Negocio: {negocio_ejemplo['nombre']}
🌐 Sitio: {negocio_ejemplo['web']}
📞 Teléfono: {negocio_ejemplo['telefono']}
📊 Calificación: {analisis['calificacion']}/10
🚨 Prioridad: {analisis['prioridad']}

Problemas detectados:
{chr(10).join('- ' + p for p in analisis['problemas'])}

---
Propuesta completa adjunta.
    """,
    to="tiagofur@gmail.com"
)

╔════════════════════════════════════════════════════════════════╗
║               🎯 LO QUE PICOCLAW PUEDE HACER AHORA             ║
╚════════════════════════════════════════════════════════════════╝

✅ SÍ puedo hacer ahora mismo:
   • Recibir URLs de negocios (me las das tú)
   • Analizar automáticamente cada sitio web
   • Calificar calidad (1-10) y detectar problemas
   • Generar propuestas personalizadas
   • Crear tareas de seguimiento (task_manager)
   • Enviar reportes por email (send_email_report)

❌ Limitación temporal:
   • web_search - API key no configurada (pero se puede arreglar)
   • Google Maps - Bloqueado por protección anti-scraping

📋 SOLUCIÓN HÍBRIDA RECOMENDADA:
   1. Tú buscas negocios en Google Maps manualmente
   2. Me pasas la lista (nombre, dirección, teléfono, web)
   3. Yo analizo automáticamente y genero propuestas
   4. Creo tareas y te envío reporte por email
""")


def formato_lead_para_picoclaw():
    """
    Retorna el formato que Picoclaw debe usar para procesar leads
    """
    return """
📋 FORMATO DE DATOS PARA PICOCLAW
─────────────────────────────────────────────────

Para que analice leads, envíame los datos en este formato:

1. **Formato simple (chat):**

   "Analiza estos sitios web:
   - Restaurante Los Abuelos: www.losabuelos.com
   - Despacho García: www.abogadosgarcia.net
   - TechStore: www.techstore.mx"

2. **Formato CSV/archivo:**

   nombre,tipo,direccion,telefono,email,web
   Restaurante Los Abuelos,Restaurante,Av. Lincoln 123,81-1234-5678,,www.losabuelos.com
   Despacho García,Legal,Calle 5 de Mayo 456,81-8765-4321,,www.abogadosgarcia.net

3. **Formato JSON:**

   [
     {
       "nombre": "Restaurante Los Abuelos",
       "tipo": "Restaurante",
       "web": "www.losabuelos.com",
       "telefono": "81-1234-5678"
     }
   ]


🚀 LO QUE HARÁ PICOCLAW CON ESTOS DATOS:

1. Para cada web:
   • web_fetch(url) → Obtener HTML
   • Analizar tecnologías, problemas
   • Calificar 1-10
   • Determinar prioridad

2. Generar:
   • CSV con todos los leads
   • Propuestas personalizadas
   • Reporte en Markdown

3. Notificar:
   • task_manager → Crear tareas de seguimiento
   • send_email_report → Enviar reporte por email
"""


if __name__ == "__main__":
    flujo_completo_ejemplo()
    
    print("\n" + "="*70)
    print(formato_lead_para_picoclaw())
