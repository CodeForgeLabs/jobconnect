import type { CheckinEntry, UserProfile, UserSearchQuery, WeeklyReport } from "./user.types"

export interface UserRepository {
  findByTelegramUserId(telegramUserId: number): Promise<UserProfile | null>
  linkTelegramToPlatformId(telegramUserId: number, platformUserId: string): Promise<UserProfile | null>

  
  checkin(telegramUserId: number, plannedHours: number): Promise<CheckinEntry>
  checkout(telegramUserId: number): Promise<CheckinEntry | null>
  getWeeklyReport(telegramUserId: number): Promise<WeeklyReport | null>
}
