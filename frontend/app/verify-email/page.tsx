"use client";

import { FormEvent, useState } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import { authApi } from "@/lib/authApi";
import { ApiError } from "@/lib/apiTypes";

export default function VerifyEmailPage() {
  const router = useRouter();
  const params = useSearchParams();
  const [email, setEmail] = useState(params.get("email") ?? "");
  const [otp, setOtp] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  async function onSubmit(event: FormEvent) {
    event.preventDefault();
    setError(null);
    setLoading(true);

    try {
      await authApi.verifyEmailOtp({ email, otp });
      router.push(`/login?email=${encodeURIComponent(email)}`);
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Verification failed");
    } finally {
      setLoading(false);
    }
  }

  return (
    <main className="min-h-screen bg-[#f7f9f7] px-4 py-14">
      <div className="mx-auto max-w-md rounded-2xl border border-[#d7ddd3] bg-white p-8 shadow-sm">
        <h1 className="text-3xl font-semibold text-[#1f1f1f]">Verify your email</h1>
        <p className="mt-2 text-sm text-[#5e6d55]">
          Enter the OTP code sent to your email to activate your account.
        </p>

        <form className="mt-7 space-y-4" onSubmit={onSubmit}>
          <input
            className="input input-bordered w-full border-[#cfd6ca] bg-white"
            type="email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            required
          />
          <input
            className="input input-bordered w-full border-[#cfd6ca] bg-white"
            value={otp}
            onChange={(e) => setOtp(e.target.value)}
            placeholder="6-digit OTP"
            required
          />
          {error && <p className="text-sm text-error">{error}</p>}
          <button
            className="btn w-full border-none bg-[#108a00] text-white hover:bg-[#0d7300]"
            type="submit"
            disabled={loading}
          >
            {loading ? "Verifying..." : "Verify Email"}
          </button>
        </form>
      </div>
    </main>
  );
}
