type Theme = 'light' | 'dark' | 'system'

const STORAGE_KEY = 'theme'

function getStoredTheme(): Theme {
  if (typeof window === 'undefined') return 'system'
  return (localStorage.getItem(STORAGE_KEY) as Theme) || 'system'
}

function resolveTheme(theme: Theme): 'light' | 'dark' {
  if (theme === 'dark') return 'dark'
  if (theme === 'light') return 'light'
  return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light'
}

function applyThemeClass(theme: Theme) {
  const root = document.documentElement
  root.classList.remove('light', 'dark')
  if (theme === 'dark') {
    root.classList.add('dark')
  } else if (theme === 'light') {
    root.classList.add('light')
  }
}

let themeListeners: Array<() => void> = []
let currentTheme: Theme = getStoredTheme()

function notifyListeners() {
  themeListeners.forEach((fn) => fn())
}

export function getTheme(): Theme {
  return currentTheme
}

export function setTheme(theme: Theme) {
  currentTheme = theme
  localStorage.setItem(STORAGE_KEY, theme)
  applyThemeClass(theme)
  notifyListeners()
}

export function subscribeToTheme(fn: () => void) {
  themeListeners.push(fn)
  return () => {
    themeListeners = themeListeners.filter((f) => f !== fn)
  }
}

export function getResolvedTheme(): 'light' | 'dark' {
  return resolveTheme(currentTheme)
}

if (typeof window !== 'undefined') {
  window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', () => {
    if (currentTheme === 'system') {
      applyThemeClass('system')
      notifyListeners()
    }
  })
}
