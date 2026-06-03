const personalNamePattern = /^[\p{L}\p{M} .'\-]+$/u;
const phonePattern = /^\+?\d+$/;
const positiveDecimalPattern = /^(?:\d+(?:\.\d*)?|\.\d+)$/;
const positiveWholeNumberPattern = /^[1-9]\d*$/;

export const personalNameMessage = (label: string) =>
  `${label} can only include letters, spaces, hyphens, apostrophes, and periods.`;

export const phoneNumberMessage = (label: string) =>
  `${label} can only include digits, with an optional leading +.`;

export const positiveDecimalMessage = (label: string) =>
  `${label} must be a valid number greater than 0.`;

export const positiveWholeNumberMessage = (label: string) =>
  `${label} must be a whole number greater than 0.`;

export const minHourlyRateMessage =
  "Minimum hourly rate cannot be greater than maximum hourly rate.";

export const minBudgetMessage =
  "Minimum budget cannot be greater than maximum budget.";

export function validatePersonalName(value: string, label: string) {
  const trimmedValue = value.trim();
  if (!trimmedValue) return null;

  return personalNamePattern.test(trimmedValue)
    ? null
    : personalNameMessage(label);
}

export function validatePhoneNumber(value: string, label: string) {
  const trimmedValue = value.trim();
  if (!trimmedValue) return null;

  return phonePattern.test(trimmedValue) ? null : phoneNumberMessage(label);
}

export function validatePositiveDecimal(value: string, label: string) {
  const trimmedValue = value.trim();

  if (
    !trimmedValue ||
    !positiveDecimalPattern.test(trimmedValue) ||
    Number(trimmedValue) <= 0
  ) {
    return positiveDecimalMessage(label);
  }

  return null;
}

export function validateOptionalPositiveDecimal(value: string, label: string) {
  return value.trim() ? validatePositiveDecimal(value, label) : null;
}

export function validatePositiveWholeNumber(value: string, label: string) {
  const trimmedValue = value.trim();

  return positiveWholeNumberPattern.test(trimmedValue)
    ? null
    : positiveWholeNumberMessage(label);
}

export function validateOptionalPositiveWholeNumber(
  value: string,
  label: string,
) {
  return value.trim() ? validatePositiveWholeNumber(value, label) : null;
}
