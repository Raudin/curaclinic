import { createContext, lazy, Suspense, useCallback, useContext, useMemo, useState, type ReactNode } from 'react'

const Toast = lazy(() => import('@/components/arc/toast/toast'))

type ToastContextValue = {
  notify: (message: string, title?: string) => void
}

const ToastContext = createContext<ToastContextValue | null>(null)

type Notice = { message: string; title: string }

export function ToastProvider({ children }: { children: ReactNode }) {
  const [notice, setNotice] = useState<Notice | null>(null)

  const notify = useCallback((message: string, title = 'CuraClinic update') => {
    setNotice({ message, title })
  }, [])

  const value = useMemo(() => ({ notify }), [notify])

  return <ToastContext.Provider value={value}>
    {children}
    <div className="fixed bottom-5 left-1/2 z-[60] w-[calc(100%-2rem)] max-w-[26rem] -translate-x-1/2">
      <Suspense fallback={null}>
        <Toast
          title={notice?.title ?? 'CuraClinic update'}
          description={notice?.message ?? ''}
          open={notice !== null}
          onOpenChange={open => { if (!open) setNotice(null) }}
          duration={3600}
        />
      </Suspense>
    </div>
  </ToastContext.Provider>
}

export function useToast(): ToastContextValue {
  const value = useContext(ToastContext)
  if (!value) throw new Error('useToast must be used inside ToastProvider')
  return value
}
