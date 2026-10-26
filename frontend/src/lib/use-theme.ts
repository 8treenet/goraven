import { useEffect, useState } from 'react'
import { useAppStore, type Theme } from '@/stores/app-store'

function resolveTheme(theme: Theme, systemIsDark: boolean): 'light' | 'dark' {
  if (theme === 'dark') return 'dark'
  if (theme === 'light') return 'light'
  return systemIsDark ? 'dark' : 'light'
}

function applyThemeClass(theme: Theme, systemIsDark: boolean) {
  const root = document.documentElement
  root.classList.remove('light', 'dark')
  const resolved = resolveTheme(theme, systemIsDark)
  if (resolved === 'dark') {
    root.classList.add('dark')
  } else {
    root.classList.add('light')
  }
}

export function useTheme(): 'light' | 'dark' {
  const theme = useAppStore((s) => s.theme)
  const [systemIsDark, setSystemIsDark] = useState(
    () => typeof window !== 'undefined' && window.matchMedia('(prefers-color-scheme: dark)').matches,
  )

  useEffect(() => {
    const mq = window.matchMedia('(prefers-color-scheme: dark)')
    const handler = (e: MediaQueryListEvent) => setSystemIsDark(e.matches)
    mq.addEventListener('change', handler)
    return () => mq.removeEventListener('change', handler)
  }, [])

  useEffect(() => {
    applyThemeClass(theme, systemIsDark)
  }, [theme, systemIsDark])

  return resolveTheme(theme, systemIsDark)
}
