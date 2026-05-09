"use client";

import { useEffect } from "react";
import { useRouter } from "next/navigation";
import { useSelector } from "react-redux";
import {
  selectIsAuthenticated,
  selectIsHydrated,
} from "@/features/login/loginSlice";

export default function ClientDashboardPage() {
  const router = useRouter();
  const isHydrated = useSelector(selectIsHydrated);
  const isAuthenticated = useSelector(selectIsAuthenticated);

  useEffect(() => {
    if (!isHydrated) return;
    if (!isAuthenticated) {
      router.replace("/login");
    }
  }, [isHydrated, isAuthenticated, router]);

  if (!isHydrated || !isAuthenticated) {
    return (
      <main className="min-h-screen bg-surface px-6 py-10">
        <div className="mx-auto max-w-4xl rounded-2xl border border-outline-variant bg-surface-container-lowest p-8">
          <p className="text-on-surface-variant">Preparing your workspace...</p>
        </div>
      </main>
    );
  }

  return (
    <main className="min-h-screen bg-[#f7f9f7] px-6 py-10">
      <div className="mx-auto max-w-4xl rounded-2xl border border-[#d7ddd3] bg-white p-8">
        <h1 className="text-3xl font-semibold text-[#1f1f1f]">Client Dashboard</h1>
        <p className="mt-3 text-[#5e6d55]">
          Welcome back. Manage profile setup and account security from your account hub.
        </p>
        <div className="mt-6 flex flex-wrap gap-3">
          <a href="/onboarding" className="btn border border-[#ccd6c4] bg-white text-[#1f1f1f] hover:bg-[#f7fbf5]">
            Continue Onboarding
          </a>
          <a href="/account" className="btn border-none bg-[#108a00] text-white hover:bg-[#0d7300]">
            Open Account Hub
          </a>
        </div>
      </div>
    </main>
  );
}
