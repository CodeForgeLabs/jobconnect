"use client";

import Link from "next/link";
import { useGetMyContractsQuery } from "@/api/contractapi";
import { useGetMyJobsQuery } from "@/api/jobsapi";
import { useGetMeQuery } from "@/api/userapi";
import {
  BadgeDollarSign,
  BriefcaseBusiness,
  ChevronRight,
  FileText,
  Handshake,
  Plus,
  TrendingUp,
  type LucideIcon,
  UserCircle2,
  UserPlus,
} from "lucide-react";
import { useRouter } from "next/navigation";

function EmptyStateBox({
  icon,
  title,
  description,
  actionLabel,
  actionHref,
}: {
  icon: LucideIcon;
  title: string;
  description: string;
  actionLabel: string;
  actionHref: string;
}) {
  const Icon = icon;

  return (
    <div className="rounded-xl border border-dashed border-outline-variant/30 bg-surface-container-low p-6 flex items-start gap-4">
      <div className="h-12 w-12 rounded-full bg-primary/10 text-primary flex items-center justify-center shrink-0">
        <Icon className="h-5 w-5" aria-hidden="true" />
      </div>
      <div className="space-y-2">
        <h3 className="font-headline text-lg font-bold text-on-surface">
          {title}
        </h3>
        <p className="text-sm text-on-surface-variant max-w-md">
          {description}
        </p>
        <Link
          href={actionHref}
          className="inline-flex items-center gap-2 rounded-lg bg-primary px-4 py-2 text-sm font-bold text-on-primary transition-all hover:opacity-90 active:scale-95"
        >
          <Plus className="h-4.5 w-4.5" aria-hidden="true" />
          {actionLabel}
        </Link>
      </div>
    </div>
  );
}

export default function ClientDashboard() {
  const router = useRouter();
  const { data: me } = useGetMeQuery();
  const { data: myJobs = [] } = useGetMyJobsQuery();
  const { data: myContracts = [] } = useGetMyContractsQuery();

  const displayName =
    [me?.first_name, me?.last_name].filter(Boolean).join(" ") || "Marcus";

  const recentJobs = myJobs.slice(0, 4);
  const activeContracts = myContracts.slice(0, 3);

  const activeJobCount =
    myJobs.filter((job) => String(job.status).toUpperCase() === "OPEN")
      .length || myJobs.length;
  const pendingProposalsCount = myJobs.reduce(
    (total, job) => total + (job.applications_count ?? 0),
    0,
  );
  const totalSpent = myContracts.reduce(
    (total, contract) => total + (contract.total_budget ?? 0),
    0,
  );

  const formatDate = (value: Date | string) => {
    if (!value) return "recently";

    const date = new Date(value);
    if (Number.isNaN(date.getTime())) {
      return "recently";
    }

    return date.toLocaleDateString("en-US", {
      month: "short",
      day: "numeric",
      year: "numeric",
    });
  };

  const formatBudget = (job: (typeof myJobs)[number]) => {
    if (job.job_type === "HOURLY") {
      return `${Number(job.hourly_rate || job.budget || 0).toLocaleString()} birr/hr`;
    }

    return `${Number(job.budget || 0).toLocaleString()} birr`;
  };

  const formatSkills = (skills: string) =>
    skills
      .split(",")
      .map((skill) => skill.trim())
      .filter(Boolean)
      .slice(0, 3);

  return (
    <>
      {/* Top NavBar */}

      {/* Main Content */}
      <main className="pt-12 pb-24 px-8 md:px-12 max-w-7xl mx-auto space-y-12">
        {/* Welcome Header & Quick Actions */}
        <header className="flex flex-col md:flex-row md:items-end justify-between gap-8">
          <div className="space-y-2">
            <span className="text-primary font-label text-sm font-bold uppercase tracking-[0.2em]">
              Dashboard Overview
            </span>
            <h1 className="text-5xl font-extrabold tracking-tight text-on-surface font-headline">
              Welcome back, {displayName}
            </h1>
          </div>
          <div className="flex gap-4 text-xs">
            <button
              onClick={() => router.push("/client/findtalent")}
              className="flex items-center gap-2 px-6 py-4 bg-secondary-container text-on-secondary-container font-bold rounded hover:bg-secondary-fixed transition-all active:scale-95"
            >
              <UserPlus className="h-5 w-5" aria-hidden="true" />
              Invite Freelancers
            </button>
            <Link
              href="/client/mypostings"
              className="flex items-center gap-2 px-8 py-4 bg-linear-to-br from-primary to-primary-container text-on-primary font-bold rounded-lg shadow-lg shadow-primary/20 hover:shadow-primary/40 transition-all active:scale-95"
            >
              <Plus className="h-5 w-5" aria-hidden="true" />
              Post a New Job
            </Link>
          </div>
        </header>

        {/* Metrics Grid */}
        <section className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
          <div className="bg-surface-container-lowest p-8 rounded-lg border border-outline-variant/10 shadow-sm hover:shadow-md transition-shadow">
            <div className="flex justify-between items-start mb-4">
              <div className="p-3 bg-surface-container rounded-lg">
                <BriefcaseBusiness
                  className="h-5 w-5 text-primary"
                  aria-hidden="true"
                />
              </div>
            </div>
            <p className="text-on-surface-variant text-sm font-medium font-label">
              Active Jobs
            </p>
            <h3 className="text-4xl font-bold text-on-surface mt-1 font-headline">
              {activeJobCount}
            </h3>
          </div>
          <div className="bg-surface-container-lowest p-8 rounded-lg border border-outline-variant/10 shadow-sm hover:shadow-md transition-shadow">
            <div className="flex justify-between items-start mb-4">
              <div className="p-3 bg-surface-container rounded-lg text-primary">
                <FileText className="h-5 w-5 text-primary" aria-hidden="true" />
              </div>
              <span className="bg-tertiary-fixed text-on-tertiary-fixed-variant px-2 py-1 rounded-full text-[10px] font-bold">
                +12%
              </span>
            </div>
            <p className="text-on-surface-variant text-sm font-medium font-label">
              Pending Proposals
            </p>
            <h3 className="text-4xl font-bold text-on-surface mt-1 font-headline">
              {pendingProposalsCount}
            </h3>
          </div>
          <div className="bg-surface-container-lowest p-8 rounded-lg border border-outline-variant/10 shadow-sm hover:shadow-md transition-shadow">
            <div className="flex justify-between items-start mb-4">
              <div className="p-3 bg-surface-container rounded-lg text-primary">
                <Handshake
                  className="h-5 w-5 text-primary"
                  aria-hidden="true"
                />
              </div>
            </div>
            <p className="text-on-surface-variant text-sm font-medium font-label">
              Active Contracts
            </p>
            <h3 className="text-4xl font-bold text-on-surface mt-1 font-headline">
              {activeContracts.length}
            </h3>
          </div>
          <div className="bg-surface-container-lowest p-8 rounded-lg border border-outline-variant/10 shadow-sm hover:shadow-md transition-shadow">
            <div className="flex justify-between items-start mb-4">
              <div className="p-3 bg-surface-container rounded-lg text-primary">
                <BadgeDollarSign
                  className="h-5 w-5 text-primary"
                  aria-hidden="true"
                />
              </div>
              <span className="text-primary font-bold text-xs">This Year</span>
            </div>
            <p className="text-on-surface-variant text-sm font-medium font-label">
              Total Spent
            </p>
            <div className="flex items-baseline gap-2">
              <h3 className="text-4xl font-bold text-on-surface mt-1 font-headline">
                {totalSpent.toLocaleString()} birr
              </h3>
              <TrendingUp
                className="h-4 w-4 text-on-tertiary-container"
                aria-hidden="true"
              />
            </div>
          </div>
        </section>

        {/* Asymmetric Body Layout */}
        <div className="grid grid-cols-1 lg:grid-cols-12 gap-12">
          {/* Recent Job Postings (Left Column) */}
          <section className="lg:col-span-7 space-y-6">
            <div className="flex items-center justify-between">
              <h2 className="text-2xl font-bold text-on-surface font-headline">
                Recent Job Postings
              </h2>
              <button
                onClick={() => {
                  router.push("/client/mypostings");
                }}
                className="text-primary text-sm font-bold hover:underline"
              >
                View All
              </button>
            </div>
            <div className="space-y-4">
              {recentJobs.length > 0 ? (
                recentJobs.map((job) => (
                  <div
                    // onClick={() => router.push(`/client/job/${job.job_id}`)}
                    key={job.title}
                    className="group bg-surface-container-low p-6 rounded-lg flex items-center justify-between hover:bg-surface-container-highest transition-all duration-200 cursor-pointer hover:translate-x-1"
                  >
                    <div className="space-y-3 pr-6 max-w-[72%]">
                      <div className="space-y-1">
                        <h4 className="font-bold text-lg text-on-surface group-hover:text-primary transition-colors font-headline">
                          {job.title}
                        </h4>
                        <p className="text-on-surface-variant text-sm line-clamp-2">
                          {job.description}
                        </p>
                      </div>

                      <div className="flex flex-wrap items-center gap-2 text-xs text-on-surface-variant">
                        <span>{formatDate(job.created_at)}</span>
                        <span className="w-1 h-1 rounded-full bg-outline-variant"></span>
                        <span className="font-medium">
                          {job.applications_count} Proposals
                        </span>
                        <span className="w-1 h-1 rounded-full bg-outline-variant"></span>
                        <span className="font-medium">{formatBudget(job)}</span>
                      </div>

                      <div className="flex flex-wrap gap-2">
                        <span className="rounded-full bg-white px-3 py-1 text-[11px] font-semibold text-primary">
                          {job.job_type}
                        </span>
                        <span className="rounded-full bg-white px-3 py-1 text-[11px] font-semibold text-on-surface-variant">
                          {job.work_mode}
                        </span>
                        {formatSkills(job.skills).map((skill) => (
                          <span
                            key={skill}
                            className="rounded-full bg-surface-container px-3 py-1 text-[11px] font-medium text-on-surface-variant"
                          >
                            {skill}
                          </span>
                        ))}
                      </div>
                    </div>
                    <div className="flex items-center gap-4">
                      <span className="bg-tertiary-fixed text-on-tertiary-fixed-variant text-[11px] font-bold px-3 py-1 rounded-full uppercase tracking-wider">
                        {job.status}
                      </span>
                      <ChevronRight
                        className="h-5 w-5 text-outline-variant"
                        aria-hidden="true"
                      />
                    </div>
                  </div>
                ))
              ) : (
                <EmptyStateBox
                  icon={BriefcaseBusiness}
                  title="No recent job postings"
                  description="Add a new job post to start receiving proposals and fill this section with matching talent."
                  actionLabel="Add new post"
                  actionHref="/client/mypostings"
                />
              )}
            </div>
          </section>

          {/* Active Contracts (Right Column) */}
          <section className="lg:col-span-5 space-y-6">
            <div className="flex items-center justify-between">
              <h2 className="text-2xl font-bold text-on-surface font-headline">
                Active Contracts
              </h2>
              <button
                onClick={() => {
                  router.push("/client/mycontracts");
                }}
                className="text-primary text-sm font-bold hover:underline"
              >
                Manage
              </button>
            </div>
            <div className="bg-surface-container-low rounded-lg divide-y divide-outline-variant/10 overflow-hidden">
              {activeContracts.length > 0 ? (
                activeContracts.map((contract) => (
                  <div key={contract.contract_id} className="p-6 space-y-4">
                    <div className="flex items-center gap-4">
                      <div className="h-12 w-12 rounded-full overflow-hidden shrink-0 bg-surface-container-high flex items-center justify-center">
                        <UserCircle2
                          className="h-6 w-6 text-primary"
                          aria-hidden="true"
                        />
                      </div>
                      <div>
                        <h5 className="font-bold text-on-surface font-headline">
                          {contract.freelancer_first_name}{" "}
                          {contract.freelancer_last_name}
                        </h5>
                        <p className="text-on-surface-variant text-sm">
                          {contract.job_title || contract.title}
                        </p>
                      </div>
                    </div>
                    <div className="grid gap-3 rounded-lg bg-surface-container-highest/50 px-4 py-3 md:grid-cols-3">
                      <div className="space-y-1">
                        <p className="text-[10px] font-bold uppercase tracking-[0.18em] text-on-surface-variant">
                          Current milestone
                        </p>
                        <span className="text-sm font-medium text-on-surface-variant">
                          {contract.milestones?.[0]?.Description ||
                            "No milestone yet"}
                        </span>
                      </div>
                      <div className="space-y-1">
                        <p className="text-[10px] font-bold uppercase tracking-[0.18em] text-on-surface-variant">
                          Budget
                        </p>
                        <span className="text-sm font-medium text-on-surface-variant">
                          {contract.total_budget?.toLocaleString()} birr
                        </span>
                      </div>
                      <div className="space-y-1">
                        <p className="text-[10px] font-bold uppercase tracking-[0.18em] text-on-surface-variant">
                          Contract details
                        </p>
                        <span className="text-sm font-medium text-on-surface-variant">
                          {contract.type} · {contract.weekly_hour_limit} hrs/wk
                        </span>
                      </div>
                    </div>
                    <div className="flex flex-wrap items-center gap-2">
                      <span className="rounded-full bg-white px-3 py-1 text-[11px] font-semibold text-primary">
                        Started {formatDate(contract.start_date)}
                      </span>
                      <span className="rounded-full bg-white px-3 py-1 text-[11px] font-semibold text-on-surface-variant">
                        {contract.milestones?.length || 0} milestones
                      </span>
                      <span className="text-on-tertiary-fixed-variant text-xs font-bold uppercase tracking-tight ml-auto">
                        {contract.status}
                      </span>
                    </div>
                  </div>
                ))
              ) : (
                <div className="p-6">
                  <EmptyStateBox
                    icon={Handshake}
                    title="No active contracts yet"
                    description="Hire your first freelancer and this section will show milestones, approvals, and progress updates."
                    actionLabel="Post a new job"
                    actionHref="/client/mypostings"
                  />
                </div>
              )}
            </div>
          </section>
        </div>
      </main>

      {/* Footer */}
    </>
  );
}
