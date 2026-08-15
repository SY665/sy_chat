import type{
    AuthData,
    LoginInput,
    LogoutData,
    RegisterInput,
} from "@/features/auth/types"

import { apiRequest } from "./client"

export function register(input:RegisterInput): Promise<AuthData>{
    return apiRequest<AuthData>('/auth/register',{
        method:'POST',
        body:JSON.stringify(input),
    })
}

export function login(input: LoginInput): Promise<AuthData> {
  return apiRequest<AuthData>('/auth/login', {
    method: 'POST',
    body: JSON.stringify(input),
  })
}

export function getCurrentUser(): Promise<AuthData> {
  return apiRequest<AuthData>('/auth/me')
}

export function logout(): Promise<LogoutData> {
  return apiRequest<LogoutData>('/auth/logout', {
    method: 'POST',
  })
}