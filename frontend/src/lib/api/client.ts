import type { ApiEnvelope } from "./types"

const apiBaseUrl = (
  import.meta.env.VITE_API_BASE_URL ||
  'http://localhost:8080/api/v1'
).replace(/\/+$/, '')

/**
 * 表示后端返回的业务错误、HTTP 错误或网络错误。
 */
export class ApiRequestError extends Error {
  // 1. 显式声明属性及其类型
  public readonly status: number
  public readonly code: string

  constructor(
    status: number,
    code: string,
    message: string,
  ) {
    super(message)
    // 2. 显式赋值
    this.status = status
    this.code = code
    this.name = 'ApiRequestError'
  }
}

/**
 * 发送请求并保留原始 Response。
 *
 * 普通 JSON 请求和流式请求共用地址、请求头及 Cookie 配置。
 */
export async function apiFetch(
  path: string,
  options: RequestInit = {},
): Promise<Response> {
  const normalizedPath = path.startsWith('/') ? path : `/${path}`
  const headers = new Headers(options.headers)

  if (options.body && !headers.has('Content-Type')) {
    headers.set('Content-Type', 'application/json')
  }

  try {
    return await fetch(`${apiBaseUrl}${normalizedPath}`, {
      ...options,
      headers,
      credentials: 'include',
    })
  } catch (error) {
    // 主动取消不是网络故障，交给调用方单独处理。
    if (isAbortError(error)) {
      throw error
    }

    throw new ApiRequestError(
      0,
      'NETWORK_ERROR',
      '无法连接到后端服务',
    )
  }
}

/**
 * 发送统一格式的 API 请求。
 *
 * 泛型 T 表示成功响应中 data 字段的类型。
 */
export async function apiRequest<T>(
  path: string,
  options: RequestInit = {},
): Promise<T> {
  const response = await apiFetch(path, options)

  let payload: ApiEnvelope<T>

  try {
    payload = (await response.json()) as ApiEnvelope<T>
  } catch {
    throw new ApiRequestError(
      response.status,
      'INVALID_RESPONSE',
      '后端返回了无法解析的数据',
    )
  }

  if (!response.ok || !payload.success) {
    throw new ApiRequestError(
      response.status,
      payload.error?.code ?? 'REQUEST_FAILED',
      payload.error?.message ?? '请求失败',
    )
  }

  if (payload.data === undefined) {
    throw new ApiRequestError(
      response.status,
      'EMPTY_RESPONSE',
      '后端响应缺少 data 字段',
    )
  }

  return payload.data
}

/**
 * 判断错误是否来自 AbortController 主动取消。
 */
export function isAbortError(error: unknown): boolean {
  return (
    error instanceof DOMException &&
    error.name === 'AbortError'
  )
}