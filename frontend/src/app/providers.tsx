import { useEffect } from 'react'
import { QueryClientProvider } from '@tanstack/react-query'
import { Toaster } from 'sonner'
import { queryClient } from './query-client'
import { useTheme } from '@/lib/use-theme'
import { useAppStore } from '@/stores/app-store'
import { preferenceApi } from '@/api'
import type { ReactNode } from 'react'

function ThemeAwareToaster() {
  const resolved = useTheme()

  return (
    <Toaster
      position="top-center"
      theme={resolved}
      toastOptions={{
        duration: 1000,
        classNames: {
          toast: '!bg-bg-layer-2 !border !border-border-custom !rounded-lg !font-sans !text-text-1',
          title: '!text-text-1 !font-medium !text-sm',
          description: '!text-text-2 !text-[13px]',
          closeButton: '!text-text-muted hover:!text-text-1 !transition-colors',
          actionButton: '!bg-text-1 !text-bg-base !rounded-md !text-[13px] !font-medium !px-3 !py-1.5 hover:!opacity-90',
          cancelButton: '!bg-bg-layer-3 !text-text-2 !rounded-md !text-[13px] !font-medium !px-3 !py-1.5 hover:!bg-bg-hover',
        },
      }}
    />
  )
}

function AppInit() {
  useEffect(() => {
    if (window.location.pathname === '/install') return
    preferenceApi.getPreference().then((res) => {
      if (res?.language) {
        useAppStore.getState().setLanguage(res.language)
      }
    }).catch(() => {})
  }, [])

  return null
}

export function Providers({ children }: { children: ReactNode }) {
  useTheme()

  return (
    <QueryClientProvider client={queryClient}>
      <AppInit />
      {children}
      <ThemeAwareToaster />
    </QueryClientProvider>
  )
}
