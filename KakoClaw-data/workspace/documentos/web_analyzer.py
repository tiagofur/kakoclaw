#!/usr/bin/env python3
"""
Analizador de Páginas Web para Servicios de Rediseño
Este script automatiza el análisis de sitios web de empresas locales.
"""

import requests
from bs4 import BeautifulSoup
import re
import json
from urllib.parse import urlparse
from datetime import datetime
import csv
import os

class WebAnalyzer:
    """Analizador de sitios web para evaluar calidad y oportunidades de rediseño"""

    def __init__(self, csv_path="documentos/empresas_66036.csv"):
        self.csv_path = csv_path
        self.technologies = {
            # Frameworks CSS
            'bootstrap': ['bootstrap.min.css', 'bootstrap.css', 'bootstrap'],
            'tailwind': ['tailwind', 'tailwindcss'],
            'bulma': ['bulma.min.css', 'bulma.css', 'bulma'],
            'foundation': ['foundation.min.css', 'foundation.css', 'foundation'],

            # Frameworks JS
            'react': ['react', 'react-dom'],
            'vue': ['vue.min.js', 'vue.js', 'vue'],
            'angular': ['angular', 'ng-app'],
            'jquery': ['jquery.min.js', 'jquery.js', 'jquery'],

            # CMS
            'wordpress': ['wp-content', 'wp-includes', 'wordpress'],
            'drupal': ['drupal', '/sites/default/'],
            'joomla': ['joomla', '/components/'],

            # Plataformas
            'wix': ['wix', 'wixstatic'],
            'squarespace': ['squarespace'],
            'shopify': ['shopify', 'cdn.shopify'],

            # Indicadores obsoletos
            'flash': ['.swf', 'application/x-shockwave-flash'],
            'html4': ['<!DOCTYPE HTML PUBLIC'],
            'tables_layout': ['<table', '<td><img'],  # Tablas para maquetación

            # Indicadores modernos
            'html5': ['<!DOCTYPE html>', '<header>', '<nav>', '<footer>'],
            'ssl': ['https://'],
            'responsive': ['viewport', 'media='],
        }

    def analyze_website(self, url):
        """Analiza un sitio web completo"""
        try:
            if not url.startswith('http'):
                url = 'https://' + url

            response = requests.get(url, timeout=10)
            soup = BeautifulSoup(response.content, 'html.parser')

            analysis = {
                'url': url,
                'status_code': response.status_code,
                'technologies': [],
                'quality_score': 0,
                'is_responsive': False,
                'has_ssl': url.startswith('https://'),
                'uses_flash': False,
                'is_obsolete': False,
                'framework': None,
                'cms': None,
                'observations': []
            }

            # Detectar tecnologías
            html_text = str(soup).lower()

            for tech, patterns in self.technologies.items():
                for pattern in patterns:
                    if pattern in html_text:
                        if tech not in analysis['technologies']:
                            analysis['technologies'].append(tech)

                        # Detalles específicos
                        if tech == 'flash':
                            analysis['uses_flash'] = True
                            analysis['is_obsolete'] = True
                            analysis['observations'].append(
                                "⚠️ CRÍTICO: Sitio usa Flash (no funciona en móviles)"
                            )
                        elif tech == 'tables_layout':
                            analysis['is_obsolete'] = True
                            analysis['observations'].append(
                                "⚠️ Usa tablas HTML para maquetación (técnica obsoleta)"
                            )
                        elif tech == 'html4':
                            analysis['is_obsolete'] = True
                            analysis['observations'].append(
                                "⚠️ Sitio en HTML4 (obsoleto)"
                            )
                        elif tech == 'wordpress':
                            analysis['cms'] = 'WordPress'
                        elif tech == 'joomla':
                            analysis['cms'] = 'Joomla'
                            analysis['is_obsolete'] = True
                            analysis['observations'].append(
                                "⚠️ Detectado Joomla (verificar versión por seguridad)"
                            )
                        elif tech in ['bootstrap', 'tailwind', 'bulma', 'foundation']:
                            analysis['framework'] = tech.capitalize()

            # Detectar responsividad
            if soup.find('meta', attrs={'name': 'viewport'}):
                analysis['is_responsive'] = True
            else:
                analysis['observations'].append(
                    "❌ No detectado viewport (puede no ser responsivo)"
                )

            # Detectar formulario
            has_form = soup.find('form')
            if not has_form:
                analysis['observations'].append(
                    "❌ No se encontró formulario de contacto"
                )

            # Detectar integración con redes sociales
            social_links = soup.find_all('a', href=re.compile(r'(facebook|instagram|twitter|linkedin|whatsapp)'))
            if len(social_links) == 0:
                analysis['observations'].append(
                    "❌ Sin integración con redes sociales"
                )

            # Detectar SEO básico
            title = soup.find('title')
            if not title or len(title.get_text()) < 10:
                analysis['observations'].append(
                    "⚠️ Título de página ausente o muy corto (mala para SEO)"
                )

            meta_desc = soup.find('meta', attrs={'name': 'description'})
            if not meta_desc:
                analysis['observations'].append(
                    "⚠️ Sin meta description (mala para SEO)"
                )

            # Calcular score
            base_score = 5

            # Bonificaciones
            if analysis['has_ssl']:
                base_score += 1
            if analysis['is_responsive']:
                base_score += 1
            if analysis['framework']:
                base_score += 1
            if analysis['cms']:
                base_score += 1

            # Penalizaciones
            if analysis['uses_flash']:
                base_score -= 4
            if analysis['is_obsolete']:
                base_score -= 2
            if not analysis['is_responsive']:
                base_score -= 1
            if not analysis['has_ssl']:
                base_score -= 1

            analysis['quality_score'] = max(1, min(10, base_score))

            # Determinar estado de diseño
            if analysis['quality_score'] >= 8:
                analysis['design_state'] = 'Excelente'
            elif analysis['quality_score'] >= 6:
                analysis['design_state'] = 'Bueno'
            elif analysis['quality_score'] >= 4:
                analysis['design_state'] = 'Regular'
            elif analysis['quality_score'] >= 2:
                analysis['design_state'] = 'Antiguo'
            else:
                analysis['design_state'] = 'Crítico'

            return analysis

        except Exception as e:
            return {
                'url': url,
                'error': str(e),
                'status_code': 'Error',
                'technologies': [],
                'quality_score': 0,
                'design_state': 'No accessible',
                'observations': [f"❌ Error al acceder al sitio: {str(e)}"]
            }

    def check_duplicate(self, url):
        """Verifica si la empresa ya está en el CSV"""
        if not os.path.exists(self.csv_path):
            return False

        with open(self.csv_path, 'r', encoding='utf-8') as f:
            reader = csv.DictReader(f)
            for row in reader:
                if row.get('Sitio_Web') == url:
                    return True
        return False

    def add_to_csv(self, company_data, analysis):
        """Añade una empresa al CSV con su análisis"""
        if self.check_duplicate(company_data['Sitio_Web']):
            return False

        # Crear CSV si no existe
        if not os.path.exists(self.csv_path):
            with open(self.csv_path, 'w', newline='', encoding='utf-8') as f:
                writer = csv.writer(f)
                writer.writerow([
                    'Nombre', 'Empresa', 'Direccion', 'Telefono', 'Sitio_Web',
                    'Tecnologias', 'Estado_Diseño', 'Calidad_Diseño',
                    'Observaciones', 'Fecha_Analisis', 'Añadido_CSV'
                ])

        # Añadir registro
        with open(self.csv_path, 'a', newline='', encoding='utf-8') as f:
            writer = csv.writer(f)
            writer.writerow([
                company_data.get('Nombre', ''),
                company_data.get('Empresa', ''),
                company_data.get('Direccion', ''),
                company_data.get('Telefono', ''),
                company_data.get('Sitio_Web', ''),
                ', '.join(analysis.get('technologies', [])),
                analysis.get('design_state', ''),
                analysis.get('quality_score', 0),
                '; '.join(analysis.get('observations', [])),
                datetime.now().strftime('%Y-%m-%d'),
                'Si'
            ])

        return True

    def generate_report(self, analyses):
        """Genera un reporte de los análisis"""
        total = len(analyses)
        with_site = sum(1 for a in analyses if a.get('status_code') == 200)
        critical = sum(1 for a in analyses if a.get('quality_score', 0) <= 2)
        obsolete = sum(1 for a in analyses if a.get('is_obsolete', False))
        avg_score = sum(a.get('quality_score', 0) for a in analyses) / total

        report = f"""
# Reporte de Análisis de Sitios Web
Fecha: {datetime.now().strftime('%Y-%m-%d %H:%M')}

## Resumen
- Total empresas analizadas: {total}
- Sitios web accesibles: {with_site} ({with_site/total*100:.0f}%)
- Sitios críticos (requieren urgente): {critical} ({critical/total*100:.0f}%)
- Sitios obsoletos: {obsolete} ({obsolete/total*100:.0f}%)
- Calidad promedio: {avg_score:.1f}/10

## Oportunidades de Negocio
"""
        if critical > 0:
            report += f"🔴 **PRIORIDAD ALTA**: {critical} empresas con sitios críticos que requieren rediseño urgente\n"
        if obsolete > 0:
            report += f"🟡 **PRIORIDAD MEDIA**: {obsolete} empresas con sitios obsoletos\n"

        return report


# Función principal para análisis en lote
def analyze_companies_batch(companies):
    """Analiza una lista de empresas"""
    analyzer = WebAnalyzer()
    results = []

    for company in companies:
        print(f"Analizando: {company['Nombre']} - {company['Sitio_Web']}")
        analysis = analyzer.analyze_website(company['Sitio_Web'])
        analyzer.add_to_csv(company, analysis)
        results.append(analysis)

    return results


# Ejemplo de uso
if __name__ == "__main__":
    # Ejemplo de empresas para analizar
    companies_example = [
        {
            'Nombre': 'Restaurante El Rincón',
            'Empresa': 'Restaurantes El Rincón S.A.',
            'Direccion': 'Av. Benito Juárez 123, 66036 Monterrey, NL',
            'Telefono': '(81) 1234-5678',
            'Sitio_Web': 'https://restaurante-elincon.com'
        },
        {
            'Nombre': 'Dentista Sonrisas San Nicolás',
            'Empresa': 'Dental Care Monterrey S.A.',
            'Direccion': 'Av. Mitras 101, 66036 Monterrey, NL',
            'Telefono': '(81) 3456-7890',
            'Sitio_Web': 'https://sonrisassannicolas.com'
        },
    ]

    print("=== Analizador de Páginas Web ===\n")
    print("Este script analiza sitios web para identificar oportunidades de rediseño.\n")
    print("Para usar:")
    print("1. Modificar la lista de empresas en 'companies_example'")
    print("2. Ejecutar: python3 web_analyzer.py")
    print("3. Los resultados se guardarán en 'documentos/empresas_66036.csv'\n")
