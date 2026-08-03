/**
 * 后端返回的结构化错误。
 */
export interface ApiError {
  code: string
  message: string
}

/**
 * 对应 Go 后端 response.Envelope。
 */
export interface ApiEnvelope<T> {
  success: boolean
  data?: T
  error?: ApiError
}