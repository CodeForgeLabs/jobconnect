"use client";

import { useEffect, useState } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import { applyLoginToken, redirectPathForRole } from "@/lib/authSession";
import { store } from "@/store/store";
import { selectUserRole } from "@/features/login/loginSlice";

export default function AuthCallbackPage() {
  const router = useRouter();
  const params = useSearchParams();
  const [message] = useState(() => {
    const token = params.get("access_token");
    const expires = params.get("access_token_expires_in_seconds");
    if (!token || !expires) {
      return "OAuth callback did not include an access token. Please retry login.";
    }
    const ttl = Number(expires);
    if (!Number.isFinite(ttl)) {
      return "Invalid token expiry in callback URL.";
    }
    return "Completing sign in...";
  });

  useEffect(() => {
    const token = params.get("access_token");
    const expires = params.get("access_token_expires_in_seconds");
    const isNewUser = params.get("is_new_user") === "true";

    if (!token || !expires) return;

    const ttl = Number(expires);
    if (!Number.isFinite(ttl)) return;

    applyLoginToken(token, ttl);
    if (isNewUser) {
      router.replace("/onboarding");
      return;
    }
    const role = selectUserRole(store.getState());
    router.replace(redirectPathForRole(role));
  }, [params, router]);

  return (
    <main className="min-h-screen bg-[#f7f9f7] px-4 py-14">
      <div className="mx-auto max-w-xl rounded-2xl border border-[#d7ddd3] bg-white p-8 shadow-sm">
        <h1 className="text-3xl font-semibold text-[#1f1f1f]">Finishing sign in</h1>
        <p className="mt-3 text-sm text-[#5e6d55]">{message}</p>
      </div>
    </main>
  );
}
