import { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { Plus, FolderOpen } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { CreateProjectModal } from '@/components/projects/CreateProjectModal'
import { getProjects } from '@/api/projects'
import { formatDate } from '@/lib/utils'
import type { Project } from '@/types'

const PAGE_SIZE = 8

export function ProjectsPage() {
  const navigate = useNavigate()
  const [projects, setProjects] = useState<Project[]>([])
  const [page, setPage] = useState(1)
  const [reloadKey, setReloadKey] = useState(0)
  const [pagination, setPagination] = useState({
    page: 1,
    limit: PAGE_SIZE,
    total: 0,
    total_pages: 1,
  })
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [showCreate, setShowCreate] = useState(false)

  useEffect(() => {
    let cancelled = false
    setLoading(true)
    setError('')
    ;(async () => {
      try {
        const { data } = await getProjects({ page, limit: PAGE_SIZE })
        if (!cancelled) {
          setProjects(data.data ?? [])
          setPagination(data.pagination)
        }
      } catch {
        if (!cancelled) setError('Failed to load projects')
      } finally {
        if (!cancelled) setLoading(false)
      }
    })()
    return () => {
      cancelled = true
    }
  }, [page, reloadKey])

  return (
    <div>
      <div className="mb-6 flex items-center justify-between">
        <h1 className="text-2xl font-bold">Projects</h1>
        <Button onClick={() => setShowCreate(true)}>
          <Plus className="mr-1 h-4 w-4" /> New Project
        </Button>
      </div>

      {loading && (
        <div className="flex justify-center py-16">
          <div className="h-8 w-8 animate-spin rounded-full border-4 border-primary border-t-transparent" />
        </div>
      )}

      {!loading && error && (
        <div className="rounded-md border border-destructive/30 bg-destructive/10 p-4 text-center">
          <p className="text-destructive">{error}</p>
          <Button
            variant="outline"
            size="sm"
            className="mt-2"
            onClick={() => {
              setReloadKey((k) => k + 1)
            }}
          >
            Retry
          </Button>
        </div>
      )}

      {!loading && !error && projects.length === 0 && (
        <div className="flex flex-col items-center justify-center rounded-lg border border-dashed py-16 text-center">
          <FolderOpen className="mb-3 h-10 w-10 text-muted-foreground" />
          <h3 className="text-lg font-medium">No projects yet</h3>
          <p className="mt-1 text-sm text-muted-foreground">Create your first project to get started</p>
          <Button className="mt-4" onClick={() => setShowCreate(true)}>
            <Plus className="mr-1 h-4 w-4" /> Create project
          </Button>
        </div>
      )}

      {!loading && !error && projects.length > 0 && (
        <>
          <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
            {projects.map((p) => (
              <Card
                key={p.id}
                className="cursor-pointer transition-shadow hover:shadow-md"
                onClick={() => navigate(`/projects/${p.id}`)}
              >
                <CardHeader className="pb-2">
                  <CardTitle className="text-base">{p.name}</CardTitle>
                  {p.description && (
                    <CardDescription className="line-clamp-2">{p.description}</CardDescription>
                  )}
                </CardHeader>
                <CardContent>
                  <p className="text-xs text-muted-foreground">Created {formatDate(p.created_at)}</p>
                </CardContent>
              </Card>
            ))}
          </div>
          {pagination.total_pages > 1 && (
            <div className="mt-6 flex flex-wrap items-center justify-center gap-3">
              <Button
                variant="outline"
                size="sm"
                disabled={page <= 1}
                onClick={() => setPage((p) => Math.max(1, p - 1))}
              >
                Previous
              </Button>
              <span className="text-sm text-muted-foreground">
                Page {pagination.page} of {pagination.total_pages} ({pagination.total} total)
              </span>
              <Button
                variant="outline"
                size="sm"
                disabled={page >= pagination.total_pages}
                onClick={() => setPage((p) => p + 1)}
              >
                Next
              </Button>
            </div>
          )}
        </>
      )}

      <CreateProjectModal
        open={showCreate}
        onClose={() => setShowCreate(false)}
        onCreated={(_project) => {
          setPage(1)
          setReloadKey((k) => k + 1)
        }}
      />
    </div>
  )
}
