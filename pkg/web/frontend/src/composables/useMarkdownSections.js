/**
 * Composable for parsing and serializing markdown files with ## sections.
 * Used by SoulSettingsTab to map structured form fields to/from SOUL.md, IDENTITY.md, USER.md.
 */

/**
 * Parse a markdown file into its title (# heading) and sections (## headings).
 * Unknown/extra sections are preserved so round-tripping doesn't lose data.
 *
 * @param {string} markdown - Raw markdown content
 * @returns {{ title: string, preamble: string, sections: Array<{ name: string, content: string }> }}
 */
export function parseMarkdownSections(markdown) {
  if (!markdown || !markdown.trim()) {
    return { title: '', preamble: '', sections: [] }
  }

  const lines = markdown.split('\n')
  let title = ''
  let preamble = ''
  const sections = []
  let currentSection = null
  let inPreamble = true

  for (const line of lines) {
    // Top-level title
    if (line.startsWith('# ') && !line.startsWith('## ') && !title) {
      title = line.replace(/^# /, '').trim()
      inPreamble = true
      continue
    }

    // Section header
    if (line.startsWith('## ')) {
      if (currentSection) {
        currentSection.content = currentSection.content.trim()
        sections.push(currentSection)
      }
      currentSection = { name: line.replace(/^## /, '').trim(), content: '' }
      inPreamble = false
      continue
    }

    // Accumulate content
    if (currentSection) {
      currentSection.content += line + '\n'
    } else if (inPreamble && title) {
      preamble += line + '\n'
    }
  }

  // Push last section
  if (currentSection) {
    currentSection.content = currentSection.content.trim()
    sections.push(currentSection)
  }

  return { title, preamble: preamble.trim(), sections }
}

/**
 * Parse a bullet list section into an array of strings.
 * Handles both "- item" and "- key: value" formats.
 *
 * @param {string} content - Section content with bullet points
 * @returns {string[]}
 */
export function parseBulletList(content) {
  if (!content) return []
  return content
    .split('\n')
    .filter(line => line.trim().startsWith('-'))
    .map(line => line.replace(/^\s*-\s*/, '').trim())
    .filter(Boolean)
}

/**
 * Serialize an array of strings back to a markdown bullet list.
 *
 * @param {string[]} items
 * @returns {string}
 */
export function serializeBulletList(items) {
  if (!items || items.length === 0) return ''
  return items.map(item => `- ${item}`).join('\n')
}

/**
 * Serialize structured data back to a markdown file.
 *
 * @param {string} title - The # title
 * @param {string} preamble - Text between title and first section (optional)
 * @param {Array<{ name: string, content: string }>} sections - Ordered sections
 * @returns {string}
 */
export function serializeToMarkdown(title, preamble, sections) {
  let md = `# ${title}\n`

  if (preamble) {
    md += `\n${preamble}\n`
  }

  for (const section of sections) {
    if (section.name && section.content) {
      md += `\n## ${section.name}\n\n${section.content}\n`
    }
  }

  return md
}

/**
 * Get a section's content by name (case-insensitive).
 *
 * @param {Array<{ name: string, content: string }>} sections
 * @param {string} name
 * @returns {string}
 */
export function getSection(sections, name) {
  const section = sections.find(s => s.name.toLowerCase() === name.toLowerCase())
  return section ? section.content : ''
}

/**
 * Set or create a section's content by name.
 * Returns a new sections array (immutable).
 *
 * @param {Array<{ name: string, content: string }>} sections
 * @param {string} name
 * @param {string} content
 * @returns {Array<{ name: string, content: string }>}
 */
export function setSection(sections, name, content) {
  const idx = sections.findIndex(s => s.name.toLowerCase() === name.toLowerCase())
  const newSections = [...sections]
  if (idx >= 0) {
    newSections[idx] = { ...newSections[idx], content }
  } else if (content) {
    newSections.push({ name, content })
  }
  return newSections
}
