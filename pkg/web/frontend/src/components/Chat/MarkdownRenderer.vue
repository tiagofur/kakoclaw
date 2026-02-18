<template>
  <div class="markdown-body" ref="containerRef" v-html="rendered"></div>
</template>

<script setup>
import { computed, ref, onMounted, onUpdated, onBeforeUnmount } from 'vue'
import MarkdownIt from 'markdown-it'
import hljs from 'highlight.js/lib/core'
import { ensureLoaded, runPython, getStatus } from '../../services/pyodideRunner'

// Register commonly used languages to keep bundle small
import javascript from 'highlight.js/lib/languages/javascript'
import python from 'highlight.js/lib/languages/python'
import bash from 'highlight.js/lib/languages/bash'
import json from 'highlight.js/lib/languages/json'
import go from 'highlight.js/lib/languages/go'
import sql from 'highlight.js/lib/languages/sql'
import xml from 'highlight.js/lib/languages/xml'
import css from 'highlight.js/lib/languages/css'
import typescript from 'highlight.js/lib/languages/typescript'
import yaml from 'highlight.js/lib/languages/yaml'
import markdown from 'highlight.js/lib/languages/markdown'
import shell from 'highlight.js/lib/languages/shell'

hljs.registerLanguage('javascript', javascript)
hljs.registerLanguage('js', javascript)
hljs.registerLanguage('python', python)
hljs.registerLanguage('py', python)
hljs.registerLanguage('bash', bash)
hljs.registerLanguage('json', json)
hljs.registerLanguage('go', go)
hljs.registerLanguage('golang', go)
hljs.registerLanguage('sql', sql)
hljs.registerLanguage('html', xml)
hljs.registerLanguage('xml', xml)
hljs.registerLanguage('css', css)
hljs.registerLanguage('typescript', typescript)
hljs.registerLanguage('ts', typescript)
hljs.registerLanguage('yaml', yaml)
hljs.registerLanguage('yml', yaml)
hljs.registerLanguage('markdown', markdown)
hljs.registerLanguage('md', markdown)
hljs.registerLanguage('shell', shell)
hljs.registerLanguage('sh', shell)

const PYTHON_LANGS = new Set(['python', 'py'])

const md = new MarkdownIt({
  html: false,
  linkify: true,
  typographer: false,
  breaks: true,
  highlight(str, lang) {
    const isPython = PYTHON_LANGS.has(lang)
    const runBtn = isPython
      ? `<button class="hljs-run-btn" data-pyodide-run>&#9654; Run</button>`
      : ''

    if (lang && hljs.getLanguage(lang)) {
      try {
        const result = hljs.highlight(str, { language: lang, ignoreIllegals: true })
        return `<pre class="hljs-code-block${isPython ? ' hljs-python-block' : ''}"><div class="hljs-code-header"><span class="hljs-code-lang">${lang}</span><span class="hljs-code-actions">${runBtn}<button class="hljs-copy-btn" data-copy>Copy</button></span></div><code class="hljs">${result.value}</code></pre>`
      } catch (_) { /* fallback */ }
    }
    // Auto-detect if no language specified
    try {
      const result = hljs.highlightAuto(str)
      return `<pre class="hljs-code-block"><div class="hljs-code-header"><span class="hljs-code-lang">code</span><span class="hljs-code-actions"><button class="hljs-copy-btn" data-copy>Copy</button></span></div><code class="hljs">${result.value}</code></pre>`
    } catch (_) { /* fallback */ }
    return ''
  }
})

const props = defineProps({
  content: { type: String, default: '' }
})

const rendered = computed(() => {
  if (!props.content) return ''
  return md.render(props.content)
})

// --- DOM event delegation for Copy and Run buttons ---
const containerRef = ref(null)

function handleContainerClick(e) {
  const copyBtn = e.target.closest('[data-copy]')
  if (copyBtn) {
    const pre = copyBtn.closest('pre')
    if (pre) {
      const code = pre.querySelector('code')
      if (code) {
        navigator.clipboard.writeText(code.textContent)
        copyBtn.textContent = 'Copied!'
        setTimeout(() => { copyBtn.textContent = 'Copy' }, 1500)
      }
    }
    return
  }

  const runBtn = e.target.closest('[data-pyodide-run]')
  if (runBtn) {
    handleRunPython(runBtn)
  }
}

async function handleRunPython(btn) {
  const pre = btn.closest('pre')
  if (!pre) return

  const code = pre.querySelector('code')
  if (!code) return
  const source = code.textContent

  // Disable button and show loading state
  btn.disabled = true
  const originalText = btn.innerHTML
  btn.innerHTML = '&#9203; Loading...'

  // Find or create output container right after the <pre>
  let outputEl = pre.nextElementSibling
  if (!outputEl || !outputEl.classList.contains('hljs-code-output')) {
    outputEl = document.createElement('div')
    outputEl.className = 'hljs-code-output'
    pre.insertAdjacentElement('afterend', outputEl)
  }
  outputEl.textContent = 'Loading Pyodide...'
  outputEl.className = 'hljs-code-output'

  try {
    // Ensure Pyodide is loaded (first call downloads WASM)
    if (getStatus() !== 'ready') {
      await ensureLoaded()
    }

    btn.innerHTML = '&#9203; Running...'
    outputEl.textContent = 'Running...'

    const { output, error, duration } = await runPython(source)

    if (error) {
      outputEl.className = 'hljs-code-output hljs-code-output-error'
      outputEl.textContent = error
    } else if (output) {
      outputEl.className = 'hljs-code-output hljs-code-output-success'
      outputEl.textContent = output
    } else {
      outputEl.className = 'hljs-code-output hljs-code-output-success'
      outputEl.textContent = '(no output)'
    }

    // Append duration info
    const durationSpan = document.createElement('span')
    durationSpan.className = 'hljs-code-output-duration'
    durationSpan.textContent = ` (${duration}ms)`
    outputEl.appendChild(durationSpan)
  } catch (err) {
    outputEl.className = 'hljs-code-output hljs-code-output-error'
    outputEl.textContent = `Failed to load Pyodide: ${err.message}`
  } finally {
    btn.disabled = false
    btn.innerHTML = '&#9654; Run'
  }
}

onMounted(() => {
  if (containerRef.value) {
    containerRef.value.addEventListener('click', handleContainerClick)
  }
})

onBeforeUnmount(() => {
  if (containerRef.value) {
    containerRef.value.removeEventListener('click', handleContainerClick)
  }
})
</script>

<style scoped>
/* Markdown body styles */
.markdown-body {
  line-height: 1.6;
  word-wrap: break-word;
  overflow-wrap: break-word;
}

.markdown-body :deep(p) {
  margin: 0.4em 0;
}

.markdown-body :deep(p:first-child) {
  margin-top: 0;
}

.markdown-body :deep(p:last-child) {
  margin-bottom: 0;
}

.markdown-body :deep(ul),
.markdown-body :deep(ol) {
  padding-left: 1.5em;
  margin: 0.4em 0;
}

.markdown-body :deep(li) {
  margin: 0.15em 0;
}

.markdown-body :deep(blockquote) {
  border-left: 3px solid rgba(139, 92, 246, 0.4);
  padding-left: 0.8em;
  margin: 0.5em 0;
  color: inherit;
  opacity: 0.85;
}

.markdown-body :deep(a) {
  color: rgb(139, 92, 246);
  text-decoration: underline;
  text-decoration-color: rgba(139, 92, 246, 0.3);
}

.markdown-body :deep(a:hover) {
  text-decoration-color: rgba(139, 92, 246, 0.8);
}

.markdown-body :deep(code:not(.hljs)) {
  background: rgba(139, 92, 246, 0.1);
  border: 1px solid rgba(139, 92, 246, 0.15);
  padding: 0.15em 0.35em;
  border-radius: 4px;
  font-size: 0.88em;
  font-family: ui-monospace, SFMono-Regular, 'SF Mono', Menlo, Consolas, monospace;
}

.markdown-body :deep(.hljs-code-block) {
  background: rgba(0, 0, 0, 0.25);
  border: 1px solid rgba(139, 92, 246, 0.15);
  border-radius: 8px;
  margin: 0.5em 0;
  overflow: hidden;
  font-size: 0.85em;
}

.markdown-body :deep(.hljs-code-header) {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0.3em 0.8em;
  background: rgba(0, 0, 0, 0.15);
  border-bottom: 1px solid rgba(139, 92, 246, 0.1);
  font-size: 0.8em;
}

.markdown-body :deep(.hljs-code-lang) {
  color: rgb(139, 92, 246);
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.markdown-body :deep(.hljs-code-actions) {
  display: flex;
  gap: 0.4em;
  align-items: center;
}

.markdown-body :deep(.hljs-copy-btn) {
  background: rgba(139, 92, 246, 0.15);
  border: 1px solid rgba(139, 92, 246, 0.2);
  color: rgba(139, 92, 246, 0.8);
  padding: 0.15em 0.5em;
  border-radius: 4px;
  cursor: pointer;
  font-size: 0.85em;
  transition: all 0.15s;
}

.markdown-body :deep(.hljs-copy-btn:hover) {
  background: rgba(139, 92, 246, 0.25);
  color: rgb(139, 92, 246);
}

.markdown-body :deep(.hljs-run-btn) {
  background: rgba(52, 211, 153, 0.15);
  border: 1px solid rgba(52, 211, 153, 0.25);
  color: rgba(52, 211, 153, 0.9);
  padding: 0.15em 0.5em;
  border-radius: 4px;
  cursor: pointer;
  font-size: 0.85em;
  transition: all 0.15s;
}

.markdown-body :deep(.hljs-run-btn:hover) {
  background: rgba(52, 211, 153, 0.25);
  color: rgb(52, 211, 153);
}

.markdown-body :deep(.hljs-run-btn:disabled) {
  opacity: 0.6;
  cursor: wait;
}

.markdown-body :deep(.hljs-code-block code) {
  display: block;
  padding: 0.7em 0.9em;
  overflow-x: auto;
  font-family: ui-monospace, SFMono-Regular, 'SF Mono', Menlo, Consolas, monospace;
  line-height: 1.5;
}

/* Python code output area */
.markdown-body :deep(.hljs-code-output) {
  background: rgba(0, 0, 0, 0.3);
  border: 1px solid rgba(139, 92, 246, 0.1);
  border-top: none;
  border-radius: 0 0 8px 8px;
  margin: -0.5em 0 0.5em;
  padding: 0.5em 0.9em;
  font-family: ui-monospace, SFMono-Regular, 'SF Mono', Menlo, Consolas, monospace;
  font-size: 0.8em;
  line-height: 1.5;
  white-space: pre-wrap;
  word-break: break-all;
  color: rgba(255, 255, 255, 0.7);
}

.markdown-body :deep(.hljs-code-output-success) {
  color: rgba(52, 211, 153, 0.9);
}

.markdown-body :deep(.hljs-code-output-error) {
  color: rgba(248, 113, 113, 0.9);
}

.markdown-body :deep(.hljs-code-output-duration) {
  color: rgba(255, 255, 255, 0.35);
  font-size: 0.85em;
}

/* Highlight.js token colors (dark theme) */
.markdown-body :deep(.hljs-keyword) { color: #c678dd; }
.markdown-body :deep(.hljs-string) { color: #98c379; }
.markdown-body :deep(.hljs-number) { color: #d19a66; }
.markdown-body :deep(.hljs-comment) { color: #5c6370; font-style: italic; }
.markdown-body :deep(.hljs-function) { color: #61afef; }
.markdown-body :deep(.hljs-title) { color: #61afef; }
.markdown-body :deep(.hljs-params) { color: #abb2bf; }
.markdown-body :deep(.hljs-built_in) { color: #e6c07b; }
.markdown-body :deep(.hljs-literal) { color: #d19a66; }
.markdown-body :deep(.hljs-type) { color: #e6c07b; }
.markdown-body :deep(.hljs-attr) { color: #d19a66; }
.markdown-body :deep(.hljs-selector-class) { color: #d19a66; }
.markdown-body :deep(.hljs-selector-id) { color: #61afef; }
.markdown-body :deep(.hljs-variable) { color: #e06c75; }
.markdown-body :deep(.hljs-meta) { color: #61afef; }
.markdown-body :deep(.hljs-addition) { color: #98c379; }
.markdown-body :deep(.hljs-deletion) { color: #e06c75; }

.markdown-body :deep(h1),
.markdown-body :deep(h2),
.markdown-body :deep(h3),
.markdown-body :deep(h4) {
  font-weight: 600;
  margin: 0.6em 0 0.3em;
}

.markdown-body :deep(h1) { font-size: 1.3em; }
.markdown-body :deep(h2) { font-size: 1.15em; }
.markdown-body :deep(h3) { font-size: 1.05em; }

.markdown-body :deep(hr) {
  border: none;
  border-top: 1px solid rgba(139, 92, 246, 0.2);
  margin: 0.8em 0;
}

.markdown-body :deep(table) {
  border-collapse: collapse;
  width: 100%;
  margin: 0.5em 0;
  font-size: 0.9em;
}

.markdown-body :deep(th),
.markdown-body :deep(td) {
  border: 1px solid rgba(139, 92, 246, 0.15);
  padding: 0.4em 0.7em;
  text-align: left;
}

.markdown-body :deep(th) {
  background: rgba(139, 92, 246, 0.08);
  font-weight: 600;
}

.markdown-body :deep(strong) { font-weight: 600; }
.markdown-body :deep(em) { font-style: italic; }

/* Image constraints */
.markdown-body :deep(img) {
  max-width: 100%;
  border-radius: 6px;
}
</style>
