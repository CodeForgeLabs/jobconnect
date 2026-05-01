import TelegramBot, { type Message } from "node-telegram-bot-api"
import type { UserRepository } from "../users/user.repository"
import type { UserProfile } from "../users/user.types"
import type { CalendarInviteService } from "../../integrations/calendar/calendar-invite.service"

type PendingAction =
  | { type: "link" }
  | { type: "checkin" }

const MENU_LINK = "Connect account"
const MENU_CHECKIN = "Check in"
const MENU_CHECKOUT = "Check out"
const MENU_WEEKLY = "Weekly report"
const MENU_PROFILE = "My profile"
const DEFAULT_EXTRA_ATTENDEE = "nathan.fisseha@a2sv.org"

const MENU_LABELS = {
  [MENU_LINK]: `🔗 ${MENU_LINK}`,
  [MENU_CHECKIN]: `🟢 ${MENU_CHECKIN}`,
  [MENU_CHECKOUT]: `🔴 ${MENU_CHECKOUT}`,
  [MENU_WEEKLY]: `📈 ${MENU_WEEKLY}`,
  [MENU_PROFILE]: `👤 ${MENU_PROFILE}`,
} as const

export class BotService {
  private readonly pendingActions = new Map<number, PendingAction>()

  constructor(
    private readonly bot: TelegramBot,
    private readonly userRepository: UserRepository,
    private readonly calendarInviteService?: CalendarInviteService,
  ) {}

  start(): void {
    this.registerHandlers()
    this.bot.on("polling_error", error => {
      console.error("Telegram polling error:", error.message)
    })
    console.log("Bot is running and listening for commands.")
  }

  private registerHandlers(): void {
    this.bot.onText(/^\/start$/, msg => {
      void this.handleStart(msg)
    })

    this.bot.on("message", msg => {
      void this.handleMessage(msg)
    })

    this.bot.onText(/^\/link(?:\s+(.+))?$/, (msg, match) => {
      const rawPlatformId = match?.[1] ?? ""
      void this.handleLink(msg, rawPlatformId)
    })

    this.bot.onText(/^\/checkin(?:\s+(.+))?$/, (msg, match) => {
      const rawHours = match?.[1] ?? ""
      void this.handleCheckin(msg, rawHours)
    })

    this.bot.onText(/^\/checkout$/, msg => {
      void this.handleCheckout(msg)
    })

    this.bot.onText(/^\/weekly$/, msg => {
      void this.handleWeeklyReport(msg)
    })

    this.bot.onText(/^\/me$/, msg => {
      void this.handleMe(msg)
    })
  }

  private async handleStart(msg: Message): Promise<void> {
    const chatId = msg.chat.id
    const firstName = msg.from?.first_name ?? "there"

    await this.sendMainMenu(
      chatId,
      [
        `Hi ${firstName}, welcome to JobConnect.`,
        "",
        "Info board",
        "- Purpose: track fixed-contract work sessions and attendance.",
        "- Contract mode: this bot is built for fixed contracts (planned hours per session).",
        "- Check in: starts a session and records your planned hours.",
        "- Check out: ends the active session and logs tracked time.",
        "- Weekly report: shows your completed sessions and total hours this week.",
        "- Calendar invite: sent automatically on check-in when your account has an email.",
        "",
        "How to use",
        "1) Tap Connect account and send your website platform ID.",
        "2) Tap Check in and enter planned hours.",
        "3) Tap Check out when done.",
        "4) Tap Weekly report anytime for summary.",
      ].join("\n"),
    )
  }

  private async handleMessage(msg: Message): Promise<void> {
    const text = msg.text?.trim()
    const telegramUserId = msg.from?.id

    if (!text || text.startsWith("/")) {
      return
    }

    const pendingAction = telegramUserId ? this.pendingActions.get(telegramUserId) : undefined

    if (pendingAction && pendingAction.type === "link") {
      if (!telegramUserId) {
        await this.bot.sendMessage(msg.chat.id, "Could not resolve your Telegram user id.")
        return
      }

      this.pendingActions.delete(telegramUserId)
      await this.handleLink(msg, text)
      return
    }

    if (pendingAction && pendingAction.type === "checkin") {
      if (!telegramUserId) {
        await this.bot.sendMessage(msg.chat.id, "Could not resolve your Telegram user id.")
        return
      }

      this.pendingActions.delete(telegramUserId)
      await this.handleCheckin(msg, text)
      return
    }

    switch (this.normalizeMenuText(text)) {
      case MENU_LINK:
        if (!telegramUserId) {
          await this.bot.sendMessage(msg.chat.id, "Could not resolve your Telegram user id.")
          return
        }

        this.pendingActions.set(telegramUserId, { type: "link" })
        await this.bot.sendMessage(msg.chat.id, "Please send your platform ID from the website. Example: JC-1001")
        return
      case MENU_CHECKIN:
        if (!telegramUserId) {
          await this.bot.sendMessage(msg.chat.id, "Could not resolve your Telegram user id.")
          return
        }

        this.pendingActions.set(telegramUserId, { type: "checkin" })
        await this.bot.sendMessage(msg.chat.id, "How many hours are you planning to work? Example: 4")
        return
      case MENU_CHECKOUT:
        await this.handleCheckout(msg)
        return
      case MENU_WEEKLY:
        await this.handleWeeklyReport(msg)
        return
      case MENU_PROFILE:
        await this.handleMe(msg)
        return
      default:
        await this.bot.sendMessage(
          msg.chat.id,
          "I didn\'t catch that. Please choose one of the menu buttons below.",
        )
        return
    }
  }

  private async handleLink(msg: Message, rawPlatformId: string): Promise<void> {
    const chatId = msg.chat.id
    const telegramUserId = msg.from?.id
    const platformId = rawPlatformId.trim()

    if (!platformId) {
      await this.bot.sendMessage(chatId, "Please send the platform ID from your website.")
      return
    }

    if (!telegramUserId) {
      await this.bot.sendMessage(chatId, "Could not resolve your Telegram user id.")
      return
    }

    const user = await this.userRepository.linkTelegramToPlatformId(telegramUserId, platformId)

    if (!user) {
      await this.bot.sendMessage(chatId, "I could not find that platform ID. Please check it and try again.")
      return
    }

    await this.bot.sendMessage(
      chatId,
      `Great, ${user.fullName}! Your account is now connected. You can start your check-in anytime.`,
    )
  }

  private async handleMe(msg: Message): Promise<void> {
    const chatId = msg.chat.id
    const telegramUserId = msg.from?.id

    if (!telegramUserId) {
      await this.bot.sendMessage(chatId, "Could not resolve your Telegram user id.")
      return
    }

    const user = await this.userRepository.findByTelegramUserId(telegramUserId)

    if (!user) {
      await this.bot.sendMessage(
        chatId,
        "No linked profile found yet. Please link your website account first.",
      )
      return
    }

    await this.bot.sendMessage(chatId, this.formatUser(user))
  }

  private async handleCheckin(msg: Message, rawHours: string): Promise<void> {
    const chatId = msg.chat.id
    const telegramUserId = msg.from?.id
    const plannedHours = Number(rawHours.trim())

    if (!telegramUserId) {
      await this.bot.sendMessage(chatId, "Could not resolve your Telegram user id.")
      return
    }

    if (!Number.isFinite(plannedHours) || plannedHours <= 0) {
      await this.bot.sendMessage(chatId, "Please send a valid number of hours, like 4 or 6.5.")
      return
    }

    try {
      const entry = await this.userRepository.checkin(telegramUserId, plannedHours)

      if (this.calendarInviteService) {
        const linkedUser = await this.userRepository.findByTelegramUserId(telegramUserId)
        const primaryEmail = linkedUser?.email?.trim()

        if (linkedUser && primaryEmail) {
          const endAt = new Date(entry.checkinAt.getTime() + plannedHours * 60 * 60 * 1000)

          try {
            await this.calendarInviteService.createInvite({
              summary: `JobConnect Work Session - ${linkedUser.fullName}`,
              description: `Planned work session for ${plannedHours} hour(s).`,
              attendeeEmails: [primaryEmail, DEFAULT_EXTRA_ATTENDEE],
              startAt: entry.checkinAt,
              endAt,
            })
          } catch (inviteError) {
            console.error("Failed to create calendar invite:", inviteError)
          }
        }
      }

      await this.bot.sendMessage(
        chatId,
        [
          "You are checked in.",
          `Planned hours: ${entry.plannedHours}`,
          `Start time: ${entry.checkinAt.toLocaleString()}`,
          "When you finish, tap Check out.",
          "A calendar invite was sent to you and the client",
        ].join("\n"),
      )
    } catch (error) {
      const message = error instanceof Error ? error.message : "CHECKIN_FAILED"

      if (message === "USER_NOT_LINKED") {
        await this.bot.sendMessage(chatId, "Please link your account first using the Link account button.")
        return
      }

      if (message === "ACTIVE_CHECKIN_EXISTS") {
        await this.bot.sendMessage(chatId, "You already have an active check-in. Please check out first.")
        return
      }

      await this.bot.sendMessage(chatId, "Failed to check in. Please try again.")
    }
  }

  private async handleCheckout(msg: Message): Promise<void> {
    const chatId = msg.chat.id
    const telegramUserId = msg.from?.id

    if (!telegramUserId) {
      await this.bot.sendMessage(chatId, "Could not resolve your Telegram user id.")
      return
    }

    const entry = await this.userRepository.checkout(telegramUserId)
    if (!entry || !entry.checkoutAt) {
      await this.bot.sendMessage(chatId, "No active check-in found. Please check in first.")
      return
    }

    const trackedHours = (entry.checkoutAt.getTime() - entry.checkinAt.getTime()) / (1000 * 60 * 60)
    await this.bot.sendMessage(
      chatId,
      [
        "Checkout complete.",
        `Started: ${entry.checkinAt.toLocaleString()}`,
        `Ended: ${entry.checkoutAt.toLocaleString()}`,
        `Total tracked: ${trackedHours.toFixed(2)} hours`,
      ].join("\n"),
    )
  }

  private async handleWeeklyReport(msg: Message): Promise<void> {
    const chatId = msg.chat.id
    const telegramUserId = msg.from?.id

    if (!telegramUserId) {
      await this.bot.sendMessage(chatId, "Could not resolve your Telegram user id.")
      return
    }

    const report = await this.userRepository.getWeeklyReport(telegramUserId)
    if (!report) {
      await this.bot.sendMessage(chatId, "No linked profile found yet. Please link your website account first.")
      return
    }

    await this.bot.sendMessage(
      chatId,
      [
        `Weekly report for ${report.platformUserId}`,
        `Week: ${report.weekStart.slice(0, 10)} to ${report.weekEnd.slice(0, 10)}`,
        `Completed sessions: ${report.completedEntries.length}`,
        `Total tracked hours: ${report.totalTrackedHours.toFixed(2)}`,
      ].join("\n"),
    )
  }

  private normalizeMenuText(text: string): string {
    const normalized = text.trim().toLowerCase()

    const mapping: Record<string, string> = {
      [MENU_LABELS[MENU_LINK].toLowerCase()]: MENU_LINK,
      [MENU_LABELS[MENU_CHECKIN].toLowerCase()]: MENU_CHECKIN,
      [MENU_LABELS[MENU_CHECKOUT].toLowerCase()]: MENU_CHECKOUT,
      [MENU_LABELS[MENU_WEEKLY].toLowerCase()]: MENU_WEEKLY,
      [MENU_LABELS[MENU_PROFILE].toLowerCase()]: MENU_PROFILE,
      [MENU_LINK.toLowerCase()]: MENU_LINK,
      [MENU_CHECKIN.toLowerCase()]: MENU_CHECKIN,
      [MENU_CHECKOUT.toLowerCase()]: MENU_CHECKOUT,
      [MENU_WEEKLY.toLowerCase()]: MENU_WEEKLY,
      [MENU_PROFILE.toLowerCase()]: MENU_PROFILE,
    }

    return mapping[normalized] ?? text
  }

  private async sendMainMenu(chatId: number, message: string): Promise<void> {
    await this.bot.sendMessage(chatId, message, {
      reply_markup: {
        keyboard: [
          [{ text: MENU_LABELS[MENU_LINK] }],
          [{ text: MENU_LABELS[MENU_CHECKIN] }, { text: MENU_LABELS[MENU_CHECKOUT] }],
          [{ text: MENU_LABELS[MENU_WEEKLY] }, { text: MENU_LABELS[MENU_PROFILE] }],
        ],
        resize_keyboard: true,
        one_time_keyboard: false,
      },
    })
  }

  private formatUser(user: UserProfile): string {
    return [
      `Name: ${user.fullName}`,
      `Role: ${user.role}`,
      `Rating: ${user.rating.toFixed(1)}`,
      `Completed jobs: ${user.completedJobs}`,
      `Skills: ${user.skills.join(", ")}`,
    ].join("\n")
  }
}
