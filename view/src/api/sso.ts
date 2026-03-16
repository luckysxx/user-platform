import { postJson } from '@/api/http'

export interface RegisterRequest {
  email: string
  username: string
  password: string
}

export interface RegisterResponse {
  email: string
  user_id: number
  username: string
}

export interface LoginRequest {
  username: string
  password: string
  app_code: string
}

export interface LoginResponse {
  access_token: string
  refresh_token: string
  user_id: number
  username: string
}

export function registerBySso(payload: RegisterRequest): Promise<RegisterResponse> {
  return postJson<RegisterRequest, RegisterResponse>('/api/v1/users/register', payload)
}

export function loginBySso(payload: LoginRequest): Promise<LoginResponse> {
  return postJson<LoginRequest, LoginResponse>('/api/v1/users/login', payload)
}
