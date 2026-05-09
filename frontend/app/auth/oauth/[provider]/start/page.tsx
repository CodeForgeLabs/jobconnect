"use client";

import { useEffect } from "react";
import { useParams, useSearchParams } from "next/navigation";
import { API_BASE_URL, AUTH_ROUTES } from "@/lib/apiConfig";

export default function OAuthStartPage() {
  const params = useParams<{ provider: string }>();
  const search = useSearchParams();

  useEffect(() => {
    const provider = params.provider;
    const role = search.get("role");
    const query = role ? `?role=${encodeURIComponent(role)}` : "";
    window.location.href = `${API_BASE_URL}${AUTH_ROUTES.oauthStart(provider)}${query}`;
  }, [params.provider, search]);

  return (
    <main className="min-h-screen bg-surface px-4 py-16">
      <div className="mx-auto max-w-xl rounded-2xl border border-outline-variant bg-surface-container-lowest p-8">
        <h1 className="text-xl font-semibold">Redirecting to OAuth provider...</h1>
      </div>
    </main>
  );
}
