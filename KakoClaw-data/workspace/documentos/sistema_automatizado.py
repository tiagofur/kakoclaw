#!/usr/bin/env python3
"""
🦞 SISTEMA AUTOMATIZADO DE LEADS - PICOCLAW

Flujo completo:
1. Buscar negocios en tu zona
2. Extraer información de contacto y web
3. Analizar calidad del sitio
4. Detectar señales de web desactualizada
5. Generar propuesta personalizada
6. Guardar leads y enviar reporte por email
"""

import re
import json
import csv
from datetime import datetime
from typing import Dict, List, Optional, Tuple

class LeadGenerator:
    """Sistema automatizado para generación de leads de rediseño web"""
    
    def __init__(self, ciudad="Monterrey", codigos_postales=None):
        self.ciudad = ciudad
        self.codigos_postales = codigos_postales or ["66035", "66036"]
        self.leads = []
        self.senales_desactualizada = [
            "HTML 4.01",
            "XHTML 1.0",
            "Flash",
            "Joomla 1.5",
            "Joomla 2.5",
            "WordPress 3",
            "PHP 5",
            "http://",
            "<meta http-equiv='X-UA-Compatible'",
            "IE=edge",
            "Dreamweaver",
            "FrontPage",
            "frameset"
        ]
        
    def analizar_web(self, url: str, html_content: str) -> Dict:
        """
        Analiza una página web y detecta señales de desactualización
        
        Args:
            url: URL del sitio a analizar
            html_content: Contenido HTML de la página
            
        Returns:
            Dict con análisis completo
        """
        analisis = {
            "url": url,
            "fecha_analisis": datetime.now().isoformat(),
            "calificacion": 0,
            "senales_encontradas": [],
            "tecnologias_detectadas": [],
            "problemas": [],
            "oportunidades": [],
            "prioridad": "Baja"
        }
        
        if not html_content or html_content.startswith("Error"):
            return {
                **analisis,
                "calificacion": 0,
                "prioridad": "URGENTE",
                "problemas": ["No se pudo acceder al sitio"]
            }
        
        html_lower = html_content.lower()
        
        # Detectar tecnologías
        tech_patterns = {
            "WordPress": ["wp-content", "wp-includes", "wordpress"],
            "Joomla": ["com_content", "joomla"],
            "Drupal": ["drupal", "sites/default"],
            "Wix": ["wix-static", "wix.com"],
            "Squarespace": ["squarespace"],
            "Shopify": ["shopify"],
            "React": ["react", "_next"],
            "Vue": ["vue"],
            "Angular": ["ng-app", "angular"],
            "Bootstrap": ["bootstrap"],
            "Tailwind": ["tailwind"],
            "jQuery": ["jquery"],
            "Google Fonts": ["fonts.googleapis"]
        }
        
        for tech, patterns in tech_patterns.items():
            if any(pattern in html_lower for pattern in patterns):
                analisis["tecnologias_detectadas"].append(tech)
        
        # Detectar señales de desactualización
        for senal in self.senales_desactualizada:
            if senal.lower() in html_lower:
                analisis["senales_encontradas"].append(senal)
                
                if "Flash" in senal:
                    analisis["problemas.append("Flash ya no está soportado")
                elif "Joomla 1.5" in senal or "Joomla 2.5" in senal:
                    analisis["problemas"].append("Versión de Joomla obsoleta y vulnerable")
                elif "HTML 4" in senal or "XHTML" in senal:
                    analisis["problemas"].append("HTML4/XHTML obsoleto, no responsive")
                elif "WordPress 3" in senal:
                    analisis["problemas"].append("WordPress muy antiguo, vulnerable")
                elif "PHP 5" in senal:
                    analisis["problemas"].append("PHP 5 obsoleto, sin soporte de seguridad")
        
        # Detectar HTTPS vs HTTP
        if "http://" in url and not "https://" in url:
            analisis["problemas"].append("No usa HTTPS - inseguro para clientes")
        elif 'content="upgrade-insecure-requests"' not in html_lower and "http://" in html_content:
            analisis["problemas"].append("Contenido mixto HTTP/HTTPS")
        
        # Detectar responsividad
        if "viewport" not in html_lower:
            analisis["problemas"].append("No responsive - no se ve bien en móviles")
        
        # Detectar meta tags básicos
        if "<meta name=\"description\"" not in html_content:
            analisis["problemas"].append("Falta meta description - mal para SEO")
        
        if "<title>" not in html_content or "<title></title>" in html_content:
            analisis["problemas"].append("Falta o vacío el título de la página")
        
        # Calificar (1-10)
        analisis["calificacion"] = self._calificar_sitio(analisis)
        
        # Determinar prioridad
        analisis["prioridad"] = self._determinar_prioridad(analisis)
        
        # Generar oportunidades
        analisis["oportunidades"] = self._generar_oportunidades(analisis)
        
        return analisis
    
    def _calificar_sitio(self, analisis: Dict) -> int:
        """Califica el sitio del 1 al 10"""
        score = 10
        
        # Restar puntos por problemas
        for problema in analisis["problemas"]:
            if "Flash" in problema:
                score -= 4
            elif "Joomla 1.5" in problema or "Joomla 2.5" in problema:
                score -= 3
            elif "no responsive" in problema:
                score -= 3
            elif "No usa HTTPS" in problema:
                score -= 2
            elif "WordPress 3" in problema:
                score -= 2
            elif "meta description" in problema:
                score -= 1
            elif "título" in problema:
                score -= 1
            elif "PHP 5" in problema:
                score -= 2
            elif "HTML4" in problema or "XHTML" in problema:
                score -= 2
        
        # Bonus por tecnologías modernas
        tech_modernas = ["React", "Vue", "Angular", "Tailwind"]
        if any(tech in analisis["tecnologias_detectadas"] for tech in tech_modernas):
            score = min(10, score + 1)
        
        return max(1, score)
    
    def _determinar_prioridad(self, analisis: Dict) -> str:
        """Determina prioridad del lead"""
        score = analisis["calificacion"]
        
        if score <= 3:
            return "URGENTE 🔴"
        elif score <= 5:
            return "ALTA 🔴"
        elif score <= 7:
            return "MEDIA 🟡"
        else:
            return "BAJA 🟢"
    
    def _generar_oportunidades(self, analisis: Dict) -> List[str]:
        """Genera lista de oportunidades de mejora"""
        oportunidades = []
        
        if "no responsive" in str(analisis["problemas"]).lower():
            oportunidades.append("✓ Rediseño responsive para móviles")
        
        if "HTTPS" in str(analisis["problemas"]):
            oportunidades.append("✓ Implementación de certificado SSL/HTTPS")
        
        if "Flash" in str(analisis["senales_encontradas"]):
            oportunidades.append("✓ Migrar de Flash a HTML5 moderno")
        
        if "Joomla 1.5" in str(analisis["senales_encontradas"]) or "Joomla 2.5" in str(analisis["senales_encontradas"]):
            oportunidades.append("✓ Actualizar o migrar de Joomla obsoleto")
        
        if any("WordPress" in s for s in analisis["tecnologias_detectadas"]) and "WordPress 3" in str(analisis["senales_encontradas"]):
            oportunidades.append("✓ Actualizar WordPress a última versión")
        
        if "meta description" in str(analisis["problemas"]):
            oportunidades.append("✓ Optimización SEO básica")
        
        if analisis["calificacion"] <= 5:
            oportunidades.append("✓ Rediseño completo de interfaz")
            oportunidades.append("✓ Mejora de velocidad de carga")
        
        return oportunidades if oportunidades else ["✓ Auditoría web completa"]
    
    def generar_propuesta(self, negocio: Dict, analisis: Dict) -> str:
        """
        Genera una propuesta personalizada para el negocio
        
        Args:
            negocio: Información del negocio (nombre, contacto, etc.)
            analisis: Resultado del análisis web
            
        Returns:
            Texto de la propuesta personalizada
        """
        score = analisis["calificacion"]
        nombre = negocio.get("nombre", "Estimado dueño")
        
        propuesta = f"""
# 🎯 Propuesta de Rediseño Web para {nombre}

## 📊 Análisis de Sitio Actual
**Sitio:** {analisis.get('url', 'No especificado')}
**Calificación:** {score}/10
**Prioridad:** {analisis['prioridad']}

## ⚠️ Problemas Detectados
"""
        
        for problema in analisis["problemas"]:
            propuesta += f"- {problema}\n"
        
        if not analisis["problemas"]:
            propuesta += "Ninguno (sitio en buen estado general)\n"
        
        propuesta += f"""
## 🚀 Oportunidades de Mejora
"""
        
        for oportunidad in analisis["oportunidades"]:
            propuesta += f"{oportunología}\n"
        
        propuesta += f"""
## 💡 Nuestra Solución

Basado en el análisis de su sitio web actual, le ofrecemos:

1. **Rediseño Responsivo** - Sitio que se ve perfecto en celulares, tablets y computadoras
2. **Tecnología Moderna** - WordPress, Wix, o desarrollo personalizado según sus necesidades
3. **Optimización SEO** - Para que aparezca mejor en Google
4. **Velocidad Optimizada** - Carga rápida para mejor experiencia de usuario
5. **HTTPS/SSL** - Sitio seguro para sus clientes

## 💰 Inversión Estimada

Based on the analysis:
- Sitio básico: $15,000 - $25,000 MXN
- Sitio profesional: $25,000 - $45,000 MXN
- Sitio avanzado con funciones: $45,000 - $80,000 MXN

---

¿Le gustaría agendar una llamada para discutir más detalles?

**Contacto:** [Tu nombre/tu empresa]
**Teléfono:** [Tu número]
**Email:** [Tu email]

---
*Análisis generado automáticamente por Picoclaw - {datetime.now().strftime('%d/%m/%Y')}*
"""
        
        return propuesta
    
    def guardar_leads(self, archivo: str = "leads_generados.csv"):
        """Guarda los leads generados en CSV"""
        if not self.leads:
            print("⚠️ No hay leads para guardar")
            return
        
        with open(archivo, 'w', newline='', encoding='utf-8') as f:
            fieldnames = [
                'fecha', 'nombre', 'tipo', 'direccion', 'telefono', 
                'email', 'web', 'calificacion', 'prioridad', 
                'problemas', 'oportunidades'
            ]
            writer = csv.DictWriter(f, fieldnames=fieldnames)
            writer.writeheader()
            
            for lead in self.leads:
                writer.writerow({
                    'fecha': datetime.now().strftime('%Y-%m-%d'),
                    'nombre': lead.get('nombre', ''),
                    'tipo': lead.get('tipo', ''),
                    'direccion': lead.get('direccion', ''),
                    'telefono': lead.get('telefono', ''),
                    'email': lead.get('email', ''),
                    'web': lead.get('web', ''),
                    'calificacion': lead.get('analisis', {}).get('calificacion', ''),
                    'prioridad': lead.get('analisis', {}).get('prioridad', ''),
                    'problemas': '; '.join(lead.get('analisis', {}).get('problemas', [])),
                    'oportunidades': '; '.join(lead.get('analisis', {}).get('oportunidades', []))
                })
        
        print(f"✅ {len(self.leads)} leads guardados en {archivo}")
    
    def generar_reporte_markdown(self) -> str:
        """Genera un reporte en formato Markdown"""
        if not self.leads:
            return "# Sin Leads\n\nNo se han generado leads aún."
        
        reporte = f"""# 📊 Reporte de Leads - {self.ciudad}
**Fecha:** {datetime.now().strftime('%d/%m/%Y %H:%M')}
**Total Leads:** {len(self.leads)}
**Códigos Postales:** {', '.join(self.codigos_postales)}

---

## 📈 Resumen por Prioridad

"""
        
        # Contar por prioridad
        prioridades = {}
        for lead in self.leads:
            prioridad = lead.get('analisis', {}).get('prioridad', 'Sin prioridad')
            prioridades[prioridad] = prioridades.get(prioridad, 0) + 1
        
        for prioridad, count in sorted(prioridades.items(), key=lambda x: x[1], reverse=True):
            reporte += f"- **{prioridad}**: {count} leads\n"
        
        reporte += f"\n## 🔴 URGENTES (Prioridad Alta)\n\n"
        
        # Listar urgentes
        urgentes = [l for l in self.leads if l.get('analisis', {}).get('prioridad') in ['URGENTE 🔴', 'ALTA 🔴']]
        
        for lead in urgentes:
            reporte += f"### {lead.get('nombre', 'Sin nombre')}\n"
            reporte += f"- **Web:** {lead.get('web', 'N/A')}\n"
            reporte += f"- **Calificación:** {lead.get('analisis', {}).get('calificacion', 'N/A')}/10\n"
            reporte += f"- **Problemas:**\n"
            for p in lead.get('analisis', {}).get('problemas', []):
                reporte += f"  - {p}\n"
            reporte += f"- **Contacto:** {lead.get('telefono', 'N/A')}\n\n"
        
        reporte += "---\n\n*Generado por Picoclaw*"
        
        return reporte
    
    def agregar_lead(self, negocio: Dict, html_content: Optional[str] = None):
        """
        Agrega un lead al sistema
        
        Args:
            negocio: Dict con {nombre, tipo, direccion, telefono, email, web}
            html_content: HTML del sitio web (opcional, para análisis)
        """
        url = negocio.get('web', '')
        
        if html_content:
            analisis = self.analizar_web(url, html_content)
        else:
            analisis = self.analizar_web(url, "")
        
        lead = {
            **negocio,
            'analisis': analisis
        }
        
        self.leads.append(lead)
        print(f"✅ Lead agregado: {negocio.get('nombre', 'Sin nombre')} - Calificación: {analisis['calificacion']}/10")
        
        return lead


def demo_con_datos_ejemplo():
    """Demo del sistema con datos de ejemplo"""
    print("🦞 Iniciando Demo del Sistema de Leads...\n")
    
    generator = LeadGenerator(ciudad="Monterrey", codigos_postales=["66035", "66036"])
    
    # Ejemplos de HTML simulado
    ejemplos_html = {
        "restaurante-viejo.com": """
        <html>
        <head>
            <title>Restaurante Viejo</title>
            <meta http-equiv="X-UA-Compatible" content="IE=EmulateIE7">
        </head>
        <body>
            <embed src="menu.swf" type="application/x-shockwave-flash">
            <p>Bienvenido a nuestro restaurante</p>
        </body>
        </html>
        """,
        
        "abogados-legacy.net": """
        <!DOCTYPE HTML PUBLIC "-//W3C//DTD HTML 4.01 Transitional//EN">
        <html>
        <head>
            <title>Despacho Jurídico</title>
            <!-- Generated by Dreamweaver -->
        </head>
        <frameset rows="80,*">
            <frame src="header.html">
            <frame src="main.html">
        </frameset>
        </html>
        """,
        
        "tienda-moderna.com": """
        <!DOCTYPE html>
        <html lang="es">
        <head>
            <meta charset="UTF-8">
            <meta name="viewport" content="width=device-width, initial-scale=1.0">
            <meta name="description" content="La mejor tienda de productos">
            <title>Tienda Moderna</title>
            <script src="https://cdn.tailwindcss.com"></script>
        </head>
        <body>
            <div id="app"></div>
            <script src="bundle.js"></script>
        </body>
        </html>
        """
    }
    
    # Negocios de ejemplo
    negocios_ejemplo = [
        {
            "nombre": "Restaurante Los Abuelos",
            "tipo": "Restaurante",
            "direccion": "Av. Lincoln 123, CP 66035",
            "telefono": "+52 81 1234-5678",
            "email": "contacto@restaurante-viejo.com",
            "web": "restaurante-viejo.com"
        },
        {
            "nombre": "Despacho Jurídico García",
            "tipo": "Servicios Legales",
            "direccion": "Calle 5 de Mayo 456, CP 66036",
            "telefono": "+52 81 8765-4321",
            "email": "info@abogados-legacy.net",
            "web": "abogados-legacy.net"
        },
        {
            "nombre": "TechStore Monterrey",
            "tipo": "Electrónica",
            "direccion": "Av. Morones Prieto 789, CP 66035",
            "telefono": "+52 81 5555-6666",
            "email": "ventas@tienda-moderna.com",
            "web": "tienda-moderna.com"
        }
    ]
    
    # Procesar leads
    for negocio in negocios_ejemplo:
        html = ejemplos_html.get(negocio['web'], '')
        lead = generator.agregar_lead(negocio, html)
        
        # Mostrar propuesta para cada lead
        propuesta = generator.generar_propuesta(negocio, lead['analisis'])
        print(f"\n{'='*60}")
        print(propuesta[:300] + "...")
        print(f"{'='*60}\n")
    
    # Guardar leads
    generator.guardar_leads("documentos/leads_demo.csv")
    
    # Generar reporte
    reporte = generator.generar_reporte_markdown()
    
    with open("documentos/reporte_demo.md", 'w', encoding='utf-8') as f:
        f.write(reporte)
    
    print("\n📄 Reporte generado: documentos/reporte_demo.md")
    print(f"\n{reporte}")
    
    return generator


if __name__ == "__main__":
    demo_con_datos_ejemplo()
