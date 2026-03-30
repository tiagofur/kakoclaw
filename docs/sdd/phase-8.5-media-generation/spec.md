# Phase 8.5 — Media & Generation Specification

**Change**: `phase-8.5-media-generation`
**Status**: Draft
**Date**: 2026-03-30

---

## Purpose

Defines requirements and acceptance scenarios for two new tools: `generate_image` (multi-provider image generation) and `read_pdf` (multi-mode PDF processing). Both are new domains — full specs, no deltas.

---

## 1. Image Generation (`generate_image`)

### Requirement: Provider dispatch with fallback

The tool MUST attempt image generation using the first configured provider. If no API key is present for that provider, it MUST fall back to the next provider in priority order (DALL-E → Fal → Replicate). If no provider has a valid key, the tool MUST return a descriptive error naming which providers were tried.

#### Scenario: DALL-E succeeds

- GIVEN `providers.openai.api_key` is configured
- WHEN the agent calls `generate_image` with a prompt
- THEN the tool returns a URL pointing to the generated image
- AND the response includes the provider name used

#### Scenario: DALL-E missing, Fal succeeds

- GIVEN `providers.openai.api_key` is absent and `providers.fal.api_key` is configured
- WHEN the agent calls `generate_image` with a prompt
- THEN the tool falls back to Fal and returns a valid image URL

#### Scenario: All providers unconfigured

- GIVEN no image provider API key is present
- WHEN the agent calls `generate_image`
- THEN the tool returns an error: "No image provider configured. Tried: dalle, fal, replicate."

### Requirement: Size and quality passthrough

The tool MUST forward `size` and `quality` parameters to the underlying provider without modification. Unsupported parameters for a given provider MUST be silently ignored (not cause errors).

#### Scenario: Size and quality forwarded

- GIVEN DALL-E is configured
- WHEN `generate_image` is called with `size: "1024x1024"` and `quality: "hd"`
- THEN the API request includes those exact values

---

## 2. PDF Processing (`read_pdf`)

### Requirement: Text mode extraction

The tool MUST extract selectable text from a standard (non-scanned) PDF when `mode: text` or `mode` is unset and text is available.

#### Scenario: Selectable text extracted

- GIVEN a PDF with selectable text
- WHEN `read_pdf` is called with `mode: text`
- THEN the tool returns the extracted text content
- AND no LLM API call is made

#### Scenario: Encrypted PDF

- GIVEN a password-protected PDF
- WHEN `read_pdf` is called with any mode
- THEN the tool returns a clear error: "PDF is encrypted and cannot be processed without a password."

### Requirement: Vision mode

The tool MUST encode PDF pages as base64 and send them to the configured LLM provider's multimodal endpoint when `mode: vision` is specified.

#### Scenario: Vision mode sends to LLM

- GIVEN a configured Claude or Gemini provider
- WHEN `read_pdf` is called with `mode: vision`
- THEN the PDF bytes are base64-encoded and submitted to the multimodal endpoint
- AND the LLM's analysis is returned as the tool result

#### Scenario: max_pages limits pages sent

- GIVEN a PDF with 20 pages and `max_pages: 5`
- WHEN `read_pdf` is called with `mode: vision`
- THEN only the first 5 pages are sent to the LLM

### Requirement: OCR mode

The tool MUST apply OCR processing when `mode: ocr` is specified, and MUST include a confidence note in the output.

#### Scenario: Scanned PDF processed via OCR

- GIVEN a scanned PDF with no selectable text
- WHEN `read_pdf` is called with `mode: ocr`
- THEN the tool returns extracted text
- AND the response includes a note indicating OCR was used (e.g., "Note: extracted via OCR; accuracy may vary")

### Requirement: Auto mode

When `mode: auto` (or `mode` is unset), the tool MUST detect whether the PDF contains selectable text and select `text` mode if so, falling back to `vision` mode otherwise.

#### Scenario: Auto selects text mode

- GIVEN a PDF with selectable text
- WHEN `read_pdf` is called with `mode: auto`
- THEN the tool uses text extraction (no LLM call)

#### Scenario: Auto falls back to vision

- GIVEN a scanned PDF with no selectable text
- WHEN `read_pdf` is called with `mode: auto`
- THEN the tool automatically uses vision mode

### Requirement: Table extraction

When `extract_tables: true` is passed, the tool MUST return a JSON array of detected tables in addition to or instead of plain text.

#### Scenario: Tables returned as JSON

- GIVEN a PDF containing tabular data and `extract_tables: true`
- WHEN `read_pdf` is called
- THEN the response includes a `tables` field containing a JSON array of table objects
- AND each table object contains headers and rows

#### Scenario: No tables found

- GIVEN a PDF with no tables and `extract_tables: true`
- WHEN `read_pdf` is called
- THEN the response includes `tables: []` and does not error

---

## Coverage Summary

| Domain | Requirements | Scenarios | Happy Path | Edge Cases | Error States |
|--------|-------------|-----------|------------|------------|--------------|
| Image Generation | 2 | 3 | Covered | Missing key fallback | All providers absent |
| PDF text/auto | 2 | 4 | Covered | Auto mode branching | Encrypted PDF |
| PDF vision | 1 | 2 | Covered | max_pages truncation | — |
| PDF OCR | 1 | 1 | Covered | — | — |
| Table extraction | 1 | 2 | Covered | Empty tables | — |
