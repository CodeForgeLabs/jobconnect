"use client";

import Link from "next/link";
import { useMemo, useState } from "react";
import { Search } from "lucide-react";

import {
  useFetchUsersQuery,
  type FetchUsersParams,
  type User,
} from "@/api/userapi";

const skillOptions = ["React", "Node.js", "Design", "Marketing", "Data"];
const quickFilters = [
  { label: "React", value: "React" },
  { label: "Node.js", value: "Node.js" },
  { label: "UI/UX", value: "UI/UX" },
  { label: "Mobile", value: "Mobile" },
];

function parseSkills(skills: User["skills"]) {
  if (Array.isArray(skills)) return skills;
  if (!skills) return [];
  return skills
    .split(",")
    .map((s) => s.trim())
    .filter(Boolean);
}

function getDisplayName(user: User) {
  return (
    [user.first_name, user.last_name].filter(Boolean).join(" ") ||
    "Anonymous Talent"
  );
}

function getHeadline(user: User) {
  return user.headline ||  "Available for new projects";
}

function UserCard({ user }: { user: User }) {
  const skills = parseSkills(user.skills).slice(0, 4);

  const initials = getDisplayName(user)
    .split(" ")
    .map((namePart: string) => namePart[0])
    .slice(0, 2)
    .join("");

  const bio = user.bio
    ? user.bio.length > 160
      ? `${user.bio.slice(0, 157)}...`
      : user.bio
    : "This talent has not added a bio yet.";

  return (
    <article className="bg-white p-6 rounded-xl border border-slate-200/70 shadow-md transition-transform duration-200 hover:-translate-y-0.5">
      <div className="flex flex-col lg:flex-row gap-4 lg:gap-6 items-start">
        <div className="shrink-0 flex gap-4 items-center">
          <div className="h-12 w-12 rounded-full overflow-hidden bg-slate-100 ring-2 ring-[#dce9ff]">
            {user.profile_picture_url ? (
              <div
                aria-label={getDisplayName(user)}
                role="img"
                className="h-full w-full bg-center bg-cover"
                style={{
                  backgroundImage: `url(${user.profile_picture_url})`,
                }}
              />
            ) : (
              <div className="h-full w-full grid place-items-center bg-gradient-to-br from-[#15157d] to-[#2e3192] text-white text-lg font-bold">
                {initials || "T"}
              </div>
            )}
          </div>
          <div className="flex flex-col gap-2 xl:flex-row xl:items-start xl:justify-between pc:hidden">
            <div>
              

              <h2 className="text-xl md:text-2xl font-extrabold tracking-tight text-[#0d1c2e]">
                {getDisplayName(user)}
              </h2>

              <p className="text-sm font-semibold text-[#15157d]">
                {getHeadline(user)}
              </p>
            </div>

            
          </div>
        </div>

        <div className="flex-1 space-y-2">
          <div className="flex flex-col gap-2 xl:flex-row xl:items-start xl:justify-between max-pc:hidden">
            <div>
              

              <h2 className="text-xl md:text-2xl font-extrabold tracking-tight text-[#0d1c2e]">
                {getDisplayName(user)}
              </h2>

              <p className="text-sm font-semibold text-[#15157d]">
                {getHeadline(user)}
              </p>
            </div>

            
          </div>

          <p className="text-sm text-[#464652] leading-relaxed">{bio}</p>

          <div className="flex flex-wrap gap-2">
            {skills.length > 0 ? (
              skills.map((skill) => (
                <span
                  key={skill}
                  className="rounded-full bg-[#e6eeff] px-3 py-1 text-xs font-semibold text-[#15157d]"
                >
                  {skill}
                </span>
              ))
            ) : (
              null
            )}
          </div>

          <div className="flex flex-wrap gap-3 pt-3 border-t border-slate-200/70 text-xs text-[#464652]">
            <span>
              Availability:{" "}
              <strong className="text-[#0d1c2e]">{user.availability || "N/A"}</strong>
            </span>

            

            <span>
              Location:{" "}
              <strong className="text-[#0d1c2e]">
                {user.location || "N/A"}
              </strong>
            </span>
          </div>
        </div>

        <div className="flex flex-row lg:flex-col gap-3 lg:justify-center">
          <Link
            href={`/freelancer/profile/${user.id}`}
            className="inline-flex items-center justify-center rounded-lg bg-[#15157d] px-4 py-2 text-sm font-bold text-white transition-all hover:shadow-lg active:scale-95"
          >
            View Profile
          </Link>

          <Link
            href={`/client/invite?user=${user.id}`}
            className="inline-flex items-center justify-center rounded-lg bg-[#d5e3fc] px-4 py-2 text-sm font-bold text-[#15157d] transition-all hover:bg-[#dce9ff] active:scale-95"
          >
            Invite to Job
          </Link>

          <div className="text-left xl:text-right">
              <p className="text-xs text-[#464652]">Hourly Rate</p>

              <p className="text-lg font-extrabold text-[#0d1c2e]">
                ${Number(user.hourly_rate || 0).toLocaleString()}/hr
              </p>

            
            </div>
        </div>
      </div>
    </article>
  );
}

export default function FindTalent() {
  const [selectedSkill, setSelectedSkill] = useState("");
  const [location, setLocation] = useState("");
  const [minHourlyRate, setMinHourlyRate] = useState(0);
  const [searchText, setSearchText] = useState("");
  const [activeQuickFilter, setActiveQuickFilter] = useState("");

  const filters = useMemo<FetchUsersParams>(() => {
    const skillValue = [searchText, selectedSkill, activeQuickFilter]
      .filter(Boolean)
      .join(", ");

    return {
      ...(skillValue ? { skills: skillValue } : {}),
      ...(location ? { location } : {}),
      ...(minHourlyRate > 0 ? { min_hourly_rate: minHourlyRate } : {}),
    };
  }, [activeQuickFilter, location, minHourlyRate, searchText, selectedSkill]);

  const {
    data: users = [],
    isLoading,
    isFetching,
    error,
  } = useFetchUsersQuery(filters);

  const handleReset = () => {
    setSelectedSkill("");
    setLocation("");
    setMinHourlyRate(0);
    setSearchText("");
    setActiveQuickFilter("");
  };

  const minRateLabel = minHourlyRate === 0 ? "Any rate" : `$${minHourlyRate}+`;

  return (
    <div className="min-h-screen bg-surface text-[#0d1c2e] selection:bg-[#2e3192] selection:text-white">
      <main className="pt-12 pb-20 px-5 md:px-8 max-w-7xl mx-auto">
        <div className="grid grid-cols-1 lg:grid-cols-[300px_minmax(0,1fr)] gap-8">
          <aside className="lg:sticky lg:top-28 self-start space-y-6">
            <div className="rounded-3xl bg-white/90 backdrop-blur border border-white shadow-[0_20px_70px_-40px_rgba(15,23,42,0.45)] p-6 space-y-6">
              <div>
                <p className="text-xs font-bold uppercase tracking-[0.2em] text-[#15157d]">
                  Filter Results
                </p>
                <h2 className="mt-2 text-2xl font-extrabold text-[#0d1c2e] font-headline">
                  Find the right talent
                </h2>
                
              </div>

              <div className="space-y-4">
                <label className="block text-sm font-semibold text-[#0d1c2e]">
                  Search skills or keywords
                </label>
                <div className="relative">
                    <Search
                      size={18}
                      className="absolute left-3 top-1/2 -translate-y-1/2 text-slate-400"
                    />

                    <input
                      value={searchText}
                      onChange={(event) => setSearchText(event.target.value)}
                      className="w-full rounded-2xl border border-slate-200 bg-white px-10 py-3 text-sm outline-none transition focus:border-[#15157d] focus:ring-2 focus:ring-[#dce9ff]"
                      placeholder="React, design, AWS..."
                      type="text"
                    />
                  </div>
              </div>

              <div className="space-y-4">
                <label className="block text-sm font-semibold text-[#0d1c2e]">
                  Skills filter
                </label>
                <select
                  value={selectedSkill}
                  onChange={(event) => setSelectedSkill(event.target.value)}
                  className="w-full rounded-2xl border border-slate-200 bg-white px-4 py-3 text-sm outline-none transition focus:border-[#15157d]"
                >
                  <option value="">All skills</option>
                  {skillOptions.map((skill) => (
                    <option key={skill} value={skill}>
                      {skill}
                    </option>
                  ))}
                </select>
              </div>

              <div className="space-y-4">
                <label className="block text-sm font-semibold text-[#0d1c2e]">
                  Location
                </label>
                <input
                  value={location}
                  onChange={(event) => setLocation(event.target.value)}
                  className="w-full rounded-2xl border border-slate-200 bg-white px-4 py-3 text-sm outline-none transition focus:border-[#15157d]"
                  placeholder="Country, city, or region"
                  type="text"
                />
              </div>

              <div className="space-y-4">
                <div className="flex items-center justify-between gap-3">
                  <label className="block text-sm font-semibold text-[#0d1c2e]">
                    Minimum hourly rate
                  </label>
                  <span className="text-xs font-bold text-[#15157d]">
                    {minRateLabel}
                  </span>
                </div>
                <input
                  value={minHourlyRate}
                  onChange={(event) =>
                    setMinHourlyRate(Number(event.target.value))
                  }
                  className="w-full accent-[#15157d]"
                  min={0}
                  max={200}
                  step={5}
                  type="range"
                />
              </div>

              <div className="space-y-3">
                <p className="text-sm font-semibold text-[#0d1c2e]">
                  Quick filters
                </p>
                <div className="flex flex-wrap gap-2">
                  {quickFilters.map((filter) => (
                    <button
                      key={filter.value}
                      type="button"
                      onClick={() => setActiveQuickFilter(filter.value)}
                      className={`rounded-full px-3 py-1.5 text-xs font-bold transition-all ${
                        activeQuickFilter === filter.value
                          ? "bg-[#15157d] text-white"
                          : "bg-[#dce9ff] text-[#15157d] hover:bg-[#cfe0ff]"
                      }`}
                    >
                      {filter.label}
                    </button>
                  ))}
                </div>
              </div>

              <div className="flex gap-3">
                <button
                  type="button"
                  onClick={handleReset}
                  className="flex-1 rounded-2xl border border-slate-200 bg-white px-4 py-3 text-sm font-bold text-[#0d1c2e] transition hover:bg-slate-50 active:scale-95"
                >
                  Reset
                </button>
                <div className="flex-1 rounded-2xl bg-[#15157d] px-4 py-3 text-center text-sm font-bold text-white">
                  {users.length} results
                </div>
              </div>
            </div>

            <div className="rounded-3xl bg-[#15157d] p-6 text-white shadow-[0_20px_70px_-40px_rgba(21,21,125,0.9)]">
              <p className="text-xs font-bold uppercase tracking-[0.2em] text-[#9da1ff]">
                Need help?
              </p>
              <h3 className="mt-2 text-2xl font-extrabold font-headline">
                Curate a shortlist faster.
              </h3>
              <p className="mt-3 text-sm text-white/75 leading-relaxed">
                Combine skills, location, and rate filters to find the best
                match for your next project.
              </p>
              <Link
                href="/client/mypostings"
                className="mt-5 inline-flex items-center justify-center rounded-xl bg-[#c6e7ff] px-4 py-2.5 text-sm font-bold text-[#001e2e] transition hover:bg-white"
              >
                Post a job
              </Link>
            </div>
          </aside>

          <section className="space-y-8">
            <header className="space-y-5">
              <div className="space-y-3 max-w-3xl">
                <p className="text-xs font-bold uppercase tracking-[0.2em] text-[#15157d]">
                  Talent Search
                </p>
                <h1 className="text-4xl md:text-5xl font-extrabold tracking-tight text-[#15157d] font-headline">
                  Discover top talent with live filters.
                </h1>
                <p className="text-base text-[#464652] leading-relaxed">
                  Search the talent directory by skills, location, and budget.
                  Results update from the backend endpoint as you change
                  filters.
                </p>
              </div>

              <div className="flex flex-wrap gap-3 items-center">
                <span className="text-xs font-bold text-[#464652] uppercase tracking-widest">
                  Popular:
                </span>
                {skillOptions.slice(0, 4).map((skill) => (
                  <button
                    key={skill}
                    type="button"
                    onClick={() => setSearchText(skill)}
                    className="rounded-full bg-[#dce9ff] px-3 py-1.5 text-[11px] font-bold uppercase tracking-wide text-[#15157d] transition hover:bg-[#15157d] hover:text-white"
                  >
                    {skill}
                  </button>
                ))}
              </div>
            </header>

            <div className="rounded-3xl bg-white/80 backdrop-blur border border-white p-4 md:p-5 shadow-[0_20px_70px_-40px_rgba(15,23,42,0.45)]">
              <div className="flex items-center justify-between gap-4 flex-wrap">
                <div>
                  <p className="text-sm font-semibold text-[#0d1c2e]">
                    Showing {users.filter((user) => user.role === "FREELANCER").length} talent profile
                    {users.length === 1 ? "" : "s"}
                  </p>
                 
                </div>

                <div className="rounded-full bg-[#eff4ff] px-4 py-2 text-xs font-semibold text-[#15157d]">
                  Filters active:{" "}
                  {Object.values(filters).filter(Boolean).length}
                </div>
              </div>
            </div>

            {isLoading ? (
              <div className="rounded-3xl bg-white p-10 text-center shadow-[0_20px_70px_-40px_rgba(15,23,42,0.45)]">
                <div className="mx-auto h-10 w-10 animate-spin rounded-full border-4 border-[#dce9ff] border-t-[#15157d]" />
                <p className="mt-4 text-sm font-semibold text-[#464652]">
                  Loading talent profiles...
                </p>
              </div>
            ) : error ? (
              <div className="rounded-3xl border border-rose-200 bg-rose-50 p-8 text-rose-700">
                <p className="font-bold">Could not load talent profiles.</p>
                <p className="mt-2 text-sm">
                  The fetch endpoint returned an error. Try adjusting the
                  filters or retrying the request.
                </p>
              </div>
            ) : users.length === 0 ? (
              <div className="rounded-3xl bg-white p-10 text-center shadow-[0_20px_70px_-40px_rgba(15,23,42,0.45)]">
                <h2 className="text-2xl font-extrabold text-[#0d1c2e] font-headline">
                  No users found
                </h2>
                <p className="mt-3 text-sm text-[#464652]">
                  Try clearing the filters or lowering the minimum hourly rate.
                </p>
                <button
                  type="button"
                  onClick={handleReset}
                  className="mt-6 rounded-xl bg-[#15157d] px-5 py-3 text-sm font-bold text-white transition active:scale-95"
                >
                  Clear filters
                </button>
              </div>
            ) : (
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                {users
                  .filter((u) => u.role === "FREELANCER")
                  .map((user) => (
                    <UserCard key={user.id} user={user} />
                  ))}
              </div>
            )}
          </section>
        </div>
      </main>
    </div>
  );
}
