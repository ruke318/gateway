import axios from 'axios'
import { ElMessage } from 'element-plus'

const request = axios.create({
  baseURL: '/admin/db',
  timeout: 10000
})

request.interceptors.request.use(config => {
  config.headers['X-Admin-Token'] = localStorage.getItem('adminToken') || 'testToken'
  return config
})

request.interceptors.response.use(
  response => {
    const res = response.data
    if (res.code !== 0) {
      ElMessage.error(res.message || '请求失败')
      return Promise.reject(res)
    }
    return res
  },
  error => {
    ElMessage.error(error.message || '网络错误')
    return Promise.reject(error)
  }
)

export default request
