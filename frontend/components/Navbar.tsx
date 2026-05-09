"use client";

import { useMemo, useState } from "react";
import Link from "next/link";
import Image from "next/image";
import { Search, X } from "lucide-react";
import { usePathname, useRouter } from "next/navigation";
import { useDispatch, useSelector } from "react-redux";
import logo from "@/assets/Background.svg";
import {
  clearAuthState,
  selectIsHydrated,
  selectIsLoggedIn,
  selectUserRole,
} from "@/features/login/loginSlice";
import { authApi } from "@/lib/authApi";

const AUTH_PAGES = new Set([
  "/login",
  "/signup",
  "/verify-email",
  "/forgot-password",
  "/reset-password",
  "/auth/callback",
]);

function isAuthRoute(pathname: string): boolean {
  if (AUTH_PAGES.has(pathname)) return true;
  return pathname.startsWith("/auth/oauth/");
}

function dashboardPath(role: string): string {
  if (role === "freelancer") return "/freelancer/dashboard";
  if (role === "client") return "/client/dashboard";
  return "/account";
}

export default function Navbar() {
  const pathname = usePathname();
  const router = useRouter();
  const dispatch = useDispatch();

  const [searchQuery, setSearchQuery] = useState("");
  const isHydrated = useSelector(selectIsHydrated);
  const isLoggedIn = useSelector(selectIsLoggedIn);
  const role = useSelector(selectUserRole);

  const authRoute = useMemo(() => isAuthRoute(pathname), [pathname]);
  const onboardingRoute = pathname.startsWith("/onboarding");

  const onLogout = async () => {
    try {
      await authApi.logoutEverywhere();
    } catch {
      // Ignore remote failure and clear local session.
    } finally {
      dispatch(clearAuthState());
      router.replace("/login");
    }
  };

  if (onboardingRoute) {
    return (
      <header className="border-b border-[#e4e8e2] bg-white">
        <div className="mx-auto flex h-16 max-w-6xl items-center justify-between px-4">
          <Link href="/" className="flex items-center gap-3">
            <Image src={logo} alt="JobConnect" className="h-8 w-8" />
            <span className="text-2xl font-semibold text-[#1f1f1f]">JobConnect</span>
          </Link>
          {isHydrated && isLoggedIn && (
            <Link href="/account" className="text-sm font-medium text-[#108a00] hover:underline">
              Account
            </Link>
          )}
        </div>
      </header>
    );
  }

  return (
    <header className="border-b border-[#e4e8e2] bg-white">
      <div className="mx-auto flex h-16 max-w-7xl items-center gap-4 px-4">
        <Link href="/" className="flex items-center gap-3">
          <Image src={logo} alt="JobConnect" className="h-8 w-8" />
          <span className="text-2xl font-semibold text-[#1f1f1f]">JobConnect</span>
        </Link>

        {isHydrated && isLoggedIn && !authRoute && (
          <div className="relative ml-2 hidden w-full max-w-md items-center md:flex">
            <Search className="pointer-events-none absolute left-3 h-4 w-4 text-[#7a7a7a]" />
            <input
              type="text"
              className="w-full rounded-full border border-[#d2d7d0] bg-white py-2 pl-9 pr-9 text-sm outline-none transition focus:border-[#108a00]"
              placeholder="Search for jobs, skills, or talent"
              value={searchQuery}
              onChange={(event) => setSearchQuery(event.target.value)}
            />
            {searchQuery && (
              <button
                type="button"
                onClick={() => setSearchQuery("")}
                className="absolute right-3 text-[#7a7a7a] hover:text-[#1f1f1f]"
              >
                <X className="h-4 w-4" />
              </button>
            )}
          </div>
        )}

        <nav className="ml-auto flex items-center gap-2">
          {!isHydrated && <div className="h-8 w-20 animate-pulse rounded-full bg-[#eef2ee]" />}

          {isHydrated && !isLoggedIn && (
            <>
              <Link
                href="/login"
                className="rounded-full px-4 py-2 text-sm font-medium text-[#1f1f1f] hover:bg-[#f3f6f3]"
              >
                Log in
              </Link>
              <Link
                href="/signup"
                className="rounded-full bg-[#108a00] px-4 py-2 text-sm font-semibold text-white hover:bg-[#0d7300]"
              >
                Sign up
              </Link>
            </>
          )}

          {isHydrated && isLoggedIn && (
            <>
              <Link
                href={dashboardPath(role)}
                className="rounded-full px-4 py-2 text-sm font-medium text-[#1f1f1f] hover:bg-[#f3f6f3]"
              >
                Dashboard
              </Link>
              <Link
                href="/account"
                className="rounded-full border border-[#d2d7d0] px-4 py-2 text-sm font-medium text-[#1f1f1f] hover:bg-[#f3f6f3]"
              >
                Account
              </Link>
              <button
                type="button"
                className="rounded-full px-4 py-2 text-sm font-medium text-[#7a2f2f] hover:bg-[#fff0f0]"
                onClick={onLogout}
              >
                Log out
              </button>
            </>
          )}
        </nav>
      </div>
    </header>
  );
}
