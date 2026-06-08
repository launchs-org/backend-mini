import { useAuth } from '@/hooks/useAuth'
import { LoginPage } from '@/pages/LoginPage'
import { DashboardPage } from '@/pages/DashboardPage'
import { Loader2 } from 'lucide-react'

const LOGIN_URL = '/auth/login'

export default function App() {
  const { state, logout, refreshUser } = useAuth()

  if (state.status === 'loading') {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-indigo-50 via-white to-zinc-50">
        <Loader2 className="w-8 h-8 animate-spin text-indigo-400" />
      </div>
    )
  }

  if (state.status === 'authenticated') {
    return (
      <DashboardPage
        user={state.user}
        token={state.token}
        onLogout={async () => {
          await logout()
          window.location.href = LOGIN_URL
        }}
        onRefresh={refreshUser}
      />
    )
  }

  return <LoginPage loginPageUrl={LOGIN_URL} />
}
