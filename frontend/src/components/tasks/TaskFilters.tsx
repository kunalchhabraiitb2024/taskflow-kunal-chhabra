import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import type { Task, TaskFilters } from '@/types'

interface TaskFiltersProps {
  filters: TaskFilters
  onChange: (filters: TaskFilters) => void
}

const STATUS_OPTIONS: { value: Task['status'] | 'all'; label: string }[] = [
  { value: 'all', label: 'All statuses' },
  { value: 'todo', label: 'Todo' },
  { value: 'in_progress', label: 'In Progress' },
  { value: 'done', label: 'Done' },
]

export function TaskFiltersBar({ filters, onChange }: TaskFiltersProps) {
  return (
    <div className="flex flex-wrap gap-3">
      <Select
        value={filters.status ?? 'all'}
        onValueChange={(v) =>
          onChange({ ...filters, status: v === 'all' ? undefined : (v as Task['status']) })
        }
      >
        <SelectTrigger className="w-40">
          <SelectValue />
        </SelectTrigger>
        <SelectContent>
          {STATUS_OPTIONS.map((o) => (
            <SelectItem key={o.value} value={o.value}>
              {o.label}
            </SelectItem>
          ))}
        </SelectContent>
      </Select>
    </div>
  )
}
