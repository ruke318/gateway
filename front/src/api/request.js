import axios from 'axios'
import { ElMessage } from 'element-plus'

// 管理后台 API (使用 JWT 认证)
const request = axios.create({
  baseURL: '/admin/db',
  timeout: 10000
})

request.interceptors.request.use(config => {
  const token = localStorage.getItem('token')
  if (token) {
    config.headers['Authorization'] = `Bearer ${token}`
  }
  return config
})

request.interceptors.response.use(
  response => {
    const res = response.data
    if (res.code !== 0) {
      ElMessage.error(res.message || '请求失败')
      // 如果是401，跳转到登录页
      if (res.code === 401) {
        localStorage.clear()
        window.location.href = '/login'
      }
      return Promise.reject(res)
    }
    return res
  },
  error => {
    if (error.response?.status === 401) {
      ElMessage.error('未登录或登录已过期')
      localStorage.clear()
      window.location.href = '/login'
    } else if (error.response?.status === 403) {
      ElMessage.error('权限不足，需要管理员权限')
    } else {
      ElMessage.error(error.message || '网络错误')
    }
    return Promise.reject(error)
  }
)

// 用户管理 API (使用 JWT)
const authRequest = axios.create({
  baseURL: '/api',
  timeout: 10000
})

authRequest.interceptors.request.use(config => {
  const token = localStorage.getItem('token')
  if (token) {
    config.headers['Authorization'] = `Bearer ${token}`
  }
  return config
})

authRequest.interceptors.response.use(
  response => {
    const res = response.data
    if (res.code !== 200) {
      ElMessage.error(res.message || '请求失败')
      // 如果是401，跳转到登录页
      if (res.code === 401) {
        localStorage.clear()
        window.location.href = '/login'
      }
      return Promise.reject(res)
    }
    return res
  },
  error => {
    if (error.response?.status === 401) {
      ElMessage.error('未登录或登录已过期')
      localStorage.clear()
      window.location.href = '/login'
    } else {
      ElMessage.error(error.message || '网络错误')
    }
    return Promise.reject(error)
  }
)

export default request
export { authRequest }
