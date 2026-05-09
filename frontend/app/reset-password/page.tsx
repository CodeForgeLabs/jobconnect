"use client";

import { FormEvent, useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { authApi } from "@/lib/authApi";

export default function ResetPasswordPage() {
  const router = useRouter();
  const [email, setEmail] = useState("");
  const [otp, setOtp] = useState("");
  const [newPassword, setNewPassword] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  async function onSubmit(event: FormEvent) {
    event.preventDefault();
    setLoading(true);
    setError(null);

    try {
      await authApi.resetPassword({ email, otp, new_password: newPassword });
      router.push("/login");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Reset failed");
    } finally {
      setLoading(false);
    }
  }

  return (
    <main className="min-h-screen bg-[#f7f9f7] px-4 py-14">
      <div className="mx-auto max-w-md rounded-2xl border border-[#d7ddd3] bg-white p-8 shadow-sm">
        <h1 className="text-3xl font-semibold text-[#1f1f1f]">Reset password</h1>
        <p className="mt-2 text-sm text-[#5e6d55]">
          Enter your email, OTP code, and your new password.
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
          <input
            className="input input-bordered w-full border-[#cfd6ca] bg-white"
            placeholder="OTP"
            value={otp}
            onChange={(e) => setOtp(e.target.value)}
            required
          />
          <input
            className="input input-bordered w-full border-[#cfd6ca] bg-white"
            type="password"
            placeholder="New password"
            value={newPassword}
            onChange={(e) => setNewPassword(e.target.value)}
            required
          />
          {error && <p className="text-sm text-error">{error}</p>}
          <button
            className="btn w-full border-none bg-[#108a00] text-white hover:bg-[#0d7300]"
            type="submit"
            disabled={loading}
          >
            {loading ? "Updating..." : "Reset password"}
          </button>
        </form>
        <p className="mt-4 text-sm text-[#5e6d55]">
          <Link href="/login" className="text-[#108a00] hover:underline">
            Back to login
          </Link>
        </p>
      </div>
    </main>
  );
}
