"use client";

import { useGetJobsQuery } from "@/api/jobsapi";
import { useGetMeQuery } from "@/api/userapi";
import {
  useGetMyContractsQuery,
  useGetWeeklyLogsMutation,
} from "@/api/contractapi";
import { useGetMyProposalsQuery } from "@/api/proposalapi";
import { useCallback, useEffect, useMemo, useState } from "react";
import {
  ArrowRight,
  BadgeDollarSign,
  CircleCheckBig,
  FileText,
  MessageSquareText,
  Sparkles,
  TrendingUp,
  UserCircle2,
  Wallet,
  Zap,
  CalendarClock,
  ShoppingBag,
} from "lucide-react";
import { useRouter } from "next/navigation";
import Jobcard from "@/components/Jobcard";

const formatDate = (value?: string) => {
  if (!value) return "Recently";

  const parsed = new Date(value);
  if (Number.isNaN(parsed.getTime())) return value;

  return parsed.toLocaleDateString("en-US", {
    month: "short",
    day: "numeric",
    year: "numeric",
  });
};

const formatMoney = (value?: number) =>
  `${Number(value ?? 0).toLocaleString()} birr`;

const formatHoursToHM = (hours?: number) => {
  const totalMinutes = Math.max(0, Math.round((hours ?? 0) * 60));
  const hrs = Math.floor(totalMinutes / 60);
  const mins = totalMinutes % 60;

  if (hrs > 0 && mins > 0) return `${hrs} hr ${mins} min`;
  if (hrs > 0) return `${hrs} hr`;
  return `${mins} min`;
};

const normalizeContractType = (value?: string) =>
  (value ?? "FIXED").toUpperCase();

const calculatePaidAmount = (
  milestones: Array<{ Amount?: number; Status?: string }>,
) =>
  milestones
    .filter((milestone) => {
      const normalized = (milestone.Status ?? "").toUpperCase();
      return normalized === "PAID" || normalized === "APPROVED";
    })
    .reduce((sum, milestone) => sum + Number(milestone.Amount || 0), 0);

const getMilestoneProgressPercent = (status?: string) => {
  const normalized = (status ?? "PENDING").toUpperCase();

  if (normalized === "PAID" || normalized === "APPROVED") return 100;
  if (normalized === "SUBMITTED") return 80;
  if (normalized === "IN_PROGRESS") return 55;
  if (normalized === "REVISION_REQUESTED") return 35;
  return 15;
};

type WeeklySession = {
  id: number;
  start_time: string;
  end_time: string;
  total_hours: number;
  is_paid: boolean;
};

type WeeklyDay = {
  day: string;
  date: string;
  total_hours: number;
  sessions: WeeklySession[];
};

type WeeklyLog = {
  week_number: number;
  week_start: string;
  week_end: string;
  total_hours: number;
  days: WeeklyDay[];
};

const normalizeWeeklyLogs = (response: unknown): WeeklyLog[] => {
  if (Array.isArray(response)) return response as WeeklyLog[];

  if (!response || typeof response !== "object") return [];

  const candidate = response as Record<string, unknown>;
  const maybeData = candidate.data ?? candidate.logs ?? candidate.weeks;

  return Array.isArray(maybeData) ? (maybeData as WeeklyLog[]) : [];
};

const pickCurrentWeeklyLog = (logs: WeeklyLog[]) => {
  if (!logs.length) return null;

  const now = new Date();
  const current = logs.find((log) => {
    const start = new Date(log.week_start);
    const end = new Date(log.week_end);
    return (
      !Number.isNaN(start.getTime()) &&
      !Number.isNaN(end.getTime()) &&
      now >= start &&
      now <= end
    );
  });

  if (current) return current;

  return [...logs].sort(
    (left, right) =>
      new Date(right.week_end).getTime() - new Date(left.week_end).getTime(),
  )[0];
};

export default function FreelancerDashboard() {
  const router = useRouter();
  const { data: me } = useGetMeQuery();
  const { data: jobs = [], isLoading } = useGetJobsQuery();
  const { data: proposals = [] } = useGetMyProposalsQuery();
  const { data: contracts = [] } = useGetMyContractsQuery();
  const [loadWeeklyLogs] = useGetWeeklyLogsMutation();
  const [hourlyWeeklyLogs, setHourlyWeeklyLogs] = useState<WeeklyLog[]>([]);
  const [loadingHourlyLogs, setLoadingHourlyLogs] = useState(false);

  const displayName =
    [me?.first_name, me?.last_name].filter(Boolean).join(" ") || "Natnael";

  const activeContracts = contracts.filter(
    (contract) => String(contract.status).toUpperCase() === "ACTIVE",
  );

  const pendingProposals = proposals.filter(
    (proposal) => String(proposal.status).toUpperCase() === "PENDING",
  );

  const recentJobs = jobs.slice(0, 3);
  const totalEarnings = contracts.reduce(
    (total, contract) => total + (contract.total_budget ?? 0),
    0,
  );
  const connectBalance = me?.connect ?? 0;

  const spotlightContract =
    activeContracts.find(
      (contract) => normalizeContractType(contract.type) === "HOURLY",
    ) ??
    activeContracts[0] ??
    contracts[0];
  const contractType = useMemo(
    () => normalizeContractType(spotlightContract?.type),
    [spotlightContract?.type],
  );
  const isHourlyContract = contractType === "HOURLY";
  const milestones = useMemo(
    () => spotlightContract?.milestones ?? [],
    [spotlightContract?.milestones],
  );
  const deadlineSource = milestones.slice(0, 3) as Array<{
    Description?: string;
    Amount?: number;
    Status?: string;
  }>;
  const paidAmount = useMemo(
    () => calculatePaidAmount(milestones),
    [milestones],
  );
  const fixedBudget =
    spotlightContract?.total_budget ??
    milestones.reduce(
      (sum, milestone) => sum + Number(milestone.Amount ?? 0),
      0,
    );
  const remainingAmount = Math.max(fixedBudget - paidAmount, 0);
  const fixedProgressPercent = fixedBudget
    ? Math.min((paidAmount / fixedBudget) * 100, 100)
    : 0;
  const hourlyWeeklyCap =
    (spotlightContract?.hourly_rate ?? 0) *
    (spotlightContract?.weekly_hour_limit ?? 0);
  const currentWeeklyLog = useMemo(
    () => pickCurrentWeeklyLog(hourlyWeeklyLogs),
    [hourlyWeeklyLogs],
  );
  const workedHoursThisWeek = currentWeeklyLog?.total_hours ?? 0;
  const remainingHoursThisWeek = Math.max(
    (spotlightContract?.weekly_hour_limit ?? 0) - workedHoursThisWeek,
    0,
  );
  const weeklyProgressPercent = spotlightContract?.weekly_hour_limit
    ? Math.min(
        (workedHoursThisWeek / spotlightContract.weekly_hour_limit) * 100,
        100,
      )
    : 0;

  const fetchHourlyWeeklyLogs = useCallback(async () => {
    if (!spotlightContract || !isHourlyContract) {
      setHourlyWeeklyLogs([]);
      return;
    }

    setLoadingHourlyLogs(true);
    try {
      const response = await loadWeeklyLogs({
        contract_id: spotlightContract.contract_id,
      }).unwrap();
      setHourlyWeeklyLogs(normalizeWeeklyLogs(response));
    } catch {
      setHourlyWeeklyLogs([]);
    } finally {
      setLoadingHourlyLogs(false);
    }
  }, [isHourlyContract, loadWeeklyLogs, spotlightContract]);

  useEffect(() => {
    void fetchHourlyWeeklyLogs();
  }, [fetchHourlyWeeklyLogs]);

  return (
    <>
      {/* Main Content Canvas */}
      <main className="pt-12 pb-16 px-8 max-w-360 mx-auto min-h-screen">
        {/* Welcome Header */}
        <header className="mb-10">
          <div className="flex flex-col sm:flex-row justify-between items-start sm:items-end gap-4">
            <div>
              <p className="text-xs font-semibold tracking-widest text-on-surface-variant mb-2 uppercase font-label">
                Freelancer Dashboard
              </p>
              <h1 className="text-4xl md:text-5xl font-extrabold text-primary tracking-tight">
                Good morning, {displayName}
              </h1>
            </div>
            <div className="flex items-center gap-3 bg-surface-container-highest px-5 py-2.5 rounded-full">
              <span className="w-2 h-2 bg-tertiary-container rounded-full animate-pulse"></span>
              <span className="text-xs font-bold text-on-tertiary-fixed-variant font-label">
                {jobs.length} Job Matches Found
              </span>
            </div>
          </div>
        </header>

        {/* Metric Cards Bento Grid */}
        <section className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-4 gap-5 mb-12">
          <div className="bg-surface-container-lowest p-6 rounded-lg shadow-sm hover:shadow-md transition-all duration-300 hover:-translate-y-1 group">
            <div className="flex justify-between items-start mb-3">
              <FileText
                className="h-8 w-8 p-2.5 text-primary bg-primary/5 rounded-2xl"
                aria-hidden="true"
              />
              <span className="text-xs font-bold text-on-surface-variant font-label">
                +{pendingProposals.length} this week
              </span>
            </div>
            <p className="text-sm font-medium text-on-surface-variant mb-1 font-label">
              Active Proposals
            </p>
            <h3 className="text-3xl font-bold text-primary">
              {pendingProposals.length}
            </h3>
          </div>

          <div className="bg-primary bg-linear-to-br from-primary to-primary-container p-6 rounded-lg shadow-lg transition-all duration-300 hover:-translate-y-1 group">
            <div className="flex justify-between items-start mb-3">
              <BadgeDollarSign
                className="h-8 w-8 p-2.5 text-white bg-white/10 rounded-2xl"
                aria-hidden="true"
              />
              <TrendingUp
                className="h-5 w-5 text-white/40"
                aria-hidden="true"
              />
            </div>
            <p className="text-sm font-medium text-white/70 mb-1 font-label">
              Earnings this month
            </p>
            <h3 className="text-3xl font-bold text-white">
              {formatMoney(totalEarnings)}
            </h3>
          </div>

          <div className="bg-surface-container-lowest p-6 rounded-lg shadow-sm hover:shadow-md transition-all duration-300 hover:-translate-y-1 group">
            <div className="flex justify-between items-start mb-3">
              <Zap
                className="h-8 w-8 p-2.5 text-tertiary-container bg-tertiary-fixed rounded-2xl"
                aria-hidden="true"
              />
              <button
                onClick={() => router.push("/freelancer/wallet")}
                className="text-xs font-bold text-primary hover:underline font-label"
              >
                Buy more
              </button>
            </div>
            <p className="text-sm font-medium text-on-surface-variant mb-1 font-label">
              Available Connects
            </p>
            <h3 className="text-3xl font-bold text-on-surface">
              {connectBalance}
            </h3>
          </div>

          <div className="bg-surface-container-lowest p-6 rounded-lg shadow-sm hover:shadow-md transition-all duration-300 hover:-translate-y-1 group">
            <div className="flex justify-between items-start mb-3">
              <CircleCheckBig
                className="h-8 w-8 p-2.5 text-secondary bg-secondary/5 rounded-2xl"
                aria-hidden="true"
              />
            </div>
            <p className="text-sm font-medium text-on-surface-variant mb-1 font-label">
              Active Contracts
            </p>
            <h3 className="text-3xl font-bold text-on-surface">
              {activeContracts.length}
            </h3>
          </div>
        </section>

        {/* Dynamic Content Layout */}
        <div className="grid grid-cols-1 lg:grid-cols-12 gap-10">
          {/* Recommended Jobs Feed */}
          <section className="lg:col-span-8 space-y-6">
            <div className="flex justify-between items-center mb-6">
              <h2 className="text-2xl font-bold text-primary">
                Recommended Jobs
              </h2>
              <button
                onClick={() => router.push("/freelancer/jobsearch")}
                className="text-sm font-bold text-primary flex items-center gap-2 group font-label"
              >
                Filter Feed
                <ArrowRight
                  className="h-4 w-4 group-hover:translate-x-1 transition-transform"
                  aria-hidden="true"
                />
              </button>
            </div>

            {isLoading ? (
              <div className="rounded-lg border border-outline-variant/10 bg-surface-container-lowest p-6 text-sm text-on-surface-variant">
                Loading jobs...
              </div>
            ) : (
              <div className="space-y-5">
                {recentJobs.length > 0 ? (
                  recentJobs.map((job, index) => (
                    <Jobcard
                      key={job.id}
                      index={index}
                      title={job.title}
                      pay={job.budget ? formatMoney(job.budget) : formatMoney(job.hourly_rate)}
                      type={(job.job_type ?? "FIXED").toUpperCase() as "FIXED" | "HOURLY"}
                      jobType={(job.job_type ?? "FIXED").toUpperCase() as "FIXED" | "HOURLY"}
                      description={job.description}
                      postTime={job.created_at}
                      tags={job.skills ? job.skills.split(",").map((s:string)=>s.trim()).filter(Boolean) : []}
                      companyName={job.company_name}
                      status={job.status}
                      skills={job.skills}
                      hourlyRate={job.hourly_rate ? formatMoney(job.hourly_rate) : undefined}
                      budget={job.budget ? formatMoney(job.budget) : undefined}
                      experienceLevel={job.experience_level}
                      createdAt={job.created_at}
                      onApply={() => router.push(`/freelancer/job/${job.id}`)}
                    />
                  ))
                ) : (
                  <div className="rounded-lg border border-dashed border-outline-variant/20 bg-surface-container-lowest p-6 text-sm text-on-surface-variant">
                    No jobs available right now.
                  </div>
                )}
              </div>
            )}
          </section>

          {/* Sidebar: Active Milestones */}
          <aside className="lg:col-span-4">
            <div className="bg-surface-container-lowest p-6 rounded-lg shadow-sm sticky top-24 space-y-6">
              <h3 className="text-lg font-bold text-primary flex items-center gap-2">
                <CalendarClock className="h-5 w-5" aria-hidden="true" />
                {isHourlyContract ? "Hourly Workload" : "Upcoming Deadlines"}
              </h3>

              {spotlightContract ? (
                <div className="grid grid-cols-2 gap-3 rounded-2xl border border-outline-variant/10 bg-surface-container-low p-4 text-sm">
                  {isHourlyContract ? (
                    <>
                      <div>
                        <p className="text-[10px] font-black uppercase tracking-[0.22em] text-on-surface-variant">
                          Worked this week
                        </p>
                        <p className="mt-1 font-bold text-on-surface">
                          {formatHoursToHM(workedHoursThisWeek)}
                        </p>
                      </div>
                      <div>
                        <p className="text-[10px] font-black uppercase tracking-[0.22em] text-on-surface-variant">
                          Remaining this week
                        </p>
                        <p className="mt-1 font-bold text-on-surface">
                          {formatHoursToHM(remainingHoursThisWeek)}
                        </p>
                      </div>
                      <div>
                        <p className="text-[10px] font-black uppercase tracking-[0.22em] text-on-surface-variant">
                          Weekly cap
                        </p>
                        <p className="mt-1 font-bold text-on-surface">
                          {formatHoursToHM(spotlightContract?.weekly_hour_limit ?? 0)}
                        </p>
                      </div>
                      <div>
                        <p className="text-[10px] font-black uppercase tracking-[0.22em] text-on-surface-variant">
                          Weekly value
                        </p>
                        <p className="mt-1 font-bold text-on-surface">
                          {formatMoney(hourlyWeeklyCap)}
                        </p>
                      </div>
                    </>
                  ) : (
                    <>
                      <div>
                        <p className="text-[10px] font-black uppercase tracking-[0.22em] text-on-surface-variant">
                          Contract budget
                        </p>
                        <p className="mt-1 font-bold text-on-surface">
                          {formatMoney(fixedBudget)}
                        </p>
                      </div>
                      <div>
                        <p className="text-[10px] font-black uppercase tracking-[0.22em] text-on-surface-variant">
                          Paid so far
                        </p>
                        <p className="mt-1 font-bold text-on-surface">
                          {formatMoney(paidAmount)}
                        </p>
                      </div>
                      <div>
                        <p className="text-[10px] font-black uppercase tracking-[0.22em] text-on-surface-variant">
                          Remaining
                        </p>
                        <p className="mt-1 font-bold text-on-surface">
                          {formatMoney(remainingAmount)}
                        </p>
                      </div>
                      <div>
                        <p className="text-[10px] font-black uppercase tracking-[0.22em] text-on-surface-variant">
                          Progress
                        </p>
                        <p className="mt-1 font-bold text-on-surface">
                          {fixedProgressPercent.toFixed(0)}%
                        </p>
                      </div>
                    </>
                  )}
                </div>
              ) : null}

              <div className="space-y-6 relative after:absolute after:top-2 after:bottom-2 after:left-1.75 after:w-0.5 after:bg-outline-variant/20">
                {isHourlyContract ? (
                  <div className="relative pl-6 group z-10">
                    <div className="absolute left-0 top-1 w-4 h-4 rounded-full bg-secondary ring-4 ring-surface-container-lowest"></div>
                    <p className="text-xs font-bold uppercase tracking-wider mb-1 font-label text-on-surface-variant">
                      {loadingHourlyLogs ? "Loading week" : "Weekly plan"}
                    </p>
                    <h5 className="text-sm font-bold text-on-surface">
                      {spotlightContract?.title ||
                        spotlightContract?.job_title ||
                        "Hourly contract"}
                    </h5>
                    <p className="text-xs text-on-surface-variant mt-1">
                      {currentWeeklyLog
                        ? `Week ${currentWeeklyLog.week_number} • ${formatDate(currentWeeklyLog.week_start)} to ${formatDate(currentWeeklyLog.week_end)}`
                        : "No weekly log found yet"}
                    </p>
                    <p className="text-xs text-on-surface-variant mt-1">
                      {formatHoursToHM(spotlightContract?.weekly_hour_limit ?? 0)} max per week
                    </p>
                    <p className="text-xs text-on-surface-variant mt-1">
                      Billed at {formatMoney(spotlightContract?.hourly_rate)} /
                      hr
                    </p>
                    <div className="mt-3 h-1.5 w-full bg-surface-container rounded-full overflow-hidden">
                      <div
                        className="h-full bg-secondary transition-all duration-300"
                        style={{ width: `${weeklyProgressPercent}%` }}
                      ></div>
                    </div>
                    <p className="text-[10px] text-right mt-1 font-bold text-secondary font-label">
                      {weeklyProgressPercent.toFixed(0)}% of weekly limit used
                    </p>
                  </div>
                ) : deadlineSource.length > 0 ? (
                  deadlineSource.map((milestone, index) => (
                    <div
                      key={`${milestone.Description ?? "milestone"}-${index}`}
                      className="relative pl-6 group z-10"
                    >
                      <div
                        className={`absolute left-0 top-1 w-4 h-4 rounded-full ring-4 ring-surface-container-lowest ${
                          index === 0 ? "bg-primary" : "bg-outline-variant"
                        }`}
                      ></div>
                      <p
                        className={`text-xs font-bold uppercase tracking-wider mb-1 font-label ${index === 0 ? "text-error" : "text-on-surface-variant"}`}
                      >
                        {index === 0 ? "Due soon" : `Milestone ${index + 1}`}
                      </p>
                      <h5 className="text-sm font-bold text-on-surface">
                        {milestone.Description || "Project milestone"}
                      </h5>
                      <p className="text-xs text-on-surface-variant mt-1">
                        Amount: {formatMoney(milestone.Amount)}
                      </p>
                      {index === 0 ? (
                        <>
                          <div className="mt-3 h-1.5 w-full bg-surface-container rounded-full overflow-hidden">
                            <div
                              className="h-full bg-primary"
                              style={{
                                width: `${getMilestoneProgressPercent(milestone.Status)}%`,
                              }}
                            ></div>
                          </div>
                          <p className="text-[10px] text-right mt-1 font-bold text-primary font-label">
                            {getMilestoneProgressPercent(milestone.Status)}%
                            Done
                          </p>
                        </>
                      ) : null}
                    </div>
                  ))
                ) : (
                  <div className="relative pl-6 group z-10">
                    <div className="absolute left-0 top-1 w-4 h-4 rounded-full bg-outline-variant ring-4 ring-surface-container-lowest"></div>
                    <p className="text-xs font-bold text-on-surface-variant uppercase tracking-wider mb-1 font-label">
                      No deadlines yet
                    </p>
                    <h5 className="text-sm font-bold text-on-surface">
                      Nothing scheduled
                    </h5>
                    <p className="text-xs text-on-surface-variant mt-1">
                      Your milestones will appear here once you are hired.
                    </p>
                  </div>
                )}
              </div>

              <div className="p-5 bg-surface-container-low rounded-xl">
                <h4 className="font-bold text-on-surface text-sm mb-4">
                  Contractor Spotlight
                </h4>
                {spotlightContract ? (
                  <div className="flex items-center gap-3">
                    <div className="w-10 h-10 rounded-full bg-surface-container-high flex items-center justify-center shrink-0">
                      <UserCircle2
                        className="h-5 w-5 text-primary"
                        aria-hidden="true"
                      />
                    </div>
                    <div className="min-w-0">
                      <p className="text-xs font-bold text-on-surface">
                        {spotlightContract.client_first_name}{" "}
                        {spotlightContract.client_last_name}
                      </p>
                      <p className="text-[10px] text-on-surface-variant truncate">
                        {spotlightContract.job_title || spotlightContract.title}
                      </p>
                    </div>
                    <button className="ml-auto text-primary flex items-center justify-center">
                      <MessageSquareText
                        className="h-5 w-5"
                        aria-hidden="true"
                      />
                    </button>
                  </div>
                ) : (
                  <p className="text-xs text-on-surface-variant">
                    No spotlight available yet.
                  </p>
                )}
              </div>
            </div>
          </aside>
        </div>
      </main>
    </>
  );
}
