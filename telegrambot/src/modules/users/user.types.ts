export type UserRole = "freelancer" | "client"

export interface UserProfile {
  id: string
  platformUserId: string
  fullName: string
  email?: string
  role: UserRole
  telegramUserId?: number
  skills: string[]
  rating: number
  completedJobs: number
}

export interface UserSearchQuery {
  skill?: string
  minRating?: number
  limit?: number
}

export interface CheckinEntry {
  id: string
  platformUserId: string
  telegramUserId: number
  plannedHours: number
  checkinAt: Date
  checkoutAt?: Date
}

export interface WeeklyReport {
  platformUserId: string
  weekStart: string
  weekEnd: string
  completedEntries: CheckinEntry[]
  totalTrackedHours: number
}
