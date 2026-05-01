import fs from "fs"
import path from "path"
import { google } from "googleapis"
import { getRequiredEnv } from "../../config/env"

const SCOPES = ["https://www.googleapis.com/auth/calendar"]

export interface CalendarInviteInput {
  summary: string
  description?: string
  attendeeEmails: string[]
  startAt: Date
  endAt: Date
}

export interface CalendarInviteResult {
  eventId: string
  htmlLink?: string
}

export interface CalendarInviteService {
  createInvite(input: CalendarInviteInput): Promise<CalendarInviteResult>
}

export class GoogleCalendarInviteService implements CalendarInviteService {
  private readonly calendarId: string
  private readonly tokenPath: string

  constructor(options?: { calendarId?: string; tokenPath?: string }) {
    this.calendarId = options?.calendarId ?? getRequiredEnv("CALENDAR_ID")
    this.tokenPath = options?.tokenPath ?? path.resolve(process.cwd(), "oauth-token.json")
  }

  async createInvite(input: CalendarInviteInput): Promise<CalendarInviteResult> {
    const auth = await this.getOAuthClient()
    const calendar = google.calendar({ version: "v3", auth })
    const attendeeEmails = input.attendeeEmails
      .map(email => email.trim())
      .filter(email => email.length > 0)

    if (attendeeEmails.length === 0) {
      throw new Error("At least one attendee email is required.")
    }

    const created = await calendar.events.insert({
      calendarId: this.calendarId,
      sendUpdates: "all",
      requestBody: {
        summary: input.summary,
        description: input.description,
        start: { dateTime: input.startAt.toISOString() },
        end: { dateTime: input.endAt.toISOString() },
        attendees: attendeeEmails.map(email => ({ email })),
      },
    })

    return {
      eventId: created.data.id ?? "",
      htmlLink: created.data.htmlLink ?? undefined,
    }
  }

  private async getOAuthClient() {
    const clientId = getRequiredEnv("GOOGLE_OAUTH_CLIENT_ID")
    const clientSecret = getRequiredEnv("GOOGLE_OAUTH_CLIENT_SECRET")
    const redirectUri = process.env.GOOGLE_OAUTH_REDIRECT_URI?.trim() || "http://localhost"

    if (!fs.existsSync(this.tokenPath)) {
      throw new Error(
        `Google OAuth token file not found at ${this.tokenPath}. Run src/test.ts once to generate it.`,
      )
    }

    const tokenRaw = fs.readFileSync(this.tokenPath, "utf8")
    const oauth2Client = new google.auth.OAuth2(clientId, clientSecret, redirectUri)
    oauth2Client.setCredentials(JSON.parse(tokenRaw))

    // Keep scope reference explicit so future auth refresh flow knows intended access.
    void SCOPES
    return oauth2Client
  }
}
