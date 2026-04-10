import { useState, useEffect, type FormEvent } from 'react'
import axios from 'axios'
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { createTask, updateTask } from '@/api/tasks'
import type { Task } from '@/types'

interface TaskFormProps {
  projectId: string
  task?: Task | null  // null/undefined = create mode, Task = edit mode
  open: boolean
  onClose: () => void
  onSaved: (task: Task) => void
}

export function TaskForm({ projectId, task, open, onClose, onSaved }: TaskFormProps) {
  const isEdit = !!task
  const [title, setTitle] = useState('')
  const [description, setDescription] = useState('')
  const [status, setStatus] = useState<Task['status']>('todo')
  const [priority, setPriority] = useState<Task['priority']>('medium')
  const [assigneeId, setAssigneeId] = useState('')
  const [dueDate, setDueDate] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  // Populate fields when editing
  useEffect(() => {
    if (task) {
      setTitle(task.title)
      setDescription(task.description ?? '')
      setStatus(task.status)
      setPriority(task.priority)
      setAssigneeId(task.assignee_id ?? '')
      setDueDate(task.due_date ? task.due_date.split('T')[0] : '')
    } else {
      setTitle('')
      setDescription('')
      setStatus('todo')
      setPriority('medium')
      setAssigneeId('')
      setDueDate('')
    }
    setError('')
  }, [task, open])

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault()
    if (!title.trim()) { setError('Title is required'); return }

    setLoading(true)
    setError('')
    const payload = {
      title: title.trim(),
      description: description || undefined,
      status,
      priority,
      assignee_id: assigneeId || undefined,
      due_date: dueDate || undefined,
    }

    try {
      let saved: Task
      if (isEdit) {
        const { data } = await updateTask(task!.id, payload)
        saved = data
      } else {
        const { data } = await createTask(projectId, payload)
        saved = data
      }
      onSaved(saved)
      onClose()
    } catch (err) {
      if (axios.isAxiosError(err)) {
        const d = err.response?.data
        setError(d?.fields ? Object.values(d.fields).join(', ') : d?.error ?? 'Failed to save task')
      } else {
        setError('Something went wrong')
      }
    } finally {
      setLoading(false)
    }
  }

  return (
    <Dialog open={open} onOpenChange={(o) => !o && onClose()}>
      <DialogContent className="sm:max-w-lg">
        <DialogHeader>
          <DialogTitle>{isEdit ? 'Edit Task' : 'New Task'}</DialogTitle>
        </DialogHeader>
        <form onSubmit={handleSubmit} className="space-y-4">
          {error && (
            <p className="rounded-md bg-destructive/10 px-3 py-2 text-sm text-destructive">{error}</p>
          )}
          <div className="space-y-1">
            <Label htmlFor="task-title">Title *</Label>
            <Input
              id="task-title"
              value={title}
              onChange={(e) => setTitle(e.target.value)}
              placeholder="Task title"
              disabled={loading}
            />
          </div>
          <div className="space-y-1">
            <Label htmlFor="task-desc">Description</Label>
            <Textarea
              id="task-desc"
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              placeholder="Optional description"
              rows={3}
              disabled={loading}
            />
          </div>
          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-1">
              <Label>Status</Label>
              <Select value={status} onValueChange={(v) => setStatus(v as Task['status'])}>
                <SelectTrigger><SelectValue /></SelectTrigger>
                <SelectContent>
                  <SelectItem value="todo">Todo</SelectItem>
                  <SelectItem value="in_progress">In Progress</SelectItem>
                  <SelectItem value="done">Done</SelectItem>
                </SelectContent>
              </Select>
            </div>
            <div className="space-y-1">
              <Label>Priority</Label>
              <Select value={priority} onValueChange={(v) => setPriority(v as Task['priority'])}>
                <SelectTrigger><SelectValue /></SelectTrigger>
                <SelectContent>
                  <SelectItem value="low">Low</SelectItem>
                  <SelectItem value="medium">Medium</SelectItem>
                  <SelectItem value="high">High</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </div>
          <div className="space-y-1">
            <Label htmlFor="task-due">Due Date</Label>
            <Input
              id="task-due"
              type="date"
              value={dueDate}
              onChange={(e) => setDueDate(e.target.value)}
              disabled={loading}
            />
          </div>
          <DialogFooter>
            <Button type="button" variant="outline" onClick={onClose} disabled={loading}>
              Cancel
            </Button>
            <Button type="submit" disabled={loading}>
              {loading ? 'Saving…' : isEdit ? 'Save changes' : 'Create task'}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
