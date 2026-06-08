import { useState, useCallback, useEffect } from 'react'
import { authbase, type UserInfo, type AuthToken } from '@/lib/authbase'

const TOKEN_KEY = 'token'

type AuthState =
  | { status: 'loading' }
  | { status: 'unauthenticated' }
  | { status: 'authenticated'; token: AuthToken; user: UserInfo }

export function useAuth() {
  const [state, setState] = useState<AuthState>({ status: 'loading' })

  useEffect(() => {
    let cancelled = false

    async function init() {
      const token = localStorage.getItem(TOKEN_KEY)
      if (!token) {
        if (!cancelled) setState({ status: 'unauthenticated' })
        return
      }

      try {
        const user = await authbase.me(token)
        if (!cancelled) setState({ status: 'authenticated', token, user })
      } catch {
        localStorage.removeItem(TOKEN_KEY)
        if (!cancelled) setState({ status: 'unauthenticated' })
      }
    }

    void init()
    return () => { cancelled = true }
  }, [])

  const logout = useCallback(async () => {
    if (state.status !== 'authenticated') return
    await authbase.logout(state.token).catch(() => {})
    localStorage.removeItem(TOKEN_KEY)
    setState({ status: 'unauthenticated' })
  }, [state])

  const refreshUser = useCallback(async () => {
    if (state.status !== 'authenticated') return
    try {
      const user = await authbase.me(state.token)
      setState({ status: 'authenticated', token: state.token, user })
    } catch {
      localStorage.removeItem(TOKEN_KEY)
      setState({ status: 'unauthenticated' })
    }
  }, [state])

  return { state, logout, refreshUser }
}
