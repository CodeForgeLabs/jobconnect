"use client";

import type { ReactNode } from "react";
import { useEffect } from "react";
import { Provider } from "react-redux";
import { store } from "@/store/store";
import { authApi } from "@/lib/authApi";
import { applyLoginToken } from "@/lib/authSession";
import { markAuthHydrated } from "@/features/login/loginSlice";

interface StoreProviderProps {
  children: ReactNode;
}

export default function StoreProvider({ children }: StoreProviderProps) {
  useEffect(() => {
    let cancelled = false;

    (async () => {
      try {
        const refreshed = await authApi.refresh();
        if (cancelled) return;
        applyLoginToken(
          refreshed.access_token,
          refreshed.access_token_expires_in_seconds
        );
      } catch {
        if (cancelled) return;
      } finally {
        if (!cancelled) {
          store.dispatch(markAuthHydrated());
        }
      }
    })();

    return () => {
      cancelled = true;
    };
  }, []);

  return <Provider store={store}>{children}</Provider>;
}
