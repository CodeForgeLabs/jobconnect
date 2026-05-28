"use client";

import Image from "next/image";
import { useRouter } from "next/navigation";
import { useMemo, useState } from "react";
import {
  BadgeAlert,
  BadgeCheck,
  Clock3,
  MessageCircle,
  RefreshCcw,
  Receipt,
  Search,
  Upload,
  Wallet,
  type LucideIcon,
} from "lucide-react";
import { type Contract, useGetMyContractsQuery ,useStartWorkSessionMutation , useEndWorkSessionMutation} from "@/api/contractapi";
import { useGetJobByIdQuery } from "@/api/jobsapi";

type ContractFilter = "ALL" | "ACTIVE" | "COMPLETED" | "CANCELLED";

type MilestoneStatusMeta = {
  label: string;
  Icon: LucideIcon;
  className: string;
  iconClassName: string;
};

const formatMoney = (value: number) =>
  new Intl.NumberFormat("en-US", {
    style: "currency",
    currency: "ETB",
    minimumFractionDigits: 2,
  }).format(value);

const formatDate = (value?: string) => {
  if (!value) return "N/A";

  const parsedDate = new Date(value);
  return Number.isNaN(parsedDate.getTime())
    ? value
    : parsedDate.toLocaleDateString("en-US", {
        month: "short",
        day: "numeric",
        year: "numeric",
      });
};

const getMilestoneStatusMeta = (status?: string): MilestoneStatusMeta => {
  const normalizedStatus = (status ?? "PENDING").toUpperCase();

  if (normalizedStatus === "APPROVED") {
    return {
      label: "Approved",
      Icon: BadgeCheck,
      className: "bg-emerald-50 text-emerald-700",
      iconClassName: "text-emerald-600",
    };
  }

  if (normalizedStatus === "PAID") {
    return {
      label: "Paid",
      Icon: Wallet,
      className: "bg-sky-50 text-sky-700",
      iconClassName: "text-sky-600",
    };
  }

  if (normalizedStatus === "IN_PROGRESS") {
    return {
      label: "In Progress",
      Icon: Clock3,
      className: "bg-amber-50 text-amber-700",
      iconClassName: "text-amber-600",
    };
  }

  if (normalizedStatus === "SUBMITTED") {
    return {
      label: "Submitted",
      Icon: Upload,
      className: "bg-blue-50 text-blue-700",
      iconClassName: "text-blue-600",
    };
  }

  if (normalizedStatus === "REVISION_REQUESTED") {
    return {
      label: "Revision Requested",
      Icon: RefreshCcw,
      className: "bg-rose-50 text-rose-700",
      iconClassName: "text-rose-600",
    };
  }

  return {
    label: "Pending",
    Icon: BadgeAlert,
    className: "bg-gray-300 text-slate-600",
    iconClassName: "text-slate-500",
  };
};

const getProgressPercent = (milestones?: { Status?: string }[]) => {
  if (!milestones?.length) return 0;

  const completedMilestones = milestones.filter((milestone) => {
    const normalizedStatus = (milestone.Status ?? "").toUpperCase();
    return normalizedStatus === "APPROVED" || normalizedStatus === "PAID";
  }).length;

  return Math.round((completedMilestones / milestones.length) * 100);
};

function ContractCard({ contract }: { contract: Contract }) {
  const { data: jobData, isLoading: isJobLoading } = useGetJobByIdQuery(
    contract.job_id,
  );

  const job = jobData?.job;
  const contractType = (
    contract.type ??
    job?.job_type ??
    "FIXED"
  ).toUpperCase();
  const isHourly = contractType === "HOURLY";
  const maxWeeklyHours =
    job?.max_weekly_hours ?? contract.weekly_hour_limit ?? 0;
  const progressPercent = getProgressPercent(contract.milestones);
  const milestones = (contract.milestones ?? []).slice(0, 4);
  const cardTitle = contract.title || contract.job_title;
  const router = useRouter();

  return (
    <div 
    onClick={()=> router.push(`/freelancer/mycontracts/${contract.contract_id}`)}
    className="group flex flex-col gap-6 rounded-2xl bg-surface p-8 shadow-xl transition-shadow hover:shadow-2xl">
      <div className="flex items-start justify-between gap-4">
        <div className="flex items-center gap-4">
          <div className="flex h-14 w-14 items-center justify-center overflow-hidden rounded-xl ring-2 ring-primary ring-offset-2">
            {contract.client_profile_picture_url ? (
              <Image
                alt={`${contract.client_first_name} ${contract.client_last_name}`}
                className="h-full w-full object-cover"
                src={contract.client_profile_picture_url}
                width={56}
                height={56}
                unoptimized
              />
            ) : (
              <span className="material-symbols-outlined text-3xl text-primary">
                apartment
              </span>
            )}
          </div>
          <div>
            <h3 className="text-2xl font-bold leading-tight text-on-surface">
              {cardTitle}
            </h3>
            <p className="font-medium text-on-surface-variant">
              {contract.client_first_name} {contract.client_last_name}
            </p>
          </div>
        </div>
        <div className="flex flex-col items-end gap-2">
          <span className="rounded-full bg-primary-container px-4 py-1.5 text-xs font-bold uppercase tracking-widest text-primary">
            {contract.status}
          </span>
          <span className="rounded-full bg-surface-container-low px-3 py-1 text-[10px] font-bold uppercase tracking-wider text-on-surface-variant">
            {contractType}
          </span>
        </div>
      </div>

      {isHourly ? (
        <div 
        
        className="rounded-2xl border border-outline-variant/50 bg-surface-container-low p-5">
          <div className="flex items-center justify-between gap-3">
            <div>
              <p className="text-xs font-bold uppercase tracking-wider text-on-surface-variant">
                Hourly workload
              </p>
              <h4 className="mt-2 text-3xl font-black text-on-surface">
                {isJobLoading ? "Loading..." : `${maxWeeklyHours} hrs / week`}
              </h4>
            </div>
            <Clock3 className="h-7 w-7 text-primary" />
          </div>
          <div className="mt-4 grid grid-cols-2 gap-3 text-sm">
            <div className="rounded-xl bg-white px-3 py-2 shadow-sm">
              <p className="text-[10px] font-bold uppercase tracking-wider text-slate-400">
                Job ID
              </p>
              <p className="font-semibold text-on-surface">{contract.job_id}</p>
            </div>
            <div className="rounded-xl bg-white px-3 py-2 shadow-sm">
              <p className="text-[10px] font-bold uppercase tracking-wider text-slate-400">
                Job title
              </p>
              <p className="truncate font-semibold text-on-surface">
                {job?.title ?? contract.job_title}
              </p>
            </div>
          </div>
          <div className="mt-4 rounded-xl bg-white px-4 py-3 text-sm text-on-surface-variant shadow-sm">
            No milestones are used for hourly contracts. Track work against the
            weekly cap above.
          </div>
        </div>
      ) : (
        <div 

        className="space-y-3">
          <div className="flex justify-between text-sm font-bold">
            <span className="text-on-surface-variant">Project Progress</span>
            <span className="text-primary">{progressPercent}%</span>
          </div>
          <div className="h-3 w-full overflow-hidden rounded-full bg-surface-container-high">
            <div
              className="h-full rounded-full bg-primary transition-all duration-1000"
              style={{ width: `${progressPercent}%` }}
            />
          </div>

          <div className="mt-4 space-y-2">
            <p className="mb-2 text-xs font-bold uppercase tracking-wider text-on-surface-variant">
              Milestones
            </p>
            {milestones.length > 0 ? (
              milestones.map((milestone, index) => {
                const statusMeta = getMilestoneStatusMeta(milestone.Status);
                const StatusIcon = statusMeta.Icon;

                return (
                  <div
                    key={milestone.ID}
                    className={`flex items-center justify-between rounded-lg px-3 py-1.5 text-xs ${statusMeta.className}`}
                  >
                    <span className="font-semibold text-on-surface">
                      {index + 1}. {milestone.Description}
                    </span>
                    <span
                      className={`flex items-center gap-1 rounded-full px-2 py-1 font-bold ${statusMeta.className}`}
                    >
                      <StatusIcon
                        className={`h-3.5 w-3.5 ${statusMeta.iconClassName}`}
                        aria-hidden="true"
                      />
                      {statusMeta.label}
                    </span>
                  </div>
                );
              })
            ) : (
              <div className="rounded-lg bg-surface-container-low px-3 py-2 text-sm text-on-surface-variant">
                No milestones yet.
              </div>
            )}
          </div>
        </div>
      )}

      <div className="grid grid-cols-2 gap-4 border-y border-outline-variant/50 py-4">
        <div>
          <p className="text-[10px] font-bold uppercase tracking-wider text-slate-400">
            Total Budget
          </p>
          <p className="text-xl font-black text-on-surface">
            {formatMoney(contract.total_budget ?? contract.weekly_hour_limit * contract.hourly_rate)}
          </p>
        </div>
        <div>
          <p className="text-[10px] font-bold uppercase tracking-wider text-slate-400">
            {contractType === "HOURLY" ? "Weekly Cap" : "Due Date"}
          </p>
          <p className="text-xl font-black text-on-surface">
            {contractType === "HOURLY"
              ? `${contract.weekly_hour_limit} hours`
              : formatDate(contract.end_date)}
          </p>
        </div>
      </div>

      <div className="mt-auto flex flex-wrap gap-3">
        <button
          type="button"
          className="flex-1 rounded-xl bg-primary py-3 text-sm font-bold text-white shadow-lg shadow-primary/20 transition-all hover:bg-blue-700 active:scale-95"
        >
          {isHourly ? "Track Time" : "Submit Work"}
        </button>
        <button
          type="button"
          className="flex flex-1 items-center justify-center gap-2 rounded-xl border border-outline-variant py-3 text-sm font-bold text-on-surface-variant transition-all hover:bg-slate-50 active:scale-95"
          onClick={(e) => {
            e.stopPropagation();
            router.push(`/messages?userid=${contract.client_id}`);
          }}
        >
          <MessageCircle className="size-4" />
          <span
          >Message Client</span>
        </button>
      </div>
    </div>
  );
}

export default function ContractsPage() {
  const [activeFilter, setActiveFilter] = useState<ContractFilter>("ACTIVE");
  const [searchTerm, setSearchTerm] = useState("");

  const {
    data: contracts = [],
    isLoading,
    isError,
    refetch,
  } = useGetMyContractsQuery();

  const filteredContracts = useMemo(() => {
    const normalizedSearch = searchTerm.trim().toLowerCase();

    return contracts.filter((contract) => {
      const matchesFilter =
        activeFilter === "ALL" ||
        contract.status.toUpperCase() === activeFilter;
      const searchableText = [
        contract.title,
        contract.job_title,
        contract.client_first_name,
        contract.client_last_name,
        contract.freelancer_first_name,
        contract.freelancer_last_name,
        contract.description,
        contract.proposal_description,
        contract.type,
      ]
        .filter(Boolean)
        .join(" ")
        .toLowerCase();

      const matchesSearch =
        !normalizedSearch || searchableText.includes(normalizedSearch);

      return matchesFilter && matchesSearch;
    });
  }, [activeFilter, contracts, searchTerm]);

  const totalActiveValue = useMemo(() => {
    return contracts
      .filter((contract) => contract.status.toUpperCase() === "ACTIVE")
      .reduce((sum, contract) => sum + (Number(contract.total_budget) || 0), 0);
  }, [contracts]);

  const upcomingDeadlines = useMemo(() => {
    return contracts.filter((contract) => {
      if (!contract.end_date) return false;

      const endDate = new Date(contract.end_date);
      if (Number.isNaN(endDate.getTime())) return false;

      const sevenDaysFromNow = new Date();
      sevenDaysFromNow.setDate(sevenDaysFromNow.getDate() + 7);

      return endDate <= sevenDaysFromNow;
    }).length;
  }, [contracts]);

  const pendingInvoices = useMemo(() => {
    return contracts.reduce((sum, contract) => {
      const milestones = contract.milestones ?? [];
      return (
        sum +
        milestones
          .filter(
            (milestone) =>
              (milestone.Status ?? "").toUpperCase() === "SUBMITTED",
          )
          .reduce(
            (milestoneSum, milestone) =>
              milestoneSum + Number(milestone.Amount || 0),
            0,
          )
      );
    }, 0);
  }, [contracts]);

  return (
    <div className="flex min-w-0 min-h-screen flex-col">
      <main className="mx-auto mt-12 w-full max-w-7xl grow p-8 lg:p-12">
        <header className="mb-12">
          <h1 className="mb-4 text-5xl font-black leading-none tracking-tighter text-on-background">
            My Contracts
          </h1>
          <p className="text-lg text-on-surface-variant">
            Manage your ongoing projects, track milestones, and keep clients
            updated.
          </p>
        </header>

        <div className="mb-8 flex flex-col justify-between gap-6 md:flex-row md:items-center">
          <div className="flex w-fit rounded-xl bg-surface-container-low p-1.5">
            {(
              [
                { label: "Active", value: "ACTIVE" as const },
                { label: "Completed", value: "COMPLETED" as const },
                { label: "Cancelled", value: "CANCELLED" as const },
              ] as const
            ).map((tab) => (
              <button
                key={tab.value}
                type="button"
                onClick={() => setActiveFilter(tab.value)}
                className={`rounded-lg px-8 py-2.5 font-bold transition-colors ${
                  activeFilter === tab.value
                    ? "bg-surface text-primary shadow-sm"
                    : "font-medium text-secondary hover:text-primary"
                }`}
              >
                {tab.label}
              </button>
            ))}
          </div>
          <div className="relative w-full md:w-96">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 text-slate-400" />
            <input
              className="w-full rounded-xl border-none bg-surface py-3 pl-12 pr-4 shadow-xl transition-shadow focus:ring-2 focus:ring-primary/20"
              placeholder="Search contracts, clients, or tags..."
              type="text"
              value={searchTerm}
              onChange={(event) => setSearchTerm(event.target.value)}
            />
          </div>
        </div>

        {isError ? (
          <div className="rounded-2xl border border-red-200 bg-red-50 p-6 text-red-700">
            <p className="font-semibold">Unable to load contracts.</p>
            <button
              type="button"
              onClick={() => refetch()}
              className="mt-3 rounded-xl bg-red-600 px-4 py-2 text-sm font-bold text-white transition-colors hover:bg-red-700"
            >
              Retry
            </button>
          </div>
        ) : null}

        {isLoading ? (
          <div className="grid grid-cols-1 gap-8 xl:grid-cols-2">
            {Array.from({ length: 4 }).map((_, index) => (
              <div
                key={index}
                className="h-112 animate-pulse rounded-2xl bg-surface shadow-xl"
              />
            ))}
          </div>
        ) : filteredContracts.length === 0 ? (
          <div className="rounded-2xl border-2 border-dashed border-outline bg-slate-50/50 p-10 text-center">
            <h2 className="text-2xl font-bold text-on-surface">
              No contracts found
            </h2>
            <p className="mx-auto mt-3 max-w-xl text-on-surface-variant">
              Try a different filter or search term, or check back once new
              contracts are created.
            </p>
          </div>
        ) : (
          <div className="grid grid-cols-1 gap-8 xl:grid-cols-2">
            {filteredContracts.map((contract) => (
              <ContractCard key={contract.contract_id} contract={contract} />
            ))}

            <div className="xl:col-span-2 grid grid-cols-1 gap-8 md:grid-cols-3">
              <div className="relative flex flex-col justify-between overflow-hidden rounded-2xl bg-primary p-8 text-white shadow-xl">
                <div className="z-10">
                  <p className="mb-1 text-sm font-bold uppercase tracking-widest text-white/70">
                    Total Active Value
                  </p>
                  <h4 className="text-4xl font-black">
                    {formatMoney(totalActiveValue)}
                  </h4>
                </div>
                <Wallet className="absolute -right-4 -bottom-4 h-28 w-28 text-white/10" />
              </div>
              <div className="flex flex-col justify-between rounded-2xl border-b-4 border-tertiary bg-surface p-8 shadow-xl">
                <div>
                  <p className="mb-1 text-sm font-bold uppercase tracking-widest text-on-surface-variant">
                    Upcoming Deadlines
                  </p>
                  <h4 className="text-4xl font-black text-on-surface">
                    {upcomingDeadlines}
                  </h4>
                </div>
                <div className="mt-4 flex items-center gap-2 text-sm font-bold text-tertiary">
                  <Clock3 className="size-4" />
                  Within 7 days
                </div>
              </div>
              <div className="flex flex-col justify-between rounded-2xl border-b-4 border-blue-400 bg-surface p-8 shadow-xl">
                <div>
                  <p className="mb-1 text-sm font-bold uppercase tracking-widest text-on-surface-variant">
                    Pending Invoices
                  </p>
                  <h4 className="text-4xl font-black text-on-surface">
                    {formatMoney(pendingInvoices)}
                  </h4>
                </div>
                <div className="mt-4 flex items-center gap-2 text-sm font-bold text-blue-400">
                  <Receipt className="size-4" />
                  Awaiting approval
                </div>
              </div>
            </div>
          </div>
        )}
      </main>

    </div>
  );
}
