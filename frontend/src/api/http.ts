import axios from 'axios'
import { toast } from 'sonner'
import { useUserStore } from '@/stores/user-store'
import type { ApiResponse } from './types'
import type { AxiosRequestConfig } from 'axios'

/*
 * ╔═══════════════════════════════════════════════════════════════╗
 * ║                    错误提示规范 (必读)                         ║
 * ╠═══════════════════════════════════════════════════════════════╣
 * ║                                                               ║
 * ║  1. 拦截器禁止弹 toast                                        ║
 * ║     拦截器只负责：认证跳转 + reject 错误                       ║
 * ║     不要在 handleGlobalError 里加 toast.error                  ║
 * ║                                                               ║
 * ║  2. 错误提示由调用方 .catch() 负责                             ║
 * ║     标准写法：.catch((err) => toast.error(err.message))        ║
 * ║     err.message 即后端返回的具体错误原因                        ║
 * ║                                                               ║
 * ║  3. 未 catch 的错误由 unhandledrejection 全局兜底              ║
 * ║     有 .catch() → 调用方弹，不触发全局                         ║
 * ║     无 .catch() → 全局弹，不会 Uncaught                        ║
 * ║     两条路径互斥，不会重复弹                                   ║
 * ║                                                               ║
 * ╚═══════════════════════════════════════════════════════════════╝
 */

const instance = axios.create({
  baseURL: '/api',
  timeout: 120_000,
  headers: { 'Content-Type': 'application/json' },
})

instance.interceptors.request.use((config) => {
  const token = useUserStore.getState().token
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

// ⚠️ 禁止在此函数内添加 toast.error —— 错误提示统一由调用方 .catch() 或全局 unhandledrejection 负责
function handleGlobalError(code: number, _msg: string) {
  switch (code) {
    case 10001:
    case 10003:
      useUserStore.getState().clearAuth()
      if (window.location.pathname !== '/login') {
        window.location.href = '/login'
      }
      break
    case 10004:
      if (window.location.pathname !== '/install') {
        window.location.href = '/install'
      }
      break
  }
}

// 全局兜底：未被 .catch() 捕获的 API 错误在此弹 toast，已 catch 的不会走到这里
window.addEventListener('unhandledrejection', (e: PromiseRejectionEvent) => {
  if (e.reason instanceof Error && e.reason.message) {
    toast.error(e.reason.message)
  }
})

instance.interceptors.response.use(
  (response) => {
    const body = response.data as ApiResponse
    if (body.code !== undefined && body.code !== 0) {
      handleGlobalError(body.code, body.msg)
      return Promise.reject(new Error(body.msg || `API error: ${body.code}`))
    }
    if (body.data !== undefined) {
      return body.data as unknown
    }
    return response.data
  },
  (error) => {
    if (error.response?.status === 401) {
      useUserStore.getState().clearAuth()
      if (window.location.pathname !== '/login') {
        window.location.href = '/login'
      }
    }
    return Promise.reject(error)
  },
)

const http = {
  get<T = unknown>(url: string, config?: AxiosRequestConfig) {
    return instance.get<T>(url, config) as Promise<T>
  },
  post<T = unknown>(url: string, data?: unknown, config?: AxiosRequestConfig) {
    return instance.post<T>(url, data, config) as Promise<T>
  },
  put<T = unknown>(url: string, data?: unknown, config?: AxiosRequestConfig) {
    return instance.put<T>(url, data, config) as Promise<T>
  },
  delete<T = unknown>(url: string, config?: AxiosRequestConfig) {
    return instance.delete<T>(url, config) as Promise<T>
  },
}

export default http
