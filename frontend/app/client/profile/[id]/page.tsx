"use client";

import Image from "next/image";
import Link from "next/link";
import { useParams } from "next/navigation";
import {
  BriefcaseBusiness,
  Building2,
  CalendarDays,
  ChevronRight,
  Loader2,
  Mail,
  MapPin,
  MessageCircle,
  UserCircle2,
} from "lucide-react";

import { type Job, useGetJobsQuery } from "@/api/jobsapi";
import { useGetUserByIdQuery } from "@/api/userapi";

const formatDate = (value?: string | Date) => {
  if (!value) return "Recently";

  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return "Recently";

  return date.toLocaleDateString("en-US", {
    month: "short",
    day: "numeric",
    year: "numeric",
  });
};

const formatBudget = (job: Job) => {
  if (job.job_type === "HOURLY") {
    return `${Number(job.hourly_rate || 0).toLocaleString()} birr/hr`;
  }

  return `${Number(job.budget || 0).toLocaleString()} birr`;
};

const parseSkills = (skills?: string) =>
  (skills || "")
    .split(",")
    .map((skill) => skill.trim())
    .filter(Boolean)
    .slice(0, 5);

export default function PublicClientProfile() {
  const params = useParams<{ id: string }>();
  const clientId = Number(params.id);
  const isValidClientId = Number.isFinite(clientId) && clientId > 0;

  const {
    data: client,
    isLoading: isLoadingClient,
    isError: isClientError,
  } = useGetUserByIdQuery(clientId, {
    skip: !isValidClientId,
  });

  const { data: jobs = [], isLoading: isLoadingJobs } = useGetJobsQuery();

  const publicClientJobs = jobs
    .filter((job) => job.created_by === clientId && !job.is_private)
    .sort(
      (a, b) =>
        new Date(b.created_at).getTime() - new Date(a.created_at).getTime(),
    );

  const openJobs = publicClientJobs.filter(
    (job) => String(job.status).toUpperCase() === "OPEN",
  );
  const totalProposals = publicClientJobs.reduce(
    (total, job) => total + (job.applications_count ?? 0),
    0,
  );

  const displayName =
    [client?.first_name, client?.last_name].filter(Boolean).join(" ") ||
    "Client";
  const initials = displayName
    .split(" ")
    .map((part) => part[0])
    .slice(0, 2)
    .join("");

  if (!isValidClientId) {
    return (
      <main className="min-h-screen bg-surface px-6 py-16">
        <div className="mx-auto max-w-3xl rounded-3xl border border-rose-200 bg-rose-50 p-8 text-rose-700">
          <h1 className="text-2xl font-black">Invalid client profile</h1>
          <p className="mt-2 text-sm">This profile URL is missing a valid id.</p>
        </div>
      </main>
    );
  }

  if (isLoadingClient) {
    return (
      <main className="min-h-screen bg-surface px-6 py-16">
        <div className="mx-auto max-w-7xl space-y-6">
          <div className="h-64 animate-pulse rounded-3xl bg-slate-200/70" />
          <div className="grid gap-6 lg:grid-cols-[0.8fr_1.2fr]">
            <div className="h-80 animate-pulse rounded-3xl bg-slate-200/70" />
            <div className="h-80 animate-pulse rounded-3xl bg-slate-200/70" />
          </div>
        </div>
      </main>
    );
  }

  if (isClientError || !client) {
    return (
      <main className="min-h-screen bg-surface px-6 py-16">
        <div className="mx-auto max-w-3xl rounded-3xl border border-rose-200 bg-rose-50 p-8 text-rose-700">
          <h1 className="text-2xl font-black">Client profile not found</h1>
          <p className="mt-2 text-sm">
            We could not load this public client profile.
          </p>
        </div>
      </main>
    );
  }

  return (
    <main className="min-h-screen bg-surface px-5 py-12 text-on-surface md:px-8">
      <div className="mx-auto max-w-7xl space-y-8">
        <header className="rounded-3xl border border-outline-variant/20 bg-surface-container-lowest p-6 shadow-sm md:p-8">
          <div className="flex flex-col gap-6 lg:flex-row lg:items-end lg:justify-between">
            <div className="flex flex-col gap-5 md:flex-row md:items-center">
              <div className="h-28 w-28 shrink-0 overflow-hidden rounded-3xl bg-surface-container-high ring-4 ring-white">
                {client.profile_picture_url ? (
                  <Image
                    src={client.profile_picture_url}
                    alt={`${displayName} profile photo`}
                    width={112}
                    height={112}
                    className="h-full w-full object-cover"
                  />
                ) : (
                  <div className="grid h-full w-full place-items-center bg-linear-to-br from-primary to-primary-container text-3xl font-black text-on-primary">
                    {initials || <UserCircle2 className="h-12 w-12" />}
                  </div>
                )}
              </div>

              <div>
                <p className="text-xs font-bold uppercase tracking-[0.24em] text-primary">
                  Public Client Profile
                </p>
                <h1 className="mt-2 text-4xl font-extrabold tracking-tight text-on-surface md:text-5xl">
                  {displayName}
                </h1>
                <p className="mt-2 max-w-2xl text-sm leading-7 text-on-surface-variant md:text-base">
                  {client.headline ||
                    "This client has not added a profile headline yet."}
                </p>
              </div>
            </div>

            <Link
              href={`/messages?userid=${client.id}`}
              className="inline-flex items-center justify-center gap-2 rounded-xl bg-primary px-5 py-3 text-sm font-bold text-on-primary transition hover:opacity-90"
            >
              <MessageCircle className="h-4 w-4" />
              Message client
            </Link>
          </div>
        </header>

        <section className="grid gap-6 md:grid-cols-3">
          <StatCard label="Public jobs" value={String(publicClientJobs.length)} />
          <StatCard label="Open jobs" value={String(openJobs.length)} />
          <StatCard label="Total proposals" value={String(totalProposals)} />
        </section>

        <div className="grid gap-8 lg:grid-cols-[0.8fr_1.2fr]">
          <aside className="space-y-6">
            <section className="rounded-3xl border border-outline-variant/20 bg-surface-container-lowest p-6 shadow-sm">
              <h2 className="text-xl font-bold text-on-surface">
                Client details
              </h2>
              <div className="mt-5 space-y-3">
                <InfoRow
                  icon={Building2}
                  label="Company"
                  value={client.company_name || "Company not listed"}
                />
                <InfoRow
                  icon={Mail}
                  label="Email"
                  value={client.email || "Email not available"}
                />
                <InfoRow
                  icon={MapPin}
                  label="Location"
                  value={client.location || "Location not added"}
                />
                <InfoRow
                  icon={CalendarDays}
                  label="Member since"
                  value={formatDate(client.created_at)}
                />
              </div>
            </section>

            <section className="rounded-3xl border border-outline-variant/20 bg-surface-container-lowest p-6 shadow-sm">
              <h2 className="text-xl font-bold text-on-surface">Profile bio</h2>
              <p className="mt-4 whitespace-pre-wrap text-sm leading-7 text-on-surface-variant">
                {client.bio ||
                  "This client has not added a public bio yet."}
              </p>
            </section>
          </aside>

          <section className="rounded-3xl border border-outline-variant/20 bg-surface-container-lowest p-6 shadow-sm">
            <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
              <div>
                <p className="text-xs font-bold uppercase tracking-[0.2em] text-primary">
                  Jobs Posted
                </p>
                <h2 className="mt-2 text-2xl font-bold text-on-surface">
                  Public job postings
                </h2>
              </div>
              <span className="rounded-full bg-primary/10 px-4 py-2 text-xs font-bold text-primary">
                Private jobs are hidden
              </span>
            </div>

            {isLoadingJobs ? (
              <div className="mt-6 rounded-2xl border border-dashed border-outline-variant/30 bg-surface-container-low p-8 text-center">
                <Loader2 className="mx-auto h-6 w-6 animate-spin text-primary" />
                <p className="mt-3 text-sm font-semibold text-on-surface-variant">
                  Loading public jobs...
                </p>
              </div>
            ) : publicClientJobs.length > 0 ? (
              <div className="mt-6 space-y-4">
                {publicClientJobs.map((job) => (
                  <article
                    key={job.id}
                    className="rounded-2xl border border-outline-variant/20 bg-surface-container-low p-5 transition hover:bg-surface-container-high"
                  >
                    <div className="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
                      <div className="min-w-0">
                        <div className="flex flex-wrap items-center gap-2">
                          <h3 className="text-lg font-bold text-on-surface">
                            {job.title}
                          </h3>
                          <span className="rounded-full bg-white px-3 py-1 text-[10px] font-black uppercase tracking-widest text-primary">
                            {job.job_type}
                          </span>
                        </div>
                        <p className="mt-2 line-clamp-2 text-sm leading-6 text-on-surface-variant">
                          {job.description}
                        </p>
                      </div>
                      <span className="shrink-0 rounded-full bg-primary px-3 py-1 text-xs font-bold uppercase tracking-wider text-on-primary">
                        {job.status}
                      </span>
                    </div>

                    <div className="mt-4 flex flex-wrap gap-2 text-xs font-semibold text-on-surface-variant">
                      <span className="rounded-full bg-white px-3 py-1">
                        {formatBudget(job)}
                      </span>
                      <span className="rounded-full bg-white px-3 py-1">
                        {job.work_mode}
                      </span>
                      <span className="rounded-full bg-white px-3 py-1">
                        Posted {formatDate(job.created_at)}
                      </span>
                      <span className="rounded-full bg-white px-3 py-1">
                        {job.applications_count ?? 0} proposals
                      </span>
                    </div>

                    {parseSkills(job.skills).length > 0 ? (
                      <div className="mt-3 flex flex-wrap gap-2">
                        {parseSkills(job.skills).map((skill) => (
                          <span
                            key={`${job.id}-${skill}`}
                            className="rounded-full bg-primary/10 px-3 py-1 text-[11px] font-bold text-primary"
                          >
                            {skill}
                          </span>
                        ))}
                      </div>
                    ) : null}

                    <Link
                      href={`/freelancer/job/${job.id}`}
                      className="mt-4 inline-flex items-center gap-2 text-sm font-bold text-primary hover:underline"
                    >
                      View job details
                      <ChevronRight className="h-4 w-4" />
                    </Link>
                  </article>
                ))}
              </div>
            ) : (
              <div className="mt-6 rounded-2xl border border-dashed border-outline-variant/30 bg-surface-container-low p-8 text-center">
                <BriefcaseBusiness className="mx-auto h-10 w-10 text-primary" />
                <h3 className="mt-4 text-lg font-bold text-on-surface">
                  No public jobs posted
                </h3>
                <p className="mx-auto mt-2 max-w-md text-sm leading-6 text-on-surface-variant">
                  This client does not have any public job postings available
                  right now.
                </p>
              </div>
            )}
          </section>
        </div>
      </div>
    </main>
  );
}

function StatCard({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-3xl border border-outline-variant/20 bg-surface-container-lowest p-6 shadow-sm">
      <p className="text-xs font-bold uppercase tracking-[0.2em] text-on-surface-variant">
        {label}
      </p>
      <p className="mt-3 text-4xl font-black text-on-surface">{value}</p>
    </div>
  );
}

function InfoRow({
  icon: Icon,
  label,
  value,
}: {
  icon: typeof Building2;
  label: string;
  value: string;
}) {
  return (
    <div className="rounded-2xl border border-outline-variant/20 bg-surface-container-low p-4">
      <div className="flex items-center gap-2">
        <Icon className="h-4 w-4 text-primary" />
        <p className="text-xs font-bold uppercase tracking-[0.18em] text-on-surface-variant">
          {label}
        </p>
      </div>
      <p className="mt-2 break-words text-sm font-semibold text-on-surface">
        {value}
      </p>
    </div>
  );
}
