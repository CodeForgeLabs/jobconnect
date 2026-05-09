"use client";

interface ChallengeFieldsProps {
  challengeId: string;
  recaptchaToken: string;
  onChallengeIdChange: (value: string) => void;
  onRecaptchaTokenChange: (value: string) => void;
  visible: boolean;
}

export default function ChallengeFields({
  challengeId,
  recaptchaToken,
  onChallengeIdChange,
  onRecaptchaTokenChange,
  visible,
}: ChallengeFieldsProps) {
  if (!visible) return null;

  return (
    <div className="rounded-lg border border-amber-300 bg-amber-50 p-4 text-sm text-amber-900">
      <p className="font-semibold">Extra verification required</p>
      <p className="mt-1 text-xs">
        This action was rate-limited. Enter challenge details and retry.
      </p>
      <div className="mt-3 grid gap-2">
        <input
          className="input input-bordered w-full"
          placeholder="Challenge ID"
          value={challengeId}
          onChange={(e) => onChallengeIdChange(e.target.value)}
        />
        <input
          className="input input-bordered w-full"
          placeholder="reCAPTCHA token"
          value={recaptchaToken}
          onChange={(e) => onRecaptchaTokenChange(e.target.value)}
        />
      </div>
    </div>
  );
}
