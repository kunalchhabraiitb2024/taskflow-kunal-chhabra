export interface User {
  id: string
  name: string
  email: string
}

export interface Project {
  id: string
  name: string
  description?: string
  owner_id: string
  created_at: string
  tasks?: Task[]
}

export interface Task {
  id: string
  title: string
  description?: string
  status: 'todo' | 'in_progress' | 'done'
  priority: 'low' | 'medium' | 'high'
  project_id: string
  assignee_id?: string
  created_by: string
  due_date?: string
  created_at: string
  updated_at: string
}

export interface AuthResponse {
  token: string
  user: User
}

export interface ApiError {
  error: string
  fields?: Record<string, string>
}

export interface TaskFilters {
  status?: Task['status']
  assignee?: string
}

/** GET /projects/:id/stats */
export interface TaskStats {
  by_status: Record<string, number>
  by_assignee: { user_id: string | null; name: string; count: number }[]
  total: number
}
