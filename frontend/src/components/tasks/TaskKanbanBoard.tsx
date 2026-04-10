import type { ReactNode } from 'react'
import {
  DndContext,
  DragEndEvent,
  PointerSensor,
  closestCorners,
  useDraggable,
  useDroppable,
  useSensor,
  useSensors,
} from '@dnd-kit/core'
import { CSS } from '@dnd-kit/utilities'
import { GripVertical, Pencil, Trash2 } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { cn, formatDate, isOverdue } from '@/lib/utils'
import { updateTask } from '@/api/tasks'
import type { Task } from '@/types'

const COLUMNS: { id: Task['status']; label: string }[] = [
  { id: 'todo', label: 'To do' },
  { id: 'in_progress', label: 'In progress' },
  { id: 'done', label: 'Done' },
]

const PRIORITY_COLORS: Record<Task['priority'], string> = {
  low: 'bg-gray-100 text-gray-600',
  medium: 'bg-yellow-100 text-yellow-700',
  high: 'bg-red-100 text-red-700',
}

function KanbanColumn({
  status,
  label,
  children,
}: {
  status: Task['status']
  label: string
  children: ReactNode
}) {
  const { setNodeRef, isOver } = useDroppable({ id: status })
  return (
    <div
      ref={setNodeRef}
      className={cn(
        'flex min-h-[200px] flex-1 flex-col rounded-lg border bg-muted/30 p-2 transition-colors',
        isOver && 'ring-2 ring-primary ring-offset-2 ring-offset-background'
      )}
    >
      <h3 className="mb-2 px-1 text-xs font-semibold uppercase tracking-wide text-muted-foreground">
        {label}
      </h3>
      <div className="flex flex-1 flex-col gap-2">{children}</div>
    </div>
  )
}

function KanbanTaskCard({
  task,
  onEdit,
  onDelete,
}: {
  task: Task
  onEdit: (t: Task) => void
  onDelete: (id: string) => void
}) {
  const { attributes, listeners, setNodeRef, transform, isDragging } = useDraggable({ id: task.id })
  const style = transform ? { transform: CSS.Translate.toString(transform) } : undefined

  return (
    <div
      ref={setNodeRef}
      style={style}
      className={cn(
        'rounded-md border bg-card p-2 shadow-sm',
        isDragging && 'z-10 opacity-90 shadow-md'
      )}
    >
      <div className="flex items-start gap-1">
        <button
          type="button"
          className="mt-0.5 cursor-grab touch-none rounded p-0.5 text-muted-foreground hover:bg-muted active:cursor-grabbing"
          {...listeners}
          {...attributes}
          aria-label="Drag to change status"
        >
          <GripVertical className="h-4 w-4" />
        </button>
        <div className="min-w-0 flex-1">
          <p className="text-sm font-medium leading-snug">{task.title}</p>
          <div className="mt-1 flex flex-wrap items-center gap-1">
            <Badge className={cn('text-[10px]', PRIORITY_COLORS[task.priority])} variant="outline">
              {task.priority}
            </Badge>
            {task.due_date && (
              <span
                className={cn(
                  'text-[10px]',
                  isOverdue(task.due_date) && task.status !== 'done'
                    ? 'text-destructive'
                    : 'text-muted-foreground'
                )}
              >
                {formatDate(task.due_date)}
              </span>
            )}
          </div>
        </div>
        <div className="flex shrink-0 gap-0.5">
          <Button variant="ghost" size="icon" className="h-7 w-7" onClick={() => onEdit(task)}>
            <Pencil className="h-3.5 w-3.5" />
          </Button>
          <Button
            variant="ghost"
            size="icon"
            className="h-7 w-7 text-destructive"
            onClick={() => onDelete(task.id)}
          >
            <Trash2 className="h-3.5 w-3.5" />
          </Button>
        </div>
      </div>
    </div>
  )
}

interface TaskKanbanBoardProps {
  tasks: Task[]
  onTaskUpdated: (t: Task) => void
  onEdit: (t: Task) => void
  onDelete: (id: string) => void
}

export function TaskKanbanBoard({ tasks, onTaskUpdated, onEdit, onDelete }: TaskKanbanBoardProps) {
  const sensors = useSensors(useSensor(PointerSensor, { activationConstraint: { distance: 6 } }))

  const handleDragEnd = async (event: DragEndEvent) => {
    const { active, over } = event
    if (!over) return
    const taskId = String(active.id)
    const newStatus = over.id as Task['status']
    if (!['todo', 'in_progress', 'done'].includes(newStatus)) return
    const task = tasks.find((t) => t.id === taskId)
    if (!task || task.status === newStatus) return
    try {
      const { data } = await updateTask(taskId, { status: newStatus })
      onTaskUpdated(data)
    } catch {
      alert('Failed to move task')
    }
  }

  return (
    <DndContext
      sensors={sensors}
      collisionDetection={closestCorners}
      onDragEnd={(e) => void handleDragEnd(e)}
    >
      <div className="flex flex-col gap-3 lg:flex-row">
        {COLUMNS.map((col) => (
          <KanbanColumn key={col.id} status={col.id} label={col.label}>
            {tasks
              .filter((t) => t.status === col.id)
              .map((task) => (
                <KanbanTaskCard key={task.id} task={task} onEdit={onEdit} onDelete={onDelete} />
              ))}
          </KanbanColumn>
        ))}
      </div>
    </DndContext>
  )
}
