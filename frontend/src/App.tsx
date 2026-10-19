import { AuthProvider } from '@/lib/auth-context'
import { ToastProvider } from '@/components/feedback/toast-provider'
import { AppRouter } from '@/app-router'

function App() {
  return <AuthProvider>
    <ToastProvider>
      <AppRouter />
    </ToastProvider>
  </AuthProvider>
}

export default App
