import { ApiRequestError } from './client'

export interface ServerSentEvent {
  event: string
  data: string
}

type EventHandler = (event: ServerSentEvent) => void

/**
 * 将 Response 的字节流解析为完整的 SSE 事件。
 *
 * 一次网络读取不一定对应一个 SSE 事件，因此需要使用 buffer
 * 保存尚未接收完整的数据。
 */
export async function readEventStream(
  response: Response,
  onEvent: EventHandler,
): Promise<void> {
  if (!response.body) {
    throw new ApiRequestError(
      response.status,
      'EMPTY_STREAM',
      '后端没有返回流式内容',
    )
  }

  const reader = response.body.getReader()
  const decoder = new TextDecoder()
  let buffer = ''

  try {
    while (true) {
      const { value, done } = await reader.read()

      buffer += decoder.decode(value, {
        stream: !done,
      })

      // 后端可能使用 CRLF，统一为换行符后再寻找事件边界。
      buffer = buffer.replace(/\r\n/g, '\n')

      let boundary = buffer.indexOf('\n\n')

      while (boundary !== -1) {
        const eventBlock = buffer.slice(0, boundary)
        buffer = buffer.slice(boundary + 2)

        dispatchEventBlock(eventBlock, onEvent)
        boundary = buffer.indexOf('\n\n')
      }

      if (done) {
        break
      }
    }

    // 兼容连接关闭前没有额外空行的最后一个事件。
    if (buffer.trim()) {
      dispatchEventBlock(buffer, onEvent)
    }
  } finally {
    reader.releaseLock()
  }
}

function dispatchEventBlock(
  block: string,
  onEvent: EventHandler,
) {
  let eventName = 'message'
  const dataLines: string[] = []

  for (const line of block.split('\n')) {
    if (line.startsWith('event:')) {
      eventName = line.slice('event:'.length).trim()
      continue
    }

    if (line.startsWith('data:')) {
      dataLines.push(
        line.slice('data:'.length).replace(/^ /, ''),
      )
    }
  }

  // 心跳或注释事件可能没有 data，不需要交给业务层。
  if (dataLines.length === 0) {
    return
  }

  onEvent({
    event: eventName,
    data: dataLines.join('\n'),
  })
}