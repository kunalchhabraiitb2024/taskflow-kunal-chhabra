import { useState, type FormEvent } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import axios from 'axios'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { useAuth } from '@/hooks/useAuth'

export function RegisterPage() {
  const { register, isAuthenticated } = useAuth()
  const navigate = useNavigate()
  const [name, setName] = useState('')
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState<string | Record<string, string>>('')
  const [loading, setLoading] = useState(false)

  if (isAuthenticated) {
    navigate('/projects', { replace: true })
    return null
  }

  const validate = (): boolean => {
    if (!name.trim()) { setError('Name is required'); return false }
    if (!email.trim() || !/\S+@\S+\.\S+/.test(email)) { setError('Enter a valid email'); return false }
    if (password.length < 6) { setError('Password must be at least 6 characters'); return false }
    return true
  }

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault()
    setError('')
    if (!validate()) return

    setLoading(true)
    try {
      await register(name, email, password)
      navigate('/projects', { replace: true })
    } catch (err) {
      if (axios.isAxiosError(err)) {
        const data = err.response?.data
        if (data?.fields) {
          setError(data.fields)
        } else {
          setError(data?.error || 'Registration failed')
        }
      } else {
        setError('Something went wrong')
      }
    } finally {
      setLoading(false)
    }
  }

  const errorStr = typeof error === 'string' ? error : ''
  const fieldErrors = typeof error === 'object' ? error : {}

  return (
    <div className="flex min-h-screen items-center justify-center bg-muted/30 px-4">
      <Card className="w-full max-w-sm">
        <CardHeader>
          <CardTitle>Create account</CardTitle>
          <CardDescription>Sign up to get started with TaskFlow</CardDescription>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleSubmit} className="space-y-4">
            {errorStr && (
              <div className="rounded-md bg-destructive/10 px-3 py-2 text-sm text-destructive">
                {errorStr}
              </div>
            )}
            <div className="space-y-1">
              <Label htmlFor="name">Name</Label>
              <Input
                id="name"
                value={name}
                onChange={(e) => setName(e.target.value)}
                placeholder="Jane Doe"
                disabled={loading}
              />
              {fieldErrors.name && <p className="text-xs text-destructive">{fieldErrors.name}</p>}
            </div>
            <div className="space-y-1">
              <Label htmlFor="email">Email</Label>
              <Input
                id="email"
                type="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                placeholder="you@example.com"
                disabled={loading}
              />
              {fieldErrors.email && <p className="text-xs text-destructive">{fieldErrors.email}</p>}
            </div>
            <div className="space-y-1">
              <Label htmlFor="password">Password</Label>
              <Input
                id="password"
                type="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                placeholder="Min 6 characters"
                disabled={loading}
              />
              {fieldErrors.password && <p className="text-xs text-destructive">{fieldErrors.password}</p>}
            </div>
            <Button type="submit" className="w-full" disabled={loading}>
              {loading ? 'Creating account…' : 'Create account'}
            </Button>
          </form>
          <p className="mt-4 text-center text-sm text-muted-foreground">
            Already have an account?{' '}
            <Link to="/login" className="text-primary underline underline-offset-4">
              Sign in
            </Link>
          </p>
        </CardContent>
      </Card>
    </div>
  )
}
