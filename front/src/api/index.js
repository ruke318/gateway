import request from './request'

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
