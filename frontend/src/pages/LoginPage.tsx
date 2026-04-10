import { useState, type FormEvent } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import axios from 'axios'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { useAuth } from '@/hooks/useAuth'

export function LoginPage() {
  const { login, isAuthenticated } = useAuth()
  const navigate = useNavigate()
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  // Already logged in → go to projects
  if (isAuthenticated) {
    navigate('/projects', { replace: true })
    return null
  }

  const validate = (): boolean => {
    if (!email.trim()) { setError('Email is required'); return false }
    if (!/\S+@\S+\.\S+/.test(email)) { setError('Enter a valid email'); return false }
    if (!password) { setError('Password is required'); return false }
    return true
  }

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault()
    setError('')
    if (!validate()) return

    setLoading(true)
    try {
      await login(email, password)
      navigate('/projects', { replace: true })
    } catch (err) {
      if (axios.isAxiosError(err)) {
        const msg = err.response?.data?.error || 'Login failed'
        setError(msg === 'invalid credentials' ? 'Invalid email or password' : msg)
      } else {
        setError('Something went wrong')
      }
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="flex min-h-screen items-center justify-center bg-muted/30 px-4">
      <Card className="w-full max-w-sm">
        <CardHeader>
          <CardTitle>Sign in</CardTitle>
          <CardDescription>Enter your email and password</CardDescription>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleSubmit} className="space-y-4">
            {error && (
              <div className="rounded-md bg-destructive/10 px-3 py-2 text-sm text-destructive">
                {error}
              </div>
            )}
            <div className="space-y-1">
              <Label htmlFor="email">Email</Label>
              <Input
                id="email"
                type="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                placeholder="you@example.com"
                autoComplete="email"
                disabled={loading}
              />
            </div>
            <div className="space-y-1">
              <Label htmlFor="password">Password</Label>
              <Input
                id="password"
                type="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                placeholder="••••••••"
                autoComplete="current-password"
                disabled={loading}
              />
            </div>
            <Button type="submit" className="w-full" disabled={loading}>
              {loading ? 'Signing in…' : 'Sign in'}
            </Button>
          </form>
          <p className="mt-4 text-center text-sm text-muted-foreground">
            No account?{' '}
            <Link to="/register" className="text-primary underline underline-offset-4">
              Register
            </Link>
          </p>
        </CardContent>
      </Card>
    </div>
  )
}
