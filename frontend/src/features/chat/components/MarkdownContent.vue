<script setup lang="ts">
import { computed, h, type VNode } from 'vue'
import MarkdownIt from 'markdown-it'
import CodeBlock from './CodeBlock.vue'

const props = defineProps<{ content: string }>()
const markdown = new MarkdownIt({
    html: false,
    linkify: true,
    breaks: true,
})

const defaultLinkOpen = markdown.renderer.rules.link_open

markdown.renderer.rules.link_open = (
    tokens, index, options, env, renderer,
) => {
    const token = tokens[index]!
    const href = token.attrGet('href')

    if (typeof href === 'string' && href.length > 0) {
        try {
            const url = new URL(href, window.location.href)
            const isExternal =
                (url.protocol === 'http:' || url.protocol === 'https:') &&
                url.origin !== window.location.origin

            if (isExternal) {
                token.attrSet('target', '_blank')
                // 新页面不持有当前聊天窗口的引用。
                token.attrSet('rel', 'noopener noreferrer')
            }
        } catch {
            // 无法解析的地址保留解析器原有的渲染行为。
        }
    }

    return defaultLinkOpen
        ? defaultLinkOpen(tokens, index, options, env, renderer)
        : renderer.renderToken(tokens, index, options)
}

const nodes = computed(() => {
    const env = {}
    const tokens = markdown.parse(props.content, env)
    let index = 0

    // 按开始、结束标记递归构建元素，保留列表、引用和表格的嵌套。
    function renderChildren(): VNode[] {
        const children: VNode[] = []
        while (index < tokens.length) {
            const token = tokens[index++]!
            if (token.nesting === -1) break

            if (token.type === 'fence' || token.type === 'code_block') {
                children.push(h(CodeBlock, {
                    key: index,
                    code: token.content,
                    language: token.info,
                }))
            } else if (token.type === 'inline') {
                children.push(h('span', {
                    innerHTML: markdown.renderer.renderInline(
                        token.children ?? [], markdown.options, env,
                    ),
                }))
            } else if (token.nesting === 1) {
                const nested = renderChildren()
                if (token.hidden) children.push(...nested)
                else children.push(h(
                    token.tag, Object.fromEntries(token.attrs ?? []), nested,
                ))
            } else if (!token.hidden && token.tag) {
                children.push(h(token.tag, Object.fromEntries(token.attrs ?? [])))
            }
        }
        return children
    }

    return renderChildren()
})

const MarkdownBody = () => nodes.value
</script>

<template>
    <div class="markdown-content">
        <MarkdownBody />
    </div>
</template>

<style scoped>
.markdown-content {
    overflow-wrap: anywhere;
    line-height: 1.625
}

.markdown-content :deep(p),
.markdown-content :deep(ul),
.markdown-content :deep(ol),
.markdown-content :deep(blockquote) {
    margin: 0.75rem 0;
}

.markdown-content :deep(h1),
.markdown-content :deep(h2),
.markdown-content :deep(h3),
.markdown-content :deep(h4),
.markdown-content :deep(h5),
.markdown-content :deep(h6) {
    margin: 1rem 0 0.5rem;
    font-weight: 600;
    line-height: 1.4;
}

.markdown-content :deep(h1) {
    font-size: 1.375rem;
}

.markdown-content :deep(h2) {
    font-size: 1.25rem;
}

.markdown-content :deep(h3) {
    font-size: 1.125rem;
}

.markdown-content :deep(h4),
.markdown-content :deep(h5),
.markdown-content :deep(h6) {
    font-size: 1rem;
}

.markdown-content :deep(blockquote) {
    border-left: 3px solid #737373;
    padding-left: 1rem;
    color: #737373;
}

/* 宽表格在消息内部滚动，避免撑宽整个聊天页面。 */
.markdown-content :deep(table) {
    display: block;
    max-width: 100%;
    overflow-x: auto;
    margin: 0.75rem 0;
    border-collapse: collapse;
}

.markdown-content :deep(th),
.markdown-content :deep(td) {
    border: 1px solid #737373;
    padding: 0.5rem 0.75rem;
    min-width: 6rem;
}

.markdown-content :deep(th) {
    font-weight: 600;
    background: rgb(115 115 115 / 12%);
}

.markdown-content :deep(hr) {
    margin: 1rem 0;
    border: 0;
    border-top: 1px solid #737373;
}

/* 只调整正文的直接子元素，避免影响代码块工具栏内部布局。 */
.markdown-content> :deep(:first-child) {
    margin-top: 0;
}

.markdown-content> :deep(:last-child) {
    margin-bottom: 0;
}

.markdown-content :deep(ul),
.markdown-content :deep(ol) {
    padding-left: 1.5rem;
}

.markdown-content :deep(ul) {
    list-style: disc;
}

.markdown-content :deep(ol) {
    list-style: decimal;
}

.markdown-content :deep(:not(pre) > code) {
    border-radius: 0.25rem;
    background: rgb(229 229 229);
    padding: 0.125rem 0.375rem;
    font-size: 0.875em;
}

.dark .markdown-content :deep(:not(pre) > code) {
    background: rgb(38 38 38);
}

.markdown-content :deep(a) {
    color: rgb(5 150 105);
    text-decoration: underline;
    text-underline-offset: 2px;
}
</style>