import { fetchEventSource } from '@microsoft/fetch-event-source'
import { useEffect, useRef } from 'react'

function eventsUrl(projectId: string) {
  const base = (import.meta.env.VITE_API_URL || '').replace(/\/$/, '')
  if (base) return `${base}/projects/${projectId}/events`
  return `/projects/${projectId}/events`
}

/** Subscribes to SSE task-change notifications for a project (requires JWT). */
export function useProjectTaskStream(projectId: string | undefined, onRefresh: () => void) {
  const onRefreshRef = useRef(onRefresh)
  onRefreshRef.current = onRefresh

  useEffect(() => {
    if (!projectId) return
    const token = localStorage.getItem('token')
    if (!token) return

    const ctrl = new AbortController()

    void fetchEventSource(eventsUrl(projectId), {
      signal: ctrl.signal,
      headers: { Authorization: `Bearer ${token}` },
      onmessage(ev) {
        if (ev.event === 'tasks') onRefreshRef.current()
      },
    })

    return () => ctrl.abort()
  }, [projectId])
}
