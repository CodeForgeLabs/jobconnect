"use client";

import { FormEvent, useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import ChallengeFields from "@/components/auth/ChallengeFields";
import { ApiError } from "@/lib/apiTypes";
import { authApi } from "@/lib/authApi";

export default function SignupPage() {
  const router = useRouter();
  const [firstName, setFirstName] = useState("");
  const [lastName, setLastName] = useState("");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [role, setRole] = useState<"client" | "freelancer">("client");
  const [acceptTerms, setAcceptTerms] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);
  const [challengeVisible, setChallengeVisible] = useState(false);
  const [challengeId, setChallengeId] = useState("");
  const [recaptchaToken, setRecaptchaToken] = useState("");
  const [challengeProof, setChallengeProof] = useState<string | undefined>(undefined);

  async function onSubmit(event: FormEvent) {
    event.preventDefault();
    setLoading(true);
    setError(null);

    try {
      await authApi.register(
        {
          email,
          password,
          first_name: firstName,
          last_name: lastName,
          role,
          accept_terms: acceptTerms,
        },
        challengeProof
      );
      sessionStorage.setItem("jc_new_signup", "1");

      router.push(`/verify-email?email=${encodeURIComponent(email)}&role=${role}`);
    } catch (err) {
      if (err instanceof ApiError && err.payload?.challenge_required) {
        setChallengeVisible(true);
        if (challengeId && recaptchaToken) {
          const solved = await authApi.solveChallenge({
            challenge_id: challengeId,
            recaptcha_token: recaptchaToken,
          });
          setChallengeProof(solved.challenge_proof);
          setError("Challenge solved. Submit again.");
        } else {
          setError("Challenge required. Fill challenge fields.");
        }
      } else {
        setError(err instanceof Error ? err.message : "Signup failed");
      }
    } finally {
      setLoading(false);
    }
  }

  return (
    <main className="min-h-screen bg-[#f7f9f7] px-4 py-14">
      <div className="mx-auto max-w-xl rounded-2xl border border-[#d7ddd3] bg-white p-8 shadow-sm">
        <h1 className="text-3xl font-semibold text-[#1f1f1f]">Create your account</h1>
        <p className="mt-2 text-sm text-[#5e6d55]">
          Join as a client hiring talent or as a freelancer looking for work.
        </p>

        <form className="mt-7 space-y-4" onSubmit={onSubmit}>
          <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
            <input
              className="input input-bordered w-full border-[#cfd6ca] bg-white"
              placeholder="First name"
              value={firstName}
              onChange={(e) => setFirstName(e.target.value)}
              required
            />
            <input
              className="input input-bordered w-full border-[#cfd6ca] bg-white"
              placeholder="Last name"
              value={lastName}
              onChange={(e) => setLastName(e.target.value)}
              required
            />
          </div>
          <input
            className="input input-bordered w-full border-[#cfd6ca] bg-white"
            type="email"
            placeholder="Work email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            required
          />
          <input
            className="input input-bordered w-full border-[#cfd6ca] bg-white"
            type="password"
            placeholder="Password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            required
          />

          <div className="grid grid-cols-2 gap-3">
            <label
              className={`cursor-pointer rounded-xl border p-3 text-center ${
                role === "client"
                  ? "border-[#108a00] bg-[#edf7eb] text-[#164814]"
                  : "border-[#d0d8cb] text-[#53634d]"
              }`}
            >
              <input type="radio" className="sr-only" checked={role === "client"} onChange={() => setRole("client")} />
              Client
            </label>
            <label
              className={`cursor-pointer rounded-xl border p-3 text-center ${
                role === "freelancer"
                  ? "border-[#108a00] bg-[#edf7eb] text-[#164814]"
                  : "border-[#d0d8cb] text-[#53634d]"
              }`}
            >
              <input type="radio" className="sr-only" checked={role === "freelancer"} onChange={() => setRole("freelancer")} />
              Freelancer
            </label>
          </div>

          <label className="flex items-center gap-2 text-sm text-[#5e6d55]">
            <input type="checkbox" className="checkbox checkbox-sm" checked={acceptTerms} onChange={(e) => setAcceptTerms(e.target.checked)} />
            I accept Terms and Privacy Policy.
          </label>

          <ChallengeFields
            visible={challengeVisible}
            challengeId={challengeId}
            recaptchaToken={recaptchaToken}
            onChallengeIdChange={setChallengeId}
            onRecaptchaTokenChange={setRecaptchaToken}
          />

          {error && <p className="text-sm text-error">{error}</p>}

          <button
            type="submit"
            className="btn w-full border-none bg-[#108a00] text-white hover:bg-[#0d7300]"
            disabled={loading || !acceptTerms}
          >
            {loading ? "Creating..." : "Create Account"}
          </button>
        </form>

        <p className="mt-4 text-sm text-[#5e6d55]">
          Already have an account?{" "}
          <Link href="/login" className="text-[#108a00] hover:underline">
            Log in
          </Link>
        </p>

        <div className="mt-7 border-t border-[#e4e8e2] pt-5">
          <p className="mb-2 text-xs font-semibold uppercase tracking-wide text-[#5e6d55]">
            Or sign up with OAuth as <span className="normal-case text-[#1f1f1f]">{role}</span>
          </p>
          <div className="flex gap-2">
            <a
              href={`/api/v1/auth/oauth/google/start?role=${role}`}
              className="btn btn-outline border-[#ccd6c4] bg-white text-[#1f1f1f] hover:bg-[#f7fbf5]"
            >
              Google
            </a>
            <a
              href={`/api/v1/auth/oauth/github/start?role=${role}`}
              className="btn btn-outline border-[#ccd6c4] bg-white text-[#1f1f1f] hover:bg-[#f7fbf5]"
            >
              GitHub
            </a>
          </div>
        </div>
      </div>
    </main>
  );
}
