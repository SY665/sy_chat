import { defineStore } from "pinia";
import { getHealth } from "@/lib/api/health";

export type ThemePreference = 'light' | 'dark' | 'system'
type ResolvedTheme = 'light' | 'dark'

const themeStorageKey = 'sy-chat-theme'

const sidebarStorageKey = 'sy-chat-sidebar-collapsed'

function readThemePreference(): ThemePreference {
    try {
        const savedTheme = window.localStorage.getItem(themeStorageKey)

        if (
            savedTheme === 'light'
            || savedTheme === 'dark'
            || savedTheme === 'system'
        ) {
            return savedTheme
        }
    } catch {
        // 本地存储不可用时使用默认设置。
    }

    return 'system'
}

function readSidebarCollapsed(): boolean {
    try {
        return window.localStorage.getItem(sidebarStorageKey) === 'true'
    } catch {
        return false
    }
}

export const useAppStore = defineStore('app', {
    state: () => ({
        themePreference: readThemePreference(),
        theme: 'light' as ResolvedTheme,
        apiStatus: 'checking' as 'checking' | 'online' | 'offline',
        sidebarCollapsed: readSidebarCollapsed(),
    }),

    getters: {
        isDark: (state) => state.theme === 'dark'
    },

    actions: {
        toggleSidebar() {
            this.sidebarCollapsed = !this.sidebarCollapsed

            try {
                window.localStorage.setItem(
                    sidebarStorageKey,
                    String(this.sidebarCollapsed),
                )
            } catch {
                // 存储不可用时，当前页面内仍然可以正常切换。
            }
        },

        initializeTheme() {
            this.applyTheme()

            const systemTheme = window.matchMedia('(prefers-color-scheme: dark)')
            systemTheme.addEventListener('change', () => {
                if (this.themePreference === 'system') {
                    this.applyTheme()
                }
            })
        },

        applyTheme() {
            const systemPrefersDark = window.matchMedia(
                '(prefers-color-scheme: dark)',
            ).matches

            const shouldUseDark =
                this.themePreference === 'dark'
                || (
                    this.themePreference === 'system'
                    && systemPrefersDark
                )

            this.theme = shouldUseDark ? 'dark' : 'light'
            document.documentElement.classList.toggle('dark', shouldUseDark)
        },
        setThemePreference(preference: ThemePreference) {
            this.themePreference = preference

            try {
                window.localStorage.setItem(themeStorageKey, preference)
            } catch {
                // 存储失败不影响当前页面切换主题。
            }

            this.applyTheme()
        },


        toggleTheme() {
            this.setThemePreference(this.isDark ? 'light' : 'dark')
        },

        async checkApiHealth() {
            this.apiStatus = 'checking'

            try {
                const health = await getHealth()

                this.apiStatus =
                    health.status === 'ok'
                        ? 'online'
                        : 'offline'
            } catch (error) {
                console.error('API health check failed:', error)
                this.apiStatus = 'offline'
            }
        },
    }
})