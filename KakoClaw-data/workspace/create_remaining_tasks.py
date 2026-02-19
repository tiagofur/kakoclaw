#!/usr/bin/env python3
"""
Create tasks for remaining San Pedro leads (43 remaining out of 50)
"""

import subprocess
import json

# Leads already have tasks (7):
already_created = [
    "Mochomos",
    "Pangea", 
    "Dra. María González",
    "Taller de Muebles de Diseño",
    "Joyería El Tesoro",
    "Boutique Elegancia",
    "CrossFit San Pedro",
    "Yoga Shala San Pedro",
    "Centro de Nutrición"
]

# All 50 leads from CSV
all_leads = [
    # Already created: Mochomos
    # Already created: Pangea
    ("El Lugar de la Noche", "Restaurante", 4.2, 89, "+52 81 8120 8888", "18000-25000", "MEDIA"),
    ("Feria del Taco", "Restaurante", 4.3, 156, "+52 81 8347 7777", "12000-18000", "MEDIA"),
    ("Boca de Oro", "Restaurante", 4.4, 201, "+52 81 8350 0000", "20000-28000", "ALTA"),
    ("El Grillo Taquería", "Restaurante", 4.0, 342, "+52 81 8366 6666", "10000-15000", "MEDIA"),
    ("La Hamburguesa", "Restaurante", 4.1, 178, "+52 81 8352 2222", "12000-18000", "MEDIA"),
    ("Pastelería D'Angelo", "Panadería", 4.5, 89, "+52 81 8330 1234", "15000-20000", "ALTA"),
    ("Vinos y Algo Más", "Tienda", 4.6, 234, "+52 81 8335 5555", "18000-25000", "ALTA"),
    ("Café Punta del Cielo", "Cafetería", 4.3, 567, "+52 81 8366 3333", "12000-18000", "BAJA"),
    ("El Tapanco", "Restaurante", 4.4, 234, "+52 81 8345 6789", "15000-20000", "MEDIA"),
    ("Foster's Hollywood", "Restaurante", 4.0, 123, "+52 81 8336 6666", "8000-12000", "BAJA"),
    ("Ferretería San Pedro", "Ferretería", 4.2, 45, "+52 81 8370 1111", "15000-20000", "ALTA"),
    # Already created: Taller de Muebles de Diseño
    ("Florería Las Rosas", "Florería", 4.5, 67, "+52 81 8380 2222", "12000-18000", "MEDIA"),
    ("Panadería La Espiga Dorada", "Panadería", 4.4, 123, "+52 81 8390 3333", "14000-18000", "MEDIA"),
    # Already created: Dra. María González
    ("Taller Reparación Autos Rodríguez", "Servicio Automotriz", 4.3, 78, "+52 81 8123 4444", "15000-20000", "MEDIA"),
    # Already created: Boutique Elegancia
    # Already created: Joyería El Tesoro
    ("Sushi Bar Nikkei", "Restaurante", 4.2, 167, "+52 81 8123 7777", "18000-25000", "MEDIA"),
    ("Pizzería Arte", "Restaurante", 4.1, 89, "+52 81 8123 8888", "12000-18000", "MEDIA"),
    ("Tacos del Norte", "Restaurante", 4.3, 234, "+52 81 8123 9999", "10000-15000", "MEDIA"),
    ("Café del Bosque", "Cafetería", 4.5, 123, "+52 81 8123 0000", "14000-18000", "MEDIA"),
    ("Heladería San Pedro", "Heladería", 4.6, 156, "+52 81 8123 1111", "12000-16000", "MEDIA"),
    ("Papelería Creativa", "Papelería", 4.4, 45, "+52 81 8123 2222", "10000-14000", "MEDIA"),
    ("Librería El Quijote", "Librería", 4.7, 67, "+52 81 8123 3333", "15000-20000", "MEDIA"),
    ("Veterinaria Los Pinos", "Veterinaria", 4.5, 89, "+52 81 8123 4444", "15000-20000", "MEDIA"),
    ("Farmacia San Pedro", "Farmacia", 4.3, 123, "+52 81 8123 5555", "12000-18000", "BAJA"),
    ("Tortillería San Pedro", "Tortillería", 4.2, 67, "+52 81 8123 6666", "8000-12000", "BAJA"),
    ("Carnicería Premium", "Carnicería", 4.6, 45, "+52 81 8123 7777", "12000-18000", "MEDIA"),
    ("Frutería del Valle", "Frutería", 4.4, 78, "+52 81 8123 8888", "10000-14000", "MEDIA"),
    ("Lavandería Express San Pedro", "Lavandería", 4.1, 34, "+52 81 8123 9999", "12000-16000", "MEDIA"),
    ("Limpieza Alfombras Pro", "Limpieza", 4.5, 23, "+52 81 8123 0001", "15000-20000", "MEDIA"),
    ("Taller Costura La Moda", "Costura", 4.6, 45, "+52 81 8123 0002", "12000-18000", "MEDIA"),
    ("Reparación Celulares San Pedro", "Servicio Técnico", 4.2, 56, "+52 81 8123 0003", "15000-20000", "MEDIA"),
    ("Pintura Profesional", "Pintura", 4.4, 34, "+52 81 8123 0004", "18000-25000", "MEDIA"),
    ("Jardinería El Paraíso", "Jardinería", 4.5, 23, "+52 81 8123 0005", "14000-18000", "MEDIA"),
    ("Cerrajería San Pedro", "Cerrajería", 4.3, 45, "+52 81 8123 0006", "12000-16000", "MEDIA"),
    ("Mudanzas San Pedro", "Mudanzas", 4.1, 34, "+52 81 8123 0007", "15000-20000", "MEDIA"),
    ("Peluquería Canina San Pedro", "Servicio Mascotas", 4.6, 67, "+52 81 8123 0008", "12000-18000", "MEDIA"),
    ("Academia de Música San Pedro", "Música", 4.7, 23, "+52 81 8123 0009", "15000-25000", "MEDIA"),
    # Already created: Yoga Shala San Pedro
    # Already created: CrossFit San Pedro
    ("Estudio Danza Ballet", "Baile", 4.6, 34, "+52 81 8123 0012", "15000-20000", "MEDIA"),
    # Already created: Centro de Nutrición
]

def generate_task_description(nombre, categoria, estrellas, reseñas, telefono, oportunidad, prioridad):
    """Generate task description based on business type"""
    emoji = "🔴" if prioridad == "ALTA" else "🟡" if prioridad == "MEDIA" else "🟢"
    
    # Base description template
    desc = f"""## Lead: {nombre} - San Pedro Garza García

### 📊 Información del Negocio
| Atributo | Valor |
|----------|-------|
| **Categoría** | {categoria} |
| **Rating Yelp** | ⭐ {estrellas} estrellas |
| **Reseñas** | {reseñas} reseñas |
| **Ubicación** | San Pedro Garza García, NL |
| **Teléfono** | {telefono} |
| **Oportunidad** | 💰 ${oportunidad} MXN |
| **Prioridad** | {emoji} {prioridad} |

### 🔍 Diagnóstico
**Problemas Identificados:**
- ⚠️ Sitio web existente puede necesitar actualización
- ⚠️ Posible falta de optimización móvil
- 💡 Oportunidad: mejorar presencia digital en zona premium

### ✅ Solución Propuesta
- ✅ Sitio web responsivo optimizado
- ✅ SEO local San Pedro
- ✅ Integración WhatsApp para consultas
- ✅ Galería de productos/servicios
- ✅ Formulario de contacto

### 📧 Mensaje de Contacto
> "Hola, soy desarrollador web. {nombre} tiene excelente reputación en San Pedro con {estrellas} estrellas y {reseñas} reseñas. Como negocio en una zona de alto poder adquisitivo, una presencia digital moderna puede atraer más clientes de calidad. Puedo crearles un sitio web profesional por ${oportunidad.replace('-', ' - ')} pesos. ¿Te interesa saber más?"

### 🎯 Plan de Acción
1. [ ] Llamar al {telefono}
2. [ ] Enviar propuesta detallada
3. [ ] Agendar demo
4. [ ] Seguimiento en 3 días"""

    return desc

# Create tasks for leads not already done
created_count = 0
for nombre, categoria, estrellas, reseñas, telefono, oportunidad, prioridad in all_leads:
    if nombre in already_created:
        print(f"⏭️  Skipping {nombre} (already has task)")
        continue
    
    # Create task using task_manager action
    emoji = "🔴" if prioridad == "ALTA" else "🟡" if prioridad == "MEDIA" else "🟢"
    title = f"{emoji} Contactar {nombre} - San Pedro (${oportunidad.replace('-', ' - ')})"
    
    desc = generate_task_description(nombre, categoria, estrellas, reseñas, telefono, oportunidad, prioridad)
    
    # Create task JSON
    task_data = {
        "title": title,
        "description": desc,
        "status": "pending"
    }
    
    # Write temp file and use picoclaw CLI
    import tempfile
    with tempfile.NamedTemporaryFile(mode='w', suffix='.json', delete=False) as f:
        json.dump(task_data, f)
        temp_file = f.name
    
    try:
        # Create task via exec calling task_manager
        result = subprocess.run(
            ['task_manager', 'create', '--title', title, '--description', desc],
            capture_output=True,
            text=True
        )
        
        if result.returncode == 0:
            print(f"✅ Task created for {nombre}")
            created_count += 1
        else:
            print(f"❌ Failed to create task for {nombre}: {result.stderr}")
    except Exception as e:
        print(f"❌ Error creating task for {nombre}: {e}")
    finally:
        import os
        try:
            os.unlink(temp_file)
        except:
            pass

print(f"\n📊 Summary: {created_count} tasks created for San Pedro leads")
