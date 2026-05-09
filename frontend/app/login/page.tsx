"use client";

import { FormEvent, useMemo, useState } from "react";
import Link from "next/link";
import { useRouter, useSearchParams } from "next/navigation";
import { ApiError } from "@/lib/apiTypes";
import { authApi } from "@/lib/authApi";
import { applyLoginToken, redirectPathForRole } from "@/lib/authSession";
import { store } from "@/store/store";
import { selectUserRole } from "@/features/login/loginSlice";
import ChallengeFields from "@/components/auth/ChallengeFields";

export default function LoginPage() {
  const router = useRouter();
  const params = useSearchParams();
  const [email, setEmail] = useState(params.get("email") ?? "");
  const [password, setPassword] = useState("");
  const [oauthRole, setOauthRole] = useState<"client" | "freelancer">("client");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [challengeVisible, setChallengeVisible] = useState(false);
  const [challengeId, setChallengeId] = useState("");
  const [recaptchaToken, setRecaptchaToken] = useState("");
  const [challengeProof, setChallengeProof] = useState<string | undefined>(undefined);

  const canSubmit = useMemo(() => email && password && !loading, [email, password, loading]);

  async function maybeSolveChallenge(err: ApiError): Promise<string | undefined> {
    if (!err.payload?.challenge_required) {
      return undefined;
    }

    setChallengeVisible(true);
    if (!challengeId || !recaptchaToken) {
      throw new Error("Challenge required. Fill challenge fields and retry.");
    }

    const solved = await authApi.solveChallenge({
      challenge_id: challengeId,
      recaptcha_token: recaptchaToken,
    });

    return solved.challenge_proof;
  }

  async function onSubmit(event: FormEvent) {
    event.preventDefault();
    setLoading(true);
    setError(null);

    try {
      const result = await authApi.login({ email, password }, challengeProof);
      applyLoginToken(result.access_token, result.access_token_expires_in_seconds);
      const firstTime = sessionStorage.getItem("jc_new_signup") === "1";
      if (firstTime) {
        sessionStorage.removeItem("jc_new_signup");
        router.push("/onboarding");
        return;
      }
      const role = selectUserRole(store.getState());
      router.push(redirectPathForRole(role));
    } catch (err) {
      if (err instanceof ApiError) {
        try {
          const proof = await maybeSolveChallenge(err);
          if (proof) {
            setChallengeProof(proof);
            setError("Challenge solved. Submit again to continue.");
          } else {
            setError(err.message);
          }
        } catch (challengeErr) {
          setError(challengeErr instanceof Error ? challengeErr.message : err.message);
        }
      } else {
        setError("Unexpected error. Please try again.");
      }
    } finally {
      setLoading(false);
    }
  }

  return (
    <main className="min-h-screen bg-[#f7f9f7] px-4 py-14">
      <div className="mx-auto max-w-md rounded-2xl border border-[#d7ddd3] bg-white p-8 shadow-sm">
        <h1 className="text-3xl font-semibold text-[#1f1f1f]">Log in to JobConnect</h1>
        <p className="mt-2 text-sm text-[#5e6d55]">Continue where you left off.</p>

        <form className="mt-7 space-y-4" onSubmit={onSubmit}>
          <input
            className="input input-bordered w-full border-[#cfd6ca] bg-white"
            type="email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            placeholder="Email"
            required
          />
          <input
            className="input input-bordered w-full border-[#cfd6ca] bg-white"
            type="password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            placeholder="Password"
            required
          />

          <ChallengeFields
            visible={challengeVisible}
            challengeId={challengeId}
            recaptchaToken={recaptchaToken}
            onChallengeIdChange={setChallengeId}
            onRecaptchaTokenChange={setRecaptchaToken}
          />

          {error && <p className="text-sm text-error">{error}</p>}

          <button
            className="btn w-full border-none bg-[#108a00] text-white hover:bg-[#0d7300]"
            disabled={!canSubmit}
            type="submit"
          >
            {loading ? "Signing in..." : "Log in"}
          </button>
        </form>

        <div className="mt-4 flex items-center justify-between text-sm">
          <Link href="/forgot-password" className="text-[#108a00] hover:underline">
            Forgot password?
          </Link>
          <Link href="/signup" className="text-[#108a00] hover:underline">
            Create account
          </Link>
        </div>

        <div className="mt-7 border-t border-[#e4e8e2] pt-5">
          <p className="mb-3 text-xs font-semibold uppercase tracking-wide text-[#5e6d55]">
            Continue with OAuth
          </p>

          <div className="mb-3 grid grid-cols-2 gap-2 rounded-xl border border-[#d8dfd3] bg-[#f8faf7] p-1">
            <button
              type="button"
              className={`rounded-lg px-3 py-2 text-sm font-medium transition ${
                oauthRole === "client"
                  ? "bg-white text-[#1f1f1f] shadow-sm"
                  : "text-[#607058] hover:bg-white"
              }`}
              onClick={() => setOauthRole("client")}
            >
              As Client
            </button>
            <button
              type="button"
              className={`rounded-lg px-3 py-2 text-sm font-medium transition ${
                oauthRole === "freelancer"
                  ? "bg-white text-[#1f1f1f] shadow-sm"
                  : "text-[#607058] hover:bg-white"
              }`}
              onClick={() => setOauthRole("freelancer")}
            >
              As Freelancer
            </button>
          </div>

          <div className="grid grid-cols-1 gap-2">
            <a
              href={`/api/v1/auth/oauth/google/start?role=${oauthRole}`}
              className="btn btn-outline justify-start border-[#ccd6c4] bg-white text-[#1f1f1f] hover:bg-[#f7fbf5]"
            >
              Continue with Google
            </a>
            <a
              href={`/api/v1/auth/oauth/github/start?role=${oauthRole}`}
              className="btn btn-outline justify-start border-[#ccd6c4] bg-white text-[#1f1f1f] hover:bg-[#f7fbf5]"
            >
              Continue with GitHub
            </a>
          </div>
        </div>
      </div>
    </main>
  );
}
