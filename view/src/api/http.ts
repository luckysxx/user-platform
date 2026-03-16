export interface ApiEnvelope<T> {
  code: number
  msg: string
  data: T | null
}

const RAW_BASE_URL = import.meta.env.VITE_API_BASE_URL?.trim() ?? ''
const API_BASE_URL = RAW_BASE_URL.replace(/\/+$/, '')

function buildApiUrl(path: string): string {
  if (/^https?:\/\//.test(path)) {
    return path
  }

  const normalizedPath = path.startsWith('/') ? path : `/${path}`
  return API_BASE_URL ? `${API_BASE_URL}${normalizedPath}` : normalizedPath
}

export async function postJson<TRequest, TResponse>(path: string, payload: TRequest): Promise<TResponse> {
  const response = await fetch(buildApiUrl(path), {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(payload),
  })

  let envelope: ApiEnvelope<TResponse> | null = null
  try {
    envelope = (await response.json()) as ApiEnvelope<TResponse>
  } catch {
    throw new Error(`服务响应解析失败（HTTP ${response.status}）`)
  }

  if (!response.ok) {
    throw new Error(envelope.msg || `请求失败（HTTP ${response.status}）`)
  }

  if (envelope.code !== 200 || envelope.data === null) {
    throw new Error(envelope.msg || '请求失败')
  }

  return envelope.data
}
