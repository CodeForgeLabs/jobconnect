import { google } from "googleapis"
import dotenv from "dotenv"
import path from "path"
import fs from "fs"
import readline from "readline"

dotenv.config({ path: path.resolve(__dirname, "../.env") })

const tokenPath = path.resolve(__dirname, "../oauth-token.json")
const scopes = ["https://www.googleapis.com/auth/calendar"]

function getRequiredEnv(name: string): string {
  const value = process.env[name]
  if (!value || value.trim().length === 0) {
    throw new Error(`Missing required env var: ${name}`)
  }

  return value.trim()
}

function askQuestion(question: string): Promise<string> {
  const rl = readline.createInterface({ input: process.stdin, output: process.stdout })
  return new Promise(resolve => {
    rl.question(question, answer => {
      rl.close()
      resolve(answer)
    })
  })
}

async function getOAuthClient() {
  const clientId = getRequiredEnv("GOOGLE_OAUTH_CLIENT_ID")
  const clientSecret = getRequiredEnv("GOOGLE_OAUTH_CLIENT_SECRET")
  const redirectUri = process.env.GOOGLE_OAUTH_REDIRECT_URI?.trim() || "http://localhost"

  const oauth2Client = new google.auth.OAuth2(clientId, clientSecret, redirectUri)

  if (fs.existsSync(tokenPath)) {
    const tokenRaw = fs.readFileSync(tokenPath, "utf8")
    oauth2Client.setCredentials(JSON.parse(tokenRaw))
    return oauth2Client
  }

  const authUrl = oauth2Client.generateAuthUrl({
    access_type: "offline",
    scope: scopes,
    prompt: "consent",
  })

  console.log("Authorize this app by visiting this URL:")
  console.log(authUrl)

  const code = (await askQuestion("\nPaste authorization code here: ")).trim()
  if (!code) {
    throw new Error("No authorization code provided.")
  }

  const { tokens } = await oauth2Client.getToken(code)
  oauth2Client.setCredentials(tokens)
  fs.writeFileSync(tokenPath, JSON.stringify(tokens, null, 2))
  console.log(`Saved OAuth token to ${tokenPath}`)

  return oauth2Client
}

async function testCalendar() {
  try {
    const auth = await getOAuthClient()
    const calendar = google.calendar({ version: "v3", auth })
    console.log("OAuth authentication loaded successfully.")

    // Include hidden calendars to avoid false negatives.
    const res = await calendar.calendarList.list({
      showHidden: true,
      showDeleted: false,
      minAccessRole: "reader",
    })

    const items = res.data.items ?? []
    console.log("Calendars accessible by user:", items.length)
    items.forEach(cal => console.log(`- ${cal.summary} (${cal.id})`))

    const calendarId = process.env.CALENDAR_ID
    if (calendarId) {
      const one = await calendar.calendars.get({ calendarId })
      console.log("Direct calendar access OK:", one.data.summary, `(${one.data.id})`)

      const eventsRes = await calendar.events.list({
        calendarId,
        timeMin: new Date().toISOString(),
        singleEvents: true,
        orderBy: "startTime",
        maxResults: 10,
      })

      const events = eventsRes.data.items ?? []
      console.log(`Upcoming events (${events.length}):`)
      if (events.length === 0) {
        console.log("- No upcoming events found.")
      }

      events.forEach((event, idx) => {
        const start = event.start?.dateTime ?? event.start?.date ?? "No start"
        const end = event.end?.dateTime ?? event.end?.date ?? "No end"
        console.log(`${idx + 1}. ${event.summary ?? "(No title)"}`)
        console.log(`   Start: ${start}`)
        console.log(`   End:   ${end}`)
      })

      const attendeeEmails = (process.env.ATTENDEE_EMAILS ?? process.env.ATTENDEE_EMAIL ?? "natnael.biruk@a2sv.org")
        .split(",")
        .map(email => email.trim())
        .filter(email => email.length > 0)
      const shouldCreateInvite = (process.env.CREATE_TEST_INVITE ?? "true").toLowerCase() === "true"

      if (shouldCreateInvite) {
        const startDate = new Date(Date.now() + 10 * 60 * 1000)
        const endDate = new Date(startDate.getTime() + 30 * 60 * 1000)

        const created = await calendar.events.insert({
          calendarId,
          sendUpdates: "all",
          requestBody: {
            summary: `JobConnect Bot Test Invite ${new Date().toISOString()}`,
            description: "Temporary invite created by telegram bot integration test.",
            start: { dateTime: startDate.toISOString() },
            end: { dateTime: endDate.toISOString() },
            attendees: attendeeEmails.map(email => ({ email })),
          },
        })

        console.log("Invite created successfully:")
        console.log("- Event ID:", created.data.id)
        console.log("- Event Link:", created.data.htmlLink)
        console.log("- Invited:", attendeeEmails.join(", "))
      } else {
        console.log("Skipping invite creation. Set CREATE_TEST_INVITE=true to send an invite.")
      }
    } else {
      console.log("Tip: set CALENDAR_ID in .env to test direct access to one calendar.")
    }
  } catch (err) {
    const anyErr = err as {
      code?: number
      message?: string
      cause?: { message?: string }
    }

    if (anyErr?.code === 401) {
      console.error("Authentication failed (401). Try refreshing the OAuth token.")
      console.error("1) Delete oauth-token.json")
      console.error("2) Run: npx ts-node test.ts")
      console.error("3) Complete consent again and paste a fresh code")
      return
    }

    console.error("Error accessing calendar:", anyErr?.cause?.message ?? anyErr?.message ?? err)
  }
}

testCalendar()