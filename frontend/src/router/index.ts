import { createRouter, createWebHistory } from 'vue-router'

import { useAuthStore } from '@/features/auth/stores/auth'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),

  routes: [
    {
      path: '/',
      redirect: {
        name: 'home',
      },
    },
    {
      path: '/chat/:conversationId?',
      name: 'home',
      component: () => import('@/pages/HomePage.vue'),
      meta: {
        requiresAuth: true,
      },
    },
    {
      path: '/share/:token',
      name: 'public-share',
      component: () => import('@/pages/PublicSharePage.vue'),
    },
    {
      path: '/login',
      name: 'login',
      component: () => import('@/pages/LoginPage.vue'),
      meta: {
        guestOnly: true,
      },
    },
    {
      path: '/register',
      name: 'register',
      component: () => import('@/pages/RegisterPage.vue'),
      meta: {
        guestOnly: true,
      },
    },
  ],
})

router.beforeEach(async (to) => {
  if (!to.meta.requiresAuth && !to.meta.guestOnly) {
    return
  }

  const authStore = useAuthStore()

  // 浏览器刷新后，通过 /auth/me 恢复 Cookie 对应的用户。
  await authStore.initialize()

  if (to.meta.requiresAuth && !authStore.isAuthenticated) {
    return {
      name: 'login',
      query: {
        redirect: to.fullPath
      }
    }
  }

  if (to.meta.guestOnly && authStore.isAuthenticated) {
    return {
      name: 'home',
    }
  }
})

export default router