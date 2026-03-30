const imageExtensions = ['.png', '.jpg', '.jpeg', '.gif', '.webp', '.svg']

export function normalizePath(path = '') {
  return String(path || '').replace(/\\/g, '/')
}

export function isImagePath(path = '') {
  const normalized = normalizePath(path).split('?')[0].split('#')[0].toLowerCase()
  return imageExtensions.some(ext => normalized.endsWith(ext))
}

export function extractImageResultData(raw) {
  if (!raw) return null

  if (typeof raw === 'string') {
    try {
      const parsed = JSON.parse(raw)
      return extractImageResultData(parsed)
    } catch {
      return isImagePath(raw) ? { url: raw } : null
    }
  }

  if (typeof raw !== 'object') return null

  const candidatePath = raw.local_path || raw.url || ''
  if (!candidatePath || !isImagePath(candidatePath)) return null

  return raw
}

export function deriveWorkspaceRelativePath(path = '', workspaceRoot = '') {
  const normalizedPath = normalizePath(path)
  const normalizedWorkspaceRoot = normalizePath(workspaceRoot).replace(/\/+$/, '')

  if (!normalizedPath) return ''

  if (normalizedWorkspaceRoot) {
    const lowerPath = normalizedPath.toLowerCase()
    const lowerRoot = normalizedWorkspaceRoot.toLowerCase()
    if (lowerPath.startsWith(lowerRoot)) {
      return normalizedPath.slice(normalizedWorkspaceRoot.length).replace(/^\/+/, '')
    }
  }

  const workspaceMarkerIndex = normalizedPath.toLowerCase().lastIndexOf('/workspace/')
  if (workspaceMarkerIndex >= 0) {
    return normalizedPath.slice(workspaceMarkerIndex + '/workspace/'.length).replace(/^\/+/, '')
  }

  return normalizedPath.replace(/^\/+/, '')
}

export function buildFilesApiUrl(path = '', token = '', options = {}) {
  const { download = false, workspaceRoot = '' } = options

  if (!path) return ''
  if (/^https?:\/\//i.test(path)) return path

  const relPath = deriveWorkspaceRelativePath(path, workspaceRoot)
  const encodedPath = encodeURIComponent(relPath).replace(/%2F/g, '/')
  const params = new URLSearchParams()

  if (download) params.set('download', 'true')
  if (token) params.set('token', token)

  const query = params.toString()
  return query ? `/api/v1/files/${encodedPath}?${query}` : `/api/v1/files/${encodedPath}`
}

export { imageExtensions }
