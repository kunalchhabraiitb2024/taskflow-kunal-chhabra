import { useState } from 'react'
import { Pencil, Trash2, Calendar } from 'lucide-react'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { cn, formatDate, isOverdue } from '@/lib/utils'
import { updateTask } from '@/api/tasks'
import type { Task } from '@/types'

const STATUS_COLORS: Record<Task['status'], string> = {
  todo: 'bg-gray-100 text-gray-700',
  in_progress: 'bg-blue-100 text-blue-700',
  done: 'bg-green-100 text-green-700',
}

const STATUS_LABELS: Record<Task['status'], string> = {
  todo: 'Todo',
  in_progress: 'In Progress',
  done: 'Done',
}

const PRIORITY_COLORS: Record<Task['priority'], string> = {
  low: 'bg-gray-100 text-gray-600',
  medium: 'bg-yellow-100 text-yellow-700',
  high: 'bg-red-100 text-red-700',
}

interface TaskCardProps {
  task: Task
  onEdit: (task: Task) => void
  onDelete: (taskId: string) => void
  onUpdate: (task: Task) => void
}

export function TaskCard({ task, onEdit, onDelete, onUpdate }: TaskCardProps) {
  const [status, setStatus] = useState<Task['status']>(task.status)
  const [updating, setUpdating] = useState(false)

  const handleStatusChange = async (newStatus: Task['status']) => {
    const prev = status
    // Optimistic update
    setStatus(newStatus)
    setUpdating(true)
    try {
      const { data } = await updateTask(task.id, { status: newStatus })
      onUpdate(data)
    } catch {
      // Revert on failure
      setStatus(prev)
    } finally {
      setUpdating(false)
    }
  }

  return (
    <div className="rounded-lg border bg-card p-4 shadow-sm transition-shadow hover:shadow-md">
      <div className="flex items-start justify-between gap-2">
        <div className="flex-1 min-w-0">
          <h3 className="font-medium leading-snug truncate">{task.title}</h3>
          {task.description && (
            <p className="mt-1 text-sm text-muted-foreground line-clamp-2">{task.description}</p>
          )}
        </div>
        <div className="flex items-center gap-1 shrink-0">
          <Button variant="ghost" size="icon" className="h-7 w-7" onClick={() => onEdit(task)}>
            <Pencil className="h-3.5 w-3.5" />
          </Button>
          <Button
            variant="ghost"
            size="icon"
            className="h-7 w-7 text-destructive hover:text-destructive"
            onClick={() => onDelete(task.id)}
          >
            <Trash2 className="h-3.5 w-3.5" />
          </Button>
        </div>
      </div>

      <div className="mt-3 flex flex-wrap items-center gap-2">
        {/* Status — optimistic dropdown */}
        <Select value={status} onValueChange={(v) => handleStatusChange(v as Task['status'])} disabled={updating}>
          <SelectTrigger className={cn('h-6 w-32 text-xs border-0 px-2', STATUS_COLORS[status])}>
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="todo">Todo</SelectItem>
            <SelectItem value="in_progress">In Progress</SelectItem>
            <SelectItem value="done">Done</SelectItem>
          </SelectContent>
        </Select>

        <Badge className={cn('text-xs', PRIORITY_COLORS[task.priority])} variant="outline">
          {task.priority}
        </Badge>

        {task.due_date && (
          <span
            className={cn(
              'flex items-center gap-1 text-xs',
              isOverdue(task.due_date) && status !== 'done' ? 'text-destructive' : 'text-muted-foreground'
            )}
          >
            <Calendar className="h-3 w-3" />
            {formatDate(task.due_date)}
          </span>
        )}
      </div>

      {/* Unused status label for screen readers */}
      <span className="sr-only">{STATUS_LABELS[status]}</span>
    </div>
  )
}
