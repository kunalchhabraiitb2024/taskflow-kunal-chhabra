import { Link } from 'react-router-dom'
import { LogOut, CheckSquare, Sun, Moon } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { useAuth } from '@/hooks/useAuth'
import { useTheme } from '@/hooks/useTheme'

export function Navbar() {
  const { user, logout } = useAuth()
  const { theme, toggle } = useTheme()

  return (
    <header className="border-b bg-white shadow-sm dark:border-gray-700 dark:bg-gray-900">
      <div className="container mx-auto flex h-14 items-center justify-between px-4">
        <Link to="/projects" className="flex items-center gap-2 font-semibold text-primary">
          <CheckSquare className="h-5 w-5" />
          TaskFlow
        </Link>

        <div className="flex items-center gap-2">
          {user && (
            <span className="hidden text-sm text-muted-foreground sm:inline">
              {user.name}
            </span>
          )}

          {/* Dark mode toggle — persists across sessions via localStorage */}
          <Button variant="ghost" size="icon" onClick={toggle} title="Toggle dark mode">
            {theme === 'dark' ? (
              <Sun className="h-4 w-4" />
            ) : (
              <Moon className="h-4 w-4" />
            )}
          </Button>

          <Button variant="ghost" size="sm" onClick={logout}>
            <LogOut className="mr-1 h-4 w-4" />
            <span className="hidden sm:inline">Logout</span>
          </Button>
        </div>
      </div>
    </header>
  )
}
