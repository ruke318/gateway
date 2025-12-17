import { createRouter, createWebHistory } from 'vue-router'
import MainLayout from '../layouts/MainLayout.vue'

const routes = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('../views/Login.vue')
  },
  {
    path: '/',
    component: MainLayout,
    redirect: '/vendor',
    children: [
      { path: 'vendor', name: 'Vendor', component: () => import('../views/Vendor.vue') },
      { path: 'organization', name: 'Organization', component: () => import('../views/Organization.vue') },
      { path: 'service', name: 'Service', component: () => import('../views/Service.vue') },
      { path: 'script', name: 'Script', component: () => import('../views/Script.vue') },
      { path: 'hook-script', name: 'HookScript', component: () => import('../views/HookScript.vue') },
      { path: 'dictionary-config', name: 'DictionaryConfig', component: () => import('../views/DictionaryConfig.vue') },
      { path: 'user', name: 'User', component: () => import('../views/User.vue') },
      { path: 'operation-log', name: 'OperationLog', component: () => import('../views/OperationLog.vue') }
    ]
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

// 路由守卫
router.beforeEach((to, from, next) => {
  const token = localStorage.getItem('token')

  if (to.path === '/login') {
    // 如果已登录，跳转到首页
    if (token) {
      next('/')
    } else {
      next()
    }
  } else {
    // 其他页面需要登录
    if (!token) {
      next('/login')
    } else {
      next()
    }
  }
})

export default router
