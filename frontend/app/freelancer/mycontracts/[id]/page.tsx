"use client";

import Image from "next/image";
import {
  useCallback,
  useEffect,
  useMemo,
  useRef,
  useState,
  type FormEvent,
} from "react";
import { useParams, useRouter } from "next/navigation";
import {
  BadgeAlert,
  BadgeCheck,
  Clock3,
  MessageCircle,
  RefreshCcw,
  Upload,
  Wallet,
  type LucideIcon,
} from "lucide-react";
import {
  type ContractMilestone,
  useEndWorkSessionMutation,
  useGetContractByIdQuery,
  useGetWeeklyHoursMutation,
  useGetWorkSessionTimeElapsedMutation,
  useGetWorkSessionTimeLogsMutation,
  useSubmitMilestoneWorkMutation,
  useStartWorkSessionMutation,
  type WorkSessionTimeLogsResponse,
} from "@/api/contractapi";
import { useGetJobByIdQuery } from "@/api/jobsapi";
import { useUploadFileMutation } from "@/api/userapi";

type MilestoneStatusMeta = {
  label: string;
  Icon: LucideIcon;
  badgeClassName: string;
};

type TimeLogEntry = {
  id: string;
  description: string;
  startedAt: string;
  endedAt: string;
  durationSeconds: number;
};

const formatMoney = (value: number) =>
  new Intl.NumberFormat("en-US", {
    style: "currency",
    currency: "ETB",
    minimumFractionDigits: 2,
  }).format(value);

const formatDate = (value?: string | Date) => {
  if (!value) return "N/A";

  const parsed = value instanceof Date ? value : new Date(String(value));
  if (Number.isNaN(parsed.getTime())) return String(value);

  return parsed.toLocaleDateString("en-US", {
    month: "short",
    day: "numeric",
    year: "numeric",
  });
};

const getMilestoneStatusMeta = (status?: string): MilestoneStatusMeta => {
  const normalized = (status ?? "PENDING").toUpperCase();

  if (normalized === "APPROVED") {
    return {
      label: "Approved",
      Icon: BadgeCheck,
      badgeClassName: "bg-emerald-100 text-emerald-700",
    };
  }

  if (normalized === "PAID") {
    return {
      label: "Paid",
      Icon: Wallet,
      badgeClassName: "bg-sky-100 text-sky-700",
    };
  }

  if (normalized === "IN_PROGRESS") {
    return {
      label: "In Progress",
      Icon: Clock3,
      badgeClassName: "bg-amber-100 text-amber-700",
    };
  }

  if (normalized === "SUBMITTED") {
    return {
      label: "Submitted",
      Icon: Upload,
      badgeClassName: "bg-blue-100 text-blue-700",
    };
  }

  if (normalized === "REVISION_REQUESTED") {
    return {
      label: "Revision Requested",
      Icon: RefreshCcw,
      badgeClassName: "bg-rose-100 text-rose-700",
    };
  }

  return {
    label: "Pending",
    Icon: BadgeAlert,
    badgeClassName: "bg-slate-100 text-slate-700",
  };
};

const calculatePaidAmount = (milestones: ContractMilestone[]) =>
  milestones
    .filter((milestone) => {
      const normalized = (milestone.Status ?? "").toUpperCase();
      return normalized === "PAID" || normalized === "APPROVED";
    })
    .reduce((sum, milestone) => sum + Number(milestone.Amount || 0), 0);

const formatDuration = (totalSeconds: number) => {
  console.log("Formatting duration for seconds:", totalSeconds);
  const hours = Math.floor(totalSeconds / 3600);
  const minutes = Math.floor((totalSeconds % 3600) / 60);
  const seconds = Math.floor(totalSeconds % 60);

  return `${String(hours).padStart(2, "0")}:${String(minutes).padStart(2, "0")}:${String(seconds).slice(0, 2)}`;
};

const normalizeElapsedSeconds = (response: unknown) => {
  if (typeof response === "number") return response;

  if (!response || typeof response !== "object") return 0;

  const candidate = response as Record<string, unknown>;
  const rawValue =
    candidate.elapsed_seconds ??
    candidate.elapsedSeconds ??
    candidate.seconds ??
    candidate.duration ??
    candidate.time_elapsed ??
    candidate.timeElapsed;

  return typeof rawValue === "number" ? rawValue : Number(rawValue || 0);
};

const normalizeWeeklyHours = (response: unknown) => {
  if (typeof response === "number") return response;

  if (!response || typeof response !== "object") return 0;

  const candidate = response as Record<string, unknown>;
  const rawValue =
    candidate.weekly_hours ??
    candidate.weeklyHours ??
    candidate.hours ??
    candidate.total_hours ??
    candidate.totalHours;

  return typeof rawValue === "number" ? rawValue : Number(rawValue || 0);
};

const normalizeTimeLogs = (response: unknown): TimeLogEntry[] => {
  if (Array.isArray(response)) {
    return response.map((item, index) => {
      if (!item || typeof item !== "object") {
        return {
          id: String(index),
          description: "Work log",
          startedAt: "",
          endedAt: "",
          durationSeconds: 0,
        };
      }

      const log = item as Record<string, unknown>;

      const durationSeconds = (() => {
        if (typeof log.duration_seconds === "number")
          return log.duration_seconds;
        if (typeof log.durationSeconds === "number") return log.durationSeconds;

        if (typeof log.TotalHours === "number")
          return Math.round(log.TotalHours * 3600);
        if (typeof log.total_hours === "number")
          return Math.round(log.total_hours * 3600);
        if (typeof log.totalHours === "number")
          return Math.round(log.totalHours * 3600);
        if (
          typeof log.TotalHours === "string" &&
          !Number.isNaN(Number(log.TotalHours))
        )
          return Math.round(Number(log.TotalHours) * 3600);

        const startRaw =
          log.start_time ??
          log.startTime ??
          log.starttime ??
          log.started_at ??
          log.startedAt;
        const endRaw =
          log.end_time ??
          log.endTime ??
          log.endtime ??
          log.ended_at ??
          log.endedAt;

        if (startRaw && endRaw) {
          const startMs = Date.parse(String(startRaw));
          const endMs = Date.parse(String(endRaw));
          if (!Number.isNaN(startMs) && !Number.isNaN(endMs)) {
            const diff = Math.floor((endMs - startMs) / 1000);
            return diff > 0 ? diff : 0;
          }
        }

        return 0;
      })();

      return {
        id: String(log.id ?? log.ID ?? index),
        description: String(
          log.description ?? log.workDescription ?? "Work log",
        ),
        startedAt: String(
          log.started_at ??
            log.startedAt ??
            log.start_time ??
            log.startTime ??
            "",
        ),
        endedAt: String(
          log.ended_at ?? log.endedAt ?? log.end_time ?? log.endTime ?? "",
        ),
        durationSeconds: durationSeconds,
      };
    });
  }

  if (!response || typeof response !== "object") return [];

  const candidate = response as Record<string, unknown>;
  const maybeLogs =
    candidate.logs ??
    candidate.time_logs ??
    candidate.timeLogs ??
    candidate.data;
  if (Array.isArray(maybeLogs)) {
    return normalizeTimeLogs(maybeLogs);
  }

  return [];
};

const formatWeeklyHours = (hours: number) => {
  if (!Number.isFinite(hours)) return "0h 0m";

  const wholeHours = Math.floor(hours);
  const minutes = Math.round((hours - wholeHours) * 60);

  return `${wholeHours}h ${minutes}m`;
};

export default function FreelancerContractDetailPage() {
  const router = useRouter();
  const params = useParams<{ id: string }>();
  const contractId = Number(params?.id);
  const isValidId = Number.isFinite(contractId) && contractId > 0;

  const [isTracking, setIsTracking] = useState(false);
  const [trackingStartedAt, setTrackingStartedAt] = useState<number | null>(
    null,
  );
  const [elapsedBaselineSeconds, setElapsedBaselineSeconds] = useState(0);
  const [currentTimeMs, setCurrentTimeMs] = useState(() => Date.now());
  const [weeklyHours, setWeeklyHours] = useState(0);
  const [timeLogs, setTimeLogs] = useState<TimeLogEntry[]>([]);
  const [sessionMessage, setSessionMessage] = useState<string | null>(null);
  const [paidHourlySoFar, setPaidHourlySoFar] = useState(0);
  const [notPaidHourly, setNotPaidHourly] = useState(0);

  const [startWorkSession, { isLoading: isStartingSession }] =
    useStartWorkSessionMutation();
  const [endWorkSession, { isLoading: isEndingSession }] =
    useEndWorkSessionMutation();
  const [fetchElapsed] = useGetWorkSessionTimeElapsedMutation();
  const [fetchTimeLogs] = useGetWorkSessionTimeLogsMutation();
  const [fetchWeeklyHours] = useGetWeeklyHoursMutation();

  const [submissionDescription, setSubmissionDescription] = useState("");
  const [submissionProjectUrl, setSubmissionProjectUrl] = useState("");
  const [submissionFile, setSubmissionFile] = useState<File | null>(null);
  const [isUploadingFile, setIsUploadingFile] = useState(false);
  const [isSubmitModalOpen, setIsSubmitModalOpen] = useState(false);
  const [selectedMilestone, setSelectedMilestone] =
    useState<ContractMilestone | null>(null);
  const [submitError, setSubmitError] = useState<string | null>(null);

  const [expandedFeedbackIds, setExpandedFeedbackIds] = useState<
    Record<number, boolean>
  >({});

  const submitDialogRef = useRef<HTMLDialogElement | null>(null);

  const [submitMilestoneWork, { isLoading: isSubmittingMilestone }] =
    useSubmitMilestoneWorkMutation();
  const [uploadFile] = useUploadFileMutation();

  const handleOpenSubmitModal = (milestone: ContractMilestone) => {
    setSelectedMilestone(milestone);
    setSubmitError(null);
    setIsSubmitModalOpen(true);
    submitDialogRef.current?.showModal();
  };

  const handleCloseSubmitModal = () => {
    setIsSubmitModalOpen(false);
    submitDialogRef.current?.close();

    setSelectedMilestone(null);
    setSubmissionDescription("");
    setSubmissionProjectUrl("");
    setSubmissionFile(null);
    setSubmitError(null);
    setIsUploadingFile(false);
  };

  const toggleFeedback = (id: number) =>
    setExpandedFeedbackIds((prev) => ({ ...prev, [id]: !prev[id] }));

  const previewFeedback = (text: string, maxLen = 120) => {
    if (!text) return "";
    const firstLine = String(text).split("\n")[0] ?? "";
    if (firstLine.length <= maxLen) return firstLine;
    return `${firstLine.slice(0, maxLen).trim()}...`;
  };

  const handleSubmitMilestone = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();

    if (!contract || !selectedMilestone) return;

    setSubmitError(null);

    try {
      let milestoneProjectUrl = submissionProjectUrl.trim();

      if (submissionFile) {
        setIsUploadingFile(true);
        const uploadResult = await uploadFile(submissionFile).unwrap();
        milestoneProjectUrl = uploadResult.secure_url;
      }

      if (!milestoneProjectUrl) {
        setSubmitError(
          "Add a project link or upload a file before submitting.",
        );
        return;
      }

      await submitMilestoneWork({
        contract_id: contract.contract_id,
        milestone_id: selectedMilestone.ID,
        description: submissionDescription.trim(),
        milestone_project_url: milestoneProjectUrl,
      }).unwrap();

      await refetch();
      handleCloseSubmitModal();
    } catch {
      setSubmitError("Unable to submit the milestone right now.");
    } finally {
      setIsUploadingFile(false);
    }
  };

  const {
    data: contract,
    isLoading,
    isError,
    refetch,
  } = useGetContractByIdQuery(contractId, {
    skip: !isValidId,
  });

  const { data: jobData, isLoading: isJobLoading } = useGetJobByIdQuery(
    contract?.job_id ?? 0,
    { skip: !contract?.job_id },
  );

  const contractType = useMemo(() => {
    if (!contract) return "FIXED";
    return (contract.type ?? jobData?.job?.job_type ?? "FIXED").toUpperCase();
  }, [contract, jobData?.job?.job_type]);

  const isHourly = contractType === "HOURLY";
  const milestones = contract?.milestones ?? [];
  const paidAmount = calculatePaidAmount(milestones);
  const remainingAmount = Math.max(
    (contract?.total_budget ?? 0) - paidAmount,
    0,
  );
  const maxWeeklyHours =
    jobData?.job?.max_weekly_hours ?? contract?.weekly_hour_limit ?? 0;

  const syncHourlySessionState = useCallback(async () => {
    if (!contract || !isHourly) return;

    const requestBody = { contract_id: contract.contract_id };

    const [elapsedResult, logsResult, weeklyResult] = await Promise.all([
      fetchElapsed(requestBody)
        .unwrap()
        .catch(() => 0),
      fetchTimeLogs(requestBody)
        .unwrap()
        .catch((): WorkSessionTimeLogsResponse => ({ time_logs: [] })),
      fetchWeeklyHours(requestBody)
        .unwrap()
        .catch(() => 0),
    ]);

    const elapsedFromServer = normalizeElapsedSeconds(elapsedResult);
    const weeklyHoursFromServer = normalizeWeeklyHours(weeklyResult);

    let hasActiveSession = false;

    for (const log of logsResult.time_logs) {
      if (log.end_time === null || log.end_time === undefined) {
        hasActiveSession = true;

        const sessionStartTime = new Date(log.start_time).getTime();

        setIsTracking(true);
        setTrackingStartedAt(sessionStartTime);

        break;
      }
    }
    let paidHourly = 0;
    let notPaidHourly = 0;

    for (const log of logsResult.time_logs) {
      if (log.IsPaid) {
        paidHourly += log.TotalHours * (contract.hourly_rate || 0);
      } else {
        notPaidHourly += log.TotalHours * (contract.hourly_rate || 0);
      }
    }
    setPaidHourlySoFar(paidHourly);
    setNotPaidHourly(notPaidHourly);

    if (!hasActiveSession) {
      setIsTracking(false);
      setTrackingStartedAt(null);
    }

    setElapsedBaselineSeconds(elapsedFromServer);
    setWeeklyHours(weeklyHoursFromServer);
    setTimeLogs(normalizeTimeLogs(logsResult));
  }, [contract, fetchElapsed, fetchTimeLogs, fetchWeeklyHours, isHourly]);

  useEffect(() => {
    if (!isHourly || !contract) return;

    const syncTimer = window.setTimeout(() => {
      void syncHourlySessionState();
    }, 0);

    return () => window.clearTimeout(syncTimer);
  }, [contract, isHourly, syncHourlySessionState]);

  useEffect(() => {
    if (!isTracking || trackingStartedAt === null) return undefined;

    const timerId = window.setInterval(() => {
      setCurrentTimeMs(Date.now());
    }, 1000);

    return () => window.clearInterval(timerId);
  }, [isTracking, trackingStartedAt]);

  const elapsedSeconds = useMemo(() => {
    if (!isTracking || trackingStartedAt === null) {
      return elapsedBaselineSeconds;
    }

    return Math.floor((currentTimeMs - trackingStartedAt) / 1000);
  }, [currentTimeMs, elapsedBaselineSeconds, isTracking, trackingStartedAt]);

  const handleStartSession = async () => {
    if (!contract || !isHourly) return;

    setSessionMessage(null);

    try {
      await startWorkSession({ contract_id: contract.contract_id }).unwrap();
      setIsTracking(true);
      setTrackingStartedAt(Date.now());
      setElapsedBaselineSeconds(0);
      await syncHourlySessionState();
      setSessionMessage("Work session started.");
    } catch {
      setSessionMessage("Unable to start the work session.");
    }
  };

  const handleEndSession = async () => {
    if (!contract || !isHourly) return;

    setSessionMessage(null);

    try {
      await endWorkSession({ contract_id: contract.contract_id }).unwrap();
      setIsTracking(false);
      setTrackingStartedAt(null);
      await syncHourlySessionState();
      setSessionMessage("Work session ended.");
    } catch {
      setSessionMessage("Unable to end the work session.");
    }
  };

  if (!isValidId) {
    return (
      <div className="mx-auto mt-16 max-w-4xl rounded-xl border border-red-200 bg-red-50 p-8 text-red-700">
        <h1 className="text-xl font-bold">Invalid contract ID</h1>
        <p className="mt-2">The contract path parameter is invalid.</p>
      </div>
    );
  }

  if (isLoading) {
    return (
      <div className="mx-auto mt-16 grid max-w-7xl grid-cols-1 gap-6 px-8 pb-16 md:grid-cols-3">
        <div className="h-56 animate-pulse rounded-xl bg-surface" />
        <div className="h-56 animate-pulse rounded-xl bg-surface md:col-span-2" />
        <div className="h-96 animate-pulse rounded-xl bg-surface md:col-span-3" />
      </div>
    );
  }

  if (isError || !contract) {
    return (
      <div className="mx-auto mt-16 max-w-4xl rounded-xl border border-red-200 bg-red-50 p-8 text-red-700">
        <h1 className="text-xl font-bold">Unable to load contract</h1>
        <p className="mt-2">Please try again.</p>
        <button
          type="button"
          onClick={() => refetch()}
          className="mt-4 rounded-lg bg-red-600 px-4 py-2 text-sm font-bold text-white"
        >
          Retry
        </button>
      </div>
    );
  }

  const activeSinceLabel = contract.start_date
    ? `Active since ${formatDate(contract.start_date)}`
    : "Contract active";

  const nextSubmittableMilestone = milestones.find((milestone) => {
    const status = (milestone.Status ?? "").toUpperCase();

    return (
      status === "PENDING" ||
      status === "IN_PROGRESS" ||
      status === "REVISION_REQUESTED"
    );
  });

  return (
    <>
      <main className="mx-auto max-w-screen-2xl space-y-12 px-8 pb-24 pt-12">
        <header className="flex px-6 flex-col items-start justify-between gap-6 md:flex-row md:items-end">
          <div className="max-w-3xl">
            <div className="mb-4 flex items-center gap-3">
              <span className="rounded-full bg-tertiary-fixed px-4 py-1 text-xs font-bold uppercase tracking-wide text-on-tertiary-fixed-variant">
                {isHourly ? "Hourly Contract" : "Fixed Price Contract"}
              </span>
              <span className="font-medium text-on-surface-variant/60">•</span>
              <span className="text-sm font-medium text-on-surface-variant">
                {activeSinceLabel}
              </span>
            </div>
            <h1 className="text-4xl font-display font-extrabold leading-tight tracking-tight text-on-surface md:text-5xl">
              {contract.title || contract.job_title}
            </h1>
            <div className="mt-6 flex items-center gap-4">
              {contract.client_profile_picture_url ? (
                <Image
                  alt={`${contract.client_first_name} ${contract.client_last_name}`}
                  src={contract.client_profile_picture_url}
                  width={48}
                  height={48}
                  className="h-12 w-12 rounded-full ring-4 ring-surface-container object-cover"
                  unoptimized
                />
              ) : (
                <div className="flex h-12 w-12 items-center justify-center rounded-full ring-4 ring-surface-container bg-surface-container-high text-on-surface-variant">
                  C
                </div>
              )}
              <div>
                <p className="text-lg font-headline font-bold text-primary">
                  {contract.client_first_name} {contract.client_last_name}
                </p>
                <p className="text-sm font-label font-bold uppercase tracking-wider text-on-surface-variant">
                  {contract.client_email}
                </p>
              </div>
            </div>
          </div>
          <div className="flex gap-4">
            <button
              onClick={() => {
                router.push(`/messages?userid=${contract.client_id}`);
              }}
              className="flex items-center gap-2 rounded-full bg-surface-container-highest px-8 py-4 font-bold text-primary transition-all duration-300 hover:bg-primary-container hover:text-white active:scale-[0.99] active:opacity-80"
            >
              <MessageCircle className="h-4 w-4" />
              Message Client
            </button>
            {isHourly && (
              <button
                onClick={
                  isHourly
                    ? isTracking
                      ? handleEndSession
                      : handleStartSession
                    : undefined
                }
                className="premium-gradient flex items-center gap-2 rounded-full px-10 py-4 font-bold text-primary shadow-xl shadow-primary/20 transition-all duration-300 hover:scale-[1.02] active:scale-[0.98]"
              >
                {isHourly ? (isTracking ? "End Session" : "Start Session") : ""}
              </button>
            )}
          </div>
        </header>

        <div className="grid grid-cols-1 gap-8 md:grid-cols-12">
          <div className="md:col-span-4 flex flex-col justify-between rounded-lg bg-surface-container-lowest p-10 shadow-[0_8px_30px_rgb(13,28,46,0.02)]">
            <div>
              <h3 className="mb-8 text-sm font-label font-black uppercase tracking-[0.2em] text-on-surface-variant">
                Financial Overview
              </h3>
              <div className="space-y-6">
                <div>
                  <p className="text-sm font-medium text-on-surface-variant">
                    Total Budget
                  </p>
                  <p className="mt-1 text-4xl font-display font-black text-on-surface">
                    {formatMoney(
                      contract.total_budget || contract.hourly_rate
                        ? (contract.hourly_rate || 0) * (maxWeeklyHours || 0)
                        : 0,
                    )}
                  </p>
                </div>
                <div className="grid grid-cols-2 gap-4">
                  <div className="rounded-md bg-surface-container-low p-4">
                    <p className="text-xs font-medium text-on-surface-variant">
                      Paid
                    </p>
                    <p className="text-xl font-headline font-bold text-primary">
                      {formatMoney(paidAmount || Math.floor(paidHourlySoFar))}
                    </p>
                  </div>
                  <div className="rounded-md bg-tertiary-fixed p-4">
                    <p className="text-xs font-medium text-on-tertiary-fixed-variant">
                      Remaining
                    </p>
                    <p className="text-xl font-headline font-bold text-on-tertiary-fixed-variant">
                      {formatMoney(remainingAmount || Math.ceil(notPaidHourly))}
                    </p>
                  </div>
                </div>
              </div>
            </div>
            <div className="mt-8 border-t border-outline-variant/20 pt-8">
              <div className="flex items-center justify-between">
                <span className="text-sm font-medium text-on-surface-variant">
                  Contract Status
                </span>
                <span className="text-primary text-sm font-bold">
                  {contract.status}
                </span>
              </div>
            </div>
          </div>

          <div className="md:col-span-8 rounded-lg bg-surface-container-lowest p-10 shadow-[0_8px_30px_rgb(13,28,46,0.02)]">
            <h3 className="mb-8 text-sm font-label font-black uppercase tracking-[0.2em] text-on-surface-variant">
              Contract Details
            </h3>
            <div className="prose prose-slate max-w-none">
              <p className="text-lg leading-relaxed text-on-surface-variant">
                {contract.description ||
                  contract.proposal_description ||
                  "No description provided."}
              </p>
              <div className="mt-8 grid grid-cols-2 gap-8 md:grid-cols-4">
                <div>
                  <p className="mb-1 text-xs font-bold uppercase tracking-widest text-on-surface-variant">
                    Status
                  </p>
                  <p className="font-headline font-bold text-on-surface">
                    {contract.status}
                  </p>
                </div>
                <div>
                  <p className="mb-1 text-xs font-bold uppercase tracking-widest text-on-surface-variant">
                    Start Date
                  </p>
                  <p className="font-headline font-bold text-on-surface">
                    {formatDate(contract.start_date)}
                  </p>
                </div>
                <div>
                  <p className="mb-1 text-xs font-bold uppercase tracking-widest text-on-surface-variant">
                    Type
                  </p>
                  <p className="font-headline font-bold text-on-surface">
                    {isHourly ? "Hourly" : "Fixed Price"}
                  </p>
                </div>
                <div>
                  <p className="mb-1 text-xs font-bold uppercase tracking-widest text-on-surface-variant">
                    Due Date
                  </p>
                  <p className="font-headline font-bold text-on-surface">
                    {formatDate(contract.end_date)}
                  </p>
                </div>
              </div>
            </div>
          </div>
        </div>

        <section className="space-y-8">
          <div className="flex items-center justify-between">
            <h2 className="text-3xl font-display font-black tracking-tight text-on-surface">
              {isHourly ? "Hourly Work Session" : "Milestones & Payments"}
            </h2>
            <div className="mx-8 hidden h-0.5 grow bg-surface-container-high md:block" />
            {/* <button className="text-primary flex items-center gap-2 font-bold transition-transform hover:translate-x-1 active:scale-[0.99] active:opacity-80">
              View History{" "}
              <span className="material-symbols-outlined text-sm">
                arrow_forward
              </span>
            </button> */}
          </div>

          {isHourly ? (
            <div className="rounded-lg bg-surface-container-lowest p-8 shadow-[0_8px_40px_rgb(13,28,46,0.03)]">
              <div className="grid grid-cols-1 gap-6 lg:grid-cols-12">
                <div className="rounded-2xl bg-surface-container-low p-6 lg:col-span-4">
                  <p className="text-xs font-bold uppercase tracking-widest text-on-surface-variant">
                    Live Session
                  </p>
                  <div className="mt-3 flex items-end gap-3">
                    <p className="text-4xl font-black text-on-surface">
                      {formatDuration(elapsedSeconds)}
                    </p>
                    <span
                      className={`mb-1 rounded-full px-3 py-1 text-[10px] font-black uppercase tracking-widest ${isTracking ? "bg-emerald-100 text-emerald-700" : "bg-slate-100 text-slate-700"}`}
                    >
                      {isTracking ? "Running" : "Idle"}
                    </span>
                  </div>
                  <p className="mt-2 text-sm text-on-surface-variant">
                    {isTracking
                      ? "Your timer is running locally in the browser and syncing with the session endpoints."
                      : "Start a session to begin tracking elapsed time."}
                  </p>
                  <div className="mt-6 flex flex-wrap gap-3">
                    <button
                      type="button"
                      onClick={handleStartSession}
                      disabled={isStartingSession || isTracking}
                      className="rounded-full bg-primary px-5 py-3 text-sm font-bold text-white transition-all hover:bg-blue-700 disabled:cursor-not-allowed disabled:opacity-50"
                    >
                      Start Session
                    </button>
                    <button
                      type="button"
                      onClick={handleEndSession}
                      disabled={isEndingSession || !isTracking}
                      className="rounded-full bg-surface-container-high px-5 py-3 text-sm font-bold text-on-surface transition-all hover:bg-surface-container-highest disabled:cursor-not-allowed disabled:opacity-50"
                    >
                      End Session
                    </button>
                  </div>
                  {sessionMessage ? (
                    <p className="mt-4 text-sm font-medium text-on-surface-variant">
                      {sessionMessage}
                    </p>
                  ) : null}
                </div>

                <div className="rounded-2xl bg-surface-container-low p-6 lg:col-span-4">
                  <p className="text-xs font-bold uppercase tracking-widest text-on-surface-variant">
                    Weekly Hours
                  </p>
                  <p className="mt-3 text-4xl font-black text-on-surface">
                    {formatWeeklyHours(weeklyHours)}
                  </p>
                  <p className="mt-2 text-sm text-on-surface-variant">
                    Current job cap:{" "}
                    {isJobLoading
                      ? "Loading..."
                      : `${maxWeeklyHours} hrs / week`}
                  </p>
                  <div className="mt-6 grid grid-cols-2 gap-4">
                    <div className="rounded-xl bg-white p-4 shadow-sm">
                      <p className="text-[10px] font-bold uppercase tracking-wider text-slate-400">
                        Hourly Rate
                      </p>
                      <p className="mt-2 text-2xl font-black text-on-surface">
                        {formatMoney(contract.hourly_rate)}
                      </p>
                    </div>
                    <div className="rounded-xl bg-white p-4 shadow-sm">
                      <p className="text-[10px] font-bold uppercase tracking-wider text-slate-400">
                        Contract ID
                      </p>
                      <p className="mt-2 text-2xl font-black text-on-surface">
                        #{contract.contract_id}
                      </p>
                    </div>
                  </div>
                </div>

                <div className="rounded-2xl bg-surface-container-low p-6 lg:col-span-4">
                  <p className="text-xs font-bold uppercase tracking-widest text-on-surface-variant">
                    Time Logs
                  </p>
                  <div className="mt-4 space-y-3">
                    {timeLogs.length > 0 ? (
                      timeLogs.slice(0, 4).map((log) => (
                        <div
                          key={log.id}
                          className="rounded-xl bg-white p-4 shadow-sm"
                        >
                          <div className="flex items-center justify-between gap-3">
                            <p className="text-sm font-bold text-on-surface">
                              {log.description}
                            </p>
                            <span className="text-xs font-bold text-primary">
                              {formatDuration(log.durationSeconds)}
                            </span>
                          </div>
                          <p className="mt-2 text-xs text-on-surface-variant">
                            {log.startedAt
                              ? formatDate(log.startedAt)
                              : "Started: N/A"}
                            {log.endedAt
                              ? ` • Ended: ${formatDate(log.endedAt)}`
                              : " • In progress"}
                          </p>
                        </div>
                      ))
                    ) : (
                      <div className="rounded-xl bg-white p-4 text-sm text-on-surface-variant shadow-sm">
                        No time logs yet.
                      </div>
                    )}
                  </div>
                </div>
              </div>
              <p className="mt-6 rounded-md bg-surface-container-low p-4 text-sm text-on-surface-variant">
                This hourly contract uses browser-ticked elapsed time plus
                session endpoints for start, end, logs, and weekly hours.
              </p>
            </div>
          ) : (
            <div className="overflow-x-auto rounded-lg bg-surface-container-lowest shadow-[0_8px_40px_rgb(13,28,46,0.03)]">
              <table className="min-w-200 w-full border-collapse text-left">
                <thead>
                  <tr className="bg-surface-container-low">
                    <th className="px-8 py-6 text-xs font-label uppercase tracking-widest text-on-surface-variant">
                      #
                    </th>
                    <th className="px-8 py-6 text-xs font-label uppercase tracking-widest text-on-surface-variant">
                      Milestone Description
                    </th>
                    <th className="px-8 py-6 text-right text-xs font-label uppercase tracking-widest text-on-surface-variant">
                      Amount
                    </th>
                    <th className="px-8 py-6 text-center text-xs font-label uppercase tracking-widest text-on-surface-variant">
                      Status
                    </th>
                    <th className="px-8 py-6 text-right text-xs font-label uppercase tracking-widest text-on-surface-variant">
                      Action
                    </th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-surface-container">
                  {milestones.length > 0 ? (
                    milestones.map((milestone, index) => {
                      const statusMeta = getMilestoneStatusMeta(
                        milestone.Status,
                      );
                      const Icon = statusMeta.Icon;
                      const canSubmit =
                        nextSubmittableMilestone?.ID === milestone.ID;

                      return (
                        <tr
                          key={milestone.ID}
                          className="transition-colors hover:bg-surface-container-low/30"
                        >
                          <td className="px-8 py-8 font-label font-bold text-on-surface-variant">
                            {index + 1}
                          </td>
                          <td className="px-8 py-8">
                            <p className="font-headline font-bold text-on-surface">
                              {milestone.Description}
                            </p>
                            <p className="mt-1 text-xs text-on-surface-variant">
                              Due {formatDate(milestone.deadline)}
                            </p>
                            {milestone.submission_url ? (
                              <div className="mt-2 space-y-1">
                                {String(milestone.submission_url)
                                  .split(",")
                                  .map((u) => u.trim())
                                  .filter(Boolean)
                                  .map((url, idx) => (
                                    <a
                                      key={idx}
                                      href={url}
                                      target="_blank"
                                      rel="noopener noreferrer"
                                      className="text-sm text-primary hover:underline block"
                                    >
                                      Submitted file {idx + 1}
                                    </a>
                                  ))}
                              </div>
                            ) : null}

                            {milestone.ClientFeedback ? (
                              <div className="mt-2">
                                {(() => {
                                  const isExpanded =
                                    !!expandedFeedbackIds[milestone.ID];
                                  const showToggle =
                                    String(milestone.ClientFeedback).includes(
                                      "\n",
                                    ) ||
                                    String(milestone.ClientFeedback).length >
                                      120;

                                  return (
                                    <>
                                      <p className="text-sm text-on-surface-variant">
                                        {isExpanded
                                          ? milestone.ClientFeedback
                                          : previewFeedback(
                                              milestone.ClientFeedback,
                                            )}
                                      </p>
                                      {showToggle ? (
                                        <button
                                          type="button"
                                          onClick={() =>
                                            toggleFeedback(milestone.ID)
                                          }
                                          className="mt-1 text-primary text-sm font-bold hover:underline"
                                        >
                                          {isExpanded ? "Less" : "More"}
                                        </button>
                                      ) : null}
                                    </>
                                  );
                                })()}
                              </div>
                            ) : null}
                          </td>
                          <td className="px-8 py-8 text-right font-headline font-bold text-on-surface">
                            {formatMoney(milestone.Amount)}
                          </td>
                          <td className="px-8 py-8 text-center">
                            <span
                              className={`inline-flex items-center gap-1 rounded-full px-4 py-1.5 text-[10px] font-black uppercase tracking-widest ${statusMeta.badgeClassName}`}
                            >
                              <Icon className="h-3.5 w-3.5" />
                              {statusMeta.label}
                            </span>
                          </td>
                          <td className="px-8 py-8 text-right">
                            {canSubmit ? (
                              <button
                                type="button"
                                onClick={() => handleOpenSubmitModal(milestone)}
                                className="text-primary text-sm font-bold hover:underline active:opacity-80"
                              >
                                Submit Milestone
                              </button>
                            ) : (
                              <span className="text-sm text-on-surface-variant">
                                {milestone.Status === "APPROVED" ||
                                milestone.Status === "PAID"
                                  ? "No action needed"
                                  : milestone.Status === "REVISION_REQUESTED"
                                    ? "Revise and resubmit"
                                    : milestone.Status === "SUBMITTED"
                                      ? "Awaiting client review"
                                      : "Submit previous milestones first"}
                              </span>
                            )}
                          </td>
                        </tr>
                      );
                    })
                  ) : (
                    <tr>
                      <td
                        colSpan={5}
                        className="px-8 py-8 text-center text-on-surface-variant"
                      >
                        No milestones found for this contract.
                      </td>
                    </tr>
                  )}
                </tbody>
              </table>
            </div>
          )}
        </section>

        <div className="grid grid-cols-1 gap-12 md:grid-cols-12">
          <div className="md:col-span-8">
            <h3 className="mb-6 text-xl font-display font-black text-on-surface">
              Freelancer&apos;s Proposal
            </h3>
            <div className="relative overflow-hidden rounded-lg bg-surface-container-low p-10">
              <div className="absolute -mr-32 -mt-32 right-0 top-0 h-64 w-64 rounded-full bg-primary/5 blur-3xl" />
              <p className="whitespace-pre-line italic leading-relaxed text-on-surface font-body">
                {contract.proposal_description || "No proposal text provided."}
              </p>
            </div>
          </div>
          <div className="space-y-8 md:col-span-4">
            {/* <div>
              <h3 className="mb-6 text-sm font-label font-black uppercase tracking-[0.2em] text-on-surface-variant">
                Attachments
              </h3>
              <div className="space-y-3">
                <div className="group flex cursor-pointer items-center justify-between rounded-md border border-outline-variant/10 bg-surface-container-lowest p-4 transition-all hover:border-primary/30">
                  <div className="flex items-center gap-3">
                    <span className="material-symbols-outlined text-primary">
                      description
                    </span>
                    <span className="text-sm font-medium">
                      contract_scope.pdf
                    </span>
                  </div>
                  <span className="material-symbols-outlined text-sm opacity-0 transition-opacity group-hover:opacity-100">
                    download
                  </span>
                </div>
              </div>
            </div> */}
            <div className="rounded-lg bg-primary p-8 text-white">
              <p className="mb-2 text-xs font-bold uppercase tracking-widest opacity-60">
                Workspace Tip
              </p>
              <p className="text-sm leading-relaxed">
                Review work updates frequently to keep payments and approvals on
                track.
              </p>
            </div>
          </div>
        </div>
      </main>

      <dialog
        ref={submitDialogRef}
        open={isSubmitModalOpen}
        onClose={handleCloseSubmitModal}
        onCancel={(event) => {
          event.preventDefault();
          handleCloseSubmitModal();
        }}
        className="w-[min(92vw,44rem)] rounded-3xl border border-outline-variant/30 bg-surface-container-lowest p-0 shadow-2xl backdrop:bg-slate-950/60"
      >
        {selectedMilestone ? (
          <form onSubmit={handleSubmitMilestone} className="p-8 md:p-10">
            <div className="flex items-start justify-between gap-6">
              <div>
                <p className="text-xs font-bold uppercase tracking-[0.2em] text-on-surface-variant">
                  Submit Milestone
                </p>
                <h3 className="mt-2 text-2xl font-display font-black text-on-surface">
                  {selectedMilestone.Description}
                </h3>
                <p className="mt-2 text-sm text-on-surface-variant">
                  Due {formatDate(selectedMilestone.deadline)} ·{" "}
                  {formatMoney(selectedMilestone.Amount)}
                </p>
              </div>
              <button
                type="button"
                onClick={handleCloseSubmitModal}
                className="rounded-full bg-surface-container-high px-3 py-2 text-sm font-bold text-on-surface transition-colors hover:bg-surface-container-highest"
              >
                Close
              </button>
            </div>

            <div className="mt-8 space-y-6">
              <label className="block space-y-2">
                <span className="text-sm font-bold text-on-surface">Notes</span>
                <textarea
                  value={submissionDescription}
                  onChange={(event) =>
                    setSubmissionDescription(event.target.value)
                  }
                  rows={5}
                  className="w-full rounded-2xl border border-outline-variant/20 bg-surface-container-low px-4 py-3 text-sm text-on-surface outline-none transition focus:border-primary"
                  placeholder="Add a short summary of the work you completed"
                />
              </label>

              <label className="block space-y-2">
                <span className="text-sm font-bold text-on-surface">
                  Project Link
                </span>
                <input
                  type="url"
                  value={submissionProjectUrl}
                  onChange={(event) =>
                    setSubmissionProjectUrl(event.target.value)
                  }
                  className="w-full rounded-2xl border border-outline-variant/20 bg-surface-container-low px-4 py-3 text-sm text-on-surface outline-none transition focus:border-primary"
                  placeholder="https://..."
                />
              </label>

              <label className="block space-y-2">
                <span className="text-sm font-bold text-on-surface">
                  Upload File
                </span>
                <input
                  type="file"
                  onChange={(event) =>
                    setSubmissionFile(event.target.files?.[0] ?? null)
                  }
                  className="w-full rounded-2xl border border-dashed border-outline-variant/30 bg-surface-container-low px-4 py-3 text-sm text-on-surface file:mr-4 file:rounded-full file:border-0 file:bg-primary file:px-4 file:py-2 file:text-sm file:font-bold file:text-white"
                />
                <p className="text-xs text-on-surface-variant">
                  If you upload a file, it will be uploaded first and the
                  resulting link will be submitted.
                </p>
              </label>

              {submitError ? (
                <div className="rounded-2xl bg-rose-50 px-4 py-3 text-sm text-rose-700">
                  {submitError}
                </div>
              ) : null}
            </div>

            <div className="mt-8 flex flex-col-reverse gap-3 sm:flex-row sm:justify-end">
              <button
                type="button"
                onClick={handleCloseSubmitModal}
                className="rounded-full bg-surface-container-high px-6 py-3 text-sm font-bold text-on-surface transition-colors hover:bg-surface-container-highest"
              >
                Cancel
              </button>
              <button
                type="submit"
                disabled={isSubmittingMilestone || isUploadingFile}
                className="rounded-full bg-primary px-6 py-3 text-sm font-bold text-white transition-all hover:bg-blue-700 disabled:cursor-not-allowed disabled:opacity-60"
              >
                {isSubmittingMilestone || isUploadingFile
                  ? "Submitting..."
                  : "Submit Milestone"}
              </button>
            </div>
          </form>
        ) : null}
      </dialog>
    </>
  );
}
