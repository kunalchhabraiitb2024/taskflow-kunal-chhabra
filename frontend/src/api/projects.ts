import api from './client'
import type { Project, TaskStats } from '@/types'

interface Paginated<T> {
  data: T[]
  pagination: { page: number; limit: number; total: number; total_pages: number }
}

export const getProjects = (params?: { page?: number; limit?: number }) =>
  api.get<Paginated<Project>>('/projects', { params })

export const getProject = (id: string) =>
  api.get<Project>(`/projects/${id}`)

export const createProject = (data: { name: string; description?: string }) =>
  api.post<Project>('/projects', data)

export const updateProject = (id: string, data: { name?: string; description?: string }) =>
  api.patch<Project>(`/projects/${id}`, data)

export const deleteProject = (id: string) =>
  api.delete(`/projects/${id}`)

export const getProjectStats = (id: string) =>
  api.get<TaskStats>(`/projects/${id}/stats`)
