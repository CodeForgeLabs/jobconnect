"use client";

import { ChangeEvent, useEffect, useMemo, useState } from "react";
import { useRouter } from "next/navigation";
import { useSelector } from "react-redux";
import {
  selectIsAuthenticated,
  selectIsHydrated,
  selectUserRole,
} from "@/features/login/loginSlice";
import { userApi } from "@/lib/userApi";

interface ProfileData {
  core?: {
    display_name?: string;
    contact_email?: string;
    contact_phone?: string;
    location?: string;
    bio?: string;
    tax_id?: string;
    avatar_url?: string;
  };
  client?: {
    company_name?: string;
    verification_status?: string;
  };
  freelancer?: {
    headline?: string;
    skills?: string[];
    hourly_rate?: number;
    availability?: string;
    verification_status?: string;
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
  steps?: Array<{
    key?: string;
    status?: string;
  }>;
}

interface SettingsData {
  ui_locale?: string;
  email_notifications_enabled?: boolean;
  push_notifications_enabled?: boolean;
}

interface HiringPreferences {
  min_hourly_rate?: number;
  max_hourly_rate?: number;
  preferred_locations?: string[];
}

interface WorkPreferences {
  preferred_project_length?: string;
  min_budget?: number;
  max_budget?: number;
  contract_types?: string[];
  weekly_capacity_hours?: number;
}

interface PresignedUploadResponse {
  storage_key: string;
  upload_url: string;
}

interface PortfolioItemSummary {
  id?: number;
}

type RoleType = "client" | "freelancer" | "admin" | "unknown";
type StepKey = "profile" | "avatar" | "preferences" | "kyc";

function parseNumber(value: string): number | undefined {
  const trimmed = value.trim();
  if (!trimmed) return undefined;
  const parsed = Number(trimmed);
  if (!Number.isFinite(parsed)) return undefined;
  return parsed;
}

function csvToArray(value: string): string[] {
  return value
    .split(",")
    .map((item) => item.trim())
    .filter(Boolean);
}

function mediaTypeForContentType(contentType: string):
  | "PORTFOLIO_MEDIA_TYPE_IMAGE"
  | "PORTFOLIO_MEDIA_TYPE_VIDEO"
  | "PORTFOLIO_MEDIA_TYPE_FILE"
  | null {
  if (contentType.startsWith("image/")) return "PORTFOLIO_MEDIA_TYPE_IMAGE";
  if (contentType.startsWith("video/")) return "PORTFOLIO_MEDIA_TYPE_VIDEO";
  if (contentType === "application/pdf") return "PORTFOLIO_MEDIA_TYPE_FILE";
  return null;
}

async function uploadToPresignedUrl(url: string, file: File) {
  const response = await fetch(url, {
    method: "PUT",
    headers: { "Content-Type": file.type },
    body: file,
  });
  if (!response.ok) {
    throw new Error("File upload failed.");
  }
}

const STEP_ORDER: StepKey[] = ["profile", "avatar", "preferences", "kyc"];

export default function OnboardingPage() {
  const router = useRouter();
  const isHydrated = useSelector(selectIsHydrated);
  const isAuthenticated = useSelector(selectIsAuthenticated);
  const role = useSelector(selectUserRole) as RoleType;

  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [message, setMessage] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);

  const [activeStep, setActiveStep] = useState(0);
  const stepKey = STEP_ORDER[activeStep];

  const [readinessPercent, setReadinessPercent] = useState(0);
  const [missingRequired, setMissingRequired] = useState<string[]>([]);
  const [recommendations, setRecommendations] = useState<string[]>([]);
  const [portfolioCount, setPortfolioCount] = useState(0);
  const [hasCv, setHasCv] = useState(false);
  const [avatarExists, setAvatarExists] = useState(false);
  const [verificationStatus, setVerificationStatus] = useState("");

  const [displayName, setDisplayName] = useState("");
  const [contactEmail, setContactEmail] = useState("");
  const [contactPhone, setContactPhone] = useState("");
  const [location, setLocation] = useState("");
  const [bio, setBio] = useState("");
  const [taxId, setTaxId] = useState("");

  const [companyName, setCompanyName] = useState("");
  const [headline, setHeadline] = useState("");
  const [skillsCsv, setSkillsCsv] = useState("");
  const [hourlyRate, setHourlyRate] = useState("");
  const [availability, setAvailability] = useState("AVAILABILITY_AS_NEEDED");

  const [uiLocale, setUiLocale] = useState("en");
  const [emailNotifications, setEmailNotifications] = useState(true);
  const [pushNotifications, setPushNotifications] = useState(true);

  const [minHourlyRate, setMinHourlyRate] = useState("");
  const [maxHourlyRate, setMaxHourlyRate] = useState("");
  const [preferredLocationsCsv, setPreferredLocationsCsv] = useState("");

  const [preferredProjectLength, setPreferredProjectLength] = useState("");
  const [minBudget, setMinBudget] = useState("");
  const [maxBudget, setMaxBudget] = useState("");
  const [contractTypesCsv, setContractTypesCsv] = useState("");
  const [weeklyCapacityHours, setWeeklyCapacityHours] = useState("");

  const [avatarFile, setAvatarFile] = useState<File | null>(null);
  const [avatarUpload, setAvatarUpload] = useState<PresignedUploadResponse | null>(null);

  const [cvFile, setCvFile] = useState<File | null>(null);
  const [cvUpload, setCvUpload] = useState<PresignedUploadResponse | null>(null);

  const [portfolioTitle, setPortfolioTitle] = useState("");
  const [portfolioDescription, setPortfolioDescription] = useState("");
  const [portfolioProjectUrl, setPortfolioProjectUrl] = useState("");
  const [portfolioRoleInProject, setPortfolioRoleInProject] = useState("");
  const [portfolioTagsCsv, setPortfolioTagsCsv] = useState("");
  const [portfolioMediaFile, setPortfolioMediaFile] = useState<File | null>(null);
  const [portfolioUpload, setPortfolioUpload] = useState<PresignedUploadResponse | null>(null);

  const isClient = role === "client";
  const isFreelancer = role === "freelancer";

  const stepTitle = useMemo(() => {
    if (stepKey === "profile") return "Profile";
    if (stepKey === "avatar") return "Avatar";
    if (stepKey === "preferences") return "Preferences";
    return "Compliance";
  }, [stepKey]);

  const stepSubtitle = useMemo(() => {
    if (stepKey === "profile") {
      return "Set your professional basics and role details.";
    }
    if (stepKey === "avatar") {
      return "Upload a clear profile photo to build trust.";
    }
    if (stepKey === "preferences") {
      if (!isClient && !isFreelancer) {
        return "Set notification defaults and complete account preferences.";
      }
      return isFreelancer
        ? "Set work preferences, upload CV, and add at least one portfolio item."
        : "Set hiring preferences and notification defaults.";
    }
    return "Confirm tax and verification details before finishing onboarding.";
  }, [isClient, isFreelancer, stepKey]);

  async function loadData() {
    setLoading(true);
    setError(null);
    try {
      const [profileRes, onboardingRes, settingsRes] = await Promise.all([
        userApi.getProfile() as Promise<{ profile?: ProfileData }>,
        userApi.getOnboardingStatus() as Promise<OnboardingData>,
        userApi.getSettings() as Promise<{ settings?: SettingsData }>,
      ]);

      const profile = profileRes.profile;
      const core = profile?.core;
      const client = profile?.client;
      const freelancer = profile?.freelancer;
      const settings = settingsRes.settings;

      setDisplayName(core?.display_name ?? "");
      setContactEmail(core?.contact_email ?? "");
      setContactPhone(core?.contact_phone ?? "");
      setLocation(core?.location ?? "");
      setBio(core?.bio ?? "");
      setTaxId(core?.tax_id ?? "");
      setAvatarExists(Boolean(core?.avatar_url));

      setCompanyName(client?.company_name ?? "");
      setHeadline(freelancer?.headline ?? "");
      setSkillsCsv((freelancer?.skills ?? []).join(", "));
      setHourlyRate(
        typeof freelancer?.hourly_rate === "number"
          ? String(freelancer.hourly_rate)
          : ""
      );
      setAvailability(freelancer?.availability ?? "AVAILABILITY_AS_NEEDED");

      if (isClient) {
        setVerificationStatus(client?.verification_status ?? "");
      } else if (isFreelancer) {
        setVerificationStatus(freelancer?.verification_status ?? "");
      } else {
        setVerificationStatus("");
      }

      setUiLocale(settings?.ui_locale ?? "en");
      setEmailNotifications(settings?.email_notifications_enabled ?? true);
      setPushNotifications(settings?.push_notifications_enabled ?? true);

      setReadinessPercent(
        onboardingRes.readiness?.percent ?? onboardingRes.completeness?.percent ?? 0
      );
      setMissingRequired(
        onboardingRes.readiness?.missing_required_fields ??
          onboardingRes.completeness?.missing_required_fields ??
          []
      );
      setRecommendations(onboardingRes.readiness?.recommendations ?? []);

      if (isClient) {
        try {
          const hiring = (await userApi.getHiringPreferences()) as {
            preferences?: HiringPreferences;
          };
          setMinHourlyRate(
            typeof hiring.preferences?.min_hourly_rate === "number"
              ? String(hiring.preferences.min_hourly_rate)
              : ""
          );
          setMaxHourlyRate(
            typeof hiring.preferences?.max_hourly_rate === "number"
              ? String(hiring.preferences.max_hourly_rate)
              : ""
          );
          setPreferredLocationsCsv(
            (hiring.preferences?.preferred_locations ?? []).join(", ")
          );
        } catch {
          // Missing preferences is expected for first-time users.
        }
      }

      if (isFreelancer) {
        try {
          const work = (await userApi.getWorkPreferences()) as {
            settings?: WorkPreferences;
          };
          setPreferredProjectLength(work.settings?.preferred_project_length ?? "");
          setMinBudget(
            typeof work.settings?.min_budget === "number"
              ? String(work.settings.min_budget)
              : ""
          );
          setMaxBudget(
            typeof work.settings?.max_budget === "number"
              ? String(work.settings.max_budget)
              : ""
          );
          setContractTypesCsv((work.settings?.contract_types ?? []).join(", "));
          setWeeklyCapacityHours(
            typeof work.settings?.weekly_capacity_hours === "number"
              ? String(work.settings.weekly_capacity_hours)
              : ""
          );
        } catch {
          // Missing preferences is expected for first-time users.
        }

        try {
          const portfolio = (await userApi.listPortfolio(1, "")) as {
            items?: PortfolioItemSummary[];
          };
          setPortfolioCount(portfolio.items?.length ?? 0);
        } catch {
          setPortfolioCount(0);
        }

        try {
          const cv = (await userApi.getCV()) as {
            cv?: { user_id?: string };
          };
          setHasCv(Boolean(cv.cv?.user_id));
        } catch {
          setHasCv(false);
        }
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to load onboarding.");
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
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [isHydrated, isAuthenticated, router, role]);

  async function runAction(name: string, fn: () => Promise<void>) {
    setSaving(true);
    setError(null);
    setMessage(null);
    try {
      await fn();
      await loadData();
      setMessage(`${name} saved.`);
      return true;
    } catch (err) {
      setError(err instanceof Error ? err.message : `${name} failed.`);
      return false;
    } finally {
      setSaving(false);
    }
  }

  async function saveProfile() {
    const rolePatch = isClient
      ? {
          company_name: companyName || undefined,
        }
      : isFreelancer
        ? {
            headline: headline || undefined,
            skills: csvToArray(skillsCsv),
            hourly_rate: parseNumber(hourlyRate),
            availability: availability || undefined,
          }
        : {};

    return runAction("Profile", async () => {
      await userApi.patchProfile({
        display_name: displayName || undefined,
        contact_email: contactEmail || undefined,
        contact_phone: contactPhone || undefined,
        location: location || undefined,
        bio: bio || undefined,
        ...rolePatch,
      });
    });
  }

  async function reserveAvatarUploadUrl() {
    if (!avatarFile) {
      setError("Select an image file first.");
      return;
    }
    await runAction("Avatar upload URL", async () => {
      const reserved = (await userApi.requestAvatarUploadUrl(
        avatarFile.name,
        avatarFile.type
      )) as PresignedUploadResponse;
      setAvatarUpload(reserved);
    });
  }

  async function uploadAvatar() {
    if (!avatarFile || !avatarUpload) {
      setError("Reserve upload URL and select an avatar first.");
      return;
    }
    await runAction("Avatar", async () => {
      await uploadToPresignedUrl(avatarUpload.upload_url, avatarFile);
      await userApi.upsertAvatar({
        storage_key: avatarUpload.storage_key,
        file_name: avatarFile.name,
        content_type: avatarFile.type,
      });
      setAvatarExists(true);
    });
  }

  async function reserveCvUploadUrl() {
    if (!cvFile) {
      setError("Select a CV file first.");
      return;
    }
    await runAction("CV upload URL", async () => {
      const reserved = (await userApi.requestCVUploadUrl(
        cvFile.name,
        cvFile.type
      )) as PresignedUploadResponse;
      setCvUpload(reserved);
    });
  }

  async function uploadCv() {
    if (!cvFile || !cvUpload) {
      setError("Reserve upload URL and select a CV first.");
      return;
    }
    await runAction("CV", async () => {
      await uploadToPresignedUrl(cvUpload.upload_url, cvFile);
      await userApi.upsertCV({
        storage_key: cvUpload.storage_key,
        file_name: cvFile.name,
        content_type: cvFile.type,
      });
      setHasCv(true);
    });
  }

  async function reservePortfolioUploadUrl() {
    if (!portfolioMediaFile) {
      setError("Choose portfolio media file first.");
      return;
    }
    await runAction("Portfolio upload URL", async () => {
      const reserved = (await userApi.requestPortfolioUploadUrl(
        portfolioMediaFile.name,
        portfolioMediaFile.type
      )) as PresignedUploadResponse;
      setPortfolioUpload(reserved);
    });
  }

  async function addPortfolioItem() {
    if (!portfolioTitle.trim()) {
      setError("Portfolio title is required.");
      return;
    }
    if (!portfolioDescription.trim()) {
      setError("Portfolio description is required.");
      return;
    }

    await runAction("Portfolio item", async () => {
      const media: Array<Record<string, unknown>> = [];

      if (portfolioMediaFile) {
        if (!portfolioUpload) {
          throw new Error("Reserve portfolio upload URL before adding media.");
        }
        const mediaType = mediaTypeForContentType(portfolioMediaFile.type);
        if (!mediaType) {
          throw new Error("Unsupported portfolio media type.");
        }
        await uploadToPresignedUrl(portfolioUpload.upload_url, portfolioMediaFile);
        media.push({
          media_type: mediaType,
          storage_key: portfolioUpload.storage_key,
          content_type: portfolioMediaFile.type,
          file_name: portfolioMediaFile.name,
        });
      }

      await userApi.createPortfolioItem({
        title: portfolioTitle.trim(),
        description: portfolioDescription.trim(),
        project_url: portfolioProjectUrl.trim() || undefined,
        role_in_project: portfolioRoleInProject.trim() || undefined,
        tags: csvToArray(portfolioTagsCsv),
        media,
      });

      setPortfolioTitle("");
      setPortfolioDescription("");
      setPortfolioProjectUrl("");
      setPortfolioRoleInProject("");
      setPortfolioTagsCsv("");
      setPortfolioMediaFile(null);
      setPortfolioUpload(null);
      setPortfolioCount((count) => count + 1);
    });
  }

  async function savePreferences() {
    return runAction("Preferences", async () => {
      if (isClient) {
        await userApi.patchHiringPreferences({
          min_hourly_rate: parseNumber(minHourlyRate),
          max_hourly_rate: parseNumber(maxHourlyRate),
          preferred_locations: csvToArray(preferredLocationsCsv),
        });
      }

      if (isFreelancer) {
        await userApi.patchWorkPreferences({
          preferred_project_length: preferredProjectLength || undefined,
          min_budget: parseNumber(minBudget),
          max_budget: parseNumber(maxBudget),
          contract_types: csvToArray(contractTypesCsv),
          weekly_capacity_hours: parseNumber(weeklyCapacityHours),
        });
      }

      await userApi.patchSettings({
        ui_locale: uiLocale || "en",
        email_notifications_enabled: emailNotifications,
        push_notifications_enabled: pushNotifications,
      });
    });
  }

  async function saveCompliance() {
    return runAction("Compliance", async () => {
      await userApi.patchProfile({
        tax_id: taxId || undefined,
      });
    });
  }

  function onAvatarFileChange(event: ChangeEvent<HTMLInputElement>) {
    setAvatarFile(event.target.files?.[0] ?? null);
    setAvatarUpload(null);
  }

  function onCvFileChange(event: ChangeEvent<HTMLInputElement>) {
    setCvFile(event.target.files?.[0] ?? null);
    setCvUpload(null);
  }

  function onPortfolioMediaFileChange(event: ChangeEvent<HTMLInputElement>) {
    setPortfolioMediaFile(event.target.files?.[0] ?? null);
    setPortfolioUpload(null);
  }

  async function onContinue() {
    if (stepKey === "profile") {
      const ok = await saveProfile();
      if (!ok) return;
    }

    if (stepKey === "avatar") {
      const avatarRequired = missingRequired.includes("avatar");
      if (avatarRequired && !avatarExists) {
        setError("Upload your avatar to continue.");
        return;
      }
    }

    if (stepKey === "preferences") {
      const ok = await savePreferences();
      if (!ok) return;

      if (isFreelancer && missingRequired.includes("cv") && !hasCv) {
        setError("Upload your CV to continue.");
        return;
      }
      if (isFreelancer && missingRequired.includes("portfolio") && portfolioCount < 1) {
        setError("Add at least one portfolio item to continue.");
        return;
      }
    }

    if (stepKey === "kyc") {
      const ok = await saveCompliance();
      if (!ok) return;

      const unresolvedBlocking =
        missingRequired.includes("core_profile") ||
        missingRequired.includes("role_profile") ||
        missingRequired.includes("avatar");

      if (unresolvedBlocking) {
        setError("Complete required profile fields before finishing.");
        return;
      }

      if (role === "freelancer") {
        router.push("/freelancer/dashboard");
      } else if (role === "client") {
        router.push("/client/dashboard");
      } else {
        router.push("/account");
      }
      return;
    }

    setActiveStep((current) => Math.min(current + 1, STEP_ORDER.length - 1));
  }

  function onBack() {
    if (activeStep === 0) {
      router.push("/account");
      return;
    }
    setActiveStep((current) => Math.max(current - 1, 0));
  }

  if (!isHydrated || loading) {
    return (
      <main className="min-h-screen bg-[#f7f9f7] px-4 py-12">
        <div className="mx-auto max-w-3xl rounded-3xl border border-[#d7ddd3] bg-white p-8">
          <p className="text-[#5e6d55]">Loading onboarding...</p>
        </div>
      </main>
    );
  }

  return (
    <main className="min-h-screen bg-[#f7f9f7] px-4 py-10">
      <div className="mx-auto max-w-3xl">
        <p className="text-center text-xs font-semibold uppercase tracking-[0.14em] text-[#5e6d55]">
          Onboarding Progress
        </p>
        <h1 className="mt-2 text-center text-4xl font-semibold text-[#1f1f1f]">
          Phase {activeStep + 1} of {STEP_ORDER.length}: {stepTitle}
        </h1>

        <div className="mt-6 grid grid-cols-4 gap-2">
          {STEP_ORDER.map((item, index) => (
            <div
              key={item}
              className={`h-1.5 rounded-full ${
                index <= activeStep ? "bg-[#108a00]" : "bg-[#d8e4d4]"
              }`}
            />
          ))}
        </div>

        <section className="mt-8 rounded-3xl border border-[#d7ddd3] bg-white p-8 shadow-sm">
          <h2 className="text-center text-3xl font-semibold text-[#1f1f1f]">{stepTitle}</h2>
          <p className="mt-2 text-center text-[#5e6d55]">{stepSubtitle}</p>

          {message && <p className="mt-4 text-center text-sm text-[#108a00]">{message}</p>}
          {error && <p className="mt-4 text-center text-sm text-[#b42318]">{error}</p>}

          {stepKey === "profile" && (
            <div className="mt-8 space-y-4">
              <input
                className="input input-bordered w-full border-[#cfd6ca] bg-white"
                placeholder="Display name"
                value={displayName}
                onChange={(event) => setDisplayName(event.target.value)}
              />

              <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
                <input
                  className="input input-bordered w-full border-[#cfd6ca] bg-white"
                  placeholder="Contact email"
                  value={contactEmail}
                  onChange={(event) => setContactEmail(event.target.value)}
                />
                <input
                  className="input input-bordered w-full border-[#cfd6ca] bg-white"
                  placeholder="Contact phone"
                  value={contactPhone}
                  onChange={(event) => setContactPhone(event.target.value)}
                />
              </div>

              <input
                className="input input-bordered w-full border-[#cfd6ca] bg-white"
                placeholder="Location"
                value={location}
                onChange={(event) => setLocation(event.target.value)}
              />

              <textarea
                className="textarea textarea-bordered h-32 w-full border-[#cfd6ca] bg-white"
                placeholder="Tell clients and freelancers about your background."
                maxLength={500}
                value={bio}
                onChange={(event) => setBio(event.target.value)}
              />

              {isClient && (
                <input
                  className="input input-bordered w-full border-[#cfd6ca] bg-white"
                  placeholder="Company name"
                  value={companyName}
                  onChange={(event) => setCompanyName(event.target.value)}
                />
              )}

              {isFreelancer && (
                <div className="space-y-4 rounded-2xl border border-[#dce5d7] bg-[#f8fbf7] p-4">
                  <input
                    className="input input-bordered w-full border-[#cfd6ca] bg-white"
                    placeholder="Headline"
                    value={headline}
                    onChange={(event) => setHeadline(event.target.value)}
                  />
                  <input
                    className="input input-bordered w-full border-[#cfd6ca] bg-white"
                    placeholder="Skills (comma separated)"
                    value={skillsCsv}
                    onChange={(event) => setSkillsCsv(event.target.value)}
                  />
                  <input
                    className="input input-bordered w-full border-[#cfd6ca] bg-white"
                    placeholder="Hourly rate"
                    type="number"
                    min="0"
                    value={hourlyRate}
                    onChange={(event) => setHourlyRate(event.target.value)}
                  />
                  <select
                    className="select select-bordered w-full border-[#cfd6ca] bg-white"
                    value={availability}
                    onChange={(event) => setAvailability(event.target.value)}
                  >
                    <option value="AVAILABILITY_AS_NEEDED">As needed</option>
                    <option value="AVAILABILITY_PART_TIME">Part time</option>
                    <option value="AVAILABILITY_FULL_TIME">Full time</option>
                    <option value="AVAILABILITY_UNAVAILABLE">Unavailable</option>
                  </select>
                </div>
              )}
            </div>
          )}

          {stepKey === "avatar" && (
            <div className="mt-8 space-y-5">
              <div className="mx-auto flex h-36 w-36 items-center justify-center rounded-full border border-[#dce5d7] bg-[#eff7ed] text-[#5e6d55]">
                {avatarExists ? "Avatar ready" : "Add photo"}
              </div>

              <label className="block cursor-pointer rounded-2xl border-2 border-dashed border-[#c7d9bf] bg-[#f8fbf7] p-8 text-center">
                <input
                  type="file"
                  className="hidden"
                  accept="image/jpeg,image/png,image/webp"
                  onChange={onAvatarFileChange}
                />
                <p className="text-base font-semibold text-[#108a00]">
                  Click to select avatar image
                </p>
                <p className="mt-1 text-xs uppercase tracking-wide text-[#6f7d67]">
                  JPG, PNG, WEBP
                </p>
                {avatarFile && (
                  <p className="mt-3 text-sm font-medium text-[#1f1f1f]">{avatarFile.name}</p>
                )}
              </label>

              <div className="grid grid-cols-1 gap-3 md:grid-cols-2">
                <button
                  type="button"
                  className="btn border border-[#ccd6c4] bg-white text-[#1f1f1f] hover:bg-[#f7fbf5]"
                  onClick={reserveAvatarUploadUrl}
                  disabled={saving}
                >
                  Reserve Upload URL
                </button>
                <button
                  type="button"
                  className="btn border-none bg-[#108a00] text-white hover:bg-[#0d7300]"
                  onClick={uploadAvatar}
                  disabled={saving}
                >
                  Upload Avatar
                </button>
              </div>
            </div>
          )}

          {stepKey === "preferences" && (
            <div className="mt-8 space-y-6">
              {isClient && (
                <div className="space-y-4 rounded-2xl border border-[#dce5d7] bg-[#f8fbf7] p-5">
                  <h3 className="text-xl font-semibold text-[#1f1f1f]">Hiring Preferences</h3>
                  <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
                    <input
                      className="input input-bordered w-full border-[#cfd6ca] bg-white"
                      placeholder="Min hourly rate"
                      type="number"
                      min="0"
                      value={minHourlyRate}
                      onChange={(event) => setMinHourlyRate(event.target.value)}
                    />
                    <input
                      className="input input-bordered w-full border-[#cfd6ca] bg-white"
                      placeholder="Max hourly rate"
                      type="number"
                      min="0"
                      value={maxHourlyRate}
                      onChange={(event) => setMaxHourlyRate(event.target.value)}
                    />
                  </div>
                  <input
                    className="input input-bordered w-full border-[#cfd6ca] bg-white"
                    placeholder="Preferred locations (comma separated)"
                    value={preferredLocationsCsv}
                    onChange={(event) => setPreferredLocationsCsv(event.target.value)}
                  />
                </div>
              )}

              {isFreelancer && (
                <>
                  <div className="space-y-4 rounded-2xl border border-[#dce5d7] bg-[#f8fbf7] p-5">
                    <h3 className="text-xl font-semibold text-[#1f1f1f]">Work Preferences</h3>
                    <select
                      className="select select-bordered w-full border-[#cfd6ca] bg-white"
                      value={preferredProjectLength}
                      onChange={(event) => setPreferredProjectLength(event.target.value)}
                    >
                      <option value="">No preference</option>
                      <option value="PROJECT_LENGTH_SHORT_TERM">Short term</option>
                      <option value="PROJECT_LENGTH_MEDIUM_TERM">Medium term</option>
                      <option value="PROJECT_LENGTH_LONG_TERM">Long term</option>
                    </select>

                    <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
                      <input
                        className="input input-bordered w-full border-[#cfd6ca] bg-white"
                        placeholder="Min budget"
                        type="number"
                        min="0"
                        value={minBudget}
                        onChange={(event) => setMinBudget(event.target.value)}
                      />
                      <input
                        className="input input-bordered w-full border-[#cfd6ca] bg-white"
                        placeholder="Max budget"
                        type="number"
                        min="0"
                        value={maxBudget}
                        onChange={(event) => setMaxBudget(event.target.value)}
                      />
                    </div>

                    <input
                      className="input input-bordered w-full border-[#cfd6ca] bg-white"
                      placeholder="Contract types (comma separated)"
                      value={contractTypesCsv}
                      onChange={(event) => setContractTypesCsv(event.target.value)}
                    />

                    <input
                      className="input input-bordered w-full border-[#cfd6ca] bg-white"
                      placeholder="Weekly capacity hours"
                      type="number"
                      min="0"
                      value={weeklyCapacityHours}
                      onChange={(event) => setWeeklyCapacityHours(event.target.value)}
                    />
                  </div>

                  <div className="space-y-4 rounded-2xl border border-[#dce5d7] bg-[#f8fbf7] p-5">
                    <h3 className="text-xl font-semibold text-[#1f1f1f]">CV Upload</h3>
                    <p className="text-sm text-[#5e6d55]">
                      Current CV status: {hasCv ? "Uploaded" : "Missing"}
                    </p>
                    <input
                      type="file"
                      className="file-input file-input-bordered w-full border-[#cfd6ca] bg-white"
                      accept="application/pdf,application/msword,application/vnd.openxmlformats-officedocument.wordprocessingml.document"
                      onChange={onCvFileChange}
                    />
                    <div className="flex flex-wrap gap-3">
                      <button
                        type="button"
                        className="btn border border-[#ccd6c4] bg-white text-[#1f1f1f] hover:bg-[#f7fbf5]"
                        onClick={reserveCvUploadUrl}
                        disabled={saving}
                      >
                        Reserve CV Upload URL
                      </button>
                      <button
                        type="button"
                        className="btn border-none bg-[#108a00] text-white hover:bg-[#0d7300]"
                        onClick={uploadCv}
                        disabled={saving}
                      >
                        Upload CV
                      </button>
                    </div>
                  </div>

                  <div className="space-y-4 rounded-2xl border border-[#dce5d7] bg-[#f8fbf7] p-5">
                    <h3 className="text-xl font-semibold text-[#1f1f1f]">Portfolio</h3>
                    <p className="text-sm text-[#5e6d55]">
                      Portfolio items: {portfolioCount}
                    </p>
                    <input
                      className="input input-bordered w-full border-[#cfd6ca] bg-white"
                      placeholder="Project title"
                      value={portfolioTitle}
                      onChange={(event) => setPortfolioTitle(event.target.value)}
                    />
                    <textarea
                      className="textarea textarea-bordered h-28 w-full border-[#cfd6ca] bg-white"
                      placeholder="Project description"
                      maxLength={200}
                      value={portfolioDescription}
                      onChange={(event) => setPortfolioDescription(event.target.value)}
                    />
                    <input
                      className="input input-bordered w-full border-[#cfd6ca] bg-white"
                      placeholder="Project URL (optional)"
                      value={portfolioProjectUrl}
                      onChange={(event) => setPortfolioProjectUrl(event.target.value)}
                    />
                    <input
                      className="input input-bordered w-full border-[#cfd6ca] bg-white"
                      placeholder="Role in project (optional)"
                      value={portfolioRoleInProject}
                      onChange={(event) => setPortfolioRoleInProject(event.target.value)}
                    />
                    <input
                      className="input input-bordered w-full border-[#cfd6ca] bg-white"
                      placeholder="Tags (comma separated)"
                      value={portfolioTagsCsv}
                      onChange={(event) => setPortfolioTagsCsv(event.target.value)}
                    />

                    <input
                      type="file"
                      className="file-input file-input-bordered w-full border-[#cfd6ca] bg-white"
                      accept="image/jpeg,image/png,image/webp,video/mp4,video/webm,application/pdf"
                      onChange={onPortfolioMediaFileChange}
                    />

                    <div className="flex flex-wrap gap-3">
                      <button
                        type="button"
                        className="btn border border-[#ccd6c4] bg-white text-[#1f1f1f] hover:bg-[#f7fbf5]"
                        onClick={reservePortfolioUploadUrl}
                        disabled={saving || !portfolioMediaFile}
                      >
                        Reserve Media Upload URL
                      </button>
                      <button
                        type="button"
                        className="btn border-none bg-[#108a00] text-white hover:bg-[#0d7300]"
                        onClick={addPortfolioItem}
                        disabled={saving}
                      >
                        Add Portfolio Item
                      </button>
                    </div>
                  </div>
                </>
              )}

              <div className="space-y-4 rounded-2xl border border-[#dce5d7] bg-[#f8fbf7] p-5">
                <h3 className="text-xl font-semibold text-[#1f1f1f]">Notifications</h3>
                <div className="grid grid-cols-1 gap-4 md:grid-cols-3">
                  <input
                    className="input input-bordered border-[#cfd6ca] bg-white"
                    placeholder="Locale (en)"
                    value={uiLocale}
                    onChange={(event) => setUiLocale(event.target.value)}
                  />
                  <label className="label cursor-pointer justify-start gap-2 rounded-lg border border-[#d7ddd3] bg-white px-3">
                    <input
                      type="checkbox"
                      className="checkbox checkbox-sm"
                      checked={emailNotifications}
                      onChange={(event) => setEmailNotifications(event.target.checked)}
                    />
                    <span className="label-text">Email notifications</span>
                  </label>
                  <label className="label cursor-pointer justify-start gap-2 rounded-lg border border-[#d7ddd3] bg-white px-3">
                    <input
                      type="checkbox"
                      className="checkbox checkbox-sm"
                      checked={pushNotifications}
                      onChange={(event) => setPushNotifications(event.target.checked)}
                    />
                    <span className="label-text">Push notifications</span>
                  </label>
                </div>
              </div>
            </div>
          )}

          {stepKey === "kyc" && (
            <div className="mt-8 space-y-5">
              <div className="rounded-2xl border border-[#dce5d7] bg-[#f8fbf7] p-5">
                <h3 className="text-xl font-semibold text-[#1f1f1f]">Tax and Verification</h3>
                <p className="mt-2 text-sm text-[#5e6d55]">
                  Verification status: {verificationStatus || "Not submitted"}
                </p>
                <input
                  className="input input-bordered mt-4 w-full border-[#cfd6ca] bg-white"
                  placeholder="Tax ID"
                  value={taxId}
                  onChange={(event) => setTaxId(event.target.value)}
                />
              </div>

              <div className="rounded-2xl border border-[#dce5d7] bg-[#f8fbf7] p-5">
                <h3 className="text-xl font-semibold text-[#1f1f1f]">Readiness Summary</h3>
                <p className="mt-2 text-sm text-[#5e6d55]">
                  Current readiness: <span className="font-semibold text-[#108a00]">{readinessPercent}%</span>
                </p>
                <p className="mt-2 text-sm text-[#5e6d55]">
                  Missing requirements:{" "}
                  {missingRequired.length > 0 ? missingRequired.join(", ") : "None"}
                </p>
                {recommendations.length > 0 && (
                  <ul className="mt-3 list-disc space-y-1 pl-5 text-sm text-[#5e6d55]">
                    {recommendations.map((item) => (
                      <li key={item}>{item}</li>
                    ))}
                  </ul>
                )}
              </div>
            </div>
          )}
        </section>

        <div className="mt-8 flex items-center justify-between">
          <button
            type="button"
            className="btn rounded-full border border-[#ccd6c4] bg-white px-6 text-[#1f1f1f] hover:bg-[#f7fbf5]"
            onClick={onBack}
            disabled={saving}
          >
            Back
          </button>
          <button
            type="button"
            className="btn rounded-full border-none bg-[#108a00] px-8 text-white hover:bg-[#0d7300]"
            onClick={onContinue}
            disabled={saving}
          >
            {stepKey === "kyc" ? "Finish Onboarding" : "Continue"}
          </button>
        </div>
      </div>
    </main>
  );
}
