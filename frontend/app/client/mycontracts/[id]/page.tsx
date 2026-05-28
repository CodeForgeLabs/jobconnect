"use client";

/* eslint-disable @next/next/no-img-element */

import Link from "next/link";
import { useParams  , useRouter} from "next/navigation";
import {
  ChangeEvent,
  useCallback,
  useEffect,
  useMemo,
  useState,
} from "react";
import {
  BadgeAlert,
  CalendarDays,
  CheckCircle2,
  Clock3,
  ExternalLink,
  FileText,
  Loader2,
  MessageCircle,
  RefreshCcw,
  Wallet,
  type LucideIcon,
} from "lucide-react";
import {
  type ContractMilestone,
  useGetContractByIdQuery,
  useGetWeeklyLogsMutation,
  usePayWeeklyLogsMutation,
  useUpdateMilestoneStatusMutation,
  useUpdateContractStatusMutation,
} from "@/api/contractapi";

type MilestoneStatusMeta = {
  label: string;
  Icon: LucideIcon;
  className: string;
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

const formatMoney = (value: number) =>
  new Intl.NumberFormat("en-US", {
    style: "currency",
    currency: "USD",
    minimumFractionDigits: 2,
  }).format(Number(value || 0));

const formatDate = (value?: string) => {
  if (!value) return "N/A";

  const parsed = new Date(value);
  if (Number.isNaN(parsed.getTime())) return value;

  return parsed.toLocaleDateString("en-US", {
    month: "short",
    day: "numeric",
    year: "numeric",
  });
};

const formatTime = (value?: string) => {
  if (!value) return "--";

  const parsed = new Date(value);
  if (Number.isNaN(parsed.getTime())) return value;

  return parsed.toLocaleTimeString("en-US", {
    hour: "numeric",
    minute: "2-digit",
  });
};

const normalizeContractType = (type?: string) =>
  (type ?? "FIXED").toUpperCase();

const getMilestoneStatusMeta = (status?: string): MilestoneStatusMeta => {
  const normalized = (status ?? "PENDING").toUpperCase();

  if (normalized === "APPROVED") {
    return {
      label: "Approved",
      Icon: CheckCircle2,
      className: "bg-emerald-100 text-emerald-700 border-emerald-200",
    };
  }

  if (normalized === "PAID") {
    return {
      label: "Paid",
      Icon: Wallet,
      className: "bg-sky-100 text-sky-700 border-sky-200",
    };
  }

  if (normalized === "SUBMITTED") {
    return {
      label: "Submitted",
      Icon: FileText,
      className: "bg-blue-100 text-blue-700 border-blue-200",
    };
  }

  if (normalized === "IN_PROGRESS") {
    return {
      label: "In Progress",
      Icon: Clock3,
      className: "bg-amber-100 text-amber-700 border-amber-200",
    };
  }

  if (normalized === "REVISION_REQUESTED") {
    return {
      label: "Revision Requested",
      Icon: RefreshCcw,
      className: "bg-rose-100 text-rose-700 border-rose-200",
    };
  }

  return {
    label: "Pending",
    Icon: BadgeAlert,
    className: "bg-slate-100 text-slate-700 border-slate-200",
  };
};

const calculatePaidAmount = (milestones: ContractMilestone[]) =>
  milestones
    .filter((milestone) => {
      const normalized = (milestone.Status ?? "").toUpperCase();
      return normalized === "PAID" || normalized === "APPROVED";
    })
    .reduce((sum, milestone) => sum + Number(milestone.Amount || 0), 0);

const normalizeWeeklyLogs = (response: unknown): WeeklyLog[] => {
  if (Array.isArray(response)) return response as WeeklyLog[];

  if (!response || typeof response !== "object") return [];

  const candidate = response as Record<string, unknown>;
  const maybeData = candidate.data ?? candidate.logs ?? candidate.weeks;

  return Array.isArray(maybeData) ? (maybeData as WeeklyLog[]) : [];
};

const StatCard = ({
  label,
  value,
  helper,
}: {
  label: string;
  value: string;
  helper?: string;
}) => (
  <div className="rounded-3xl border border-white/60 bg-white/80 p-5 shadow-[0_18px_50px_rgba(15,23,42,0.08)] backdrop-blur">
    <p className="text-[10px] font-black uppercase tracking-[0.25em] text-slate-500">
      {label}
    </p>
    <p className="mt-3 text-2xl font-black tracking-tight text-slate-900">
      {value}
    </p>
    {helper ? <p className="mt-2 text-xs text-slate-500">{helper}</p> : null}
  </div>
);




export default function ContractManagement() {
  const params = useParams<{ id: string }>();
  const router = useRouter();
  const contractId = Number(params?.id);
  const isValidId = Number.isFinite(contractId) && contractId > 0;

  const {
    data: contract,
    isLoading,
    isError,
    refetch,
  } = useGetContractByIdQuery(contractId, {
    skip: !isValidId,
  });

  const [loadWeeklyLogs] = useGetWeeklyLogsMutation();
  const [payWeeklyLogs] = usePayWeeklyLogsMutation();
  const [updateMilestoneStatus] = useUpdateMilestoneStatusMutation();
  const [updateContractStatus] = useUpdateContractStatusMutation();
  const [weeklyLogs, setWeeklyLogs] = useState<WeeklyLog[]>([]);
  const [pageMessage, setPageMessage] = useState<string | null>(null);
  const [loadingWeeklyLogs, setLoadingWeeklyLogs] = useState(false);
  const [payingWeekKey, setPayingWeekKey] = useState<string | null>(null);
  const [requestingMilestoneId, setRequestingMilestoneId] = useState<
    number | null
  >(null);
  const [milestoneFeedbacks, setMilestoneFeedbacks] = useState<
  Record<number, string>
>({});

  const contractType = useMemo(
    () => normalizeContractType(contract?.type),
    [contract?.type],
  );
  const isHourly = contractType === "HOURLY";
  const milestones = useMemo(
    () => contract?.milestones ?? [],
    [contract?.milestones],
  );
  const paidAmount = useMemo(
    () => calculatePaidAmount(milestones),
    [milestones],
  );
  const remainingAmount = Math.max(
    (contract?.total_budget ?? 0) - paidAmount,
    0,
  );
  const progressPercent = contract?.total_budget
    ? Math.min((paidAmount / contract.total_budget) * 100, 100)
    : 0;
  const freelancerName = [
    contract?.freelancer_first_name,
    contract?.freelancer_last_name,
  ]
    .filter(Boolean)
    .join(" ")
    .trim();

  const syncWeeklyLogs = useCallback(async () => {
    if (!contract || !isHourly) return;

    setLoadingWeeklyLogs(true);
    try {
      const response = await loadWeeklyLogs({
        contract_id: contract.contract_id,
      }).unwrap();
      setWeeklyLogs(normalizeWeeklyLogs(response));
    } catch {
      setWeeklyLogs([]);
      setPageMessage("Unable to load weekly work logs right now.");
    } finally {
      setLoadingWeeklyLogs(false);
    }
  }, [contract, isHourly, loadWeeklyLogs]);

  useEffect(() => {
    if (!isHourly || !contract) return;
    void syncWeeklyLogs();
  }, [contract, isHourly, syncWeeklyLogs]);

  const[feedBeckError, setFeedBackError] = useState<string | null>(null);

  const handleRequestChanges = async (milestone: ContractMilestone) => {
    setPageMessage(null);
    setRequestingMilestoneId(milestone.ID);
    if (!milestoneFeedbacks[milestone.ID] || milestoneFeedbacks[milestone.ID].trim() === "") {
      setFeedBackError("Feedback is required to request changes.");
      setRequestingMilestoneId(null);
      return;
    }
    try {
      await updateMilestoneStatus({
        milestoneId: milestone.ID,
        newStatus: "REVISION_REQUESTED",
        
      }).unwrap();
      setPageMessage(`Requested changes for ${milestone.Description}.`);
      await refetch();
    } catch {
      setPageMessage("Unable to request changes right now.");
    } finally {
      setRequestingMilestoneId(null);
    }
  };

  const handlePayWeek = async (week: WeeklyLog) => {
    if (!contract) return;

    const weekKey = `${week.week_number}-${new Date(week.week_start).getFullYear()}`;
    setPageMessage(null);
    setPayingWeekKey(weekKey);

    try {
      await payWeeklyLogs({
        contract_id: contract.contract_id,
        week_number: week.week_number,
        year: new Date(week.week_start).getFullYear(),
      }).unwrap();

      setPageMessage(`Paid weekly logs for week ${week.week_number}.`);
      await syncWeeklyLogs();
      await refetch();
    } catch {
      setPageMessage("Unable to process weekly payment right now.");
    } finally {
      setPayingWeekKey(null);
    }
  };

  if (!isValidId) {
    return (
      <div className="mx-auto mt-16 max-w-3xl rounded-3xl border border-rose-200 bg-rose-50 p-8 text-rose-700 shadow-sm">
        <h1 className="text-xl font-black">Invalid contract ID</h1>
        <p className="mt-2 text-sm">The contract URL is missing a valid id.</p>
      </div>
    );
  }

  if (isLoading) {
    return (
      <div className="mx-auto mt-12 grid max-w-7xl gap-6 px-4 pb-16 md:px-8 lg:px-10">
        <div className="h-44 animate-pulse rounded-4xl bg-slate-200/70" />
        <div className="grid gap-6 lg:grid-cols-2">
          <div className="h-64 animate-pulse rounded-4xl bg-slate-200/70" />
          <div className="h-64 animate-pulse rounded-4xl bg-slate-200/70" />
        </div>
        <div className="h-96 animate-pulse rounded-4xl bg-slate-200/70" />
      </div>
    );
  }

  if (isError || !contract) {
    return (
      <div className="mx-auto mt-16 max-w-3xl rounded-3xl border border-rose-200 bg-rose-50 p-8 text-rose-700 shadow-sm">
        <h1 className="text-xl font-black">Unable to load contract</h1>
        <p className="mt-2 text-sm">Please try again.</p>
        <button
          type="button"
          onClick={() => refetch()}
          className="mt-5 rounded-2xl bg-rose-600 px-4 py-2.5 text-sm font-bold text-white"
        >
          Retry
        </button>
      </div>
    );
  }

  const activeSinceLabel = contract.start_date
    ? `Active since ${formatDate(contract.start_date)}`
    : "Contract active";

  const contractSummary = contract.description || contract.proposal_description;
  const statusLabel = contract.status || "ACTIVE";

 // ...existing code...

const handleUpdateContractStatus = async (
  newStatus: string,
  milestoneId: number
) => {
  if (!contract) return;

  await updateMilestoneStatus({
    milestoneId: milestoneId, // use the key expected by UpdateContractStatusRequest
    newStatus: newStatus      // use the key expected by UpdateContractStatusRequest
  }).unwrap();
};

// ...existing code...


  return (
    <div className="min-h-screen bg-surface text-slate-900 selection:bg-amber-200 selection:text-slate-900">
      <main className="mx-auto max-w-7xl px-4 pb-16 pt-10 md:px-8 lg:px-10">
        <header className="mb-8 flex flex-col gap-6 rounded-4xl border border-white/70 bg-white/75 p-6 shadow-[0_24px_80px_rgba(15,23,42,0.08)] backdrop-blur md:p-8 lg:flex-row lg:items-end lg:justify-between">
          <div className="max-w-3xl">
            <nav className="mb-4 flex items-center gap-2 text-xs font-semibold uppercase tracking-[0.24em] text-slate-500">
              <span>Contracts</span>
              <span>/</span>
              <span>{contractType === "HOURLY" ? "Hourly" : "Fixed"}</span>
            </nav>
            <div className="flex flex-wrap items-center gap-3">
              <span className="rounded-full bg-slate-900 px-3 py-1 text-[10px] font-black uppercase tracking-[0.22em] text-white">
                {contractType === "HOURLY"
                  ? "Hourly Contract"
                  : "Fixed Contract"}
              </span>
              <span className="rounded-full border border-slate-200 bg-white px-3 py-1 text-[10px] font-black uppercase tracking-[0.22em] text-slate-600">
                {statusLabel}
              </span>
              <span className="text-sm text-slate-500">{activeSinceLabel}</span>
            </div>
            <h1 className="mt-4 text-3xl font-black tracking-tight text-slate-950 md:text-5xl">
              {contract.title || contract.job_title}
            </h1>
            {contractSummary ? (
              <p className="mt-4 max-w-3xl text-sm leading-7 text-slate-600 md:text-base">
                {contractSummary}
              </p>
            ) : null}
          </div>

          <div className="flex flex-wrap items-center gap-3">
            <button className="inline-flex items-center gap-2 rounded-2xl border border-slate-200 bg-white px-5 py-3 text-sm font-bold text-slate-700 transition hover:border-slate-300 hover:bg-slate-50"
            onClick={() => {
              router.push(`/messages?userid=${contract.freelancer_id}`);
            }}
            >
              <MessageCircle className="h-4 w-4" />
              Message freelancer
            </button>
            <Link
              href="/client/mycontracts"
              className="inline-flex items-center gap-2 rounded-2xl bg-slate-950 px-5 py-3 text-sm font-bold text-white transition hover:bg-slate-800"
            >
              Back to contracts
            </Link>
          </div>
        </header>

        {pageMessage ? (
          <div className="mb-6 rounded-2xl border border-amber-200 bg-amber-50 px-4 py-3 text-sm font-medium text-amber-800">
            {pageMessage}
          </div>
        ) : null}

        <section className="grid gap-6 lg:grid-cols-[1.15fr_0.85fr]">
          <div className="rounded-4xl border border-white/70 bg-white/80 p-6 shadow-[0_24px_80px_rgba(15,23,42,0.08)] backdrop-blur md:p-8">
            <div className="flex items-start justify-between gap-4">
              <div>
                <p className="text-[10px] font-black uppercase tracking-[0.25em] text-slate-500">
                  Freelancer
                </p>
                <h2 className="mt-3 text-2xl font-black tracking-tight text-slate-950">
                  {freelancerName || contract.freelancer_email}
                </h2>
                <p className="mt-2 text-sm text-slate-500">
                  {contract.freelancer_headline || contract.job_title}
                </p>
              </div>

              <div className="h-16 w-16 overflow-hidden rounded-2xl border border-slate-200 bg-slate-100">
                {contract.freelancer_profile_picture_url ? (
                  <img
                    src={contract.freelancer_profile_picture_url}
                    alt={freelancerName || contract.freelancer_email}
                    className="h-full w-full object-cover"
                  />
                ) : (
                  <div className="flex h-full w-full items-center justify-center text-lg font-black text-slate-500">
                    {freelancerName?.charAt(0) || "F"}
                  </div>
                )}
              </div>
            </div>

            <div className="mt-6 grid gap-4 sm:grid-cols-3">
              <StatCard
                label="Total budget"
                value={formatMoney(contract.total_budget || contract.hourly_rate * contract.weekly_hour_limit )}
                helper={
                  contractType === "HOURLY" ? "Budget ceiling" : "Fixed price"
                }
              />
              <StatCard
                label="Paid so far"
                value={formatMoney(paidAmount)}
                helper={
                  contractType === "HOURLY"
                    ? "Amount paid based on approved weekly logs"
                    : "Approved milestones"
                }
              />
              <StatCard
                label="Remaining"
                value={formatMoney(remainingAmount)}
                helper={
                  contractType === "HOURLY"
                    ? "Available balance"
                    : "Pending payout"
                }
              />
            </div>

            <div className="mt-6 rounded-[1.75rem] border border-slate-200 bg-linear-to-r from-slate-950 to-slate-800 p-5 text-white">
              <div className="flex items-center justify-between gap-4">
                <div>
                  <p className="text-[10px] font-black uppercase tracking-[0.25em] text-white/60">
                    Progress
                  </p>
                  <p className="mt-2 text-2xl font-black tracking-tight">
                    {Math.round(progressPercent)}%
                  </p>
                </div>
                <p className="max-w-md text-right text-sm text-white/70">
                  {contractType === "HOURLY"
                    ? "Hourly work is tracked weekly and can be paid once logs are ready."
                    : "Fixed work is managed milestone by milestone with submission review."}
                </p>
              </div>
              <div className="mt-4 h-3 overflow-hidden rounded-full bg-white/10">
                <div
                  className="h-full rounded-full bg-amber-300 transition-all duration-500"
                  style={{ width: `${Math.max(progressPercent, 8)}%` }}
                />
              </div>
            </div>
          </div>

          <div className="rounded-4xl border border-white/70 bg-white/80 p-6 shadow-[0_24px_80px_rgba(15,23,42,0.08)] backdrop-blur md:p-8">
            <div className="flex items-center justify-between gap-4">
              <div>
                <p className="text-[10px] font-black uppercase tracking-[0.25em] text-slate-500">
                  Contract terms
                </p>
                <h3 className="mt-3 text-xl font-black tracking-tight text-slate-950">
                  {contractType === "HOURLY"
                    ? "Hourly workflow"
                    : "Fixed workflow"}
                </h3>
              </div>
              <div className="rounded-full border border-slate-200 bg-slate-50 px-3 py-1 text-xs font-black uppercase tracking-[0.22em] text-slate-600">
                ID #{contract.contract_id}
              </div>
            </div>

            <div className="mt-6 grid gap-4 sm:grid-cols-2">
              <div className="rounded-2xl bg-slate-50 p-4">
                <p className="text-[10px] font-black uppercase tracking-[0.22em] text-slate-500">
                  Start date
                </p>
                <p className="mt-2 text-sm font-bold text-slate-900">
                  {formatDate(contract.start_date)}
                </p>
              </div>
              <div className="rounded-2xl bg-slate-50 p-4">
                <p className="text-[10px] font-black uppercase tracking-[0.22em] text-slate-500">
                  End date
                </p>
                <p className="mt-2 text-sm font-bold text-slate-900">
                  {formatDate(contract.end_date)}
                </p>
              </div>
              <div className="rounded-2xl bg-slate-50 p-4">
                <p className="text-[10px] font-black uppercase tracking-[0.22em] text-slate-500">
                  Hourly rate
                </p>
                <p className="mt-2 text-sm font-bold text-slate-900">
                  {formatMoney(contract.hourly_rate)}
                </p>
              </div>
              <div className="rounded-2xl bg-slate-50 p-4">
                <p className="text-[10px] font-black uppercase tracking-[0.22em] text-slate-500">
                  Weekly limit
                </p>
                <p className="mt-2 text-sm font-bold text-slate-900">
                  {contract.weekly_hour_limit || "—"} hrs
                </p>
              </div>
            </div>

            <div className="mt-6 rounded-2xl border border-slate-200 bg-amber-50 p-4 text-sm leading-7 text-slate-700">
              {contractType === "HOURLY"
                ? "Hourly contracts show weekly work logs below. Each week can be reviewed and paid after the logged sessions are verified."
                : "Fixed contracts show submitted milestone files below. You can review each submission and request changes from the milestone card."}
            </div>
          </div>
        </section>

        {contractType === "HOURLY" ? (
          <section className="mt-8 rounded-4xl border border-white/70 bg-white/80 p-6 shadow-[0_24px_80px_rgba(15,23,42,0.08)] backdrop-blur md:p-8">
            <div className="flex flex-col gap-3 sm:flex-row sm:items-end sm:justify-between">
              <div>
                <p className="text-[10px] font-black uppercase tracking-[0.25em] text-slate-500">
                  Weekly hours
                </p>
                <h2 className="mt-3 text-2xl font-black tracking-tight text-slate-950">
                  Weekly work logs and payment
                </h2>
              </div>
              <button
                type="button"
                onClick={() => void syncWeeklyLogs()}
                className="inline-flex items-center gap-2 self-start rounded-2xl border border-slate-200 bg-white px-4 py-2.5 text-sm font-bold text-slate-700 transition hover:bg-slate-50"
              >
                {loadingWeeklyLogs ? (
                  <Loader2 className="h-4 w-4 animate-spin" />
                ) : (
                  <CalendarDays className="h-4 w-4" />
                )}
                Refresh week
              </button>
            </div>

            {loadingWeeklyLogs ? (
              <div className="mt-6 rounded-3xl border border-dashed border-slate-200 bg-slate-50 p-8 text-center text-sm text-slate-500">
                Loading weekly logs...
              </div>
            ) : weeklyLogs.length > 0 ? (
              <div className="mt-6 space-y-6">
                {weeklyLogs.map((week) => {
                  const weekKey = `${week.week_number}-${new Date(week.week_start).getFullYear()}`;
                  const allSessionsPaid = week.days.every((day) =>
                    day.sessions.every((session) => session.is_paid),
                  );

                  return (
                    <article
                      key={weekKey}
                      className="overflow-hidden rounded-[1.75rem] border border-slate-200 bg-slate-50"
                    >
                      <div className="flex flex-col gap-4 border-b border-slate-200 bg-linear-to-r from-slate-950 to-slate-800 p-5 text-white md:flex-row md:items-center md:justify-between">
                        <div>
                          <p className="text-[10px] font-black uppercase tracking-[0.25em] text-white/60">
                            Week {week.week_number}
                          </p>
                          <h3 className="mt-2 text-xl font-black tracking-tight">
                            {formatDate(week.week_start)} -{" "}
                            {formatDate(week.week_end)}
                          </h3>
                          <p className="mt-2 text-sm text-white/70">
                            {week.total_hours.toFixed(2)} total hours logged
                          </p>
                        </div>

                        <button
                          type="button"
                          onClick={() => void handlePayWeek(week)}
                          disabled={
                            allSessionsPaid || payingWeekKey === weekKey
                          }
                          className="inline-flex items-center justify-center gap-2 rounded-2xl bg-amber-300 px-5 py-3 text-sm font-black text-slate-950 transition hover:bg-amber-200 disabled:cursor-not-allowed disabled:opacity-50"
                        >
                          {payingWeekKey === weekKey ? (
                            <Loader2 className="h-4 w-4 animate-spin" />
                          ) : (
                            <Wallet className="h-4 w-4" />
                          )}
                          {allSessionsPaid ? "Paid" : "Pay weekly logs"}
                        </button>
                      </div>

                      <div className="grid gap-4 p-5">
                        {week.days.map((day) => (
                          <div
                            key={`${weekKey}-${day.date}`}
                            className="rounded-2xl border border-slate-200 bg-white p-4"
                          >
                            <div className="flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
                              <div>
                                <p className="text-sm font-black text-slate-900">
                                  {day.day}
                                </p>
                                <p className="text-xs text-slate-500">
                                  {formatDate(day.date)}
                                </p>
                              </div>
                              <p className="text-sm font-bold text-slate-700">
                                {day.total_hours.toFixed(2)} hrs
                              </p>
                            </div>

                            <div className="mt-4 space-y-3">
                              {day.sessions.map((session) => (
                                <div
                                  key={session.id}
                                  className="flex flex-col gap-3 rounded-2xl bg-slate-50 px-4 py-3 sm:flex-row sm:items-center sm:justify-between"
                                >
                                  <div>
                                    <p className="text-sm font-bold text-slate-900">
                                      Session #{session.id}
                                    </p>
                                    <p className="text-xs text-slate-500">
                                      {formatTime(session.start_time)} -{" "}
                                      {formatTime(session.end_time)}
                                    </p>
                                  </div>
                                  <div className="flex items-center gap-3">
                                    <span className="rounded-full bg-slate-900 px-3 py-1 text-[10px] font-black uppercase tracking-[0.22em] text-white">
                                      {session.total_hours.toFixed(2)} hrs
                                    </span>
                                    <span
                                      className={`rounded-full border px-3 py-1 text-[10px] font-black uppercase tracking-[0.22em] ${session.is_paid ? "border-emerald-200 bg-emerald-50 text-emerald-700" : "border-amber-200 bg-amber-50 text-amber-700"}`}
                                    >
                                      {session.is_paid ? "Paid" : "Unpaid"}
                                    </span>
                                  </div>
                                </div>
                              ))}
                            </div>
                          </div>
                        ))}
                      </div>
                    </article>
                  );
                })}
              </div>
            ) : (
              <div className="mt-6 rounded-[1.75rem] border border-dashed border-slate-200 bg-slate-50 p-10 text-center">
                <p className="text-sm font-semibold text-slate-500">
                  No weekly logs found for this contract yet.
                </p>
              </div>
            )}
          </section>
        ) : (
          <section className="mt-8 rounded-4xl border border-white/70 bg-white/80 p-6 shadow-[0_24px_80px_rgba(15,23,42,0.08)] backdrop-blur md:p-8">
            <div className="flex flex-col gap-3 sm:flex-row sm:items-end sm:justify-between">
              <div>
                <p className="text-[10px] font-black uppercase tracking-[0.25em] text-slate-500">
                  Milestones
                </p>
                <h2 className="mt-3 text-2xl font-black tracking-tight text-slate-950">
                  Submitted files and change requests
                </h2>
              </div>
              <div className="rounded-full border border-slate-200 bg-slate-50 px-4 py-2 text-xs font-black uppercase tracking-[0.22em] text-slate-600">
                {milestones.length} milestone
                {milestones.length === 1 ? "" : "s"}
              </div>
            </div>

            <div className="mt-6 space-y-4">
              {milestones.length > 0 ? (
                milestones.map((milestone) => {
                  const meta = getMilestoneStatusMeta(milestone.Status);
                  // const canRequestChanges = ![
                  //   "PAID",
                  //   "REVISION_REQUESTED",
                  // ].includes((milestone.Status ?? "").toUpperCase());

                  return (
                    <article
                      key={milestone.ID}
                      className="rounded-[1.75rem] border border-slate-200 bg-slate-50 p-5 md:p-6"
                    >
                      <div className="flex flex-col gap-5 lg:flex-row lg:items-start lg:justify-between">
                        <div className="max-w-4xl">
                          <div className="flex flex-wrap items-center gap-3">
                            <h3 className="text-xl font-black tracking-tight text-slate-950">
                              {milestone.Description}
                            </h3>
                            <span
                              className={`inline-flex items-center gap-2 rounded-full border px-3 py-1 text-[10px] font-black uppercase tracking-[0.22em] ${meta.className}`}
                            >
                              <meta.Icon className="h-3.5 w-3.5" />
                              {meta.label}
                            </span>
                          </div>

                          <p className="mt-3 max-w-4xl text-sm leading-7 text-slate-600">
                            {milestone.WorkDescription ||
                           
                              "No additional milestone work description provided."}
                          </p>
                          {
                            milestone.Status == "SUBMITTED" ? (
                              <div className="mt-4">
                            <label
                              htmlFor={`feedback-${milestone.ID}`}
                              className="block text-sm font-medium text-red-700"
                            >
                              {feedBeckError }
                            </label>
                            <textarea
                              id={`feedback-${milestone.ID}`}
                              rows={3}
                              className="mt-1 block w-full rounded-md border border-slate-300 bg-slate-50 py-2 px-3 text-sm placeholder:text-slate-500 focus:outline-none focus:ring-2 focus:ring-blue-500"
                              placeholder="Enter your feedback for requesting changes..."
                              value={milestoneFeedbacks[milestone.ID] || ""}
                              onChange={(e) =>
                                setMilestoneFeedbacks({
                                  ...milestoneFeedbacks,
                                  [milestone.ID]: e.target.value,
                                })
                              }
                            />
                          </div>
                            ): null
                          }
                          

                          <div className="mt-5 grid gap-3 sm:grid-cols-3">
                            <div className="rounded-2xl bg-white p-4">
                              <p className="text-[10px] font-black uppercase tracking-[0.22em] text-slate-500">
                                Due date
                              </p>
                              <p className="mt-2 text-sm font-bold text-slate-900">
                                {formatDate(milestone.Due_date)}
                              </p>
                            </div>
                            <div className="rounded-2xl bg-white p-4">
                              <p className="text-[10px] font-black uppercase tracking-[0.22em] text-slate-500">
                                Amount
                              </p>
                              <p className="mt-2 text-sm font-bold text-slate-900">
                                {formatMoney(milestone.Amount)}
                              </p>
                            </div>
                            <div className="rounded-2xl bg-white p-4">
                              <p className="text-[10px] font-black uppercase tracking-[0.22em] text-slate-500">
                                Submission
                              </p>
                              {milestone.submission_url ? (
                                <a
                                  href={milestone.submission_url}
                                  target="_blank"
                                  rel="noreferrer"
                                  className="mt-2 inline-flex items-center gap-2 text-sm font-bold text-slate-950 underline decoration-slate-300 underline-offset-4 transition hover:text-slate-700"
                                >
                                  View submitted file
                                  <ExternalLink className="h-4 w-4" />
                                </a>
                              ) : (
                                <p className="mt-2 text-sm font-semibold text-slate-500">
                                  Awaiting submission
                                </p>
                              )}
                            </div>
                          </div>
                        </div>

                        <div className="flex shrink-0 flex-col gap-3 lg:items-end">
                          <button
                            type="button"
                            onClick={() => void handleRequestChanges(milestone)}
                            disabled={
                             milestone.Status !== "SUBMITTED" || requestingMilestoneId === milestone.ID
                            }
                            className="inline-flex items-center justify-center gap-2 rounded-2xl border border-slate-200 bg-white px-5 py-3 text-sm font-bold text-slate-700 transition hover:bg-slate-100 disabled:cursor-not-allowed disabled:opacity-50"
                          >
                            {requestingMilestoneId === milestone.ID ? (
                              <Loader2 className="h-4 w-4 animate-spin" />
                            ) : (
                              <RefreshCcw className="h-4 w-4" />
                            )}
                            Request changes
                          </button>

                          {milestone.submission_url ? (
                            <>
                            {/* <a
                              href={milestone.submission_url}
                              target="_blank"
                              rel="noreferrer"
                              className="inline-flex items-center justify-center gap-2 rounded-2xl bg-slate-950 px-5 py-3 text-sm font-bold text-white transition hover:bg-slate-800"
                            >
                              Review file
                              <FileText className="h-4 w-4" />
                            </a> */}
                            <button 
                            
                              className="inline-flex items-center justify-center gap-2 rounded-2xl bg-emerald-300 px-5 py-3 text-sm font-bold text-slate-950 transition hover:bg-emerald-200"
                              onClick={() => void handleUpdateContractStatus("APPROVED", milestone.ID)}
                            >

                              Approve and pay 
                            </button>
                            </>
                          ) : null}
                        </div>
                      </div>
                    </article>
                  );
                })
              ) : (
                <div className="rounded-[1.75rem] border border-dashed border-slate-200 bg-slate-50 p-10 text-center">
                  <p className="text-sm font-semibold text-slate-500">
                    No milestones are attached to this fixed contract.
                  </p>
                </div>
              )}
            </div>
          </section>
        )}
      </main>
    </div>
  );
}
