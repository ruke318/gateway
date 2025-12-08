import { createRouter, createWebHistory } from 'vue-router'
import MainLayout from '../layouts/MainLayout.vue'

const routes = [
  {
    path: '/',
    component: MainLayout,
    redirect: '/vendor',
    children: [
      { path: 'vendor', name: 'Vendor', component: () => import('../views/Vendor.vue') },
      { path: 'organization', name: 'Organization', component: () => import('../views/Organization.vue') },
      { path: 'service', name: 'Service', component: () => import('../views/Service.vue') },
      { path: 'script', name: 'Script', component: () => import('../views/Script.vue') },
      { path: 'hook-script', name: 'HookScript', component: () => import('../views/HookScript.vue') }
    ]
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

export default router
