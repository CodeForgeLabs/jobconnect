"use client";

import { FormEvent, useState } from "react";
import Link from "next/link";
import { authApi } from "@/lib/authApi";

export default function ForgotPasswordPage() {
  const [email, setEmail] = useState("");
  const [loading, setLoading] = useState(false);
  const [message, setMessage] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);

  async function onSubmit(event: FormEvent) {
    event.preventDefault();
    setLoading(true);
    setError(null);
    setMessage(null);

    try {
      await authApi.forgotPassword({ email });
      setMessage("If your email exists, a reset OTP has been sent.");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Request failed");
    } finally {
      setLoading(false);
    }
  }

  return (
    <main className="min-h-screen bg-[#f7f9f7] px-4 py-14">
      <div className="mx-auto max-w-md rounded-2xl border border-[#d7ddd3] bg-white p-8 shadow-sm">
        <h1 className="text-3xl font-semibold text-[#1f1f1f]">Forgot password</h1>
        <p className="mt-2 text-sm text-[#5e6d55]">
          Enter your account email and we will send a reset OTP.
        </p>
        <form className="mt-7 space-y-4" onSubmit={onSubmit}>
          <input
            className="input input-bordered w-full border-[#cfd6ca] bg-white"
            type="email"
            placeholder="Email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            required
          />
          {message && <p className="text-sm text-success">{message}</p>}
          {error && <p className="text-sm text-error">{error}</p>}
          <button
            className="btn w-full border-none bg-[#108a00] text-white hover:bg-[#0d7300]"
            type="submit"
            disabled={loading}
          >
            {loading ? "Sending..." : "Send reset OTP"}
          </button>
        </form>
        <p className="mt-4 text-sm text-[#5e6d55]">
          <Link href="/reset-password" className="text-[#108a00] hover:underline">
            Already have OTP?
          </Link>
        </p>
      </div>
    </main>
  );
}
