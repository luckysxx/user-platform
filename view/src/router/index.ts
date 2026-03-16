import { createRouter, createWebHistory } from 'vue-router'
import SsoRegisterView from '@/views/SsoRegisterView.vue'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      name: 'sso-register',
      component: SsoRegisterView,
    },
  ],
})

export default router
