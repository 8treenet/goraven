import { create } from 'zustand'
import { persist } from 'zustand/middleware'

export type Theme = 'light' | 'dark' | 'system'
export type Language = 'zh' | 'en'

function migrateTheme(): Theme {
  if (typeof window === 'undefined') return 'system'
  return (localStorage.getItem('theme') as Theme) || 'system'
}

interface AppState {
  language: Language
  theme: Theme

  setLanguage: (lang: Language) => void
  setTheme: (t: Theme) => void
}

export const useAppStore = create<AppState>()(
  persist(
    (set) => ({
      language: 'en',
      theme: migrateTheme(),

      setLanguage: (language) => set({ language }),
      setTheme: (theme) => set({ theme }),
    }),
    {
      name: 'raven-app',
      partialize: (state) => ({
        language: state.language,
        theme: state.theme,
      }),
    },
  ),
)
