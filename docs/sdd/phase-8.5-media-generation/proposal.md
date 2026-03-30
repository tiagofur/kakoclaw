# Proposal: Phase 8.5 — Media & Generation

**Change**: `phase-8.5-media-generation`
**Status**: Draft
**Inspired by**: OpenClaw (https://github.com/openclaw/openclaw)
**Date**: 2026-03-30

---

## Intent

MakoClaw lacks any media generation capabilities. Users cannot generate images from prompts, nor process PDFs via vision APIs. This gap limits use cases in creative automation, document intelligence, and multimodal workflows. Both features are available in peer frameworks (OpenClaw) and are increasingly expected in AI agent toolkits.

## Scope

### In Scope
- `generate_image` tool with pluggable providers (DALL-E 3/2, Fal.ai, Replicate)
- `ImageProvider` interface mirroring `LLMProvider` extensibility pattern
- Provider config under `providers.dalle`, `providers.fal`, `providers.replicate` in `config.json`
- `read_pdf` tool with modes: `text`, `ocr`, `vision`, `auto`
- Native vision path: send PDF bytes to Claude/Gemini multimodal API directly
- OCR fallback for scanned PDFs (no selectable text)
- Table extraction to JSON/CSV
- PDF metadata extraction (title, author, pages, date)
- Batch PDF processing (multiple files per call)
- Tool registration in `NewAgentLoop`

### Out of Scope
- Video or audio generation
- Image editing / inpainting
- PDF creation / writing
- Fine-tuning or custom model hosting

## Approach

Follow the existing `Tool` interface pattern in `pkg/tools/`. Introduce a parallel `ImageProvider` interface in `pkg/providers/` with `Generate(ctx, prompt, options) (*ImageResult, error)`. Each image backend (DALL-E, Fal, Replicate) implements this interface. Provider selection mirrors `CreateProvider` auto-selection logic.

For PDF: use a single `read_pdf` tool that branches internally based on `mode`. Vision mode marshals file bytes as base64 and passes them to the configured LLM provider's multimodal endpoint (Anthropic natively supports PDF; Gemini via file API). OCR fallback uses a lightweight Go PDF text extractor; table extraction post-processes LLM vision output.

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `pkg/tools/image.go` | New | `generate_image` tool + `ImageProvider` dispatch |
| `pkg/tools/pdf.go` | New | `read_pdf` tool with multi-mode processing |
| `pkg/providers/image_provider.go` | New | `ImageProvider` interface + DALL-E/Fal/Replicate impls |
| `pkg/config/config.go` | Modified | Add `ProvidersConfig.DALLE`, `Fal`, `Replicate` fields |
| `pkg/agent/loop.go` | Modified | Register `generate_image` and `read_pdf` in `NewAgentLoop` |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Fal.ai / Replicate API breaking changes | Med | Version-pin client; abstract behind interface |
| PDF vision token cost unexpectedly high | Med | Warn user in tool description; add `max_pages` param |
| OCR accuracy poor for complex layouts | Med | Document limitation; default to `auto` which prefers vision |
| Binary PDF dependency bloat | Low | Use pure-Go library (e.g., `pdfcpu`) to avoid CGO |

## Rollback Plan

Both tools are additive. Removing them requires: deleting `pkg/tools/image.go`, `pkg/tools/pdf.go`, `pkg/providers/image_provider.go`, removing the three provider config fields from `pkg/config/config.go`, and unregistering the two tools from `NewAgentLoop`. No DB migrations involved.

## Dependencies

- DALL-E: existing OpenAI API key (`providers.openai.api_key`)
- Fal.ai: new `providers.fal.api_key`
- Replicate: new `providers.replicate.api_key`
- PDF vision: requires Claude or Gemini provider configured (already supported)
- Pure-Go PDF lib: `pdfcpu` or `ledongthuc/pdfcontent` (no CGO)

## Success Criteria

- [ ] `generate_image` produces a URL/base64 for a given prompt via at least one provider
- [ ] Provider selection falls back gracefully when API key is absent
- [ ] `read_pdf text` extracts selectable text from a standard PDF
- [ ] `read_pdf vision` sends PDF to Claude/Gemini and returns LLM analysis
- [ ] `read_pdf ocr` handles a scanned PDF with no selectable text
- [ ] `extract_tables: true` returns JSON array of detected tables
- [ ] All tools registered and callable from agent loop
- [ ] No CGO dependencies introduced
- [ ] Unit tests cover provider dispatch and PDF mode branching

## Next Steps

- `sdd-spec` and `sdd-design` can run in parallel (both depend only on this proposal)
