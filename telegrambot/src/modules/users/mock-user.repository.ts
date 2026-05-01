import type { CheckinEntry, UserProfile, UserSearchQuery, WeeklyReport } from "./user.types"

import type { UserRepository } from "./user.repository"

const MOCK_USERS: UserProfile[] = [
  {
    id: "u-001",
    platformUserId: "JC-1001",
    email: "natnael.biruk@a2sv.org",
    fullName: "Natnael Biruk",
    role: "freelancer",
    skills: ["typescript", "node.js", "postgres"],
    rating: 4.9,
    completedJobs: 37,
  },
  {
    id: "u-002",
    platformUserId: "JC-1002",
    email: "marta.bekele@a2sv.org",
    fullName: "Marta Bekele",
    role: "freelancer",
    skills: ["go", "grpc", "docker"],
    rating: 4.8,
    completedJobs: 29,
  },
 
]

export class MockUserRepository implements UserRepository {
  private readonly users: UserProfile[]
  private readonly checkins: CheckinEntry[]

  constructor(seed: UserProfile[] = MOCK_USERS) {
    this.users = [...seed]
    this.checkins = []
  }

  async findByTelegramUserId(telegramUserId: number): Promise<UserProfile | null> {
    const user = this.users.find(item => item.telegramUserId === telegramUserId)
    return user ?? null
  }

  async linkTelegramToPlatformId(telegramUserId: number, platformUserId: string): Promise<UserProfile | null> {
    const normalizedId = platformUserId.trim().toUpperCase()
    const user = this.users.find(item => item.platformUserId.toUpperCase() === normalizedId)

    if (!user) {
      return null
    }

    user.telegramUserId = telegramUserId
    console.log(`Linked Telegram user ${telegramUserId} to platform ID ${platformUserId}`)
    return user
  }

 

 

  async checkin(telegramUserId: number, plannedHours: number): Promise<CheckinEntry> {
    const user = await this.findByTelegramUserId(telegramUserId)
    if (!user) {
      throw new Error("USER_NOT_LINKED")
    }

    const existing = this.checkins.find(entry => entry.telegramUserId === telegramUserId && !entry.checkoutAt)
    if (existing) {
      throw new Error("ACTIVE_CHECKIN_EXISTS")
    }

    const entry: CheckinEntry = {
      id: `ci-${Date.now()}`,
      platformUserId: user.platformUserId,
      telegramUserId,
      plannedHours,
      checkinAt: new Date(),
    }

    this.checkins.push(entry)
    return entry
  }

  async checkout(telegramUserId: number): Promise<CheckinEntry | null> {
    const active = this.checkins.find(entry => entry.telegramUserId === telegramUserId && !entry.checkoutAt)
    if (!active) {
      return null
    }

    active.checkoutAt = new Date()
    return active
  }

  async getWeeklyReport(telegramUserId: number): Promise<WeeklyReport | null> {
    const user = await this.findByTelegramUserId(telegramUserId)
    if (!user) {
      return null
    }

    const now = new Date()
    const start = this.getStartOfWeek(now)
    const end = new Date(start)
    end.setDate(end.getDate() + 6)
    end.setHours(23, 59, 59, 999)

    const completedEntries = this.checkins.filter(entry => {
      if (entry.telegramUserId !== telegramUserId || !entry.checkoutAt) {
        return false
      }

      return entry.checkinAt >= start && entry.checkinAt <= end
    })

    const totalTrackedHours = completedEntries.reduce((sum, entry) => {
      if (!entry.checkoutAt) {
        return sum
      }

      const ms = entry.checkoutAt.getTime() - entry.checkinAt.getTime()
      return sum + ms / (1000 * 60 * 60)
    }, 0)

    return {
      platformUserId: user.platformUserId,
      weekStart: start.toISOString(),
      weekEnd: end.toISOString(),
      completedEntries,
      totalTrackedHours,
    }
  }

  private getStartOfWeek(date: Date): Date {
    const start = new Date(date)
    const day = start.getDay()
    const diff = day === 0 ? -6 : 1 - day
    start.setDate(start.getDate() + diff)
    start.setHours(0, 0, 0, 0)
    return start
  }
}
