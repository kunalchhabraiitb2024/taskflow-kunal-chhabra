import { useState, type FormEvent } from 'react'
import axios from 'axios'
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import { createProject } from '@/api/projects'
import type { Project } from '@/types'

interface CreateProjectModalProps {
  open: boolean
  onClose: () => void
  onCreated: (project: Project) => void
}

export function CreateProjectModal({ open, onClose, onCreated }: CreateProjectModalProps) {
  const [name, setName] = useState('')
  const [description, setDescription] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault()
    if (!name.trim()) { setError('Project name is required'); return }

    setLoading(true)
    setError('')
    try {
      const { data } = await createProject({
        name: name.trim(),
        description: description.trim() || undefined,
      })
      onCreated(data)
      setName('')
      setDescription('')
      onClose()
    } catch (err) {
      if (axios.isAxiosError(err)) {
        setError(err.response?.data?.error ?? 'Failed to create project')
      } else {
        setError('Something went wrong')
      }
    } finally {
      setLoading(false)
    }
  }

  return (
    <Dialog open={open} onOpenChange={(o) => !o && onClose()}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>New Project</DialogTitle>
        </DialogHeader>
        <form onSubmit={handleSubmit} className="space-y-4">
          {error && (
            <p className="rounded-md bg-destructive/10 px-3 py-2 text-sm text-destructive">{error}</p>
          )}
          <div className="space-y-1">
            <Label htmlFor="project-name">Name *</Label>
            <Input
              id="project-name"
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="My awesome project"
              disabled={loading}
              autoFocus
            />
          </div>
          <div className="space-y-1">
            <Label htmlFor="project-desc">Description</Label>
            <Textarea
              id="project-desc"
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              placeholder="Optional description"
              rows={3}
              disabled={loading}
            />
          </div>
          <DialogFooter>
            <Button type="button" variant="outline" onClick={onClose} disabled={loading}>
              Cancel
            </Button>
            <Button type="submit" disabled={loading}>
              {loading ? 'Creating…' : 'Create project'}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
