import { useEffect } from 'react'
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Separator } from '@/components/ui/separator'
import { IconUploadDialog } from '@/components/IconUploadDialog'
import { authbase, type UserInfo } from '@/lib/authbase'
import { LogOut, Mail, ShieldCheck, User } from 'lucide-react'

type Props = {
  user: UserInfo
  token: string
  onLogout: () => void
  onRefresh: () => void
}

const PROVIDER_LABEL: Record<string, string> = {
  google: 'Google',
  github: 'GitHub',
  discord: 'Discord',
  microsoftonline: 'Microsoft',
  basic: 'メール / パスワード',
}

export function DashboardPage({ user, token, onLogout, onRefresh }: Props) {
  useEffect(() => {
    onRefresh()
  }, [])

  const initials = user.name
    .split(' ')
    .map((n) => n[0])
    .join('')
    .toUpperCase()
    .slice(0, 2)

  return (
    <div className="min-h-screen bg-gradient-to-br from-indigo-50 via-white to-zinc-50 flex items-center justify-center p-4">
      <div className="w-full max-w-md space-y-4">

        {/* ヘッダー */}
        <div className="flex items-center justify-between px-1">
          <div className="flex items-center gap-2">
            <div className="flex items-center justify-center w-7 h-7 rounded-lg bg-indigo-600">
              <ShieldCheck className="w-4 h-4 text-white" />
            </div>
            <span className="text-sm font-semibold text-zinc-700">AuthBase</span>
          </div>
          <Button variant="ghost" size="sm" onClick={onLogout} className="text-zinc-500 hover:text-red-500">
            <LogOut className="w-4 h-4" />
            ログアウト
          </Button>
        </div>

        {/* プロフィールカード */}
        <Card>
          <CardContent className="pt-6">
            <div className="flex flex-col items-center gap-4">
              <div className="relative">
                <Avatar className="h-20 w-20 ring-4 ring-indigo-100">
                  <AvatarImage
                    src={`${authbase.iconUrl(user.user_id)}?t=${Date.now()}`}
                    alt={user.name}
                  />
                  <AvatarFallback className="text-xl">{initials}</AvatarFallback>
                </Avatar>
                <IconUploadDialog user={user} token={token} onUpdated={onRefresh} />
              </div>
              <div className="text-center">
                <h2 className="text-lg font-semibold text-zinc-900">{user.name}</h2>
                <p className="text-sm text-zinc-500">{user.email}</p>
              </div>
            </div>
          </CardContent>
        </Card>

        {/* ユーザー情報カード */}
        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="text-base">アカウント情報</CardTitle>
          </CardHeader>
          <CardContent className="space-y-0">
            <InfoRow
              icon={<User className="w-4 h-4 text-zinc-400" />}
              label="ユーザー ID"
              value={user.user_id}
              mono
            />
            <Separator />
            <InfoRow
              icon={<Mail className="w-4 h-4 text-zinc-400" />}
              label="メールアドレス"
              value={user.email}
            />
            <Separator />
            <InfoRow
              icon={<ShieldCheck className="w-4 h-4 text-zinc-400" />}
              label="認証方法"
              value={PROVIDER_LABEL[user.prov_code] ?? user.prov_code}
            />
          </CardContent>
        </Card>

      </div>
    </div>
  )
}

type InfoRowProps = {
  icon: React.ReactNode
  label: string
  value: string
  mono?: boolean
}

function InfoRow({ icon, label, value, mono = false }: InfoRowProps) {
  return (
    <div className="flex items-center gap-3 py-3">
      <span className="flex-shrink-0">{icon}</span>
      <div className="flex-1 min-w-0">
        <p className="text-xs text-zinc-400 mb-0.5">{label}</p>
        <p className={`text-sm text-zinc-700 truncate ${mono ? 'font-mono text-xs' : ''}`}>
          {value}
        </p>
      </div>
    </div>
  )
}
