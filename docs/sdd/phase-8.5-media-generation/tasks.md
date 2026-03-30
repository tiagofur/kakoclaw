# Tasks: Phase 8.5 — PDF Multi-Mode Processing

> **Scope**: Extend `pkg/tools/pdf.go` with `mode`, `max_pages`, `extract_tables`, `file_paths`, and `question` params, plus vision/OCR/auto branching. Image generation infrastructure is already complete — no image tasks.

---

## Phase 1: Foundation — Struct & Parameter Extension

- [ ] 1.1 Add `provider providers.LLMProvider` and `model string` fields to `PDFTool` struct in `pkg/tools/pdf.go`; add `SetProvider(p providers.LLMProvider, model string)` method; keep `NewPDFTool` signature unchanged.
  - **RED**: `TestPDFTool_SetProvider` — assert `tool.provider != nil` after calling `SetProvider`; `go test ./pkg/tools/ -run TestPDFTool_SetProvider` → FAIL
  - **GREEN**: implement fields + method → `go test ./pkg/tools/ -run TestPDFTool_SetProvider` → PASS
  - commit: `feat(pdf): add provider field and SetProvider method to PDFTool`

- [ ] 1.2 Replace `Parameters()` map in `PDFTool` with extended schema: add `mode` (enum `text|ocr|vision|auto`, default `auto`), `file_paths` (`[]string`), `max_pages` (`integer`, default 20), `extract_tables` (`boolean`), `question` (`string`). Keep `file_path` and `pages` for backwards compat; mark neither `file_path` nor `file_paths` as required (validate in Execute).
  - **RED**: `TestPDFTool_Parameters` — assert all 5 new keys present in schema; run → FAIL
  - **GREEN**: update `Parameters()` → run → PASS
  - commit: `feat(pdf): extend Parameters schema with mode/max_pages/extract_tables/file_paths/question`

---

## Phase 2: Core — Mode Dispatch Logic

- [ ] 2.1 Add helper `extractPDFPages(filePath string, maxPages int) (string, bool, error)` in `pkg/tools/pdf.go`: opens via `ledongthuc/pdf`, extracts up to `maxPages` pages, returns `(text, hasText, err)` where `hasText = strings.TrimSpace(text) != ""`.
  - **RED**: `TestExtractPDFPages_MaxPages` — create a test PDF with 3 pages (use `pdf.Open` on a real multi-page fixture, or mock with a file that causes `NumPage()` > `maxPages`); assert only `maxPages` pages returned. Actually: test with `maxPages=1` on a valid text PDF and verify only 1 page of content returned.
  - **GREEN**: implement → PASS
  - commit: `feat(pdf): add extractPDFPages helper with max_pages support`

- [ ] 2.2 Add `buildVisionMessages(data []byte, question string, maxPages int) []providers.Message` in `pkg/tools/pdf.go`. Returns a single-element slice: `Message{Role:"user", Content: "<base64-pdf-data-URI>||<question>"}`. **Implementation note**: since `providers.Message.Content` is a plain `string`, encode the vision request as a specially-prefixed string: `"PDF_VISION_BASE64:<b64data>||<question>"`. The `ClaudeProvider` will NOT receive this as a document block through the generic path — instead, `executeSingleVision` will call `providers.LLMProvider.Chat` with a structured approach.
  - **Decision required before coding**: `providers.Message.Content` is `string`. Vision needs structured content blocks (document + text). Options:
    - **A** — Add `ContentBlocks []map[string]interface{}` to `providers.Message`; `ClaudeProvider.buildClaudeParams` reads it when non-nil. Enables proper API call.
    - **B** — PDFTool uses Claude SDK directly (breaks provider abstraction).
    - **C** — Encode as special string prefix; add a `VisionProvider` interface to providers.
  - **Task 2.2**: Implement **Option A**: add `ContentBlocks []ContentBlock` to `providers.Message` in `pkg/providers/types.go` where `ContentBlock` is `{ Type, MediaType, Data, Text string }`. Update `ClaudeProvider.buildClaudeParams` user-message branch: if `msg.ContentBlocks != nil`, build `anthropic.NewUserMessage` with mixed block list (document + text blocks).
  - **RED**: `TestClaudeProvider_VisionMessage` — create a `Message` with `ContentBlocks` set; mock `anthropic.Client`; assert it does NOT fall through to the plain-text branch. Use a real `ClaudeProvider` with a table-driven test + `MockProvider` (mock provider ignores content blocks but compiles).
  - **GREEN**: add `ContentBlock` type + `ContentBlocks` field + `buildClaudeParams` branch → PASS
  - commit: `feat(providers): add ContentBlock type and ContentBlocks field to Message for multimodal`

- [ ] 2.3 Add `executeSingleVision(ctx, absPath, question string, maxPages int) (string, error)` in `PDFTool`: reads file bytes, base64-encodes, builds `[]providers.Message` using `ContentBlocks`, calls `t.provider.Chat(ctx, msgs, nil, t.model, nil)`, returns response content. Returns `"vision mode unavailable: no provider configured"` error if `t.provider == nil`.
  - **RED**: `TestPDFTool_VisionMode_NoProvider` — call with `mode:vision`, no SetProvider; expect specific error string → FAIL
  - **GREEN**: implement → PASS
  - **RED**: `TestPDFTool_VisionMode_MockProvider` — call with `mode:vision`, inject mock provider; assert mock receives a message with non-empty `ContentBlocks` → FAIL (need to capture call). Create a `capturingProvider` in test file that records messages passed to Chat.
  - **GREEN**: implement → PASS
  - commit: `feat(pdf): implement vision mode execution with ContentBlocks`

- [ ] 2.4 Add `executeText(ctx, absPath string, maxPages int) (string, error)` in `PDFTool`: calls `extractPDFPages`, wraps "PDF is encrypted" detection. If the `ledongthuc/pdf` library returns an error containing "encrypted", return the spec-required message: `"PDF is encrypted and cannot be processed without a password."`.
  - **RED**: `TestPDFTool_TextMode_Encrypted` — write a file named `*.pdf` containing bytes that trigger the encrypted error from the library (or create a minimal encrypted PDF header); assert exact error string → FAIL
  - **GREEN**: implement encrypted detection + wrap → PASS
  - commit: `feat(pdf): implement text mode with encrypted PDF detection`

- [ ] 2.5 Add `executeOCR(ctx, absPath string, maxPages int) (string, error)` in `PDFTool`: delegates to `executeSingleVision` with a fixed OCR prompt (`"Extract all text from this document. Be thorough and accurate."`), then appends the OCR note: `"\n\nNote: extracted via OCR; accuracy may vary"`.
  - **RED**: `TestPDFTool_OCRMode_Note` — inject mock provider returning `"hello world"`; assert result contains the note suffix → FAIL
  - **GREEN**: implement → PASS
  - commit: `feat(pdf): implement OCR mode as vision with OCR prompt + accuracy note`

- [ ] 2.6 Add `executeAuto(ctx, absPath string, maxPages int) (string, error)` in `PDFTool`: calls `extractPDFPages`; if `hasText` → return text; else → delegate to `executeSingleVision`.
  - **RED**: `TestPDFTool_AutoMode_FallsBackToVision` — inject a non-text (empty text) PDF path + mock provider; assert mock provider is called → FAIL
  - **RED**: `TestPDFTool_AutoMode_UsesText` — inject a valid text PDF; assert provider is NOT called → FAIL
  - **GREEN**: implement → both PASS
  - commit: `feat(pdf): implement auto mode with text-first, vision-fallback branching`

---

## Phase 3: Table Extraction & Batch Processing

- [ ] 3.1 Add `extractTables(ctx context.Context, absPath, currentText string) (string, error)` in `PDFTool`: if `currentText` is non-empty (text mode), prompts provider with `"Given this PDF text, extract all tables as a JSON array. Each table: {headers: [...], rows: [[...]]}. If no tables, return []. Reply with JSON only.\n\n<text>"`. If text empty, re-runs vision path with table extraction prompt. Returns the JSON string (or `"[]"` on empty).
  - **RED**: `TestExtractTables_NoTables` — mock provider returns `"[]"`; assert result is `"[]"` without error → FAIL
  - **GREEN**: implement → PASS
  - **RED**: `TestExtractTables_WithTables` — mock provider returns JSON array; assert passed through unchanged → FAIL
  - **GREEN**: implement → PASS
  - commit: `feat(pdf): add table extraction via LLM prompt`

- [ ] 3.2 Update `Execute` in `PDFTool` to dispatch mode param, support `file_paths` batch, and integrate table extraction:
  - Parse `mode` (default `"auto"`), `maxPages` (default `20`), `extractTables bool`, `question string` from `args`
  - Resolve single `file_path` OR `file_paths []string`; return error if both empty
  - For each resolved path: call `validatePath`, then dispatch `executeText`/`executeVision`/`executeOCR`/`executeAuto` based on `mode`
  - If `extractTables && t.provider != nil`: call `extractTables()` and append result under `\n\nTables:\n<json>`
  - If batch (`file_paths`): join per-file results separated by `"\n\n---\n\n"` with filename headers
  - **RED**: `TestPDFTool_Execute_ModeText` — args with `mode:"text"`, valid text PDF; assert returns content → FAIL
  - **RED**: `TestPDFTool_Execute_Batch` — args with `file_paths:["a.pdf","b.pdf"]`, mock provider; assert both filenames appear in output → FAIL
  - **RED**: `TestPDFTool_Execute_ExtractTables_NoProvider` — `extract_tables:true`, no SetProvider; assert no panic, tables section absent → FAIL
  - **GREEN**: implement full Execute dispatch → all PASS
  - commit: `feat(pdf): update Execute with multi-mode dispatch, batch support, and table extraction`

---

## Phase 4: Wiring & Integration

- [ ] 4.1 Update `NewAgentLoop` in `pkg/agent/loop.go`: after `tools.NewPDFTool(workspace, restrict)` (line ~272), call `pdfTool.SetProvider(provider, cfg.Agents.Defaults.Model)`. Store `pdfTool` in a local var before registering.
  - Verify compiles: `go build ./pkg/agent/` → no errors
  - commit: `feat(agent): wire LLM provider into PDFTool for vision/OCR modes`

- [ ] 4.2 Update `Description()` in `PDFTool` to reflect new capabilities: `"Extract content from PDF files. Modes: text (selectable text), vision (multimodal LLM analysis), ocr (scanned PDFs), auto (default: text if available, vision otherwise). Supports batch processing via file_paths and table extraction."`
  - commit: `docs(pdf): update read_pdf tool description for multi-mode capabilities`

---

## Phase 5: Test Coverage for Spec Scenarios

- [ ] 5.1 `TestPDFTool_Spec_EncryptedPDF` — error message matches spec exactly: `"PDF is encrypted and cannot be processed without a password."` (covers spec scenario "Encrypted PDF").

- [ ] 5.2 `TestPDFTool_Spec_MaxPagesVision` — inject mock provider; set `max_pages:2` on a 5-page PDF byte fixture (fake, just ensure `extractPDFPages` caps at 2); assert mock received ≤ N-page payload (covers spec scenario "max_pages limits pages sent").

- [ ] 5.3 `TestPDFTool_Spec_OCRNote` — mock provider, `mode:ocr`; assert output contains `"Note: extracted via OCR; accuracy may vary"` (covers spec scenario "Scanned PDF processed via OCR").

- [ ] 5.4 `TestPDFTool_Spec_TablesEmpty` — mock returns `"[]"`; `extract_tables:true`; assert output includes `Tables:` section with `[]` and does NOT error (covers spec scenario "No tables found").

- [ ] 5.5 `TestPDFTool_Spec_TablesWithData` — mock returns `[{"headers":["A"],"rows":[["1"]]}]`; assert JSON present in output (covers spec scenario "Tables returned as JSON").

- [ ] 5.6 Run full test suite: `go test ./pkg/tools/ -run TestPDF -v` → all PASS; `go test ./pkg/providers/ -run TestClaudeProvider_VisionMessage -v` → PASS
  - commit: `test(pdf): add spec scenario coverage for all read_pdf acceptance criteria`

---

## Summary

| Phase | Tasks | Focus |
|-------|-------|-------|
| Phase 1 | 2 | Struct extension & schema |
| Phase 2 | 6 | Mode dispatch (text/vision/OCR/auto) |
| Phase 3 | 2 | Table extraction & batch |
| Phase 4 | 2 | Loop wiring & description |
| Phase 5 | 6 | Spec scenario test coverage |
| **Total** | **18** | |

### Implementation Order

Phase 1 first (compiler baseline), then Phase 2 in order (each function used by the next), then Phase 3 (depends on Phase 2's helpers), Phase 4 (depends on Phase 1 + 2), Phase 5 last (full coverage run). Each task ends with a commit.

### Key Decision Recorded in 2.2

`providers.Message` needs `ContentBlocks []ContentBlock` to support vision properly without breaking the provider abstraction. This is a small but cross-cutting change touching `pkg/providers/types.go` and `pkg/providers/claude_provider.go`. Do it in task 2.2 before 2.3.
