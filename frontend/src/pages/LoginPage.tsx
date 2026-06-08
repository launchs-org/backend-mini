import { LogIn } from 'lucide-react'

type Props = {
  loginPageUrl: string
}

export function LoginPage({ loginPageUrl }: Props) {
  return (
    <div className="min-h-screen bg-gradient-to-br from-indigo-50 via-white to-zinc-50 flex items-center justify-center p-4">
      <div className="w-full max-w-sm space-y-6 text-center">
        <div>
          <div className="inline-flex items-center justify-center w-12 h-12 rounded-xl bg-indigo-600 mb-4">
            <LogIn className="w-6 h-6 text-white" />
          </div>
          <h1 className="text-2xl font-bold text-zinc-900">AuthBase</h1>
          <p className="text-sm text-zinc-500 mt-1">続けるにはログインが必要です</p>
        </div>

        <a
          href={loginPageUrl}
          className="flex items-center justify-center gap-2 w-full rounded-xl bg-indigo-600 px-4 py-3 text-sm font-medium text-white shadow hover:bg-indigo-700 transition-colors"
        >
          <LogIn className="w-4 h-4" />
          ログイン
        </a>
      </div>
    </div>
  )
}
