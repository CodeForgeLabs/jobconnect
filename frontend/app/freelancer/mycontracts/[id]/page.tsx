"use client";

import Image from "next/image";
import { useMemo } from "react";
import { useParams } from "next/navigation";
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
  useGetContractByIdQuery,
} from "@/api/contractapi";
import { useGetJobByIdQuery } from "@/api/jobsapi";

type MilestoneStatusMeta = {
  label: string;
  Icon: LucideIcon;
  badgeClassName: string;
};

const formatMoney = (value: number) =>
  new Intl.NumberFormat("en-US", {
    style: "currency",
    currency: "ETB",
    minimumFractionDigits: 2,
  }).format(value);

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

export default function FreelancerContractDetailPage() {
  const params = useParams<{ id: string }>();
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

  return (
    <>
      

      <main className="mx-auto max-w-screen-2xl space-y-12 px-8 pb-24 pt-12">
        <header className="flex flex-col items-start justify-between gap-6 md:flex-row md:items-end">
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
                <p className="text-sm font-label font-bold uppercase tracking-wider text-on-surface-variant">
                  Client
                </p>
                <p className="text-lg font-headline font-bold text-primary">
                  {contract.client_first_name} {contract.client_last_name}
                </p>
              </div>
            </div>
          </div>
          <div className="flex gap-4">
            <button className="flex items-center gap-2 rounded-full bg-surface-container-highest px-8 py-4 font-bold text-primary transition-all duration-300 hover:bg-primary-container hover:text-white active:scale-[0.99] active:opacity-80">
              <MessageCircle className="h-4 w-4" />
              Message Client
            </button>
            <button className="premium-gradient flex items-center gap-2 rounded-full px-10 py-4 font-bold text-white shadow-xl shadow-primary/20 transition-all duration-300 hover:scale-[1.02] active:scale-[0.98]">
              <span
                className="material-symbols-outlined"
                style={{ fontVariationSettings: "'FILL' 1" }}
              >
                send
              </span>
              {isHourly ? "Track Time" : "Submit Work"}
            </button>
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
                    {formatMoney(contract.total_budget)}
                  </p>
                </div>
                <div className="grid grid-cols-2 gap-4">
                  <div className="rounded-md bg-surface-container-low p-4">
                    <p className="text-xs font-medium text-on-surface-variant">
                      Paid
                    </p>
                    <p className="text-xl font-headline font-bold text-primary">
                      {formatMoney(paidAmount)}
                    </p>
                  </div>
                  <div className="rounded-md bg-tertiary-fixed p-4">
                    <p className="text-xs font-medium text-on-tertiary-fixed-variant">
                      Remaining
                    </p>
                    <p className="text-xl font-headline font-bold text-on-tertiary-fixed-variant">
                      {formatMoney(remainingAmount)}
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
            <button className="text-primary flex items-center gap-2 font-bold transition-transform hover:translate-x-1 active:scale-[0.99] active:opacity-80">
              View History{" "}
              <span className="material-symbols-outlined text-sm">
                arrow_forward
              </span>
            </button>
          </div>

          {isHourly ? (
            <div className="rounded-lg bg-surface-container-lowest p-8 shadow-[0_8px_40px_rgb(13,28,46,0.03)]">
              <div className="grid grid-cols-1 gap-6 md:grid-cols-3">
                <div className="rounded-md bg-surface-container-low p-6">
                  <p className="text-xs font-bold uppercase tracking-widest text-on-surface-variant">
                    Max Weekly Hours
                  </p>
                  <p className="mt-2 text-3xl font-black text-on-surface">
                    {isJobLoading ? "Loading..." : maxWeeklyHours}
                  </p>
                </div>
                <div className="rounded-md bg-surface-container-low p-6">
                  <p className="text-xs font-bold uppercase tracking-widest text-on-surface-variant">
                    Hourly Rate
                  </p>
                  <p className="mt-2 text-3xl font-black text-on-surface">
                    {formatMoney(contract.hourly_rate)}
                  </p>
                </div>
                <div className="rounded-md bg-surface-container-low p-6">
                  <p className="text-xs font-bold uppercase tracking-widest text-on-surface-variant">
                    Contract ID
                  </p>
                  <p className="mt-2 text-3xl font-black text-on-surface">
                    #{contract.contract_id}
                  </p>
                </div>
              </div>
              <p className="mt-6 rounded-md bg-surface-container-low p-4 text-sm text-on-surface-variant">
                This is an hourly contract. Milestones are not used. Track work
                sessions and weekly cap instead.
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
                              Due {formatDate(milestone.Due_date)}
                            </p>
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
                            <button className="text-primary text-sm font-bold hover:underline active:opacity-80">
                              Submit Milestone
                            </button>
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
            <div>
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
            </div>
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

      <footer className="w-full bg-surface-container py-16 text-on-surface">
        <div className="mx-auto flex max-w-screen-2xl flex-col items-start justify-between gap-8 px-12 md:flex-row md:items-center">
          <div className="flex flex-col gap-4">
            <span className="text-lg font-display font-bold text-primary">
              JobConnect
            </span>
            <p className="text-sm font-body text-on-surface-variant">
              © 2024 JobConnect. Architecting the future of work.
            </p>
          </div>
          <div className="flex flex-wrap gap-8">
            <a
              className="text-sm font-label font-bold text-on-surface-variant transition-all duration-200 hover:translate-x-1 hover:text-primary"
              href="#"
            >
              Privacy Policy
            </a>
            <a
              className="text-sm font-label font-bold text-on-surface-variant transition-all duration-200 hover:translate-x-1 hover:text-primary"
              href="#"
            >
              Terms of Service
            </a>
            <a
              className="text-sm font-label font-bold text-on-surface-variant transition-all duration-200 hover:translate-x-1 hover:text-primary"
              href="#"
            >
              Help Center
            </a>
            <a
              className="text-sm font-label font-bold text-on-surface-variant transition-all duration-200 hover:translate-x-1 hover:text-primary"
              href="#"
            >
              Career Advice
            </a>
          </div>
        </div>
      </footer>
    </>
  );
}
