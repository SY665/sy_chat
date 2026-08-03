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
 * 发送统一格式的 API 请求。
 *
 * 泛型 T 表示成功响应中 data 字段的类型。
 */
export async function apiRequest<T>(
  path: string,
  options: RequestInit = {},
): Promise<T> {
  const normalizedPath = path.startsWith('/') ? path : `/${path}`
  const headers = new Headers(options.headers)

  // 请求包含 body 时，默认按照 JSON 发送。
  if (options.body && !headers.has('Content-Type')) {
    headers.set('Content-Type', 'application/json')
  }

  let response: Response

  try {
    response = await fetch(`${apiBaseUrl}${normalizedPath}`, {
      ...options,
      headers,

      // 后续登录使用 HttpOnly Cookie，因此所有请求允许携带凭证。
      credentials: 'include',
    })
  } catch {
    throw new ApiRequestError(
      0,
      'NETWORK_ERROR',
      '无法连接到后端服务',
    )
  }

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