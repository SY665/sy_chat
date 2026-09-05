import type { ChatMessage } from '@/features/chat/types'

function createSafeFilename(title: string): string {
    const filename = title
        .trim()
        .replace(/[<>:"/\\|?*\u0000-\u001F]/g, '-')
        .replace(/\s+/g, ' ')
        .slice(0, 80)

    return filename || 'SY Chat 会话'
}

export function exportConversationAsMarkdown(
    title: string,
    messages: readonly ChatMessage[],
) {
    const sections = messages
        .filter((message) => message.content.trim() !== '')
        .map((message) => {
            const speaker = message.role === 'user' ? '用户' : 'SY Chat'
            return `## ${speaker}\n\n${message.content.trim()}`
        })

    const markdown = [
        `# ${title}`,
        '',
        `> 导出时间：${new Date().toLocaleString('zh-CN')}`,
        '',
        ...sections,
        '',
    ].join('\n')

    // BOM 能改善 Windows 文本编辑器打开中文 Markdown 时的兼容性。
    const blob = new Blob(
        ['\uFEFF', markdown],
        { type: 'text/markdown;charset=utf-8' },
    )
    const url = URL.createObjectURL(blob)
    const link = document.createElement('a')

    link.href = url
    link.download = `${createSafeFilename(title)}.md`
    document.body.appendChild(link)
    link.click()
    link.remove()
    URL.revokeObjectURL(url)
}