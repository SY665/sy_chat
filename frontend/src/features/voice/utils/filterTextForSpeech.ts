/**
 * 移除代码、图片地址和 Markdown 标记，避免朗读无意义内容。
 */
export function filterTextForSpeech(content: string): string {
    return content
        .replace(/```[\s\S]*?```/g, '')
        .replace(/`[^`]*`/g, '')
        .replace(/!\[[^\]]*]\([^)]+\)/g, '')
        .replace(/\[([^\]]+)]\([^)]+\)/g, '$1')
        .replace(/\|[^\n]+\|(\n\|[-:\s|]+\|)?(\n\|[^\n]+\|)*/g, '')
        .replace(/https?:\/\/[^\s)]+/g, '')
        .replace(/<[^>]+>/g, '')
        .replace(/^#{1,6}\s+/gm, '')
        .replace(/[*_~>#-]/g, '')
        .replace(/[ \t]+/g, ' ')
        .replace(/\n{3,}/g, '\n\n')
        .trim()
}