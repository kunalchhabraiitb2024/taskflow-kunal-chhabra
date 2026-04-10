import { useState, useEffect, useCallback } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { ArrowLeft, Plus, Trash2, ClipboardList, LayoutGrid, List } from 'lucide-react'
import axios from 'axios'
import { Button } from '@/components/ui/button'
import { TaskCard } from '@/components/tasks/TaskCard'
import { TaskKanbanBoard } from '@/components/tasks/TaskKanbanBoard'
import { TaskForm } from '@/components/tasks/TaskForm'
import { TaskFiltersBar } from '@/components/tasks/TaskFilters'
import { deleteProject, getProject, getProjectStats } from '@/api/projects'
import { deleteTask } from '@/api/tasks'
import { useAuth } from '@/hooks/useAuth'
import { useProjectTaskStream } from '@/hooks/useProjectTaskStream'
import { Badge } from '@/components/ui/badge'
import type { Project, Task, TaskFilters, TaskStats } from '@/types'

type ViewMode = 'list' | 'board'

export function ProjectDetailPage() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const { user } = useAuth()

  const [project, setProject] = useState<Project | null>(null)
  const [allTasks, setAllTasks] = useState<Task[]>([])
  const [stats, setStats] = useState<TaskStats | null>(null)
  const [filters, setFilters] = useState<TaskFilters>({})
  const [viewMode, setViewMode] = useState<ViewMode>('board')
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  const [formOpen, setFormOpen] = useState(false)
  const [editingTask, setEditingTask] = useState<Task | null>(null)

  const refreshProjectAndStats = useCallback(async () => {
    if (!id) return
    try {
      const { data } = await getProject(id)
      setProject(data)
      setAllTasks(data.tasks ?? [])
      const { data: st } = await getProjectStats(id)
      setStats(st)
    } catch {
      /* silent refresh for SSE */
    }
  }, [id])

  const fetchProject = useCallback(async () => {
    if (!id) return
    setLoading(true)
    setError('')
    try {
      const { data } = await getProject(id)
      setProject(data)
      setAllTasks(data.tasks ?? [])
      try {
        const { data: st } = await getProjectStats(id)
        setStats(st)
      } catch {
        setStats(null)
      }
    } catch (err) {
      if (axios.isAxiosError(err) && err.response?.status === 404) {
        setError('Project not found')
      } else {
        setError('Failed to load project')
      }
    } finally {
      setLoading(false)
    }
  }, [id])

  useEffect(() => {
    void fetchProject()
  }, [fetchProject])

  useProjectTaskStream(id, refreshProjectAndStats)

  const visibleTasks = allTasks.filter((t) => {
    if (filters.status && t.status !== filters.status) return false
    if (filters.assignee && t.assignee_id !== filters.assignee) return false
    return true
  })

  const handleTaskSaved = (saved: Task) => {
    setAllTasks((prev) => {
      const idx = prev.findIndex((t) => t.id === saved.id)
      if (idx >= 0) {
        const next = [...prev]
        next[idx] = saved
        return next
      }
      return [saved, ...prev]
    })
    void refreshProjectAndStats()
  }

  const handleTaskUpdate = (updated: Task) => {
    setAllTasks((prev) => prev.map((t) => (t.id === updated.id ? updated : t)))
    void refreshProjectAndStats()
  }

  const handleDeleteTask = async (taskId: string) => {
    if (!confirm('Delete this task?')) return
    try {
      await deleteTask(taskId)
      setAllTasks((prev) => prev.filter((t) => t.id !== taskId))
      void refreshProjectAndStats()
    } catch {
      alert('Failed to delete task')
    }
  }

  const handleDeleteProject = async () => {
    if (!confirm(`Delete project "${project?.name}" and all its tasks? This cannot be undone.`)) return
    try {
      await deleteProject(id!)
      navigate('/projects', { replace: true })
    } catch {
      alert('Failed to delete project')
    }
  }

  const isOwner = project && user && project.owner_id === user.id

  if (loading) {
    return (
      <div className="flex justify-center py-16">
        <div className="h-8 w-8 animate-spin rounded-full border-4 border-primary border-t-transparent" />
      </div>
    )
  }

  if (error) {
    return (
      <div className="rounded-md border border-destructive/30 bg-destructive/10 p-8 text-center">
        <p className="text-destructive">{error}</p>
        <Button variant="outline" className="mt-3" onClick={() => navigate('/projects')}>
          Back to projects
        </Button>
      </div>
    )
  }

  if (!project) return null

  return (
    <div>
      <div className="mb-6">
        <Button variant="ghost" size="sm" className="mb-3 -ml-2" onClick={() => navigate('/projects')}>
          <ArrowLeft className="mr-1 h-4 w-4" /> All projects
        </Button>
        <div className="flex flex-wrap items-start justify-between gap-4">
          <div>
            <h1 className="text-2xl font-bold">{project.name}</h1>
            {project.description && (
              <p className="mt-1 text-muted-foreground">{project.description}</p>
            )}
            {stats && (
              <div className="mt-3 flex flex-wrap items-center gap-2">
                <span className="text-xs text-muted-foreground">Stats (API):</span>
                <Badge variant="secondary" className="text-xs">
                  total {stats.total}
                </Badge>
                {Object.entries(stats.by_status).map(([k, v]) => (
                  <Badge key={k} variant="outline" className="text-xs">
                    {k.replace('_', ' ')}: {v}
                  </Badge>
                ))}
              </div>
            )}
          </div>
          <div className="flex flex-wrap items-center gap-2">
            <div className="mr-1 flex rounded-md border p-0.5">
              <Button
                type="button"
                variant={viewMode === 'board' ? 'secondary' : 'ghost'}
                size="sm"
                className="h-8"
                onClick={() => setViewMode('board')}
              >
                <LayoutGrid className="mr-1 h-4 w-4" /> Board
              </Button>
              <Button
                type="button"
                variant={viewMode === 'list' ? 'secondary' : 'ghost'}
                size="sm"
                className="h-8"
                onClick={() => setViewMode('list')}
              >
                <List className="mr-1 h-4 w-4" /> List
              </Button>
            </div>
            <Button onClick={() => { setEditingTask(null); setFormOpen(true) }}>
              <Plus className="mr-1 h-4 w-4" /> Add Task
            </Button>
            {isOwner && (
              <Button variant="outline" size="icon" onClick={handleDeleteProject} title="Delete project">
                <Trash2 className="h-4 w-4 text-destructive" />
              </Button>
            )}
          </div>
        </div>
      </div>

      <div className="mb-4">
        <TaskFiltersBar filters={filters} onChange={setFilters} />
      </div>

      <p className="mb-4 text-sm text-muted-foreground">
        {visibleTasks.length} task{visibleTasks.length !== 1 ? 's' : ''}
        {filters.status ? ` with status "${filters.status.replace('_', ' ')}"` : ''}
        <span className="ml-2 hidden sm:inline">· Live sync when tasks change (SSE)</span>
      </p>

      {viewMode === 'board' ? (
        visibleTasks.length === 0 ? (
          <div className="flex flex-col items-center justify-center rounded-lg border border-dashed py-16 text-center">
            <ClipboardList className="mb-3 h-10 w-10 text-muted-foreground" />
            <h3 className="text-lg font-medium">No tasks</h3>
            <p className="mt-1 text-sm text-muted-foreground">
              {filters.status ? 'No tasks match the current filter.' : 'Add a task to get started.'}
            </p>
            {!filters.status && (
              <Button className="mt-4" onClick={() => { setEditingTask(null); setFormOpen(true) }}>
                <Plus className="mr-1 h-4 w-4" /> Add task
              </Button>
            )}
          </div>
        ) : (
          <TaskKanbanBoard
            tasks={visibleTasks}
            onTaskUpdated={handleTaskUpdate}
            onEdit={(t) => { setEditingTask(t); setFormOpen(true) }}
            onDelete={handleDeleteTask}
          />
        )
      ) : visibleTasks.length === 0 ? (
        <div className="flex flex-col items-center justify-center rounded-lg border border-dashed py-16 text-center">
          <ClipboardList className="mb-3 h-10 w-10 text-muted-foreground" />
          <h3 className="text-lg font-medium">No tasks</h3>
          <p className="mt-1 text-sm text-muted-foreground">
            {filters.status ? 'No tasks match the current filter.' : 'Add a task to get started.'}
          </p>
          {!filters.status && (
            <Button className="mt-4" onClick={() => { setEditingTask(null); setFormOpen(true) }}>
              <Plus className="mr-1 h-4 w-4" /> Add task
            </Button>
          )}
        </div>
      ) : (
        <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
          {visibleTasks.map((task) => (
            <TaskCard
              key={task.id}
              task={task}
              onEdit={(t) => { setEditingTask(t); setFormOpen(true) }}
              onDelete={handleDeleteTask}
              onUpdate={handleTaskUpdate}
            />
          ))}
        </div>
      )}

      <TaskForm
        projectId={id!}
        task={editingTask}
        open={formOpen}
        onClose={() => setFormOpen(false)}
        onSaved={handleTaskSaved}
      />
    </div>
  )
}
