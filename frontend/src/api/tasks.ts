import api from './client'
import type { Task, TaskFilters } from '@/types'

interface Paginated<T> {
  data: T[]
  pagination: { page: number; limit: number; total: number; total_pages: number }
}

export const getTasks = (projectId: string, filters?: TaskFilters) =>
  api.get<Paginated<Task>>(`/projects/${projectId}/tasks`, { params: filters })

export const createTask = (projectId: string, data: Partial<Omit<Task, 'id' | 'project_id' | 'created_by' | 'created_at' | 'updated_at'>>) =>
  api.post<Task>(`/projects/${projectId}/tasks`, data)

export const updateTask = (id: string, data: Partial<Omit<Task, 'id' | 'project_id' | 'created_by' | 'created_at' | 'updated_at'>>) =>
  api.patch<Task>(`/tasks/${id}`, data)

export const deleteTask = (id: string) =>
  api.delete(`/tasks/${id}`)
