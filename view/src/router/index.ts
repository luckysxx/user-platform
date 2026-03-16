import { createRouter, createWebHistory } from 'vue-router'
import SsoLoginView from '@/views/SsoLoginView.vue'
import SsoRegisterView from '@/views/SsoRegisterView.vue'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/register',
      alias: '/',
      name: 'sso-register',
      component: SsoRegisterView,
    },
    {
      path: '/login',
      name: 'sso-login',
      component: SsoLoginView,
    },
  ],
})

export default router
