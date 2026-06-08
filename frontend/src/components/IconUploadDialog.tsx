import { useRef, useState } from 'react'
import * as Dialog from '@radix-ui/react-dialog'
import { Button } from '@/components/ui/button'
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar'
import { Camera, Loader2, X } from 'lucide-react'
import { authbase, type UserInfo } from '@/lib/authbase'

type Props = {
  user: UserInfo
  token: string
  onUpdated: () => void
}

export function IconUploadDialog({ user, token, onUpdated }: Props) {
  const [open, setOpen] = useState(false)
  const [preview, setPreview] = useState<string | null>(null)
  const [file, setFile] = useState<File | null>(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const inputRef = useRef<HTMLInputElement>(null)

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const f = e.target.files?.[0]
    if (!f) return
    setFile(f)
    setPreview(URL.createObjectURL(f))
    setError(null)
  }

  const handleUpload = async () => {
    if (!file) return
    setLoading(true)
    setError(null)
    try {
      await authbase.uploadIcon(token, file)
      onUpdated()
      setOpen(false)
      setPreview(null)
      setFile(null)
    } catch (e) {
      setError(e instanceof Error ? e.message : 'アップロードに失敗しました')
    } finally {
      setLoading(false)
    }
  }

  const initials = user.name
    .split(' ')
    .map((n) => n[0])
    .join('')
    .toUpperCase()
    .slice(0, 2)

  return (
    <Dialog.Root open={open} onOpenChange={setOpen}>
      <Dialog.Trigger asChild>
        <button
          className="absolute bottom-0 right-0 flex h-7 w-7 items-center justify-center rounded-full bg-indigo-600 text-white shadow-md hover:bg-indigo-700 transition-colors"
          aria-label="アイコンを変更"
        >
          <Camera className="w-3.5 h-3.5" />
        </button>
      </Dialog.Trigger>

      <Dialog.Portal>
        <Dialog.Overlay className="fixed inset-0 bg-black/40 backdrop-blur-sm z-40" />
        <Dialog.Content className="fixed left-1/2 top-1/2 z-50 w-full max-w-sm -translate-x-1/2 -translate-y-1/2 rounded-xl bg-white p-6 shadow-xl focus:outline-none">
          <div className="flex items-center justify-between mb-5">
            <Dialog.Title className="text-base font-semibold text-zinc-900">
              アイコンを変更
            </Dialog.Title>
            <Dialog.Close asChild>
              <button className="rounded-md p-1 text-zinc-400 hover:text-zinc-600 hover:bg-zinc-100 transition-colors">
                <X className="w-4 h-4" />
              </button>
            </Dialog.Close>
          </div>

          <div className="flex flex-col items-center gap-5">
            <Avatar className="h-24 w-24 ring-4 ring-indigo-100">
              <AvatarImage
                src={preview ?? `${authbase.iconUrl(user.user_id)}?t=${Date.now()}`}
                alt={user.name}
              />
              <AvatarFallback className="text-2xl">{initials}</AvatarFallback>
            </Avatar>

            <input
              ref={inputRef}
              type="file"
              accept="image/*"
              className="hidden"
              onChange={handleFileChange}
            />

            <Button
              variant="outline"
              size="sm"
              onClick={() => inputRef.current?.click()}
              className="w-full"
            >
              <Camera className="w-4 h-4" />
              画像を選択
            </Button>

            {error && (
              <p className="w-full text-sm text-red-500 bg-red-50 border border-red-200 rounded-lg px-3 py-2">
                {error}
              </p>
            )}
          </div>

          <div className="flex gap-2 mt-6">
            <Dialog.Close asChild>
              <Button variant="outline" className="flex-1" disabled={loading}>
                キャンセル
              </Button>
            </Dialog.Close>
            <Button
              className="flex-1"
              onClick={handleUpload}
              disabled={!file || loading}
            >
              {loading ? <Loader2 className="w-4 h-4 animate-spin" /> : '保存'}
            </Button>
          </div>
        </Dialog.Content>
      </Dialog.Portal>
    </Dialog.Root>
  )
}
