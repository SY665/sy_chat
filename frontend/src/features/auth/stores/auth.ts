import { defineStore } from "pinia";

import {
    getCurrentUser,
    login as loginRequest,
    logout as logoutRequest,
    register as registerRequest,
} from "@/lib/api/auth"
import { ApiRequestError } from '@/lib/api/client'
import type {
    AuthUser,
    LoginInput,
    RegisterInput,
} from '@/features/auth/types'

type AuthStatus = 'idle' | 'loading' | 'authenticated' | 'guest'

export const useAuthStore = defineStore('auth', {
    state: () => ({
        user: null as AuthUser | null,
        status: 'idle' as AuthStatus,
        initialized: false,
        error: null as string | null,
    }),

    getters: {
        isAuthenticated: (state) => state.status === 'authenticated',
        isLoading: (state) => state.status === 'loading',
    },

    actions: {
        // initialize 通过 HttpOnly Cookie 恢复刷新前的登录状态。
        async initialize() {
            if (this.initialized) {
                return
            }

            this.status = 'loading'
            this.error = null

            try {
                const data = await getCurrentUser()

                this.user = data.user
                this.status = 'authenticated'
            } catch (error) {
                this.user = null
                this.status = 'guest'

                // 未登录是正常状态，不需要向用户显示错误。
                if (!(error instanceof ApiRequestError && error.status === 401)) {
                    this.error = getErrorMessage(error)
                }
            } finally {
                this.initialized = true
            }
        },

        async register(input: RegisterInput) {
            this.status = 'loading'
            this.error = null

            try {
                const data = await registerRequest(input)

                this.user = data.user
                this.status = 'authenticated'
                this.initialized = true

                return data.user
            } catch (error) {
                this.user = null
                this.status = 'guest'
                this.error = getErrorMessage(error)
                throw error
            }
        },

        async login(input: LoginInput) {
            this.status = 'loading'
            this.error = null

            try {
                const data = await loginRequest(input)

                this.user = data.user
                this.status = 'authenticated'
                this.initialized = true

                return data.user
            } catch (error) {
                this.user = null
                this.status = 'guest'
                this.error = getErrorMessage(error)
                throw error
            }
        },

        async logout() {
            this.error = null

            try {
                await logoutRequest()

                // 只有后端成功删除 Cookie 后，才清空前端用户状态。
                this.user = null
                this.status = 'guest'
                this.initialized = true
            } catch (error) {
                this.error = getErrorMessage(error)
                throw error
            }
        },

        clearError() {
            this.error = null
        },
    },
})

function getErrorMessage(error: unknown): string {
    if (error instanceof ApiRequestError) {
        return error.message
    }

    return '请求失败，请稍后重试'
}

