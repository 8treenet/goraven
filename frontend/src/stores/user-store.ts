import { create } from 'zustand'
import { persist } from 'zustand/middleware'

export interface CurrentUser {
  userId: string
  username: string
  nickname: string
  avatar: string
  role: number
  email: string
}

interface UserState {
  token: string | null
  currentUser: CurrentUser | null
  isAuthenticated: boolean
  isAdmin: boolean

  setAuth: (token: string, user: CurrentUser) => void
  clearAuth: () => void
}

export const useUserStore = create<UserState>()(
  persist(
    (set) => ({
      token: null,
      currentUser: null,
      isAuthenticated: false,
      isAdmin: false,

      setAuth: (token, user) =>
        set({
          token,
          currentUser: user,
          isAuthenticated: true,
          isAdmin: user.role === 1,
        }),

      clearAuth: () =>
        set({
          token: null,
          currentUser: null,
          isAuthenticated: false,
          isAdmin: false,
        }),
    }),
    {
      name: 'user-storage',
      partialize: (state) => ({
        token: state.token,
        currentUser: state.currentUser,
        isAuthenticated: state.isAuthenticated,
        isAdmin: state.isAdmin,
      }),
    },
  ),
)
