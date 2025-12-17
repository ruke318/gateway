import request, { authRequest } from './request'

// ============ 认证和用户管理 API (使用 JWT) ============

// 登录
export const login = (data) => authRequest.post('/auth/login', data)

// 获取当前用户信息
export const getCurrentUser = () => authRequest.get('/auth/me')

// 修改密码
export const changePassword = (data) => authRequest.post('/auth/change-password', data)

// 用户管理
export const getUsers = (params) => authRequest.get('/users', { params })
export const createUser = (data) => authRequest.post('/users', data)
export const updateUser = (id, data) => authRequest.put(`/users/${id}`, data)
export const deleteUser = (id) => authRequest.delete(`/users/${id}`)
export const resetUserPassword = (id, data) => authRequest.post(`/users/${id}/reset-password`, data)

// 操作日志
export const getOperationLogs = (params) => authRequest.get('/operation-logs', { params })
export const getOperationLog = (id) => authRequest.get(`/operation-logs/${id}`)
export const getOperationStatistics = () => authRequest.get('/operation-logs/statistics')

// ============ 管理后台 API (使用 X-Admin-Token) ============

// 厂商
export const vendorApi = {
  list: (params) => request.get('/vendors', { params }),
  get: (id) => request.get(`/vendor/${id}`),
  create: (data) => request.post('/vendor', data),
  update: (id, data) => request.put(`/vendor/${id}`, data),
  delete: (id) => request.delete(`/vendor/${id}`)
}

// 机构
export const organizationApi = {
  list: (params) => request.get('/organizations', { params }),
  get: (id) => request.get(`/organization/${id}`),
  create: (data) => request.post('/organization', data),
  update: (id, data) => request.put(`/organization/${id}`, data),
  delete: (id) => request.delete(`/organization/${id}`)
}

// 接口
export const serviceApi = {
  list: (params) => request.get('/services', { params }),
  get: (id) => request.get(`/service/${id}`),
  create: (data) => request.post('/service', data),
  update: (id, data) => request.put(`/service/${id}`, data),
  delete: (id) => request.delete(`/service/${id}`)
}

// 公共函数库
export const scriptApi = {
  list: (params) => request.get('/scripts', { params }),
  get: (id) => request.get(`/script/${id}`),
  create: (data) => request.post('/script', data),
  update: (id, data) => request.put(`/script/${id}`, data),
  delete: (id) => request.delete(`/script/${id}`)
}

// Hook脚本
export const hookScriptApi = {
  list: (params) => request.get('/hook-scripts', { params }),
  get: (id) => request.get(`/hook-script/${id}`),
  create: (data) => request.post('/hook-script', data),
  update: (id, data) => request.put(`/hook-script/${id}`, data),
  delete: (id) => request.delete(`/hook-script/${id}`)
}

// 接口Hook关联
export const serviceHookApi = {
  list: () => request.get('/service-hooks'),
  listByService: (serviceId) => request.get(`/service-hooks?service_id=${serviceId}`),
  get: (id) => request.get(`/service-hook/${id}`),
  create: (data) => request.post('/service-hook', data),
  update: (id, data) => request.put(`/service-hook/${id}`, data),
  delete: (id) => request.delete(`/service-hook/${id}`)
}

// 重载函数库
export const reloadLibrary = () => request.post('/reload-library')

// 字典配置
export const dictionaryConfigApi = {
  list: (params) => request.get('/dictionary-configs', { params }),
  get: (id) => request.get(`/dictionary-config/${id}`),
  create: (data) => request.post('/dictionary-config', data),
  update: (id, data) => request.put(`/dictionary-config/${id}`, data),
  delete: (id) => request.delete(`/dictionary-config/${id}`),
  batchCreate: (data) => request.post('/dictionary-configs/batch', data)
}

// 重载字典
export const reloadDictionary = () => request.post('/reload-dictionary')

