# Design: Phase 8.5 — Media & Generation

## Technical Approach

Most image-generation infrastructure already exists. `ImageProvider` interface, `OpenAIImageProvider` (DALL-E), `FalImageProvider`, `ReplicateImageProvider`, and `ImageGenerateTool` are all live in `pkg/tools/image_provider.go` and `pkg/tools/image.go`. Config structs (`ImageProvidersConfig`, `ImageProviderConfig`) exist in `pkg/config/config.go` and are already wired into `NewAgentLoop`. The `read_pdf` tool exists in `pkg/tools/pdf.go` using `ledongthuc/pdf` but only supports text mode — no vision, OCR, auto, metadata, table extraction, or batch.

**Work remaining:** extend `PDFTool` with multi-mode dispatch; the image side needs no structural changes.

## Architecture Decisions

| Area | Option A | Option B | Decision |
|------|----------|----------|----------|
| PDF library | `ledongthuc/pdf` (already present, pure-Go) | `pdfcpu` (more features, still pure-Go) | Keep `ledongthuc/pdf` — already imported, zero CGO, sufficient for text extraction. For OCR fallback, detect empty text output and route to vision. |
| OCR implementation | External `tesseract` binary | Vision LLM as OCR | Vision LLM — no binary dependency, already satisfies "no CGO". Scanned PDFs go to vision path automatically in `auto` mode. |
| Vision path — PDF format | Convert PDF pages to PNG first | Send raw PDF bytes base64-encoded | Raw PDF base64 — Anthropic natively supports `application/pdf` as a document block; Gemini supports it via inline data. Eliminates image conversion entirely. |
| Table extraction | Parse HTML tables from LLM output | Prompt LLM to return JSON tables | Prompt-driven JSON — simpler, no fragile HTML parsing. Tool returns tables as a JSON array when `extract_tables: true`. |
| Batch processing | Loop over files in tool, aggregate results | Single file per call | Loop inside tool — agent calls once with `file_paths: [...]`, tool aggregates; simpler for the LLM. |
| Multi-mode branching | Separate tools per mode | Single tool with `mode` param | Single tool with `mode` param — matches proposal, consistent with existing `pages` param pattern. |

## Data Flow

### PDF Vision Mode

```
Agent calls read_pdf(mode="vision", file_path="report.pdf")
    │
    ▼
PDFTool.Execute()
    │  reads file bytes → base64-encodes
    │
    ▼
buildVisionMessages(bytes, prompt, pages)
    │  constructs []providers.Message with content blocks:
    │    [{ type:"document", source:{ type:"base64", media_type:"application/pdf", data:"..." } },
    │     { type:"text", text:"Extract and summarize..." }]
    │
    ▼
providers.LLMProvider.Chat(ctx, messages, nil, model, opts)
    │  (injected at construction via PDFTool.provider field)
    │
    ▼
LLM response text returned to agent
```

### PDF Auto Mode

```
PDFTool.Execute(mode="auto")
    │
    ├── ExtractPDFText(path, pages)  ← attempt text extraction
    │       │
    │       ├── text non-empty → return text (mode="text" path)
    │       └── text empty / all whitespace
    │               │
    │               └── fall through to vision path
```

### Image Generation (existing, no change)

```
Agent → ImageGenerateTool → ImageProvider.Generate() → URL/LocalPath → saved to workspace
```

## File Changes

| File | Action | Description |
|------|--------|-------------|
| `pkg/tools/pdf.go` | Modify | Add `mode`, `extract_tables`, `file_paths`, `max_pages`, `question` params; add vision/OCR/auto branching; inject `providers.LLMProvider` |
| `pkg/tools/image_provider.go` | No change | Already complete |
| `pkg/tools/image.go` | No change | Already complete |
| `pkg/config/config.go` | No change | `ImageProvidersConfig` already has Fal, Replicate, OpenAI fields |
| `pkg/agent/loop.go` | No change | `NewPDFTool` already registered; image tools already registered |

## Interfaces / Contracts

`PDFTool` gains a `provider` field (optional — nil disables vision mode):

```go
type PDFTool struct {
    workspace string
    restrict  bool
    provider  providers.LLMProvider  // nil → vision mode unavailable
    model     string                 // model to use for vision calls
}

func NewPDFTool(workspace string, restrict bool) *PDFTool
func (t *PDFTool) SetProvider(p providers.LLMProvider, model string)
```

Extended `Parameters()` map:

```go
"mode":           enum["text","ocr","vision","auto"]  // default "auto"
"file_paths":     []string                            // batch; mutually exclusive with file_path
"max_pages":      integer                             // cap to avoid token overrun
"extract_tables": boolean                             // prompt LLM for JSON tables
"question":       string                             // vision-mode question override
```

Vision message format (Anthropic document block):

```json
{
  "role": "user",
  "content": [
    { "type": "document",
      "source": { "type": "base64", "media_type": "application/pdf", "data": "<b64>" } },
    { "type": "text", "text": "<question>" }
  ]
}
```

## Testing Strategy

| Layer | What to Test | Approach |
|-------|-------------|----------|
| Unit | `parsePageRange`, text extraction, mode branching logic | Table-driven tests in `pdf_test.go`, mock `LLMProvider` |
| Unit | Vision path builds correct message structure | Assert message content blocks in mock provider call |
| Unit | Auto mode falls back to vision when text empty | Inject a scanned-only PDF fixture (all-image) |
| Integration | `ImageGenerateTool` + each provider | Existing `image_test.go`; add Fal/Replicate mock HTTP server tests |

## Migration / Rollout

No migration required. `NewPDFTool` signature is unchanged; `SetProvider` is additive. Calling code in `loop.go` adds one `SetProvider` call after construction.

## Open Questions

- [ ] Gemini vision path: does the current `HTTPProvider` Chat method accept document content blocks, or does it need a separate code path? Needs investigation before implementing vision for Gemini.
- [ ] `max_pages` default value: proposal mentions the concern about token cost but does not specify a sensible default. Suggest 20 — needs sign-off.
