import { createRouter, createWebHistory } from 'vue-router'
import AdminLayout from '../layouts/AdminLayout.vue'
import {redirectToPortal, ensureSession, clearToken} from '../utils/auth'

const APP_TITLE = 'MaterialCore - 素材中心'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/auth/callback',
      name: 'AuthCallback',
      component: () => import('../views/AuthCallback.vue'),
      meta: { public: true },
    },
    {
      path: '/auth/logout',
      name: 'AuthLogout',
      component: () => import('../views/AuthLogout.vue'),
      meta: { public: true },
    },
    {
      path: '/m/photo-upload',
      name: 'MobilePhotoUpload',
      component: () => import('../views/MobilePhotoUpload.vue'),
      meta: { public: true, title: '扫码上传' },
    },
    {
      path: '/',
      component: AdminLayout,
      redirect: '/dashboard',
      children: [
        { path: 'dashboard', name: 'Dashboard', component: () => import('../views/Dashboard.vue'), meta: { title: '工作台' } },
        { path: 'materials', name: 'Materials', component: () => import('../views/Materials.vue'), meta: { title: '素材库' } },
        { path: 'categories', name: 'Categories', component: () => import('../views/Categories.vue'), meta: { title: '分类管理' } },
        { path: 'inquiry', name: 'Inquiry', component: () => import('../views/Inquiry.vue'), meta: { title: '询盘工作台' } },
      ],
    },
  ],
})

router.beforeEach(async (to) => {
  if (to.meta.public) return true
  const ok = await ensureSession()
  if (!ok) {
    clearToken()
    redirectToPortal()
    return false
  }
  return true
})

router.afterEach((to) => {
  const page = to.meta.title as string | undefined
  document.title = page ? `${page} - ${APP_TITLE}` : APP_TITLE
})

export default router
