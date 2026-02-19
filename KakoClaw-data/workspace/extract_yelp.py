#!/usr/bin/env python3
import urllib.parse
import requests
import json
import time
import random
from bs4 import BeautifulSoup

# Lista de categorías para buscar en San Pedro Garza García
categorias = [
    "restaurants",
    "shopping",
    "home_services",
    "localservices",
    "automotive",
    "health",
    "beautysvc",
    "arts",
    "realestate"
]

todos_resultados = []

for categoria in categorias:
    print(f"\n{'='*60}")
    print(f"🔍 Buscando categoría: {categoria}")
    print(f"{'='*60}")

    # Yelp search URL
    query = f"San Pedro Garza García, Nuevo León"
    encoded_query = urllib.parse.quote(query)
    url = f"https://www.yelp.com/search?find_desc={categoria}&find_loc={encoded_query}&start=0"

    print(f"URL: {url}")

    try:
        # Headers para simular un navegador real
        headers = {
            "User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36",
            "Accept": "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8",
            "Accept-Language": "es-ES,es;q=0.8,en-US;q=0.5,en;q=0.3",
            "Accept-Encoding": "gzip, deflate, br",
            "DNT": "1",
            "Connection": "keep-alive",
            "Upgrade-Insecure-Requests": "1"
        }

        response = requests.get(url, headers=headers, timeout=30)

        if response.status_code == 200:
            soup = BeautifulSoup(response.text, 'html.parser')

            # Buscar enlaces de negocios en el HTML
            links = soup.find_all('a', href=True)
            for link in links:
                href = link.get('href', '')
                if '/biz/' in href and link.text:
                    nombre = link.text.strip()
                    if nombre and len(nombre) > 3 and len(nombre) < 100:
                        # Evitar duplicados
                        if not any(r['nombre'] == nombre for r in todos_resultados):
                            todos_resultados.append({
                                'categoria': categoria,
                                'nombre': nombre,
                                'url': f"https://www.yelp.com{href}" if href.startswith('/') else href,
                                'rating': 0,
                                'review_count': 0
                            })

            # Guardar el HTML crudo para análisis
            with open(f'/home/picoclaw/.picoclaw/workspace/documentos/yelp_sanpedro_{categoria}.html', 'w', encoding='utf-8') as f:
                f.write(response.text)

            print(f"✅ Guardado: yelp_sanpedro_{categoria}.html")

        else:
            print(f"❌ Error HTTP {response.status_code}")

    except Exception as e:
        print(f"❌ Error: {e}")

    time.sleep(random.uniform(2, 4))

# Guardar todos los resultados
with open('/home/picoclaw/.picoclaw/workspace/documentos/yelp_sanpedro_raw.json', 'w', encoding='utf-8') as f:
    json.dump(todos_resultados, f, ensure_ascii=False, indent=2)

print(f"\n{'='*60}")
print(f"📊 Total resultados encontrados: {len(todos_resultados)}")
print(f"{'='*60}")
for r in todos_resultados[:30]:
    print(f"  - {r['nombre']} [{r['categoria']}]")
