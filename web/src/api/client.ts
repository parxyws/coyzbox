import axios, { type AxiosInstance, type InternalAxiosRequestConfig } from "axios"

const API_BASE = import.meta.env.VITE_API_URL || "/api/v1"

function getToken(): string | null {
  return localStorage.getItem("token")
}

const client: AxiosInstance = axios.create({
  baseURL: API_BASE,
  headers: { "Content-Type": "application/json" },
})

client.interceptors.request.use((config: InternalAxiosRequestConfig) => {
  const token = getToken()
  if (token) {
    config.headers.set("Authorization", `Bearer ${token}`)
  }
  return config
})

client.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.data) {
      const data = error.response.data
      return Promise.reject({
        message: data.message || "An unexpected error occurred",
        errors: data.errors,
        status: error.response.status,
      })
    }
    return Promise.reject({
      message: "Network error",
      status: 0,
    })
  },
)

export const api = {
  get: <T>(endpoint: string, params?: Record<string, string | number | undefined>) =>
    client.get<T>(endpoint, { params }).then((res) => res.data),

  post: <T>(endpoint: string, body?: unknown) =>
    client.post<T>(endpoint, body).then((res) => res.data),

  put: <T>(endpoint: string, body?: unknown) =>
    client.put<T>(endpoint, body).then((res) => res.data),

  patch: <T>(endpoint: string, body?: unknown) =>
    client.patch<T>(endpoint, body).then((res) => res.data),

  delete: <T>(endpoint: string) =>
    client.delete<T>(endpoint).then((res) => res.data),
}

export { client }
