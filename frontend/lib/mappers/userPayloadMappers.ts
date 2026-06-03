import type {
  AvailabilityValue,
  PatchHiringPreferencesRequestPayload,
  PatchWorkPreferencesRequestPayload,
  PortfolioMediaTypeValue,
  ProjectLengthValue,
  StringListPayload,
} from "@/lib/contracts/user";

export function toOptionalNumber(value: string): number | undefined {
  const trimmed = value.trim();
  if (!trimmed) return undefined;
  const parsed = Number(trimmed);
  if (!Number.isFinite(parsed)) return undefined;
  return parsed;
}

export function toStringList(values: string[]): StringListPayload | undefined {
  const normalized = values.map((item) => item.trim()).filter(Boolean);
  if (normalized.length === 0) return undefined;
  return { values: normalized };
}

export function csvToValues(value: string): string[] {
  return value
    .split(",")
    .map((item) => item.trim())
    .filter(Boolean);
}

export function normalizeAvailability(value: string): AvailabilityValue | undefined {
  const normalized = value.trim().toUpperCase();
  if (!normalized) return undefined;
  if (normalized.startsWith("AVAILABILITY_")) {
    return normalized as AvailabilityValue;
  }

  switch (normalized) {
    case "FULL_TIME":
      return "AVAILABILITY_FULL_TIME";
    case "PART_TIME":
      return "AVAILABILITY_PART_TIME";
    case "AS_NEEDED":
      return "AVAILABILITY_AS_NEEDED";
    case "UNAVAILABLE":
      return "AVAILABILITY_UNAVAILABLE";
    default:
      return undefined;
  }
}

export function normalizeProjectLength(value: string): ProjectLengthValue | undefined {
  const normalized = value.trim().toUpperCase();
  if (!normalized) return undefined;
  if (normalized.startsWith("PROJECT_LENGTH_")) {
    return normalized as ProjectLengthValue;
  }

  switch (normalized) {
    case "SHORT_TERM":
    case "SHORT":
      return "PROJECT_LENGTH_SHORT_TERM";
    case "MEDIUM_TERM":
    case "MEDIUM":
      return "PROJECT_LENGTH_MEDIUM_TERM";
    case "LONG_TERM":
    case "LONG":
      return "PROJECT_LENGTH_LONG_TERM";
    default:
      return "PROJECT_LENGTH_UNSPECIFIED";
  }
}

export function mediaTypeForContentType(contentType: string): PortfolioMediaTypeValue | null {
  if (contentType.startsWith("image/")) return "PORTFOLIO_MEDIA_TYPE_IMAGE";
  if (contentType.startsWith("video/")) return "PORTFOLIO_MEDIA_TYPE_VIDEO";
  if (contentType === "application/pdf") return "PORTFOLIO_MEDIA_TYPE_FILE";
  return null;
}

export function buildWorkPreferencesPayload(input: {
  preferredProjectLength: string;
  minBudget: string;
  maxBudget: string;
  contractTypes?: string[];
  contractTypesCsv?: string;
  weeklyCapacityHours: string;
}): PatchWorkPreferencesRequestPayload {
  const normalizedContractTypes = Array.isArray(input.contractTypes)
    ? input.contractTypes.map((item) => item.trim()).filter(Boolean)
    : csvToValues(input.contractTypesCsv ?? "");
  return {
    preferred_project_length: normalizeProjectLength(input.preferredProjectLength),
    min_budget: toOptionalNumber(input.minBudget),
    max_budget: toOptionalNumber(input.maxBudget),
    contract_types: toStringList(normalizedContractTypes),
    weekly_capacity_hours: toOptionalNumber(input.weeklyCapacityHours),
  };
}

export function buildHiringPreferencesPayload(input: {
  minHourlyRate: string;
  maxHourlyRate: string;
  preferredLocations: string[];
}): PatchHiringPreferencesRequestPayload {
  return {
    min_hourly_rate: toOptionalNumber(input.minHourlyRate),
    max_hourly_rate: toOptionalNumber(input.maxHourlyRate),
    preferred_locations: toStringList(input.preferredLocations),
  };
}

