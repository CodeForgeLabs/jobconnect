"use client";

import Link from "next/link";
import { useMemo, useState } from "react";
import { Loader2, Search, Send, Star, X } from "lucide-react";

import {
  type FetchUsersParams,
  type User,
  useFetchUsersQuery,
} from "@/api/userapi";
import {
  type Job,
  useGetMyJobsQuery,
  useInviteToJobMutation,
} from "@/api/jobsapi";
import { useRouter } from "next/navigation";

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
  return user.headline || "Available for new projects";
}

function getJobBudgetLabel(job: Job) {
  if (job.job_type === "HOURLY") {
    return `$${Number(job.hourly_rate || 0).toLocaleString()}/hr`;
  }

  return `$${Number(job.budget || 0).toLocaleString()} fixed`;
}

interface Userwithreview extends User {
  average_rating: number;
  total_reviews: number;
}

function UserCard({
  user,
  onInvite,
}: {
  user: Userwithreview;
  onInvite: (user: Userwithreview) => void;
}) {
  const skills = parseSkills(user.skills).slice(0, 4);
  const router = useRouter();

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
    <article
      onClick={() => {
        router.push(`/freelancer/profile/${user.id}`);
      }}
      className="bg-white p-6 rounded-2xl border border-slate-200/80 shadow-sm hover:shadow-md transition-all duration-200 flex flex-col justify-between gap-5"
    >
      {/* Main Content Area */}
      <div className="flex gap-4 items-start">
        {/* Profile Picture / Initials Avatar */}
        <div className="h-14 w-14 rounded-full overflow-hidden bg-slate-100 ring-4 ring-slate-100 shrink-0 shadow-sm">
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
            <div className="h-full w-full grid place-items-center bg-linear-to-br from-[#15157d] to-[#2e3192] text-white text-xl font-bold">
              {initials || "T"}
            </div>
          )}
        </div>

        {/* Name, Headline & Rating Data */}
        <div className="space-y-1 min-w-0 flex-1">
          <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-1">
            <h2 className="text-xl font-black tracking-tight text-[#0d1c2e] truncate">
              {getDisplayName(user)}
            </h2>

            {/* Dynamic Bayesian Rating Badge */}
            <div className="flex items-center gap-1 shrink-0">
              <div className="flex items-center gap-0.5 bg-amber-50 text-amber-700 px-2 py-0.5 rounded-lg border border-amber-200/50 text-xs font-black">
                <Star size={13} className="fill-amber-500 text-amber-500" />
                <span>{Number(user.average_rating || 0).toFixed(1)}</span>
              </div>
              <span className="text-[11px] font-bold text-slate-400">
                ({user.total_reviews || 0})
              </span>
            </div>
          </div>

          <p className="text-xs font-bold text-[#15157d] tracking-wide uppercase">
            {getHeadline(user)}
          </p>
        </div>
      </div>

      {/* Bio excerpt description */}
      <p className="text-sm text-[#464652] leading-relaxed line-clamp-3">
        {bio}
      </p>

      {/* Skills Pills tags row */}
      <div className="flex flex-wrap gap-1.5">
        {skills.length > 0 ? (
          skills.map((skill) => (
            <span
              key={skill}
              className="rounded-xl bg-[#e6eeff] px-2.5 py-1 text-xs font-bold text-[#15157d]"
            >
              {skill}
            </span>
          ))
        ) : (
          <span className="text-xs italic text-slate-400">
            No skills specified
          </span>
        )}
      </div>

      {/* Meta Footer Row */}
      <div className="pt-4 border-t border-slate-100 flex flex-col sm:flex-row gap-4 items-stretch sm:items-center justify-between w-full mt-auto">
        {/* Core details data descriptors */}
        <div className="flex gap-4 text-[11px] text-[#464652] font-medium">
          <span>
            Location:{" "}
            <strong className="text-[#0d1c2e] font-bold">
              {user.location || "N/A"}
            </strong>
          </span>
        </div>

        {/* Rates & Actions Controls Wrapper */}
        <div className="flex items-center justify-between sm:justify-end gap-3 native-row">
          <div className="text-left sm:text-right sm:mr-1">
            <p className="text-[10px] font-bold text-slate-400 uppercase tracking-wider">
              Hourly Rate
            </p>
            <p className="text-base font-black text-[#0d1c2e]">
              ${Number(user.hourly_rate || 0).toLocaleString()}
              <span className="text-xs font-normal text-slate-500">/hr</span>
            </p>
          </div>

          <div className="flex gap-2">
            <button
              type="button"
              onClick={(e) => {
                e.stopPropagation();
                onInvite(user);
              }}
              className="inline-flex items-center justify-center rounded-xl bg-[#15157d] px-3 py-2 text-xs font-extrabold text-white transition hover:bg-[#2e3192] hover:shadow-md active:scale-95"
            >
              Invite
            </button>
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
  const [selectedTalent, setSelectedTalent] = useState<User | null>(null);
  const [selectedJobId, setSelectedJobId] = useState<number | null>(null);
  const [inviteMessage, setInviteMessage] = useState<string | null>(null);
  const [inviteError, setInviteError] = useState<string | null>(null);

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

  const { data: users = [], isLoading, error } = useFetchUsersQuery(filters);
  console.log("Fetched users with filters:", filters, users);
  const { data: myJobs = [], isLoading: isLoadingMyJobs } = useGetMyJobsQuery();
  const [inviteToJob, { isLoading: isInviting }] = useInviteToJobMutation();

  const sortedFreelancers = useMemo(() => {
    // 1. Filter for freelancers and safely cast to your extended interface
    const freelancers = (users as Userwithreview[]).filter(
      (u) => u.role === "FREELANCER",
    );

    // 2. Set Bayesian constants
    const m = 5; // Dampening threshold (number of reviews needed to fully trust a rating)
    const C = 3.5; // Platform baseline rating for brand new profiles

    // 3. Execute the weighted sort
    return [...freelancers].sort((a, b) => {
      const rA = a.average_rating ?? 0;
      const vA = a.total_reviews ?? 0;
      const scoreA = (vA * rA + m * C) / (vA + m);

      const rB = b.average_rating ?? 0;
      const vB = b.total_reviews ?? 0;
      const scoreB = (vB * rB + m * C) / (vB + m);

      return scoreB - scoreA; // Sorts descending (highest weighted score first)
    });
  }, [users]);

  const privateJobs = useMemo(
    () =>
      myJobs.filter(
        (job) =>
          job.is_private && (job.status ?? "OPEN").toUpperCase() !== "CLOSED",
      ),
    [myJobs],
  );

  const handleReset = () => {
    setSelectedSkill("");
    setLocation("");
    setMinHourlyRate(0);
    setSearchText("");
    setActiveQuickFilter("");
  };

  const openInviteModal = (user: User) => {
    setSelectedTalent(user);
    setSelectedJobId(null);
    setInviteMessage(null);
    setInviteError(null);
  };

  const closeInviteModal = () => {
    if (isInviting) return;

    setSelectedTalent(null);
    setSelectedJobId(null);
    setInviteMessage(null);
    setInviteError(null);
  };

  const handleInviteToPrivateJob = async () => {
    if (!selectedTalent || !selectedJobId) {
      setInviteError("Select one private job you created before sending.");
      return;
    }

    const selectedJob = privateJobs.find((job) => job.id === selectedJobId);

    if (!selectedJob) {
      setInviteError(
        "Invitation works only on private jobs you created. Select a private job from your postings.",
      );
      return;
    }

    setInviteError(null);
    setInviteMessage(null);

    try {
      await inviteToJob({
        job_id: selectedJob.id,
        user_id: selectedTalent.id,
      }).unwrap();
      setInviteMessage(
        `Invitation sent to ${getDisplayName(selectedTalent)} for ${selectedJob.title}.`,
      );
    } catch {
      setInviteError("Unable to send this invitation right now.");
    }
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
                  {users.filter((user) => user.role === "FREELANCER").length}{" "}
                  results
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
                    Showing{" "}
                    {users.filter((user) => user.role === "FREELANCER").length}{" "}
                    talent profile
                    {users.filter((user) => user.role === "FREELANCER")
                      .length === 1
                      ? ""
                      : "s"}
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
                {sortedFreelancers
                  .filter((u) => u.role === "FREELANCER")
                  .map((user) => (
                    <UserCard
                      key={user.id}
                      user={user}
                      onInvite={openInviteModal}
                    />
                  ))}
              </div>
            )}
          </section>
        </div>
      </main>

      {selectedTalent ? (
        <div
          className="fixed inset-0 z-50 flex items-center justify-center bg-slate-950/60 p-4"
          role="dialog"
          aria-modal="true"
          aria-labelledby="invite-modal-title"
        >
          <div className="w-full max-w-2xl overflow-hidden rounded-3xl bg-white shadow-2xl">
            <div className="flex items-start justify-between gap-4 border-b border-slate-200 px-6 py-5">
              <div>
                <p className="text-xs font-black uppercase tracking-[0.24em] text-[#15157d]">
                  Private job invitation
                </p>
                <h2
                  id="invite-modal-title"
                  className="mt-2 text-2xl font-extrabold text-[#0d1c2e]"
                >
                  Invite {getDisplayName(selectedTalent)}
                </h2>
              </div>
              <button
                type="button"
                onClick={closeInviteModal}
                disabled={isInviting}
                className="rounded-full border border-slate-200 p-2 text-slate-500 transition hover:bg-slate-50 disabled:cursor-not-allowed disabled:opacity-60"
                aria-label="Close invite modal"
              >
                <X className="h-4 w-4" />
              </button>
            </div>

            <div className="space-y-5 px-6 py-5">
              <div className="rounded-2xl border border-amber-200 bg-amber-50 px-4 py-3 text-sm font-bold leading-6 text-amber-900">
                Invitation works only on private jobs you created.
              </div>

              {isLoadingMyJobs ? (
                <div className="rounded-2xl border border-slate-200 bg-slate-50 p-8 text-center">
                  <Loader2 className="mx-auto h-6 w-6 animate-spin text-[#15157d]" />
                  <p className="mt-3 text-sm font-semibold text-[#464652]">
                    Loading your private jobs...
                  </p>
                </div>
              ) : privateJobs.length === 0 ? (
                <div className="rounded-2xl border border-slate-200 bg-slate-50 p-6">
                  <h3 className="text-lg font-extrabold text-[#0d1c2e]">
                    No private jobs available
                  </h3>
                  <p className="mt-2 text-sm leading-6 text-[#464652]">
                    You can only invite talent to private jobs that you created.
                    Create or update a posting as private before sending an
                    invitation.
                  </p>
                  <Link
                    href="/client/mypostings"
                    className="mt-4 inline-flex items-center justify-center rounded-xl bg-[#15157d] px-4 py-2.5 text-sm font-bold text-white transition hover:bg-[#0f0f5d]"
                  >
                    Manage postings
                  </Link>
                </div>
              ) : (
                <div className="space-y-3">
                  <p className="text-sm font-bold text-[#0d1c2e]">
                    Select one private job you created:
                  </p>
                  <div className="max-h-80 space-y-3 overflow-y-auto pr-1">
                    {privateJobs.map((job) => (
                      <label
                        key={job.id}
                        className={`block cursor-pointer rounded-2xl border p-4 transition ${
                          selectedJobId === job.id
                            ? "border-[#15157d] bg-[#eff4ff]"
                            : "border-slate-200 bg-white hover:bg-slate-50"
                        }`}
                      >
                        <div className="flex items-start gap-3">
                          <input
                            type="radio"
                            name="private-job"
                            value={job.id}
                            checked={selectedJobId === job.id}
                            onChange={() => setSelectedJobId(job.id)}
                            className="mt-1 h-4 w-4 accent-[#15157d]"
                          />
                          <div className="min-w-0 flex-1">
                            <div className="flex flex-wrap items-center gap-2">
                              <h3 className="font-extrabold text-[#0d1c2e]">
                                {job.title}
                              </h3>
                              <span className="rounded-full bg-[#15157d] px-2.5 py-1 text-[10px] font-black uppercase tracking-widest text-white">
                                Private
                              </span>
                            </div>
                            <p className="mt-2 line-clamp-2 text-sm leading-6 text-[#464652]">
                              {job.description}
                            </p>
                            <div className="mt-3 flex flex-wrap gap-3 text-xs font-semibold text-[#464652]">
                              <span>{job.job_type}</span>
                              <span>{getJobBudgetLabel(job)}</span>
                              <span>{job.location || "Remote"}</span>
                            </div>
                          </div>
                        </div>
                      </label>
                    ))}
                  </div>
                </div>
              )}

              {inviteError ? (
                <div className="rounded-2xl border border-rose-200 bg-rose-50 px-4 py-3 text-sm font-semibold text-rose-700">
                  {inviteError}
                </div>
              ) : null}

              {inviteMessage ? (
                <div className="rounded-2xl border border-emerald-200 bg-emerald-50 px-4 py-3 text-sm font-semibold text-emerald-700">
                  {inviteMessage}
                </div>
              ) : null}
            </div>

            <div className="flex flex-col-reverse gap-3 border-t border-slate-200 px-6 py-5 sm:flex-row sm:justify-end">
              <button
                type="button"
                onClick={closeInviteModal}
                disabled={isInviting}
                className="rounded-xl border border-slate-200 bg-white px-5 py-3 text-sm font-bold text-[#0d1c2e] transition hover:bg-slate-50 disabled:cursor-not-allowed disabled:opacity-60"
              >
                Close
              </button>
              <button
                type="button"
                onClick={() => void handleInviteToPrivateJob()}
                disabled={
                  isInviting ||
                  isLoadingMyJobs ||
                  privateJobs.length === 0 ||
                  selectedJobId === null
                }
                className="inline-flex items-center justify-center gap-2 rounded-xl bg-[#15157d] px-5 py-3 text-sm font-bold text-white transition hover:bg-[#0f0f5d] disabled:cursor-not-allowed disabled:opacity-60"
              >
                {isInviting ? (
                  <Loader2 className="h-4 w-4 animate-spin" />
                ) : (
                  <Send className="h-4 w-4" />
                )}
                {isInviting ? "Sending invitation..." : "Send invitation"}
              </button>
            </div>
          </div>
        </div>
      ) : null}
    </div>
  );
}
