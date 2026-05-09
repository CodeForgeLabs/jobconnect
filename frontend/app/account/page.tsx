"use client";

import { FormEvent, useEffect, useMemo, useState } from "react";
import { useRouter } from "next/navigation";
import { useSelector } from "react-redux";
import {
  clearAuthState,
  selectAuthUser,
  selectIsAuthenticated,
  selectIsHydrated,
  selectUserRole,
} from "@/features/login/loginSlice";
import { authApi } from "@/lib/authApi";
import { userApi } from "@/lib/userApi";
import { store } from "@/store/store";

interface SessionItem {
  session_id: string;
  created_at: string;
  expires_at: string;
  last_used_at?: string;
}

interface ProfileData {
  core?: {
    display_name?: string;
    location?: string;
    contact_phone?: string;
    bio?: string;
  };
}

interface OnboardingData {
  readiness?: {
    percent?: number;
    missing_required_fields?: string[];
    recommendations?: string[];
  };
  completeness?: {
    percent?: number;
    missing_required_fields?: string[];
  };
}

interface SettingsData {
  ui_locale?: string;
  email_notifications_enabled?: boolean;
  push_notifications_enabled?: boolean;
}

function safeDate(value?: string): string {
  if (!value) return "-";
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return value;
  return date.toLocaleString();
}

export default function AccountPage() {
  const router = useRouter();
  const isHydrated = useSelector(selectIsHydrated);
  const isAuthenticated = useSelector(selectIsAuthenticated);
  const role = useSelector(selectUserRole);
  const authUser = useSelector(selectAuthUser);

  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [message, setMessage] = useState<string | null>(null);

  const [profile, setProfile] = useState<ProfileData | null>(null);
  const [onboarding, setOnboarding] = useState<OnboardingData | null>(null);
  const [settings, setSettings] = useState<SettingsData | null>(null);
  const [sessions, setSessions] = useState<SessionItem[]>([]);

  const [newEmail, setNewEmail] = useState("");
  const [emailOtp, setEmailOtp] = useState("");

  const readinessPercent = useMemo(
    () => onboarding?.readiness?.percent ?? onboarding?.completeness?.percent ?? 0,
    [onboarding]
  );
  const missingFields = useMemo(
    () =>
      onboarding?.readiness?.missing_required_fields ??
      onboarding?.completeness?.missing_required_fields ??
      [],
    [onboarding]
  );

  async function loadData() {
    setLoading(true);
    setError(null);
    try {
      const [profileRes, onboardingRes, settingsRes, sessionsRes] = await Promise.all([
        userApi.getProfile() as Promise<{ profile?: ProfileData }>,
        userApi.getOnboardingStatus() as Promise<OnboardingData>,
        userApi.getSettings() as Promise<{ settings?: SettingsData }>,
        authApi.listSessions() as Promise<{ sessions?: SessionItem[] }>,
      ]);

      setProfile(profileRes?.profile ?? null);
      setOnboarding(onboardingRes ?? null);
      setSettings(settingsRes?.settings ?? null);
      setSessions(sessionsRes?.sessions ?? []);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to load account details.");
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    if (!isHydrated) return;
    if (!isAuthenticated) {
      router.replace("/login");
      return;
    }
    loadData();
  }, [isHydrated, isAuthenticated, router]);

  async function runAction(name: string, fn: () => Promise<void>) {
    setSaving(true);
    setError(null);
    setMessage(null);
    try {
      await fn();
      await loadData();
      setMessage(`${name} completed.`);
    } catch (err) {
      setError(err instanceof Error ? err.message : `${name} failed.`);
    } finally {
      setSaving(false);
    }
  }

  async function onLogoutEverywhere() {
    await authApi.logoutEverywhere();
    store.dispatch(clearAuthState());
    router.replace("/login");
  }

  async function onRequestEmailChange(event: FormEvent) {
    event.preventDefault();
    await runAction("Email change request", async () => {
      await authApi.requestEmailChange(newEmail);
    });
  }

  async function onConfirmEmailChange(event: FormEvent) {
    event.preventDefault();
    await runAction("Email change confirmation", async () => {
      await authApi.confirmEmailChange(emailOtp);
    });
  }

  if (!isHydrated || loading) {
    return (
      <main className="min-h-screen bg-[#f7f9f7] px-4 py-10">
        <div className="mx-auto max-w-6xl rounded-2xl border border-[#d7ddd3] bg-white p-8">
          <p className="text-[#5e6d55]">Loading account...</p>
        </div>
      </main>
    );
  }

  return (
    <main className="min-h-screen bg-[#f7f9f7] px-4 py-10">
      <div className="mx-auto max-w-6xl space-y-6">
        <section className="rounded-2xl border border-[#d7ddd3] bg-white p-6">
          <div className="flex flex-wrap items-start justify-between gap-4">
            <div>
              <h1 className="text-3xl font-semibold text-[#1f1f1f]">Account Hub</h1>
              <p className="mt-1 text-sm text-[#5e6d55]">
                {authUser?.email ?? "Unknown email"} | {role}
              </p>
            </div>

            <div className="flex flex-wrap gap-2">
              <button
                type="button"
                className="btn border-[#cdd7c7] bg-white text-[#1f1f1f] hover:bg-[#f7fbf5]"
                onClick={() => router.push("/onboarding")}
              >
                Continue Onboarding
              </button>
              <button
                type="button"
                className="btn border-none bg-[#108a00] text-white hover:bg-[#0d7300]"
                onClick={() => router.push(role === "freelancer" ? "/freelancer/dashboard" : "/client/dashboard")}
              >
                Open Dashboard
              </button>
              <button
                type="button"
                className="btn border border-[#e9c7c7] bg-[#fff7f7] text-[#8f1d1d] hover:bg-[#ffecec]"
                onClick={() => runAction("Logout everywhere", onLogoutEverywhere)}
                disabled={saving}
              >
                Logout Everywhere
              </button>
            </div>
          </div>

          <div className="mt-5">
            <div className="mb-2 flex items-center justify-between text-sm text-[#41543d]">
              <span>Readiness progress</span>
              <span className="font-semibold">{readinessPercent}%</span>
            </div>
            <progress className="progress w-full" max={100} value={readinessPercent} />
            {missingFields.length > 0 && (
              <p className="mt-2 text-sm text-[#7b4d00]">
                Missing: {missingFields.join(", ")}
              </p>
            )}
          </div>

          {message && <p className="mt-3 text-sm text-[#108a00]">{message}</p>}
          {error && <p className="mt-3 text-sm text-[#b42318]">{error}</p>}
        </section>

        <section className="grid gap-6 md:grid-cols-2">
          <div className="rounded-2xl border border-[#d7ddd3] bg-white p-5">
            <h2 className="text-xl font-semibold text-[#1f1f1f]">Profile Summary</h2>
            <div className="mt-4 space-y-2 text-sm text-[#41543d]">
              <p>
                <span className="font-semibold text-[#1f1f1f]">Display Name:</span>{" "}
                {profile?.core?.display_name || "-"}
              </p>
              <p>
                <span className="font-semibold text-[#1f1f1f]">Location:</span>{" "}
                {profile?.core?.location || "-"}
              </p>
              <p>
                <span className="font-semibold text-[#1f1f1f]">Phone:</span>{" "}
                {profile?.core?.contact_phone || "-"}
              </p>
              <p>
                <span className="font-semibold text-[#1f1f1f]">Bio:</span>{" "}
                {profile?.core?.bio || "-"}
              </p>
            </div>
          </div>

          <div className="rounded-2xl border border-[#d7ddd3] bg-white p-5">
            <h2 className="text-xl font-semibold text-[#1f1f1f]">Notification Settings</h2>
            <div className="mt-4 space-y-2 text-sm text-[#41543d]">
              <p>
                <span className="font-semibold text-[#1f1f1f]">Locale:</span>{" "}
                {settings?.ui_locale || "-"}
              </p>
              <p>
                <span className="font-semibold text-[#1f1f1f]">Email Notifications:</span>{" "}
                {settings?.email_notifications_enabled ? "Enabled" : "Disabled"}
              </p>
              <p>
                <span className="font-semibold text-[#1f1f1f]">Push Notifications:</span>{" "}
                {settings?.push_notifications_enabled ? "Enabled" : "Disabled"}
              </p>
            </div>
          </div>
        </section>

        <section className="grid gap-6 md:grid-cols-2">
          <form
            className="rounded-2xl border border-[#d7ddd3] bg-white p-5"
            onSubmit={onRequestEmailChange}
          >
            <h2 className="text-xl font-semibold text-[#1f1f1f]">Request Email Change</h2>
            <p className="mt-1 text-sm text-[#5e6d55]">
              Send an OTP to your new email address.
            </p>
            <input
              className="input input-bordered mt-4 w-full border-[#cfd6ca] bg-white"
              placeholder="new-email@example.com"
              value={newEmail}
              onChange={(event) => setNewEmail(event.target.value)}
              required
            />
            <button
              type="submit"
              className="btn mt-3 border-none bg-[#108a00] text-white hover:bg-[#0d7300]"
              disabled={saving}
            >
              Request OTP
            </button>
          </form>

          <form
            className="rounded-2xl border border-[#d7ddd3] bg-white p-5"
            onSubmit={onConfirmEmailChange}
          >
            <h2 className="text-xl font-semibold text-[#1f1f1f]">Confirm Email Change</h2>
            <p className="mt-1 text-sm text-[#5e6d55]">
              Enter the OTP sent to your new address.
            </p>
            <input
              className="input input-bordered mt-4 w-full border-[#cfd6ca] bg-white"
              placeholder="OTP code"
              value={emailOtp}
              onChange={(event) => setEmailOtp(event.target.value)}
              required
            />
            <button
              type="submit"
              className="btn mt-3 border-none bg-[#108a00] text-white hover:bg-[#0d7300]"
              disabled={saving}
            >
              Confirm
            </button>
          </form>
        </section>

        <section className="rounded-2xl border border-[#d7ddd3] bg-white p-5">
          <h2 className="text-xl font-semibold text-[#1f1f1f]">Active Sessions</h2>
          <div className="mt-4 overflow-auto rounded-xl border border-[#e4e8e2]">
            <table className="table">
              <thead>
                <tr>
                  <th>Session ID</th>
                  <th>Created</th>
                  <th>Last Used</th>
                  <th>Expires</th>
                  <th>Action</th>
                </tr>
              </thead>
              <tbody>
                {sessions.length === 0 && (
                  <tr>
                    <td colSpan={5} className="text-center text-sm text-[#5e6d55]">
                      No active sessions found.
                    </td>
                  </tr>
                )}
                {sessions.map((session) => (
                  <tr key={session.session_id}>
                    <td className="max-w-[220px] truncate">{session.session_id}</td>
                    <td>{safeDate(session.created_at)}</td>
                    <td>{safeDate(session.last_used_at)}</td>
                    <td>{safeDate(session.expires_at)}</td>
                    <td>
                      <button
                        type="button"
                        className="btn btn-sm border border-[#cdd7c7] bg-white text-[#1f1f1f] hover:bg-[#f7fbf5]"
                        onClick={() =>
                          runAction("Session revoked", async () => {
                            await authApi.revokeSession(session.session_id);
                          })
                        }
                        disabled={saving}
                      >
                        Revoke
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </section>
      </div>
    </main>
  );
}
