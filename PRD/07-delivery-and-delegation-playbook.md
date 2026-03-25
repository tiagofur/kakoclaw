# PRD 07 - Delivery and Delegation Playbook

## Objetivo

Permitir que cualquier persona o agente ejecute trabajo con calidad consistente, sin generar soluciones aisladas ni deuda de integracion.

## Flujo obligatorio para trabajo sustancial

1. Proposal: problema, alcance, riesgos y exito.
2. Spec: requerimientos funcionales y escenarios.
3. Design: decisiones tecnicas, contratos, trade-offs.
4. Tasks: breakdown ejecutable y verificable.
5. Apply: implementacion por lotes pequenos.
6. Verify: validacion contra spec + pruebas.

## Definition of Ready (DoR)

Una tarea esta lista para ser delegada si tiene:

- Objetivo de negocio claro.
- Alcance y no-alcance definidos.
- Dependencias identificadas.
- Criterios de aceptacion verificables.
- Riesgos conocidos (seguridad, performance, UX).

## Definition of Done (DoD)

Una tarea solo se cierra cuando cumple:

- Requerimientos funcionales validados.
- Sin regresiones criticas en seguridad ni multiusuario.
- Tests relevantes agregados o actualizados.
- Documentacion actualizada en este hub PRD o docs tecnicos.
- Evidencia de validacion (logs, capturas, resultados de test).

## Plantilla de delegacion (humano o agente)

### Input minimo

- Contexto del problema.
- Resultado esperado.
- Archivos o modulos objetivo.
- Restricciones tecnicas.
- Criterios de aceptacion.

### Output esperado

- Resumen ejecutivo de cambios.
- Lista de archivos tocados.
- Riesgos detectados.
- Evidencia de validacion.
- Pendientes o follow-ups.

## Checklist anti-parche

Antes de mergear cualquier cambio:

1. Respeta arquitectura y ownership de modulos.
2. No duplica logica existente.
3. No rompe patrones UX/UI globales.
4. No introduce bypass de permisos.
5. No deja deuda documental.

## Escalamiento

- Si toca auth/storage/permisos: review de seguridad obligatoria.
- Si toca UI transversal: review de design system obligatoria.
- Si toca features core: validar impacto en roadmap y KPI.
