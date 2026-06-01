import { api } from "./client"
import type { LoginRequest, RegisterRequest, AuthResponse, User, ApiResponse } from "@/types"

export async function login(data: LoginRequest): Promise<AuthResponse> {
  return api.post<AuthResponse>("/auth/login", data)
}

export async function register(data: RegisterRequest): Promise<ApiResponse<null>> {
  return api.post<ApiResponse<null>>("/auth/register", data)
}

export async function verifyEmail(token: string): Promise<ApiResponse<null>> {
  return api.post<ApiResponse<null>>("/auth/verify-email", { token })
}

export async function forgotPassword(email: string): Promise<ApiResponse<null>> {
  return api.post<ApiResponse<null>>("/auth/forgot-password", { email })
}

export async function logout(): Promise<ApiResponse<null>> {
  return api.post<ApiResponse<null>>("/auth/logout")
}

export async function getMe(): Promise<ApiResponse<User>> {
  return api.get<ApiResponse<User>>("/auth/me")
}
