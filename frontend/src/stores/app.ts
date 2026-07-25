import { defineStore } from "pinia";

type Theme = 'light' | 'dark'

export const useAppStore = defineStore('app', {
    state: () => ({
        theme: 'light' as Theme
    }),

    getters: {
        isDark: (state) => state.theme === 'dark'
    },

    actions: {
        initializeTheme() {
            const savedTheme = localStorage.getItem('sy-chat-theme')

            if (savedTheme === 'light' || savedTheme === 'dark') {
                this.theme = savedTheme
            } else {
                this.theme = window.matchMedia('(prefers-color-scheme: dark)').matches
                    ? 'dark'
                    : 'light'
            }
            this.applyTheme()
        },

        applyTheme() {
            document.documentElement.classList.toggle('dark', this.isDark)
        },

        toggleTheme() {
            this.theme = this.isDark ? 'light' : 'dark'
            localStorage.setItem('sy-chat-theme', this.theme)
            this.applyTheme()
        }
    }
})