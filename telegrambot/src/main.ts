import TelegramBot from "node-telegram-bot-api"
import { getRequiredEnv } from "./config/env"
import { BotService } from "./modules/session/bot.service"
import { MockUserRepository } from "./modules/users/mock-user.repository"
import { GoogleCalendarInviteService } from "./integrations/calendar/calendar-invite.service"

function bootstrap(): void {
  const token = getRequiredEnv("TELEGRAM_BOT_TOKEN")
  const bot = new TelegramBot(token, { polling: true })
  const userRepository = new MockUserRepository()
  const calendarInviteService = new GoogleCalendarInviteService()
  const service = new BotService(bot, userRepository, calendarInviteService)

  service.start()
}

bootstrap()
