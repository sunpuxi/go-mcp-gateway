import { createContext, useContext, useState, useEffect, type ReactNode } from 'react'
import { apiClient } from '../api/client'

interface AuthContextType {
  isAuthenticated: boolean
  login: (key: string) => void
  logout: () => void
}

const AuthContext = createContext<AuthContextType>({
  isAuthenticated: false,
  login: () => {},
  logout: () => {},
})

export function AuthProvider({ children }: { children: ReactNode }) {
  const [isAuthenticated, setIsAuthenticated] = useState(false)

  useEffect(() => {
    const saved = localStorage.getItem('admin_api_key')
    if (saved) {
      apiClient.setApiKey(saved)
      setIsAuthenticated(true)
    }
  }, [])

  const login = (key: string) => {
    apiClient.setApiKey(key)
    setIsAuthenticated(true)
  }

  const logout = () => {
    apiClient.clearApiKey()
    setIsAuthenticated(false)
  }

  return (
    <AuthContext.Provider value={{ isAuthenticated, login, logout }}>
      {children}
    </AuthContext.Provider>
  )
}

export function useAuth() {
  return useContext(AuthContext)
}
