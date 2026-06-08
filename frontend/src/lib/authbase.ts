/**
 * AuthBase SDK
 *
 * 使い方:
 *   const client = new AuthBaseClient('/auth')
 *
 *   // ログインページへリダイレクト（未認証時）
 *   client.redirectToLogin()
 *
 *   // コールバック処理（/ui/ に戻ってきたとき）
 *   const token = await client.handleCallback()  // bridge_token を交換
 *
 *   // セッション取得
 *   const user = await client.me(token)
 */

// ─── 型定義 ─────────────────────────────────────────────

export type AuthToken = string

export type UserInfo = {
  user_id: string
  name: string
  email: string
  prov_code: string
  prov_uid: string
}

export type AccessTokenClaims = {
  user_id: string
  name: string
  email: string
  labels: string[]
  prov_code: string
  prov_uid: string
  exp: number
}

export type Provider = 'google' | 'github' | 'discord' | 'microsoftonline' | 'basic'

export type AuthBaseError = {
  error: string
  status: number
}

// ─── エラークラス ────────────────────────────────────────

export class AuthBaseRequestError extends Error {
  readonly status: number
  constructor(message: string, status: number) {
    super(message)
    this.name = 'AuthBaseRequestError'
    this.status = status
  }
}

// ─── クライアント ────────────────────────────────────────

export class AuthBaseClient {
  readonly baseUrl: string

  constructor(baseUrl: string = '/auth') {
    this.baseUrl = baseUrl.replace(/\/$/, '')
  }

  // ── 内部ヘルパー ──────────────────────────────────────

  private async request<T>(
    path: string,
    init: RequestInit = {},
  ): Promise<T> {
    const res = await fetch(`${this.baseUrl}${path}`, {
      ...init,
      headers: {
        'Content-Type': 'application/json',
        ...init.headers,
      },
    })

    if (!res.ok) {
      let message = res.statusText
      try {
        const body = (await res.json()) as { error?: string }
        if (body.error) message = body.error
      } catch {
        // ignore
      }
      throw new AuthBaseRequestError(message, res.status)
    }

    const text = await res.text()
    return (text ? JSON.parse(text) : {}) as T
  }

  private sessionHeader(token: AuthToken): HeadersInit {
    return { Authorization: token }
  }

  private bearerHeader(token: AuthToken): HeadersInit {
    return { Authorization: `Bearer ${token}` }
  }

  // ── リダイレクト認証フロー ────────────────────────────

  /**
   * ログインページへリダイレクトする。
   * 認証後、callbackUrl（デフォルト: 現在の URL）に bridge_token クエリ付きで戻る。
   */
  redirectToLogin(loginUrl: string = `${this.baseUrl}/login`): never {
    window.location.href = loginUrl
    throw new Error('redirecting')
  }

  /**
   * URL の bridge_token クエリパラメータを取得して session token に交換する。
   * 成功したら URL からパラメータを除去して token を返す。
   * bridge_token がなければ null を返す。
   */
  async handleCallback(): Promise<AuthToken | null> {
    const params = new URLSearchParams(window.location.search)
    const bridgeToken = params.get('bridge_token')
    if (!bridgeToken) return null

    const { refresh_token: token } = await this.request<{ refresh_token: string }>(
      '/bridge/exchange',
      { headers: this.bearerHeader(bridgeToken) },
    )

    // URL からクエリパラメータを除去
    params.delete('bridge_token')
    const newSearch = params.toString()
    const newUrl = window.location.pathname + (newSearch ? `?${newSearch}` : '')
    window.history.replaceState(null, '', newUrl)

    return token
  }

  // ── Basic 認証 ────────────────────────────────────────

  async login(email: string, password: string): Promise<AuthToken> {
    const { token } = await this.request<{ token: string }>('/basic/login', {
      method: 'POST',
      body: JSON.stringify({ email, password }),
    })
    return token
  }

  async signup(name: string, email: string, password: string): Promise<AuthToken> {
    const { token } = await this.request<{ token: string }>('/basic/signup', {
      method: 'POST',
      body: JSON.stringify({ name, email, password }),
    })
    return token
  }

  // ── セッション ────────────────────────────────────────

  async me(token: AuthToken): Promise<UserInfo> {
    return this.request<UserInfo>('/me', {
      headers: this.sessionHeader(token),
      // Content-Type 不要なので上書き
      method: 'GET',
    })
  }

  async logout(token: AuthToken): Promise<void> {
    await this.request('/logout', {
      method: 'POST',
      headers: this.sessionHeader(token),
    })
  }

  // ── アクセストークン ──────────────────────────────────

  async getAccessToken(sessionToken: AuthToken): Promise<AuthToken> {
    const { token } = await this.request<{ token: string }>('/token', {
      method: 'GET',
      headers: this.sessionHeader(sessionToken),
    })
    return token
  }

  async getUserInfo(accessToken: AuthToken): Promise<AccessTokenClaims> {
    return this.request<AccessTokenClaims>('/userinfo', {
      method: 'GET',
      headers: this.bearerHeader(accessToken),
    })
  }

  // ── ユーザー情報 ──────────────────────────────────────

  async getPublicInfo(userId: string): Promise<{ user_id: string; name: string }> {
    return this.request<{ user_id: string; name: string }>(`/info/${userId}`, {
      method: 'GET',
      headers: {},
    })
  }

  iconUrl(userId: string): string {
    return `${this.baseUrl}/icon/${userId}`
  }

  async uploadIcon(token: AuthToken, file: File): Promise<void> {
    const form = new FormData()
    form.append('file', file)
    const res = await fetch(`${this.baseUrl}/icon`, {
      method: 'POST',
      headers: { Authorization: token },
      body: form,
    })
    if (!res.ok) {
      let message = res.statusText
      try {
        const body = (await res.json()) as { error?: string }
        if (body.error) message = body.error
      } catch {
        // ignore
      }
      throw new AuthBaseRequestError(message, res.status)
    }
  }

  // ── プロバイダー ──────────────────────────────────────

  async getProviders(): Promise<Provider[]> {
    return this.request<Provider[]>('/login/providers', { method: 'GET', headers: {} })
  }

  oauthUrl(provider: Exclude<Provider, 'basic'>): string {
    return `${this.baseUrl}/oauth/${provider}`
  }
}

// ─── デフォルトインスタンス ──────────────────────────────

export const authbase = new AuthBaseClient('/auth')
