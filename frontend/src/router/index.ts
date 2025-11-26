import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const router = createRouter({
  history: createWebHistory(),
  scrollBehavior: () => ({ top: 0 }),
  routes: [
    // Public
    { path: '/', component: () => import('@/pages/Home.vue') },
    { path: '/search', component: () => import('@/pages/Search.vue') },
    { path: '/posts/:id', component: () => import('@/pages/PostDetail.vue') },
    { path: '/users/:id', component: () => import('@/pages/UserProfile.vue') },

    // Auth
    { path: '/login', component: () => import('@/pages/auth/Login.vue'), meta: { guestOnly: true } },
    { path: '/register', component: () => import('@/pages/auth/Register.vue'), meta: { guestOnly: true } },
    { path: '/forgot-password', component: () => import('@/pages/auth/ForgotPassword.vue'), meta: { guestOnly: true } },

    // OAuth callbacks
    { path: '/auth/google', component: () => import('@/pages/oauth/GoogleCallback.vue') },
    { path: '/auth/github', component: () => import('@/pages/oauth/GitHubCallback.vue') },

    // Authenticated
    { path: '/posts/new', component: () => import('@/pages/PostCreate.vue'), meta: { requiresAuth: true } },
    { path: '/posts/:id/edit', component: () => import('@/pages/PostEdit.vue'), meta: { requiresAuth: true } },
    { path: '/notifications', component: () => import('@/pages/Notifications.vue'), meta: { requiresAuth: true } },
    { path: '/messages', component: () => import('@/pages/Messages.vue'), meta: { requiresAuth: true } },
    { path: '/messages/:id', component: () => import('@/pages/Conversation.vue'), meta: { requiresAuth: true } },
    { path: '/settings', component: () => import('@/pages/Settings.vue'), meta: { requiresAuth: true } },

    // Admin
    { path: '/admin', component: () => import('@/pages/admin/AdminStats.vue'), meta: { requiresAdmin: true } },
    { path: '/admin/users', component: () => import('@/pages/admin/AdminUsers.vue'), meta: { requiresAdmin: true } },

    // 404
    { path: '/:pathMatch(.*)*', component: () => import('@/pages/NotFound.vue') },
  ],
})

router.beforeEach(async (to) => {
  const auth = useAuthStore()

  if (to.meta.requiresAdmin) {
    if (!auth.isLoggedIn) return { path: '/login', query: { redirect: to.fullPath } }
    // Ensure user is loaded to check role
    if (!auth.user) await auth.fetchMe()
    if (!auth.isAdmin) return { path: '/' }
  }

  if (to.meta.requiresAuth) {
    if (!auth.isLoggedIn) return { path: '/login', query: { redirect: to.fullPath } }
  }

  if (to.meta.guestOnly && auth.isLoggedIn) {
    return { path: '/' }
  }
})

export default router
