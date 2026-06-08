const BASE = '/auth'

export type UserInfo = {
  user_id: string
  name: string
  email: string
  prov_code: string
  prov_uid: string
}

export type Provider = 'google' | 'github' | 'discord' | 'microsoftonline' | 'basic'

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(`${BASE}${path}`, init)
  if (!res.ok) {
    const body = await res.json().catch(() => ({ error: res.statusText }))
    throw new Error((body as { error: string }).error ?? res.statusText)
  }
  return res.json() as Promise<T>
}

export const api = {
  login(email: string, password: string) {
    return request<{ token: string }>('/basic/login', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ email, password }),
    })
  },

  signup(name: string, email: string, password: string) {
    return request<{ token: string }>('/basic/signup', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ name, email, password }),
    })
  },

  me(token: string) {
    return request<UserInfo>('/me', {
      headers: { Authorization: token },
    })
  },

  logout(token: string) {
    return request<{ message: string }>('/logout', {
      method: 'POST',
      headers: { Authorization: token },
    })
  },

  uploadIcon(token: string, file: File) {
    const form = new FormData()
    form.append('file', file)
    return request<{ result: string }>('/icon', {
      method: 'POST',
      headers: { Authorization: token },
      body: form,
    })
  },

  iconUrl(userId: string) {
    return `${BASE}/icon/${userId}`
  },

  providers() {
    return request<Provider[]>('/login/providers')
  },
}
