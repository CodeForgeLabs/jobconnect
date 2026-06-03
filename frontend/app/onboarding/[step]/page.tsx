"use client";

import { ChangeEvent, FormEvent, KeyboardEvent, useEffect, useMemo, useRef, useState } from "react";
import { useParams, useRouter } from "next/navigation";
import { useSelector } from "react-redux";
import { Trash2, UserCircle2, X } from "lucide-react";
import {
  selectIsAuthenticated,
  selectIsHydrated,
  selectUserRole,
} from "@/features/login/loginSlice";
import PageShell from "@/components/ui/PageShell";
import SectionCard from "@/components/ui/SectionCard";
import { InputField, TextAreaField } from "@/components/ui/FormField";
import UploadDropzone from "@/components/ui/UploadDropzone";
import { PrimaryButton, SecondaryButton } from "@/components/ui/Buttons";
import InlineAlert from "@/components/ui/InlineAlert";
import { userApi } from "@/lib/userApi";
import {
  GetMyVerificationStatusResponsePayload,
  GetOnboardingStatusResponsePayload,
  GetProfileResponsePayload,
  GetSettingsResponsePayload,
  HiringPreferencesResponsePayload,
  PortfolioMediaInputPayload,
  PortfolioItemPayload,
  WorkPreferencesResponsePayload,
} from "@/lib/contracts/user";
import {
  buildHiringPreferencesPayload,
  buildWorkPreferencesPayload,
  normalizeAvailability,
  toOptionalNumber,
} from "@/lib/mappers/userPayloadMappers";
import { getApiErrorMessage } from "@/lib/apiTypes";
import {
  minBudgetMessage,
  minHourlyRateMessage,
  validateOptionalPositiveDecimal,
  validateOptionalPositiveWholeNumber,
  validatePersonalName,
  validatePhoneNumber,
} from "@/lib/fieldValidation";

type RoleType = "client" | "freelancer" | "admin" | "unknown";
type StepKey = "welcome" | "profile" | "avatar" | "preferences" | "kyc" | "cv" | "portfolio" | "review";
type PortfolioMode = "list" | "edit";
type ExistingPortfolioMedia = PortfolioMediaInputPayload & { local_id: string };
type PendingPortfolioMedia = { local_id: string; file: File; preview_url: string };

const CLIENT_STEPS: StepKey[] = ["welcome", "profile", "avatar", "preferences", "kyc", "review"];
const FREELANCER_STEPS: StepKey[] = ["welcome", "profile", "avatar", "preferences", "kyc", "cv", "portfolio", "review"];
const EMPTY_BASELINES: Record<StepKey, string> = {
  welcome: "",
  profile: "",
  avatar: "",
  preferences: "",
  kyc: "",
  cv: "",
  portfolio: "",
  review: "",
};

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

function normalizeVerificationStatus(value: string) {
  return value.trim().toLowerCase().replace(/^verification_status_/, "");
}

export default function OnboardingStepPage() {
  const router = useRouter();
  const params = useParams<{ step: string }>();
  const isHydrated = useSelector(selectIsHydrated);
  const isAuthenticated = useSelector(selectIsAuthenticated);
  const role = useSelector(selectUserRole) as RoleType;

  const [initialLoading, setInitialLoading] = useState(true);
  const [isSavingStep, setIsSavingStep] = useState(false);
  const [isUploadingAvatar, setIsUploadingAvatar] = useState(false);
  const [isNavigatingStep, setIsNavigatingStep] = useState(false);
  const [message, setMessage] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);

  const [readinessPercent, setReadinessPercent] = useState(0);
  const [missingRequired, setMissingRequired] = useState<string[]>([]);
  const [recommendations, setRecommendations] = useState<string[]>([]);

  const [displayName, setDisplayName] = useState("");
  const [contactEmail, setContactEmail] = useState("");
  const [contactPhone, setContactPhone] = useState("");
  const [location, setLocation] = useState("");
  const [bio, setBio] = useState("");

  const [companyName, setCompanyName] = useState("");
  const [headline, setHeadline] = useState("");
  const [skills, setSkills] = useState<string[]>([]);
  const [skillsInput, setSkillsInput] = useState("");
  const [hourlyRate, setHourlyRate] = useState("");
  const [availability, setAvailability] = useState("AVAILABILITY_AS_NEEDED");

  const [taxId, setTaxId] = useState("");
  const [initialTaxId, setInitialTaxId] = useState("");
  const [verificationStatus, setVerificationStatus] = useState("");
  const [verificationLegalName, setVerificationLegalName] = useState("");
  const [verificationCountryCode, setVerificationCountryCode] = useState("");
  const [verificationDocumentType, setVerificationDocumentType] = useState("");
  const [verificationDocumentNumberMasked, setVerificationDocumentNumberMasked] = useState("");
  const [verificationSubmissionNote, setVerificationSubmissionNote] = useState("");
  const [verificationEvidenceFile, setVerificationEvidenceFile] = useState<File | null>(null);
  const [verificationEvidenceUrl, setVerificationEvidenceUrl] = useState("");

  const [uiLocale, setUiLocale] = useState("en");
  const [emailNotifications, setEmailNotifications] = useState(true);
  const [pushNotifications, setPushNotifications] = useState(true);

  const [minHourlyRate, setMinHourlyRate] = useState("");
  const [maxHourlyRate, setMaxHourlyRate] = useState("");
  const [preferredLocations, setPreferredLocations] = useState<string[]>([]);
  const [preferredLocationInput, setPreferredLocationInput] = useState("");

  const [preferredProjectLength, setPreferredProjectLength] = useState("");
  const [minBudget, setMinBudget] = useState("");
  const [maxBudget, setMaxBudget] = useState("");
  const [contractTypes, setContractTypes] = useState<string[]>([]);
  const [weeklyCapacityHours, setWeeklyCapacityHours] = useState("");

  const [avatarFile, setAvatarFile] = useState<File | null>(null);
  const [avatarPreviewUrl, setAvatarPreviewUrl] = useState<string | null>(null);
  const [avatarExists, setAvatarExists] = useState(false);
  const [avatarUrl, setAvatarUrl] = useState("");

  const [cvFile, setCvFile] = useState<File | null>(null);
  const [cvPreviewUrl, setCvPreviewUrl] = useState<string | null>(null);
  const [cvExists, setCvExists] = useState(false);
  const [cvDownloadUrl, setCvDownloadUrl] = useState("");
  const [cvContentType, setCvContentType] = useState("");
  const [isCvPreviewOpen, setIsCvPreviewOpen] = useState(false);

  const [portfolioItems, setPortfolioItems] = useState<PortfolioItemPayload[]>([]);
  const [portfolioMode, setPortfolioMode] = useState<PortfolioMode>("list");
  const [selectedPortfolioId, setSelectedPortfolioId] = useState<string>("");
  const [portfolioTitle, setPortfolioTitle] = useState("");
  const [portfolioDescription, setPortfolioDescription] = useState("");
  const [portfolioProjectUrl, setPortfolioProjectUrl] = useState("");
  const [portfolioRoleInProject, setPortfolioRoleInProject] = useState("");
  const [portfolioTags, setPortfolioTags] = useState<string[]>([]);
  const [portfolioTagInput, setPortfolioTagInput] = useState("");
  const [existingPortfolioMedia, setExistingPortfolioMedia] = useState<ExistingPortfolioMedia[]>([]);
  const [pendingPortfolioMedia, setPendingPortfolioMedia] = useState<PendingPortfolioMedia[]>([]);
  const pendingPortfolioMediaRef = useRef<PendingPortfolioMedia[]>([]);
  const [stepBaselines, setStepBaselines] = useState<Record<StepKey, string>>(EMPTY_BASELINES);
  const isBusy = isSavingStep || isUploadingAvatar || isNavigatingStep;

  const isClient = role === "client";
  const isFreelancer = role === "freelancer";
  const steps = useMemo(() => (isFreelancer ? FREELANCER_STEPS : CLIENT_STEPS), [isFreelancer]);

  const currentStep = (params.step ?? "welcome") as StepKey;
  const currentIndex = steps.indexOf(currentStep);
  const safeIndex = currentIndex >= 0 ? currentIndex : 0;
  const safeStep = steps[safeIndex];
  const normalizedVerificationStatus = normalizeVerificationStatus(verificationStatus);
  const profileFieldErrors = {
    displayName: validatePersonalName(displayName, "Display name"),
    contactPhone: validatePhoneNumber(contactPhone, "Contact phone"),
    hourlyRate: isFreelancer
      ? validateOptionalPositiveDecimal(hourlyRate, "Hourly rate")
      : null,
  };
  const preferenceFieldErrors: Record<string, string | null> = {
    minHourlyRate: isClient
      ? validateOptionalPositiveDecimal(minHourlyRate, "Min hourly rate")
      : null,
    maxHourlyRate: isClient
      ? validateOptionalPositiveDecimal(maxHourlyRate, "Max hourly rate")
      : null,
    minBudget: isFreelancer
      ? validateOptionalPositiveDecimal(minBudget, "Min budget")
      : null,
    maxBudget: isFreelancer
      ? validateOptionalPositiveDecimal(maxBudget, "Max budget")
      : null,
    weeklyCapacityHours: isFreelancer
      ? validateOptionalPositiveWholeNumber(
          weeklyCapacityHours,
          "Weekly capacity hours",
        )
      : null,
  };
  const kycFieldErrors = {
    legalName: validatePersonalName(verificationLegalName, "Legal name"),
  };

  if (
    isClient &&
    minHourlyRate.trim() &&
    maxHourlyRate.trim() &&
    !preferenceFieldErrors.minHourlyRate &&
    !preferenceFieldErrors.maxHourlyRate &&
    Number(minHourlyRate) > Number(maxHourlyRate)
  ) {
    preferenceFieldErrors.minHourlyRate = minHourlyRateMessage;
  }

  if (
    isFreelancer &&
    minBudget.trim() &&
    maxBudget.trim() &&
    !preferenceFieldErrors.minBudget &&
    !preferenceFieldErrors.maxBudget &&
    Number(minBudget) > Number(maxBudget)
  ) {
    preferenceFieldErrors.minBudget = minBudgetMessage;
  }

  function createLocalId(prefix: string) {
    return `${prefix}-${Date.now()}-${Math.random().toString(36).slice(2, 8)}`;
  }

  function isImageContentType(contentType?: string) {
    return Boolean(contentType?.startsWith("image/"));
  }

  function isImagePortfolioMedia(media?: PortfolioMediaInputPayload) {
    if (!media) return false;
    if (media.media_type === "PORTFOLIO_MEDIA_TYPE_IMAGE") return true;
    return isImageContentType(media.content_type);
  }

  function normalizeExistingPortfolioMedia(items: PortfolioItemPayload["media"] = []): ExistingPortfolioMedia[] {
    return (items ?? [])
      .filter((item) => isImagePortfolioMedia(item))
      .map((item) => ({
        local_id: createLocalId("existing-media"),
        media_type: "PORTFOLIO_MEDIA_TYPE_IMAGE",
        storage_key: item.storage_key,
        external_url: item.external_url,
        file_name: item.file_name,
        content_type: item.content_type,
        size_bytes: item.size_bytes,
        width: item.width,
        height: item.height,
      }));
  }

  function buildPortfolioSignature(input: {
    mode: PortfolioMode;
    selectedId: string;
    title: string;
    description: string;
    projectUrl: string;
    roleInProject: string;
    tags: string[];
    existingMedia: ExistingPortfolioMedia[];
    pendingMedia: PendingPortfolioMedia[];
  }) {
    const safeTags = Array.isArray(input.tags) ? input.tags : [];
    const safeExistingMedia = Array.isArray(input.existingMedia) ? input.existingMedia : [];
    const safePendingMedia = Array.isArray(input.pendingMedia) ? input.pendingMedia : [];
    return JSON.stringify({
      mode: input.mode,
      selectedPortfolioId: input.mode === "edit" ? input.selectedId : "",
      title: input.title.trim(),
      description: input.description.trim(),
      projectUrl: input.projectUrl.trim(),
      roleInProject: input.roleInProject.trim(),
      tags: safeTags.map((item) => item.trim().toLowerCase()).filter(Boolean),
      existingMedia: safeExistingMedia.map(
        (item) => `${item.storage_key ?? ""}|${item.external_url ?? ""}|${item.file_name ?? ""}`
      ),
      pendingMedia: safePendingMedia.map(
        (item) => `${item.file.name}:${item.file.size}:${item.file.type}`
      ),
    });
  }

  const isStepDirty = useMemo(() => {
    if (safeStep === "welcome" || safeStep === "review") return false;
    let current = "";
    switch (safeStep) {
      case "profile":
        current = JSON.stringify({
          displayName: displayName.trim(),
          contactEmail: contactEmail.trim(),
          contactPhone: contactPhone.trim(),
          location: location.trim(),
          bio: bio.trim(),
          companyName: isClient ? companyName.trim() : "",
          headline: isFreelancer ? headline.trim() : "",
          skills: isFreelancer ? [...skills].map((item) => item.trim().toLowerCase()).sort() : [],
          hourlyRate: isFreelancer ? hourlyRate.trim() : "",
          availability: isFreelancer ? availability : "",
        });
        break;
      case "avatar":
        current = avatarFile ? `selected:${avatarFile.name}:${avatarFile.size}:${avatarFile.type}` : "";
        break;
      case "preferences":
        current = JSON.stringify(
          isClient
            ? {
                minHourlyRate: minHourlyRate.trim(),
                maxHourlyRate: maxHourlyRate.trim(),
                preferredLocations: preferredLocations.map((item) => item.trim().toLowerCase()).filter(Boolean),
                uiLocale: uiLocale.trim(),
                emailNotifications,
                pushNotifications,
              }
            : {
                preferredProjectLength: preferredProjectLength.trim(),
                minBudget: minBudget.trim(),
                maxBudget: maxBudget.trim(),
                contractTypes: [...contractTypes].map((item) => item.trim().toLowerCase()).sort(),
                weeklyCapacityHours: weeklyCapacityHours.trim(),
                uiLocale: uiLocale.trim(),
                emailNotifications,
                pushNotifications,
              }
        );
        break;
      case "kyc":
        current = JSON.stringify({
          taxId: taxId.trim(),
          legalName: verificationLegalName.trim(),
          countryCode: verificationCountryCode.trim().toUpperCase(),
          documentType: verificationDocumentType.trim(),
          documentNumberMasked: verificationDocumentNumberMasked.trim(),
          submissionNote: verificationSubmissionNote.trim(),
          evidence: verificationEvidenceFile
            ? `${verificationEvidenceFile.name}:${verificationEvidenceFile.size}:${verificationEvidenceFile.type}`
            : "",
        });
        break;
      case "cv":
        current = cvFile ? `selected:${cvFile.name}:${cvFile.size}:${cvFile.type}` : "";
        break;
      case "portfolio":
        current = buildPortfolioSignature({
          mode: portfolioMode,
          selectedId: selectedPortfolioId || "",
          title: portfolioTitle,
          description: portfolioDescription,
          projectUrl: portfolioProjectUrl,
          roleInProject: portfolioRoleInProject,
          tags: portfolioTags,
          existingMedia: existingPortfolioMedia,
          pendingMedia: pendingPortfolioMedia,
        });
        break;
      default:
        current = "";
    }
    return current !== stepBaselines[safeStep];
  }, [
    safeStep,
    stepBaselines,
    displayName,
    contactEmail,
    contactPhone,
    location,
    bio,
    companyName,
    headline,
    skills,
    hourlyRate,
    availability,
    minHourlyRate,
    maxHourlyRate,
    preferredLocations,
    preferredProjectLength,
    minBudget,
    maxBudget,
    contractTypes,
    weeklyCapacityHours,
    isClient,
    isFreelancer,
    taxId,
    verificationLegalName,
    verificationCountryCode,
    verificationDocumentType,
    verificationDocumentNumberMasked,
    verificationSubmissionNote,
    verificationEvidenceFile,
    uiLocale,
    emailNotifications,
    pushNotifications,
    avatarFile,
    cvFile,
    portfolioTitle,
    portfolioDescription,
    portfolioProjectUrl,
    portfolioRoleInProject,
    portfolioTags,
    portfolioMode,
    selectedPortfolioId,
    existingPortfolioMedia,
    pendingPortfolioMedia,
  ]);

  async function loadData(isInitial = false) {
    if (isInitial) {
      setInitialLoading(true);
    }
    setError(null);

    try {
      const [profileRes, onboardingRes, settingsRes] = await Promise.all([
        userApi.getProfile() as Promise<GetProfileResponsePayload>,
        userApi.getOnboardingStatus() as Promise<GetOnboardingStatusResponsePayload>,
        userApi.getSettings() as Promise<GetSettingsResponsePayload>,
      ]);

      const profile = profileRes.profile;
      const settings = settingsRes.settings;

      setDisplayName(profile?.core?.display_name ?? "");
      setContactEmail(profile?.core?.contact_email ?? "");
      setContactPhone(profile?.core?.contact_phone ?? "");
      setLocation(profile?.core?.location ?? "");
      setBio(profile?.core?.bio ?? "");
      setTaxId(profile?.core?.tax_id ?? "");
      setInitialTaxId(profile?.core?.tax_id ?? "");
      setVerificationStatus(profile?.core?.verification_status ?? "");
      let nextVerificationLegalName = "";
      let nextVerificationCountryCode = "";
      let nextVerificationDocumentType = "";
      let nextVerificationDocumentNumberMasked = "";
      let nextVerificationSubmissionNote = "";
      let nextVerificationEvidenceUrl = "";
      try {
        const verificationRes = (await userApi.getMyVerificationStatus()) as GetMyVerificationStatusResponsePayload;
        const request = verificationRes.request;
        nextVerificationLegalName = request?.legal_name ?? "";
        nextVerificationCountryCode = (request?.country_code ?? "").toUpperCase();
        nextVerificationDocumentType = request?.document_type ?? "";
        nextVerificationDocumentNumberMasked = request?.document_number_masked ?? "";
        nextVerificationSubmissionNote = request?.submission_note ?? "";
        nextVerificationEvidenceUrl = request?.evidence_url ?? "";
        setVerificationStatus(request?.status ?? profile?.core?.verification_status ?? "");
      } catch {
        // Keep profile-based fallback when no verification request exists.
        setVerificationStatus(profile?.core?.verification_status ?? "");
      }
      setVerificationLegalName(nextVerificationLegalName);
      setVerificationCountryCode(nextVerificationCountryCode);
      setVerificationDocumentType(nextVerificationDocumentType);
      setVerificationDocumentNumberMasked(nextVerificationDocumentNumberMasked);
      setVerificationSubmissionNote(nextVerificationSubmissionNote);
      setVerificationEvidenceUrl(nextVerificationEvidenceUrl);
      setVerificationEvidenceFile(null);

      let persistedAvatarUrl = profile?.core?.avatar_url ?? "";
      let persistedAvatarExists = Boolean(persistedAvatarUrl);
      try {
        const avatarRes = await userApi.getAvatar();
        if (avatarRes.avatar?.download_url) {
          persistedAvatarUrl = avatarRes.avatar.download_url;
          persistedAvatarExists = true;
        } else if (avatarRes.avatar?.user_id) {
          persistedAvatarExists = true;
        }
      } catch {
        // Ignore avatar fetch errors and fall back to profile core.
      }
      setAvatarExists(persistedAvatarExists);
      setAvatarUrl(persistedAvatarUrl);

      setCompanyName(profile?.client?.company_name ?? "");
      setHeadline(profile?.freelancer?.headline ?? "");
      setSkills((profile?.freelancer?.skills ?? []).map((item) => item.trim()).filter(Boolean));
      setSkillsInput("");
      setHourlyRate(
        typeof profile?.freelancer?.hourly_rate === "number"
          ? String(profile.freelancer.hourly_rate)
          : ""
      );
      setAvailability(profile?.freelancer?.availability ?? "AVAILABILITY_AS_NEEDED");

      setUiLocale(settings?.ui_locale ?? "en");
      setEmailNotifications(settings?.email_notifications_enabled ?? true);
      setPushNotifications(settings?.push_notifications_enabled ?? true);

      setReadinessPercent(onboardingRes.readiness?.percent ?? 0);
      const nextMissingRequired = onboardingRes.readiness?.missing_required_fields ?? [];
      setMissingRequired(nextMissingRequired);
      setRecommendations(onboardingRes.readiness?.recommendations ?? []);
      if (nextMissingRequired.length === 0) {
        router.replace("/account");
        return;
      }

      let nextMinHourlyRate = "";
      let nextMaxHourlyRate = "";
      let nextPreferredLocations: string[] = [];
      let nextPreferredProjectLength = "";
      let nextMinBudget = "";
      let nextMaxBudget = "";
      let nextContractTypes: string[] = [];
      let nextWeeklyCapacityHours = "";
      let nextPortfolioItems: PortfolioItemPayload[] = [];
      const portfolioListBaseline = buildPortfolioSignature({
        mode: "list",
        selectedId: "",
        title: "",
        description: "",
        projectUrl: "",
        roleInProject: "",
        tags: [],
        existingMedia: [],
        pendingMedia: [],
      });

      if (isClient) {
        try {
          const hiringRes = (await userApi.getHiringPreferences()) as HiringPreferencesResponsePayload;
          nextMinHourlyRate =
            typeof hiringRes.preferences?.min_hourly_rate === "number"
              ? String(hiringRes.preferences.min_hourly_rate)
              : "";
          nextMaxHourlyRate =
            typeof hiringRes.preferences?.max_hourly_rate === "number"
              ? String(hiringRes.preferences.max_hourly_rate)
              : "";
          nextPreferredLocations = hiringRes.preferences?.preferred_locations ?? [];
          setMinHourlyRate(nextMinHourlyRate);
          setMaxHourlyRate(nextMaxHourlyRate);
          setPreferredLocations(nextPreferredLocations);
        } catch {
          // Ignore empty state.
        }
      } else {
        setPreferredLocations([]);
      }

      if (isFreelancer) {
        try {
          const workRes = (await userApi.getWorkPreferences()) as WorkPreferencesResponsePayload;
          nextPreferredProjectLength = workRes.settings?.preferred_project_length ?? "";
          nextMinBudget =
            typeof workRes.settings?.min_budget === "number" ? String(workRes.settings.min_budget) : "";
          nextMaxBudget =
            typeof workRes.settings?.max_budget === "number" ? String(workRes.settings.max_budget) : "";
          nextContractTypes = (workRes.settings?.contract_types ?? [])
            .map((item) => item.trim().toLowerCase())
            .filter(Boolean);
          nextWeeklyCapacityHours =
            typeof workRes.settings?.weekly_capacity_hours === "number"
              ? String(workRes.settings.weekly_capacity_hours)
              : "";
          setPreferredProjectLength(nextPreferredProjectLength);
          setMinBudget(nextMinBudget);
          setMaxBudget(nextMaxBudget);
          setContractTypes(nextContractTypes);
          setWeeklyCapacityHours(nextWeeklyCapacityHours);
        } catch {
          // Ignore empty state.
        }

        try {
          const cvRes = await userApi.getCV();
          setCvExists(Boolean(cvRes.cv?.user_id));
          setCvDownloadUrl(cvRes.cv?.download_url ?? "");
          setCvContentType(cvRes.cv?.content_type ?? "");
        } catch {
          setCvExists(false);
          setCvDownloadUrl("");
          setCvContentType("");
        }

        try {
          const portfolioRes = await userApi.listPortfolio(20, "");
          nextPortfolioItems = portfolioRes.items ?? [];
          setPortfolioItems(nextPortfolioItems);
        } catch {
          setPortfolioItems([]);
          nextPortfolioItems = [];
        }

        setPortfolioMode("list");
        setSelectedPortfolioId("");
        setPortfolioTitle("");
        setPortfolioDescription("");
        setPortfolioProjectUrl("");
        setPortfolioRoleInProject("");
        setPortfolioTags([]);
        setPortfolioTagInput("");
        setExistingPortfolioMedia([]);
        clearPendingPortfolioMedia();
      }

      setStepBaselines({
        ...EMPTY_BASELINES,
        profile: JSON.stringify({
          displayName: (profile?.core?.display_name ?? "").trim(),
          contactEmail: (profile?.core?.contact_email ?? "").trim(),
          contactPhone: (profile?.core?.contact_phone ?? "").trim(),
          location: (profile?.core?.location ?? "").trim(),
          bio: (profile?.core?.bio ?? "").trim(),
          companyName: isClient ? (profile?.client?.company_name ?? "").trim() : "",
          headline: isFreelancer ? (profile?.freelancer?.headline ?? "").trim() : "",
          skills: isFreelancer ? [...(profile?.freelancer?.skills ?? [])].map((item) => item.trim().toLowerCase()).filter(Boolean).sort() : [],
          hourlyRate:
            isFreelancer && typeof profile?.freelancer?.hourly_rate === "number"
              ? String(profile.freelancer.hourly_rate).trim()
              : "",
          availability: isFreelancer ? (profile?.freelancer?.availability ?? "AVAILABILITY_AS_NEEDED") : "",
        }),
        avatar: "",
        preferences: JSON.stringify(
          isClient
            ? {
                minHourlyRate: nextMinHourlyRate.trim(),
                maxHourlyRate: nextMaxHourlyRate.trim(),
                preferredLocations: nextPreferredLocations.map((item) => item.trim().toLowerCase()).filter(Boolean),
                uiLocale: (settings?.ui_locale ?? "en").trim(),
                emailNotifications: settings?.email_notifications_enabled ?? true,
                pushNotifications: settings?.push_notifications_enabled ?? true,
              }
            : {
                preferredProjectLength: nextPreferredProjectLength.trim(),
                minBudget: nextMinBudget.trim(),
                maxBudget: nextMaxBudget.trim(),
                contractTypes: [...nextContractTypes].sort(),
                weeklyCapacityHours: nextWeeklyCapacityHours.trim(),
                uiLocale: (settings?.ui_locale ?? "en").trim(),
                emailNotifications: settings?.email_notifications_enabled ?? true,
                pushNotifications: settings?.push_notifications_enabled ?? true,
              }
        ),
        kyc: JSON.stringify({
          taxId: (profile?.core?.tax_id ?? "").trim(),
          legalName: nextVerificationLegalName.trim(),
          countryCode: nextVerificationCountryCode.trim().toUpperCase(),
          documentType: nextVerificationDocumentType.trim(),
          documentNumberMasked: nextVerificationDocumentNumberMasked.trim(),
          submissionNote: nextVerificationSubmissionNote.trim(),
          evidence: "",
        }),
        cv: "",
        portfolio: isFreelancer ? portfolioListBaseline : "",
      });
    } catch (err) {
      setError(getApiErrorMessage(err, "Failed to load onboarding."));
    } finally {
      if (isInitial) {
        setInitialLoading(false);
      }
    }
  }

  useEffect(() => {
    if (!isHydrated) return;
    if (!isAuthenticated) {
      router.replace("/login");
      return;
    }
    loadData(true);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [isHydrated, isAuthenticated, role]);

  useEffect(() => {
    if (steps.indexOf(currentStep) < 0) {
      router.replace(`/onboarding/${steps[0]}`);
    }
  }, [currentStep, router, steps]);

  useEffect(() => {
    setIsNavigatingStep(false);
  }, [safeStep]);

  useEffect(() => {
    return () => {
      if (avatarPreviewUrl) {
        URL.revokeObjectURL(avatarPreviewUrl);
      }
      if (cvPreviewUrl) {
        URL.revokeObjectURL(cvPreviewUrl);
      }
      pendingPortfolioMediaRef.current.forEach((item) => {
        URL.revokeObjectURL(item.preview_url);
      });
    };
  }, [avatarPreviewUrl, cvPreviewUrl]);

  useEffect(() => {
    pendingPortfolioMediaRef.current = pendingPortfolioMedia;
  }, [pendingPortfolioMedia]);

  function clearPendingPortfolioMedia() {
    setPendingPortfolioMedia((previous) => {
      previous.forEach((item) => URL.revokeObjectURL(item.preview_url));
      return [];
    });
  }

  function setPortfolioStepBaseline(signature: string) {
    setStepBaselines((previous) => ({ ...previous, portfolio: signature }));
  }

  function openPortfolioListMode() {
    clearPendingPortfolioMedia();
    setPortfolioMode("list");
    setSelectedPortfolioId("");
    setPortfolioTitle("");
    setPortfolioDescription("");
    setPortfolioProjectUrl("");
    setPortfolioRoleInProject("");
    setPortfolioTags([]);
    setPortfolioTagInput("");
    setExistingPortfolioMedia([]);
    const baseline = buildPortfolioSignature({
      mode: "list",
      selectedId: "",
      title: "",
      description: "",
      projectUrl: "",
      roleInProject: "",
      tags: [],
      existingMedia: [],
      pendingMedia: [],
    });
    setPortfolioStepBaseline(baseline);
  }

  function openPortfolioEditMode(item?: PortfolioItemPayload) {
    clearPendingPortfolioMedia();
    const normalizedMedia = normalizeExistingPortfolioMedia(item?.media);
    const selectedId = item?.id ? String(item.id) : "";
    const title = item?.title ?? "";
    const description = item?.description ?? "";
    const projectUrl = item?.project_url ?? "";
    const roleInProject = item?.role_in_project ?? "";
    const tags = (Array.isArray(item?.tags) ? item?.tags : []).filter((tag) => tag.trim().length > 0);

    setPortfolioMode("edit");
    setSelectedPortfolioId(selectedId);
    setPortfolioTitle(title);
    setPortfolioDescription(description);
    setPortfolioProjectUrl(projectUrl);
    setPortfolioRoleInProject(roleInProject);
    setPortfolioTags(tags);
    setPortfolioTagInput("");
    setExistingPortfolioMedia(normalizedMedia);
    const baseline = buildPortfolioSignature({
      mode: "edit",
      selectedId,
      title,
      description,
      projectUrl,
      roleInProject,
      tags,
      existingMedia: normalizedMedia,
      pendingMedia: [],
    });
    setPortfolioStepBaseline(baseline);
  }

  async function runAction(name: string, action: () => Promise<void>, successMessage?: string) {
    setIsSavingStep(true);
    setError(null);
    setMessage(null);

    try {
      await action();
      setMessage(successMessage ?? `${name} saved.`);
      await loadData(false);
      return true;
    } catch (err) {
      setError(getApiErrorMessage(err, `${name} failed.`));
      return false;
    } finally {
      setIsSavingStep(false);
    }
  }

  async function saveProfile() {
    if (Object.values(profileFieldErrors).some(Boolean)) {
      setError("Fix the highlighted fields before saving.");
      return false;
    }

    return runAction("Profile", async () => {
      await userApi.patchProfile({
        display_name: displayName || undefined,
        contact_email: contactEmail || undefined,
        contact_phone: contactPhone || undefined,
        location: location || undefined,
        bio: bio || undefined,
        company_name: isClient ? companyName || undefined : undefined,
        headline: isFreelancer ? headline || undefined : undefined,
        skills: isFreelancer ? skills.map((item) => item.trim()).filter(Boolean) : undefined,
        hourly_rate: isFreelancer ? toOptionalNumber(hourlyRate) : undefined,
        availability: isFreelancer ? normalizeAvailability(availability) : undefined,
      });
    });
  }

  async function saveAvatar() {
    if (!avatarFile) {
      setError("Select an avatar image before saving.");
      return false;
    }

    setIsUploadingAvatar(true);
    try {
      return await runAction("Avatar", async () => {
        const reserved = await userApi.requestAvatarUploadUrl(avatarFile.name, avatarFile.type);
        await uploadToPresignedUrl(reserved.upload_url, avatarFile);
        const upserted = await userApi.upsertAvatar({
          storage_key: reserved.storage_key,
          file_name: avatarFile.name,
          content_type: avatarFile.type,
        });
        if (upserted.avatar_url) {
          setAvatarUrl(upserted.avatar_url);
          setAvatarExists(true);
        }
        setAvatarFile(null);
        if (avatarPreviewUrl) {
          URL.revokeObjectURL(avatarPreviewUrl);
        }
        setAvatarPreviewUrl(null);
      });
    } finally {
      setIsUploadingAvatar(false);
    }
  }

  async function savePreferences() {
    if (Object.values(preferenceFieldErrors).some(Boolean)) {
      setError("Fix the highlighted fields before saving.");
      return false;
    }

    return runAction("Preferences", async () => {
      if (isClient) {
        await userApi.patchHiringPreferences(
          buildHiringPreferencesPayload({
            minHourlyRate,
            maxHourlyRate,
            preferredLocations,
          })
        );
      }
      await userApi.patchSettings({
        ui_locale: uiLocale || "en",
        email_notifications_enabled: emailNotifications,
        push_notifications_enabled: pushNotifications,
      });
      if (isFreelancer) {
        await userApi.patchWorkPreferences(
          buildWorkPreferencesPayload({
            preferredProjectLength,
            minBudget,
            maxBudget,
            contractTypes,
            weeklyCapacityHours,
          })
        );
      }
    });
  }

  async function saveKYC() {
    if (Object.values(kycFieldErrors).some(Boolean)) {
      setError("Fix the highlighted fields before saving.");
      return false;
    }

    return runAction("KYC", async () => {
      const trimmedTaxId = taxId.trim();
      const taxIdChanged = trimmedTaxId !== initialTaxId.trim();
      const taxIdLockedByBackend =
        normalizedVerificationStatus === "submitted" ||
        normalizedVerificationStatus === "pending_review" ||
        normalizedVerificationStatus === "verified";
      if (trimmedTaxId && taxIdChanged && !taxIdLockedByBackend) {
        await userApi.patchProfile({ tax_id: trimmedTaxId });
      }

      const hasVerificationFormInput =
        verificationLegalName.trim().length > 0 ||
        verificationCountryCode.trim().length > 0 ||
        verificationDocumentType.trim().length > 0 ||
        verificationDocumentNumberMasked.trim().length > 0 ||
        verificationSubmissionNote.trim().length > 0;

      const legalName = verificationLegalName.trim();
      const countryCode = verificationCountryCode.trim().toUpperCase();
      const documentType = verificationDocumentType.trim();
      const documentNumberMasked = verificationDocumentNumberMasked.trim();
      const submissionNote = verificationSubmissionNote.trim();

      if (!legalName || !countryCode || !documentType || !documentNumberMasked) {
        throw new Error("Complete legal name, country code, document type, and document number.");
      }

      if (!hasVerificationFormInput && !verificationEvidenceFile) {
        return;
      }

      let evidenceKey = verificationEvidenceUrl.trim();
      if (verificationEvidenceFile) {
        const reservation = await userApi.requestVerificationEvidenceUploadUrl(
          verificationEvidenceFile.name,
          verificationEvidenceFile.type
        );
        await uploadToPresignedUrl(reservation.upload_url, verificationEvidenceFile);
        evidenceKey = reservation.storage_key;
        setVerificationEvidenceUrl(reservation.storage_key);
      }
      if (!evidenceKey) {
        throw new Error("Upload a verification document before submitting verification.");
      }

      await userApi.submitVerification({
        legal_name: legalName,
        country_code: countryCode,
        document_type: documentType,
        document_number_masked: documentNumberMasked,
        evidence_url: evidenceKey,
        submission_note: submissionNote || undefined,
      });

      setVerificationEvidenceFile(null);
    });
  }

  async function saveCV() {
    if (!cvFile) {
      setError("Select a CV file before saving.");
      return false;
    }

    return runAction("CV", async () => {
      const reserved = await userApi.requestCVUploadUrl(cvFile.name, cvFile.type);
      await uploadToPresignedUrl(reserved.upload_url, cvFile);
      const upserted = await userApi.upsertCV({
        storage_key: reserved.storage_key,
        file_name: cvFile.name,
        content_type: cvFile.type,
      });
      setCvExists(true);
      setCvDownloadUrl(upserted.cv?.download_url ?? "");
      setCvContentType(upserted.cv?.content_type ?? cvFile.type);
      setCvFile(null);
      if (cvPreviewUrl) {
        URL.revokeObjectURL(cvPreviewUrl);
      }
      setCvPreviewUrl(null);
    });
  }

  async function savePortfolio() {
    if (portfolioMode !== "edit") {
      return true;
    }

    if (!portfolioTitle.trim() || !portfolioDescription.trim()) {
      setError("Portfolio title and description are required.");
      return false;
    }

    return runAction("Portfolio", async () => {
      const uploadedMedia: PortfolioMediaInputPayload[] = [];
      for (const pending of pendingPortfolioMedia) {
        const reserved = await userApi.requestPortfolioUploadUrl(pending.file.name, pending.file.type);
        await uploadToPresignedUrl(reserved.upload_url, pending.file);
        uploadedMedia.push({
          media_type: "PORTFOLIO_MEDIA_TYPE_IMAGE",
          storage_key: reserved.storage_key,
          content_type: pending.file.type,
          file_name: pending.file.name,
        });
      }

      const existingMediaPayload: PortfolioMediaInputPayload[] = existingPortfolioMedia.map((item) => ({
        media_type: "PORTFOLIO_MEDIA_TYPE_IMAGE",
        storage_key: item.storage_key,
        external_url: item.external_url,
        file_name: item.file_name,
        content_type: item.content_type,
        size_bytes: item.size_bytes,
        width: item.width,
        height: item.height,
      }));

      const payload = {
        title: portfolioTitle.trim(),
        description: portfolioDescription.trim(),
        project_url: portfolioProjectUrl.trim() || undefined,
        role_in_project: portfolioRoleInProject.trim() || undefined,
        tags: portfolioTags.map((item) => item.trim()).filter(Boolean),
        media: [...existingMediaPayload, ...uploadedMedia],
      };

      if (selectedPortfolioId) {
        await userApi.updatePortfolioItem(selectedPortfolioId, payload);
      } else {
        await userApi.createPortfolioItem(payload);
      }

      clearPendingPortfolioMedia();
      setPortfolioMode("list");
      setSelectedPortfolioId("");
      setPortfolioTitle("");
      setPortfolioDescription("");
      setPortfolioProjectUrl("");
      setPortfolioRoleInProject("");
      setPortfolioTags([]);
      setPortfolioTagInput("");
      setExistingPortfolioMedia([]);
    });
  }

  async function saveCurrentStep() {
    switch (safeStep) {
      case "profile":
        return saveProfile();
      case "avatar":
        return saveAvatar();
      case "preferences":
        return savePreferences();
      case "kyc":
        return saveKYC();
      case "cv":
        return saveCV();
      case "portfolio":
        return savePortfolio();
      case "welcome":
      case "review":
      default:
        setMessage("Nothing to save on this phase.");
        return true;
    }
  }

  function goToIndex(index: number) {
    const clamped = Math.min(Math.max(index, 0), steps.length - 1);
    setIsNavigatingStep(true);
    router.push(`/onboarding/${steps[clamped]}`);
  }

  async function onContinue() {
    if (safeStep === "review") {
      if (missingRequired.length > 0) {
        setError(`Complete required fields first: ${missingRequired.join(", ")}.`);
        return;
      }
      setIsNavigatingStep(true);
      router.push(isFreelancer ? "/freelancer/dashboard" : isClient ? "/client/dashboard" : "/account");
      return;
    }

    if (safeStep !== "welcome" && isStepDirty) {
      const ok = await saveCurrentStep();
      if (!ok) return;
    }

    goToIndex(safeIndex + 1);
  }

  function onSkip() {
    if (safeStep === "review") return;
    goToIndex(safeIndex + 1);
  }

  function onBack() {
    if (safeIndex === 0) {
      router.push("/account");
      return;
    }
    goToIndex(safeIndex - 1);
  }

  function onAvatarFileChange(event: ChangeEvent<HTMLInputElement>) {
    const file = event.target.files?.[0] ?? null;
    setAvatarFile(file);
    if (avatarPreviewUrl) {
      URL.revokeObjectURL(avatarPreviewUrl);
    }
    setAvatarPreviewUrl(file ? URL.createObjectURL(file) : null);
  }

  async function onDeleteAvatar() {
    setError(null);
    setMessage(null);

    if (avatarPreviewUrl || avatarFile) {
      if (avatarPreviewUrl) {
        URL.revokeObjectURL(avatarPreviewUrl);
      }
      setAvatarPreviewUrl(null);
      setAvatarFile(null);
      setMessage("Selected avatar removed.");
      return;
    }

    if (!avatarExists) {
      return;
    }

    await runAction("Avatar", async () => {
      await userApi.removeAvatar();
      setAvatarExists(false);
      setAvatarUrl("");
    }, "Avatar deleted.");
  }

  function onCvFileChange(event: ChangeEvent<HTMLInputElement>) {
    const file = event.target.files?.[0] ?? null;
    setCvFile(file);
    if (cvPreviewUrl) {
      URL.revokeObjectURL(cvPreviewUrl);
    }
    setCvPreviewUrl(file ? URL.createObjectURL(file) : null);
    if (file) {
      setCvContentType(file.type);
    }
  }

  async function onDeleteCv() {
    setError(null);
    setMessage(null);

    if (cvFile || cvPreviewUrl) {
      if (cvPreviewUrl) {
        URL.revokeObjectURL(cvPreviewUrl);
      }
      setCvFile(null);
      setCvPreviewUrl(null);
      setMessage("Selected CV removed.");
      return;
    }

    if (!cvExists) {
      return;
    }

    await runAction("CV", async () => {
      await userApi.removeCV();
      setCvExists(false);
      setCvDownloadUrl("");
      setCvContentType("");
      setIsCvPreviewOpen(false);
    }, "CV deleted.");
  }

  function onVerificationEvidenceFileChange(event: ChangeEvent<HTMLInputElement>) {
    setVerificationEvidenceFile(event.target.files?.[0] ?? null);
  }

  function onPortfolioFileChange(event: ChangeEvent<HTMLInputElement>) {
    const files = Array.from(event.target.files ?? []);
    if (files.length === 0) return;

    const invalid = files.find((file) => !isImageContentType(file.type));
    if (invalid) {
      setError("Unsupported media type. Use JPG, PNG, or WEBP images.");
      event.currentTarget.value = "";
      return;
    }

    setError(null);
    const mapped = files.map((file) => ({
      local_id: createLocalId("pending-media"),
      file,
      preview_url: URL.createObjectURL(file),
    }));
    setPendingPortfolioMedia((previous) => [...previous, ...mapped]);
    event.currentTarget.value = "";
  }

  function onRemoveExistingPortfolioMedia(localId: string) {
    setExistingPortfolioMedia((previous) => previous.filter((item) => item.local_id !== localId));
  }

  function onRemovePendingPortfolioMedia(localId: string) {
    setPendingPortfolioMedia((previous) => {
      const target = previous.find((item) => item.local_id === localId);
      if (target) {
        URL.revokeObjectURL(target.preview_url);
      }
      return previous.filter((item) => item.local_id !== localId);
    });
  }

  function onEditPortfolioItem(itemId: string) {
    const item = portfolioItems.find((entry) => String(entry.id ?? "") === itemId);
    if (!item) {
      return;
    }
    setError(null);
    setMessage(null);
    openPortfolioEditMode(item);
  }

  function onAddNewPortfolio() {
    setError(null);
    setMessage(null);
    openPortfolioEditMode(undefined);
  }

  function onDiscardPortfolioEdit() {
    setError(null);
    setMessage(null);
    openPortfolioListMode();
  }

  async function onDeletePortfolioItem() {
    if (!selectedPortfolioId) {
      return;
    }

    await runAction("Portfolio", async () => {
      await userApi.deletePortfolioItem(selectedPortfolioId);
      clearPendingPortfolioMedia();
      setPortfolioMode("list");
      setSelectedPortfolioId("");
      setPortfolioTitle("");
      setPortfolioDescription("");
      setPortfolioProjectUrl("");
      setPortfolioRoleInProject("");
      setPortfolioTags([]);
      setPortfolioTagInput("");
      setExistingPortfolioMedia([]);
    }, "Project deleted.");
  }

  function addSkill(rawValue: string) {
    const normalized = rawValue.trim();
    if (!normalized) return;
    const exists = skills.some((item) => item.trim().toLowerCase() === normalized.toLowerCase());
    if (exists) return;
    setSkills((previous) => [...previous, normalized]);
    setSkillsInput("");
  }

  function onSkillsKeyDown(event: KeyboardEvent<HTMLInputElement>) {
    if (event.key !== "Enter") return;
    event.preventDefault();
    addSkill(skillsInput);
  }

  function removeSkill(index: number) {
    setSkills((previous) => previous.filter((_, itemIndex) => itemIndex !== index));
  }

  function addPreferredLocation(rawValue: string) {
    const normalized = rawValue.trim();
    if (!normalized) return;
    const alreadyExists = preferredLocations.some(
      (item) => item.trim().toLowerCase() === normalized.toLowerCase()
    );
    if (alreadyExists) return;
    setPreferredLocations((previous) => [...previous, normalized]);
    setPreferredLocationInput("");
  }

  function onPreferredLocationKeyDown(event: KeyboardEvent<HTMLInputElement>) {
    if (event.key !== "Enter") return;
    event.preventDefault();
    addPreferredLocation(preferredLocationInput);
  }

  function removePreferredLocation(index: number) {
    setPreferredLocations((previous) => previous.filter((_, itemIndex) => itemIndex !== index));
  }

  function toggleContractType(type: "fixed" | "hourly") {
    setContractTypes((previous) => {
      const has = previous.includes(type);
      if (has) {
        return previous.filter((item) => item !== type);
      }
      return [...previous, type];
    });
  }

  function addPortfolioTag(rawValue: string) {
    const normalized = rawValue.trim();
    if (!normalized) return;
    const exists = portfolioTags.some((item) => item.trim().toLowerCase() === normalized.toLowerCase());
    if (exists) return;
    setPortfolioTags((previous) => [...previous, normalized]);
    setPortfolioTagInput("");
  }

  function onPortfolioTagKeyDown(event: KeyboardEvent<HTMLInputElement>) {
    if (event.key !== "Enter") return;
    event.preventDefault();
    addPortfolioTag(portfolioTagInput);
  }

  function removePortfolioTag(index: number) {
    setPortfolioTags((previous) => previous.filter((_, itemIndex) => itemIndex !== index));
  }

  const phaseTitle = useMemo(() => {
    switch (safeStep) {
      case "welcome":
        return "Welcome";
      case "profile":
        return "Profile";
      case "avatar":
        return "Avatar";
      case "preferences":
        return "Preferences";
      case "kyc":
        return "KYC";
      case "cv":
        return "CV";
      case "portfolio":
        return "Portfolio";
      default:
        return "Review";
    }
  }, [safeStep]);

  const cvViewUrl = cvPreviewUrl || cvDownloadUrl;
  const normalizedCvType = cvContentType.trim().toLowerCase();
  const isCvPdf = normalizedCvType === "application/pdf" || cvViewUrl.toLowerCase().endsWith(".pdf");
  const canShowCvActions = Boolean(cvFile || cvExists);
  const selectedPortfolioItem = portfolioItems.find((item) => String(item.id ?? "") === selectedPortfolioId);
  const portfolioDisplayItems = portfolioItems.map((item) => ({
    id: String(item.id ?? ""),
    title: item.title?.trim() || "Untitled Project",
    role: item.role_in_project?.trim() || "Role not specified",
    tags: Array.isArray(item.tags) ? item.tags : [],
    previewUrl: (item.media ?? []).find((media) => isImagePortfolioMedia(media))?.external_url ?? "",
    isSelected: String(item.id ?? "") === selectedPortfolioId,
  }));

  if (!isHydrated || initialLoading) {
    return (
      <PageShell contentClassName="max-w-4xl">
        <SectionCard title="Loading onboarding">
          <p className="text-sm text-[var(--jc-ink-muted)]">Preparing your onboarding phase...</p>
        </SectionCard>
      </PageShell>
    );
  }

  return (
    <PageShell contentClassName="max-w-5xl">
      <section className="jc-auth-frame mx-auto max-w-4xl p-6 md:p-10">
        <p className="text-center text-xs font-semibold uppercase tracking-[0.18em] text-[var(--jc-ink-muted)]">Onboarding Progress</p>
        <h1 className="mt-3 text-center text-3xl font-semibold text-[var(--jc-ink)] md:text-5xl">
          Phase {safeIndex + 1} of {steps.length}: {phaseTitle}
        </h1>
        <div className="mt-6 h-2 rounded-full bg-[var(--jc-border)]">
          <div className="h-2 rounded-full bg-[var(--jc-accent)]" style={{ width: `${((safeIndex + 1) / steps.length) * 100}%` }} />
        </div>

        <div className="mt-8 rounded-2xl border border-[var(--jc-border)] bg-[var(--jc-surface-alt)] p-5 md:p-6">
          {message ? <InlineAlert tone="success" className="mb-4">{message}</InlineAlert> : null}
          {error ? <InlineAlert tone="error" className="mb-4">{error}</InlineAlert> : null}

          {safeStep === "welcome" ? (
            <div className="space-y-3 text-sm text-[var(--jc-ink-muted)]">
              <p>We will guide you through each required onboarding phase.</p>
              <p>Save &amp; Continue updates changed fields and moves you to the next phase.</p>
            </div>
          ) : null}

          {safeStep === "profile" ? (
            <div className="space-y-5">
              <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
                <InputField label="Display name" value={displayName} onChange={(e) => setDisplayName(e.target.value)} error={profileFieldErrors.displayName} />
                <InputField label="Location" value={location} onChange={(e) => setLocation(e.target.value)} />
              </div>

              <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
                <InputField label="Contact email" type="email" value={contactEmail} onChange={(e) => setContactEmail(e.target.value)} />
                <InputField label="Contact phone" value={contactPhone} onChange={(e) => setContactPhone(e.target.value)} inputMode="tel" error={profileFieldErrors.contactPhone} />
              </div>

              <TextAreaField label="Bio" value={bio} onChange={(e) => setBio(e.target.value)} rows={4} maxLength={500} />

              {isClient ? <InputField label="Company name" value={companyName} onChange={(e) => setCompanyName(e.target.value)} /> : null}
              {isFreelancer ? (
                <div className="space-y-4">
                  <InputField label="Headline" value={headline} onChange={(e) => setHeadline(e.target.value)} />
                  <div className="space-y-2">
                    <label className="block">
                      <span className="mb-1.5 block text-sm font-medium text-[var(--jc-ink)]">Skills</span>
                      <div className="flex flex-wrap items-center gap-2 rounded-xl border border-[var(--jc-border)] bg-[var(--jc-surface)] p-2">
                        {skills.map((skill, index) => (
                          <span
                            key={`${skill}-${index}`}
                            className="inline-flex items-center gap-2 rounded-full bg-[var(--jc-accent-soft)] px-3 py-1 text-sm text-[var(--jc-ink)]"
                          >
                            {skill}
                            <button
                              type="button"
                              aria-label={`Remove ${skill}`}
                              onClick={() => removeSkill(index)}
                              className="text-[var(--jc-ink-muted)] hover:text-[var(--jc-ink)]"
                            >
                              ×
                            </button>
                          </span>
                        ))}
                        <input
                          value={skillsInput}
                          onChange={(event) => setSkillsInput(event.target.value)}
                          onKeyDown={onSkillsKeyDown}
                          placeholder="Add a skill and press Enter"
                          className="h-10 min-w-[220px] flex-1 bg-transparent px-2 text-[var(--jc-ink)] outline-none"
                        />
                      </div>
                    </label>
                    <p className="text-xs text-[var(--jc-ink-muted)]">Add and remove skills individually.</p>
                  </div>
                  <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
                    <InputField label="Hourly rate" type="number" value={hourlyRate} onChange={(e) => setHourlyRate(e.target.value)} error={profileFieldErrors.hourlyRate} />
                    <label className="block">
                      <span className="mb-1.5 block text-sm font-medium text-[var(--jc-ink)]">Availability</span>
                      <select
                        className="h-12 w-full rounded-xl border border-[var(--jc-border)] bg-[var(--jc-surface)] px-4 text-[var(--jc-ink)]"
                        value={availability}
                        onChange={(e) => setAvailability(e.target.value)}
                      >
                        <option value="AVAILABILITY_AS_NEEDED">As needed</option>
                        <option value="AVAILABILITY_PART_TIME">Part time</option>
                        <option value="AVAILABILITY_FULL_TIME">Full time</option>
                        <option value="AVAILABILITY_UNAVAILABLE">Unavailable</option>
                      </select>
                    </label>
                  </div>
                </div>
              ) : null}
            </div>
          ) : null}

          {safeStep === "avatar" ? (
            <div className="space-y-4">
              <div className="mx-auto w-fit">
                <div className="relative">
                  <div className="flex h-40 w-40 items-center justify-center overflow-hidden rounded-full bg-[#dbe7ff]">
                    {avatarPreviewUrl || avatarUrl ? (
                      <img
                        src={avatarPreviewUrl || avatarUrl}
                        alt="Avatar"
                        className="h-full w-full object-cover"
                      />
                    ) : (
                      <div className="flex flex-col items-center gap-2 text-[#475569]">
                        <UserCircle2 className="h-10 w-10" />
                        <span className="text-xs font-semibold tracking-wide">ADD PHOTO</span>
                      </div>
                    )}
                  </div>
                  {(avatarPreviewUrl || avatarExists) ? (
                    <button
                      type="button"
                      aria-label="Delete avatar"
                      onClick={onDeleteAvatar}
                      className="absolute bottom-1 right-1 inline-flex h-9 w-9 items-center justify-center rounded-full bg-white text-[#e45353] shadow-md ring-1 ring-black/5 transition hover:bg-[#fff5f5]"
                    >
                      <Trash2 className="h-4 w-4" />
                    </button>
                  ) : null}
                </div>
              </div>
              <p className="text-sm text-[var(--jc-ink-muted)]">Upload a clear professional avatar.</p>
              <UploadDropzone
                label="Upload profile photo"
                helper="JPG, PNG, WEBP"
                accept="image/jpeg,image/png,image/webp"
                onChange={onAvatarFileChange}
                fileName={avatarFile?.name}
              />
            </div>
          ) : null}

          {safeStep === "preferences" ? (
            <div className="space-y-4">
              {isClient ? (
                <>
                  <InputField label="Min hourly rate" type="number" value={minHourlyRate} onChange={(e) => setMinHourlyRate(e.target.value)} error={preferenceFieldErrors.minHourlyRate} />
                  <InputField label="Max hourly rate" type="number" value={maxHourlyRate} onChange={(e) => setMaxHourlyRate(e.target.value)} error={preferenceFieldErrors.maxHourlyRate} />
                  <div className="space-y-2">
                    <label className="block">
                      <span className="mb-1.5 block text-sm font-medium text-[var(--jc-ink)]">Preferred locations</span>
                      <div className="flex flex-wrap items-center gap-2 rounded-xl border border-[var(--jc-border)] bg-[var(--jc-surface)] p-2">
                        {preferredLocations.map((locationValue, index) => (
                          <span
                            key={`${locationValue}-${index}`}
                            className="inline-flex items-center gap-2 rounded-full bg-[var(--jc-accent-soft)] px-3 py-1 text-sm text-[var(--jc-ink)]"
                          >
                            {locationValue}
                            <button
                              type="button"
                              aria-label={`Remove ${locationValue}`}
                              onClick={() => removePreferredLocation(index)}
                              className="text-[var(--jc-ink-muted)] hover:text-[var(--jc-ink)]"
                            >
                              ×
                            </button>
                          </span>
                        ))}
                        <input
                          value={preferredLocationInput}
                          onChange={(event) => setPreferredLocationInput(event.target.value)}
                          onKeyDown={onPreferredLocationKeyDown}
                          placeholder="Type a location and press Enter"
                          className="h-10 min-w-[220px] flex-1 bg-transparent px-2 text-[var(--jc-ink)] outline-none"
                        />
                      </div>
                    </label>
                    <p className="text-xs text-[var(--jc-ink-muted)]">Add multiple locations. Duplicate entries are ignored.</p>
                  </div>
                </>
              ) : null}

              {isFreelancer ? (
                <>
                  <label className="block">
                    <span className="mb-1.5 block text-sm font-medium text-[var(--jc-ink)]">Preferred project length</span>
                    <select
                      className="h-12 w-full rounded-xl border border-[var(--jc-border)] bg-[var(--jc-surface)] px-4 text-[var(--jc-ink)]"
                      value={preferredProjectLength}
                      onChange={(e) => setPreferredProjectLength(e.target.value)}
                    >
                      <option value="PROJECT_LENGTH_UNSPECIFIED">No preference</option>
                      <option value="PROJECT_LENGTH_SHORT_TERM">Short term</option>
                      <option value="PROJECT_LENGTH_MEDIUM_TERM">Medium term</option>
                      <option value="PROJECT_LENGTH_LONG_TERM">Long term</option>
                    </select>
                  </label>
                  <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
                    <InputField label="Min budget" type="number" value={minBudget} onChange={(e) => setMinBudget(e.target.value)} error={preferenceFieldErrors.minBudget} />
                    <InputField label="Max budget" type="number" value={maxBudget} onChange={(e) => setMaxBudget(e.target.value)} error={preferenceFieldErrors.maxBudget} />
                  </div>
                  <div className="space-y-2">
                    <span className="block text-sm font-medium text-[var(--jc-ink)]">Contract types</span>
                    <div className="flex flex-wrap gap-2">
                      <button
                        type="button"
                        onClick={() => toggleContractType("fixed")}
                        className={`rounded-full border px-4 py-2 text-sm font-semibold transition ${
                          contractTypes.includes("fixed")
                            ? "border-[var(--jc-accent)] bg-[var(--jc-accent-soft)] text-[var(--jc-ink)]"
                            : "border-[var(--jc-border)] bg-[var(--jc-surface)] text-[var(--jc-ink-muted)] hover:text-[var(--jc-ink)]"
                        }`}
                      >
                        Fixed
                      </button>
                      <button
                        type="button"
                        onClick={() => toggleContractType("hourly")}
                        className={`rounded-full border px-4 py-2 text-sm font-semibold transition ${
                          contractTypes.includes("hourly")
                            ? "border-[var(--jc-accent)] bg-[var(--jc-accent-soft)] text-[var(--jc-ink)]"
                            : "border-[var(--jc-border)] bg-[var(--jc-surface)] text-[var(--jc-ink-muted)] hover:text-[var(--jc-ink)]"
                        }`}
                      >
                        Hourly
                      </button>
                    </div>
                    <p className="text-xs text-[var(--jc-ink-muted)]">Select one or both.</p>
                  </div>
                  <InputField label="Weekly capacity hours" type="number" value={weeklyCapacityHours} onChange={(e) => setWeeklyCapacityHours(e.target.value)} error={preferenceFieldErrors.weeklyCapacityHours} />
                </>
              ) : null}
              <InputField label="UI locale" value={uiLocale} onChange={(e) => setUiLocale(e.target.value)} />
              <label className="flex items-center gap-2 text-sm text-[var(--jc-ink)]">
                <input type="checkbox" checked={emailNotifications} onChange={(e) => setEmailNotifications(e.target.checked)} />
                Email notifications
              </label>
              <label className="flex items-center gap-2 text-sm text-[var(--jc-ink)]">
                <input type="checkbox" checked={pushNotifications} onChange={(e) => setPushNotifications(e.target.checked)} />
                Push notifications
              </label>
            </div>
          ) : null}

          {safeStep === "kyc" ? (
            <div className="space-y-4">
              <InputField label="Tax ID" value={taxId} onChange={(e) => setTaxId(e.target.value)} />
              <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
                <InputField label="Legal name" value={verificationLegalName} onChange={(e) => setVerificationLegalName(e.target.value)} error={kycFieldErrors.legalName} />
                <InputField
                  label="Country code"
                  value={verificationCountryCode}
                  onChange={(e) => setVerificationCountryCode(e.target.value.toUpperCase())}
                  placeholder="e.g. ET"
                  className="placeholder:opacity-60"
                />
              </div>
              <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
                <InputField
                  label="Document type"
                  value={verificationDocumentType}
                  onChange={(e) => setVerificationDocumentType(e.target.value)}
                  placeholder="e.g. passport"
                  className="placeholder:opacity-60"
                />
                <InputField label="Document number (masked)" value={verificationDocumentNumberMasked} onChange={(e) => setVerificationDocumentNumberMasked(e.target.value)} />
              </div>
              <TextAreaField
                label="Submission note (optional)"
                value={verificationSubmissionNote}
                onChange={(e) => setVerificationSubmissionNote(e.target.value)}
                rows={3}
                maxLength={240}
              />
              <UploadDropzone
                label="Upload verification evidence"
                helper="JPG, PNG, WEBP, PDF"
                accept="image/jpeg,image/png,image/webp,application/pdf"
                onChange={onVerificationEvidenceFileChange}
                fileName={verificationEvidenceFile?.name}
              />
              <p className="text-xs text-[var(--jc-ink-muted)]">Verification status: {verificationStatus || "unverified"}</p>
              {normalizedVerificationStatus === "submitted" ? (
                <InlineAlert tone="success">Submitted. You can update and resubmit if needed.</InlineAlert>
              ) : null}
              {normalizedVerificationStatus === "pending_review" ? (
                <InlineAlert tone="warning">Under review. Your latest submission is being reviewed.</InlineAlert>
              ) : null}
              {normalizedVerificationStatus === "verified" ? (
                <InlineAlert tone="success">Verified. Editing and resubmitting will trigger a new review.</InlineAlert>
              ) : null}
              {(normalizedVerificationStatus === "rejected" || normalizedVerificationStatus === "reverification_required") ? (
                <InlineAlert tone="warning">Please update details and resubmit.</InlineAlert>
              ) : null}
              {verificationEvidenceUrl ? (
                <p className="text-xs text-[var(--jc-ink-muted)]">Latest evidence reference: {verificationEvidenceUrl}</p>
              ) : null}
            </div>
          ) : null}

          {safeStep === "cv" ? (
            <div className="space-y-4">
              {canShowCvActions ? (
                <div className="flex flex-wrap items-center gap-3">
                  <SecondaryButton type="button" onClick={() => setIsCvPreviewOpen(true)} disabled={!cvViewUrl || isBusy}>
                    View
                  </SecondaryButton>
                  <SecondaryButton type="button" onClick={onDeleteCv} disabled={isBusy}>
                    Remove
                  </SecondaryButton>
                </div>
              ) : null}
              <UploadDropzone
                label="Upload CV"
                helper="PDF or DOC"
                accept="application/pdf,application/msword,application/vnd.openxmlformats-officedocument.wordprocessingml.document"
                onChange={onCvFileChange}
                fileName={cvFile?.name}
              />
              <p className="text-xs text-[var(--jc-ink-muted)]">Current CV: {cvExists ? "Uploaded" : "Missing"}</p>
            </div>
          ) : null}

          {safeStep === "portfolio" ? (
            <div className="space-y-5">
              {portfolioMode === "list" ? (
                <div className="space-y-4">
                  <div className="flex flex-wrap items-center justify-between gap-3">
                    <div>
                      <p className="text-sm font-semibold text-[var(--jc-ink)]">Your Portfolio</p>
                      <p className="text-xs text-[var(--jc-ink-muted)]">Pick a project to edit or add a new one.</p>
                    </div>
                    <SecondaryButton type="button" onClick={onAddNewPortfolio} disabled={isBusy}>
                      Add New Project
                    </SecondaryButton>
                  </div>

                  {portfolioDisplayItems.length === 0 ? (
                    <InlineAlert tone="warning">No portfolio projects yet. Add your first project to continue.</InlineAlert>
                  ) : (
                    <div className="space-y-3">
                      {portfolioDisplayItems.map((item) => (
                        <div
                          key={item.id}
                          className={`rounded-2xl border p-4 transition ${
                            item.isSelected
                              ? "border-[var(--jc-accent)] bg-[var(--jc-surface)]"
                              : "border-[var(--jc-border)] bg-[var(--jc-surface)]"
                          }`}
                        >
                          <div className="flex flex-wrap items-start justify-between gap-4">
                            <div className="flex items-center gap-3">
                              {item.previewUrl ? (
                                <img src={item.previewUrl} alt={item.title} className="h-14 w-14 rounded-lg object-cover" />
                              ) : (
                                <div className="flex h-14 w-14 items-center justify-center rounded-lg bg-[var(--jc-surface-alt)] text-xs text-[var(--jc-ink-muted)]">
                                  IMG
                                </div>
                              )}
                              <div className="space-y-1">
                                <p className="text-sm font-semibold text-[var(--jc-ink)]">{item.title}</p>
                                <p className="text-xs text-[var(--jc-ink-muted)]">{item.role}</p>
                                {item.tags.length > 0 ? (
                                  <div className="flex flex-wrap gap-1">
                                    {item.tags.slice(0, 3).map((tag) => (
                                      <span
                                        key={`${item.id}-${tag}`}
                                        className="rounded-full bg-[var(--jc-accent-soft)] px-2 py-0.5 text-[11px] text-[var(--jc-ink)]"
                                      >
                                        {tag}
                                      </span>
                                    ))}
                                  </div>
                                ) : null}
                              </div>
                            </div>
                            <div className="flex items-center gap-2">
                              <PrimaryButton type="button" onClick={() => onEditPortfolioItem(item.id)} disabled={isBusy}>
                                Edit
                              </PrimaryButton>
                            </div>
                          </div>
                        </div>
                      ))}
                    </div>
                  )}
                  <p className="text-xs text-[var(--jc-ink-muted)]">Current portfolio items: {portfolioItems.length}</p>
                </div>
              ) : (
                <form className="space-y-4" onSubmit={(event: FormEvent) => event.preventDefault()}>
                  <div className="flex flex-wrap items-center justify-between gap-3">
                    <div>
                      <p className="text-sm font-semibold text-[var(--jc-ink)]">
                        {selectedPortfolioId ? "Edit Project" : "Add New Project"}
                      </p>
                      <p className="text-xs text-[var(--jc-ink-muted)]">
                        {selectedPortfolioItem?.title ? `Editing: ${selectedPortfolioItem.title}` : "Create a strong portfolio entry."}
                      </p>
                    </div>
                    <SecondaryButton type="button" onClick={onDiscardPortfolioEdit} disabled={isBusy}>
                      Back to List
                    </SecondaryButton>
                  </div>

                  <InputField label="Project title" value={portfolioTitle} onChange={(e) => setPortfolioTitle(e.target.value)} />
                  <TextAreaField label="Description" value={portfolioDescription} onChange={(e) => setPortfolioDescription(e.target.value)} rows={4} maxLength={200} />
                  <InputField label="Project URL" value={portfolioProjectUrl} onChange={(e) => setPortfolioProjectUrl(e.target.value)} />
                  <InputField label="Role in project" value={portfolioRoleInProject} onChange={(e) => setPortfolioRoleInProject(e.target.value)} />
                  <div className="space-y-2">
                    <label className="block">
                      <span className="mb-1.5 block text-sm font-medium text-[var(--jc-ink)]">Tags</span>
                      <div className="flex flex-wrap items-center gap-2 rounded-xl border border-[var(--jc-border)] bg-[var(--jc-surface)] p-2">
                        {portfolioTags.map((tag, index) => (
                          <span
                            key={`${tag}-${index}`}
                            className="inline-flex items-center gap-2 rounded-full bg-[var(--jc-accent-soft)] px-3 py-1 text-sm text-[var(--jc-ink)]"
                          >
                            {tag}
                            <button
                              type="button"
                              aria-label={`Remove ${tag}`}
                              onClick={() => removePortfolioTag(index)}
                              className="text-[var(--jc-ink-muted)] hover:text-[var(--jc-ink)]"
                            >
                              ×
                            </button>
                          </span>
                        ))}
                        <input
                          value={portfolioTagInput}
                          onChange={(event) => setPortfolioTagInput(event.target.value)}
                          onKeyDown={onPortfolioTagKeyDown}
                          placeholder="Add a tag and press Enter"
                          className="h-10 min-w-[200px] flex-1 bg-transparent px-2 text-[var(--jc-ink)] outline-none"
                        />
                      </div>
                    </label>
                    <p className="text-xs text-[var(--jc-ink-muted)]">Add multiple tags and remove any one instantly.</p>
                  </div>

                  <div className="space-y-3">
                    <div className="flex flex-wrap items-center justify-between gap-2">
                      <p className="text-sm font-semibold text-[var(--jc-ink)]">Project media</p>
                      <label className="inline-flex cursor-pointer items-center rounded-xl border border-dashed border-[var(--jc-border-strong)] bg-[var(--jc-surface)] px-3 py-2 text-sm font-semibold text-[var(--jc-accent)] hover:bg-[var(--jc-surface-raised)]">
                        Add images
                        <input
                          type="file"
                          multiple
                          accept="image/jpeg,image/png,image/webp"
                          className="hidden"
                          onChange={onPortfolioFileChange}
                        />
                      </label>
                    </div>
                    <p className="text-xs text-[var(--jc-ink-muted)]">JPG, PNG, WEBP. You can add and remove multiple images.</p>
                    <div className="grid grid-cols-2 gap-3 md:grid-cols-4">
                      {existingPortfolioMedia.map((media) => (
                        <div key={media.local_id} className="rounded-xl border border-[var(--jc-border)] bg-[var(--jc-surface)] p-2">
                          <div className="relative overflow-hidden rounded-lg bg-[var(--jc-surface-alt)]">
                            {media.external_url ? (
                              <img src={media.external_url} alt={media.file_name || "Portfolio media"} className="h-24 w-full object-cover" />
                            ) : (
                              <div className="flex h-24 items-center justify-center text-xs text-[var(--jc-ink-muted)]">Image</div>
                            )}
                          </div>
                          <div className="mt-2 flex items-center justify-between gap-2">
                            <p className="truncate text-xs text-[var(--jc-ink-muted)]">{media.file_name || "Uploaded image"}</p>
                            <button
                              type="button"
                              onClick={() => onRemoveExistingPortfolioMedia(media.local_id)}
                              className="text-xs font-semibold text-[#e45353] hover:underline"
                            >
                              Remove
                            </button>
                          </div>
                        </div>
                      ))}

                      {pendingPortfolioMedia.map((media) => (
                        <div key={media.local_id} className="rounded-xl border border-[var(--jc-border)] bg-[var(--jc-surface)] p-2">
                          <div className="relative overflow-hidden rounded-lg bg-[var(--jc-surface-alt)]">
                            <img src={media.preview_url} alt={media.file.name} className="h-24 w-full object-cover" />
                          </div>
                          <div className="mt-2 flex items-center justify-between gap-2">
                            <p className="truncate text-xs text-[var(--jc-ink-muted)]">{media.file.name}</p>
                            <button
                              type="button"
                              onClick={() => onRemovePendingPortfolioMedia(media.local_id)}
                              className="text-xs font-semibold text-[#e45353] hover:underline"
                            >
                              Remove
                            </button>
                          </div>
                        </div>
                      ))}
                    </div>
                  </div>

                  <div className="flex flex-wrap items-center justify-end gap-2">
                    {selectedPortfolioId ? (
                      <SecondaryButton type="button" onClick={onDeletePortfolioItem} disabled={isBusy}>
                        Delete Project
                      </SecondaryButton>
                    ) : null}
                    <SecondaryButton type="button" onClick={onDiscardPortfolioEdit} disabled={isBusy}>
                      Discard
                    </SecondaryButton>
                    <PrimaryButton type="button" onClick={savePortfolio} disabled={isBusy}>
                      Save Project
                    </PrimaryButton>
                  </div>
                </form>
              )}
            </div>
          ) : null}

          {safeStep === "review" ? (
            <div className="space-y-4">
              <p className="text-sm text-[var(--jc-ink-muted)]">Readiness: {readinessPercent}%</p>
              {missingRequired.length > 0 ? (
                <InlineAlert tone="warning">Missing required fields: {missingRequired.join(", ")}</InlineAlert>
              ) : (
                <InlineAlert tone="success">All required readiness fields are complete.</InlineAlert>
              )}
              {recommendations.length > 0 ? (
                <ul className="list-disc space-y-1 pl-4 text-sm text-[var(--jc-ink-muted)]">
                  {recommendations.map((item) => (
                    <li key={item}>{item}</li>
                  ))}
                </ul>
              ) : null}
            </div>
          ) : null}
        </div>

        <div className="mt-8 flex flex-wrap items-center justify-between gap-3">
          <SecondaryButton type="button" onClick={onBack} disabled={isBusy}>Back</SecondaryButton>
          <div className="flex items-center gap-3">
            {safeStep !== "review" ? (
              <SecondaryButton type="button" onClick={onSkip} disabled={isBusy}>Skip</SecondaryButton>
            ) : null}
            <PrimaryButton type="button" onClick={onContinue} disabled={isBusy}>
              {safeStep === "review" ? "Finish" : isBusy ? "Saving..." : isStepDirty ? "Save & Continue" : "Continue"}
            </PrimaryButton>
          </div>
        </div>
      </section>
      {isCvPreviewOpen && cvViewUrl ? (
        <div
          className="fixed inset-0 z-50 flex items-center justify-center bg-black/45 p-4"
          role="dialog"
          aria-modal="true"
          aria-label="CV preview"
          onClick={(event) => {
            if (event.target === event.currentTarget) {
              setIsCvPreviewOpen(false);
            }
          }}
        >
          <div className="w-full max-w-2xl rounded-2xl bg-[var(--jc-surface)] p-4 shadow-xl">
            <div className="mb-3 flex items-center justify-between">
              <p className="text-sm font-semibold text-[var(--jc-ink)]">CV Preview</p>
              <button
                type="button"
                aria-label="Close CV preview"
                onClick={() => setIsCvPreviewOpen(false)}
                className="inline-flex h-8 w-8 items-center justify-center rounded-full text-[var(--jc-ink-muted)] hover:bg-[var(--jc-surface-alt)] hover:text-[var(--jc-ink)]"
              >
                <X className="h-4 w-4" />
              </button>
            </div>
            {isCvPdf ? (
              <iframe
                title="CV preview"
                src={cvViewUrl}
                className="h-[420px] w-full rounded-xl border border-[var(--jc-border)]"
              />
            ) : (
              <div className="rounded-xl border border-[var(--jc-border)] bg-[var(--jc-surface-alt)] p-4 text-sm text-[var(--jc-ink-muted)]">
                <p>Preview is available for PDF files only.</p>
                <a href={cvViewUrl} target="_blank" rel="noreferrer" className="mt-2 inline-block font-semibold text-[var(--jc-accent)] hover:underline">
                  Open or download this CV
                </a>
              </div>
            )}
          </div>
        </div>
      ) : null}
    </PageShell>
  );
}
