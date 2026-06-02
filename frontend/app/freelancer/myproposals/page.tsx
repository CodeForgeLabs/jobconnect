"use client";
import Link from "next/link";
import { useMemo, useState } from "react";
import { Cable, Clover, MailOpen } from "lucide-react";
import { useGetJobByIdQuery } from "@/api/jobsapi";
import { useGetMeQuery } from "@/api/userapi";
import { useRouter } from "next/navigation";
import {
  type Proposal,
  type ProposalStatus,
  useGetMyProposalsQuery,
} from "@/api/proposalapi";

const ITEMS_PER_PAGE = 4;

type ProposalTab = "PENDING" | "INVITED" | "REJECTED" | "HIRED";

type ProposalRow = {
  amount: string;
  company: string;
  id: number;
  icon: string;
  iconBg: string;
  jobId: number;
  status: ProposalStatus;
  statusColor: string;
  time: string;
  title: string;
  type: string;
};

const formatMoney = (value?: number) => {
  if (typeof value !== "number") return "N/A";

  return value.toLocaleString(undefined, {
    maximumFractionDigits: 0,
  });
};

const formatProposalTime = (proposal?: Proposal) => {
  if (!proposal?.created_at) return "Submitted recently";

  const createdAt = new Date(proposal.created_at);
  if (Number.isNaN(createdAt.getTime())) return "Submitted recently";

  return `Submitted ${createdAt.toLocaleDateString(undefined, {
    month: "short",
    day: "numeric",
    year: "numeric",
  })}`;
};

const getStatusColor = (status: ProposalStatus) => {
  switch (status) {
    case "INVITED":
      return "bg-emerald-100 text-emerald-800 dark:bg-emerald-950/40 dark:text-emerald-400";
    case "REJECTED":
      return "bg-surface-container-highest text-on-surface-variant";
    case "HIRED":
      return "bg-primary/10 text-primary";
    case "PENDING":
    default:
      return "bg-amber-100 text-amber-800 dark:bg-amber-950/40 dark:text-amber-400";
  }
};

const getJobIcon = (title: string, category?: string) => {
  const text = `${title} ${category ?? ""}`.toLowerCase();

  if (text.includes("design") || text.includes("ui") || text.includes("ux")) {
    return {
      icon: "brush",
      iconBg:
        "bg-indigo-100 dark:bg-indigo-950/40 text-indigo-600 dark:text-indigo-400",
    };
  }

  if (
    text.includes("market") ||
    text.includes("campaign") ||
    text.includes("social")
  ) {
    return {
      icon: "campaign",
      iconBg:
        "bg-orange-100 dark:bg-orange-950/40 text-orange-600 dark:text-orange-400",
    };
  }

  if (
    text.includes("python") ||
    text.includes("data") ||
    text.includes("automation") ||
    text.includes("developer") ||
    text.includes("react") ||
    text.includes("engineer")
  ) {
    return { icon: "code", iconBg: "bg-primary/10 text-primary" };
  }

  return {
    icon: "work",
    iconBg: "bg-pink-100 dark:bg-pink-950/40 text-pink-600 dark:text-pink-400",
  };
};

const buildProposalRow = (
  proposal: Proposal,
  job?: {
    budget: number;
    category: string;
    company_name: string;
    hourly_rate: number;
    job_type: string;
    title: string;
  },
): ProposalRow => {
  // buildProposalRow can receive an undefined `job` when the job
  // list doesn't include the job for this proposal. Handle gracefully.
  const jobType = job?.job_type === "HOURLY" ? "Hourly" : "Fixed Price";
  const amount =
    job?.job_type === "HOURLY"
      ? `$${formatMoney(job.hourly_rate)}/hr`
      : job
        ? `$${formatMoney(job.budget)}`
        : "N/A";
  const iconData = getJobIcon(job?.title ?? "Proposal", job?.category);

  return {
    amount,
    company: job?.company_name ?? "Company unavailable",
    id: proposal.id,
    icon: iconData.icon,
    iconBg: iconData.iconBg,
    jobId: proposal.job_id,
    status: proposal.status,
    statusColor: getStatusColor(proposal.status),
    time: formatProposalTime(proposal),
    title: job?.title ?? `Job #${proposal.job_id}`,
    type: jobType,
  };
};

export default function ProposalsView() {
  const [activeTab, setActiveTab] = useState<ProposalTab>("PENDING");
  const [currentPage, setCurrentPage] = useState(0);
  const router = useRouter();

  const { data: proposalsData = [], isLoading: proposalsLoading } =
    useGetMyProposalsQuery();
  const { data: meData } = useGetMeQuery();

  // Filter raw proposals by active tab and compute pagination from raw data.
  const filteredProposalsData = useMemo(() => {
    return proposalsData.filter((p) => p.status === activeTab);
  }, [activeTab, proposalsData]);

  const tabCounts = useMemo(() => {
    return proposalsData.reduce(
      (counts, proposal) => {
        if (proposal.status === "PENDING") counts.PENDING += 1;
        if (proposal.status === "INVITED") counts.INVITED += 1;
        if (proposal.status === "REJECTED") counts.REJECTED += 1;
        if (proposal.status === "HIRED") counts.HIRED += 1;
        return counts;
      },
      { PENDING: 0, INVITED: 0, REJECTED: 0, HIRED: 0 },
    );
  }, [proposalsData]);

  const totalPages = Math.max(
    1,
    Math.ceil(filteredProposalsData.length / ITEMS_PER_PAGE),
  );
  const safeCurrentPage = Math.min(currentPage, totalPages - 1);
  const canGoPrevious = safeCurrentPage > 0;
  const canGoNext = safeCurrentPage < totalPages - 1;

  const visibleProposals = useMemo(() => {
    const start = safeCurrentPage * ITEMS_PER_PAGE;
    return filteredProposalsData.slice(start, start + ITEMS_PER_PAGE);
  }, [filteredProposalsData, safeCurrentPage]);

  const totalListings = filteredProposalsData.length;
  const showingStart =
    totalListings === 0 ? 0 : safeCurrentPage * ITEMS_PER_PAGE + 1;
  const showingEnd = Math.min(
    (safeCurrentPage + 1) * ITEMS_PER_PAGE,
    totalListings,
  );
  const successRate =
    proposalsData.length > 0
      ? `${Math.round(
          (proposalsData.filter((proposal) => proposal.status === "HIRED")
            .length /
            proposalsData.length) *
            100,
        )}%`
      : "0%";
  const connectCount = meData?.connect ?? 0;

  const ProposalItem = ({ proposal }: { proposal: Proposal }) => {
    const { data: jobResp } = useGetJobByIdQuery(proposal.job_id);
    const job = jobResp?.job;
    const row = buildProposalRow(proposal, job);

    return (
      <div
        key={row.id}
        className="p-5 md:p-6 flex flex-col sm:flex-row sm:items-center justify-between gap-4 hover:bg-surface-container-low/40 transition-colors group"
      >
        <div className="flex gap-4 items-start min-w-0">
          <div
            className={`w-12 h-12 rounded-xl flex items-center justify-center shrink-0 ${row.iconBg}`}
          >
            <span className="material-symbols-outlined text-xl">
              {row.icon}
            </span>
          </div>
          <div className="space-y-1 min-w-0">
            <div className="flex flex-wrap items-center gap-2">
              <h3 className="font-bold text-base text-on-surface font-headline truncate group-hover:text-primary transition-colors">
                {row.title}
              </h3>
              <span
                className={`text-[10px] font-extrabold tracking-wide uppercase px-2 py-0.5 rounded-md ${row.statusColor}`}
              >
                {row.status}
              </span>
            </div>
            <p className="text-on-surface-variant text-xs font-medium">
              {row.company}
              <span className="mx-1.5 text-outline">•</span>
              {row.time}
            </p>
          </div>
        </div>

        <div className="flex items-center justify-between sm:justify-end gap-6 sm:pl-0 pl-16">
          <div className="sm:text-right">
            <p className="font-extrabold text-sm font-headline text-on-surface">
              {row.amount}
            </p>
            <p className="text-outline text-[11px] font-medium tracking-wide uppercase mt-0.5">
              {row.type}
            </p>
          </div>

          <div className="flex items-center gap-2">
            <Link
              href={`/freelancer/job/${row.jobId}`}
              className="px-4 py-2 border border-outline-variant hover:border-outline hover:bg-surface text-on-surface-variant hover:text-on-surface rounded-xl text-xs font-bold transition-all"
            >
              View Proposal
            </Link>
            <Link
              href={`/freelancer/job/${row.jobId}`}
              aria-label="Open proposal details"
              className="p-2 text-outline hover:text-on-surface rounded-lg transition-colors"
            >
              delete
            </Link>
          </div>
        </div>
      </div>
    );
  };

  return (
    <div
      className={`min-h-screen flex flex-col bg-surface text-on-surface transition-colors duration-200 selection:bg-primary-fixed selection:text-primary`}
    >
      {/* Main Workspace Frame */}
      <main className="flex-1 max-w-6xl w-full mx-auto px-4 py-8 md:py-12 space-y-8">
        {/* Page Identity Dashboard Block */}
        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
          <div>
            <h1 className="text-3xl font-black tracking-tight font-headline">
              My Proposals
            </h1>
            <p className="text-on-surface-variant text-sm mt-1">
              Track and manage your active market listings.
            </p>
          </div>
          <button
            onClick={() => {
              router.push("/freelancer/jobsearch");
            }}
            className="bg-primary text-white px-5 py-3 rounded-xl font-bold text-sm hover:shadow-lg hover:shadow-primary/20 active:scale-98 transition-all flex items-center justify-center gap-2 w-full sm:w-auto"
          >
            <span className="material-symbols-outlined text-lg">+</span>
            Find New Work
          </button>
        </div>

        {/* Analytics Grid Block */}
        <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
          <StatCard
            label="Total"
            value={proposalsData.length}
            colorClass="bg-primary/10 text-primary"
          />
          <StatCard
            label="Success"
            value={successRate}
            colorClass="bg-primary/10 text-primary"
          />
          <StatCard
            label="Connects"
            value={connectCount}
            colorClass="bg-primary text-white shadow-sm"
          />
        </div>

        {/* Dynamic Interactive Document Manager Block */}
        <div className="bg-surface-container-lowest border border-outline-variant/30 rounded-2xl shadow-xs overflow-hidden">
          {/* Internal Tab Filter Ribbon */}
          <div className="flex border-b border-outline-variant/20 px-4 md:px-6 overflow-x-auto scrollbar-hide bg-surface-container-low">
            <TabButton
              label="PENDING"
              count={tabCounts.PENDING}
              active={activeTab === "PENDING"}
              onClick={() => {
                setActiveTab("PENDING");
                setCurrentPage(0);
              }}
            />
            <TabButton
              label="INVITED"
              count={tabCounts.INVITED}
              active={activeTab === "INVITED"}
              onClick={() => {
                setActiveTab("INVITED");
                setCurrentPage(0);
              }}
            />
            <TabButton
              label="REJECTED"
              count={tabCounts.REJECTED}
              active={activeTab === "REJECTED"}
              onClick={() => {
                setActiveTab("REJECTED");
                setCurrentPage(0);
              }}
            />
            <TabButton
              label="HIRED"
              count={tabCounts.HIRED}
              active={activeTab === "HIRED"}
              onClick={() => {
                setActiveTab("HIRED");
                setCurrentPage(0);
              }}
            />
          </div>

          {/* Proposals List Segment */}
          <div className="divide-y divide-outline-variant/10">
            {proposalsLoading ? (
              <div className="p-6 text-sm text-on-surface-variant">
                Loading proposals...
              </div>
            ) : visibleProposals.length > 0 ? (
              visibleProposals.map((proposal) => (
                <ProposalItem proposal={proposal} key={proposal.id} />
              ))
            ) : (
              <div className="p-6 text-sm text-on-surface-variant">
                No proposals found for this tab.
              </div>
            )}
          </div>

          {/* Paginated Footer System */}
          <footer className="flex items-center justify-between px-6 py-4 border-t border-outline-variant/20 bg-surface-container-low">
            <span className="text-xs text-on-surface-variant font-medium">
              Showing {showingEnd - showingStart} of {totalListings} listings
            </span>
            <div className="flex gap-1.5">
              <button
                type="button"
                onClick={() => setCurrentPage((page) => Math.max(0, page - 1))}
                disabled={!canGoPrevious}
                className="p-2 border border-outline-variant/40 rounded-xl text-on-surface-variant flex items-center transition-all disabled:opacity-40 disabled:cursor-not-allowed hover:border-outline bg-surface-container-lowest hover:text-on-surface"
              >
                <span className="material-symbols-outlined text-base">
                  {"<<"}
                </span>
              </button>
              <button
                type="button"
                onClick={() =>
                  setCurrentPage((page) => Math.min(totalPages - 1, page + 1))
                }
                disabled={!canGoNext}
                className="p-2 border border-outline-variant/40 rounded-xl text-on-surface-variant flex items-center transition-all disabled:opacity-40 disabled:cursor-not-allowed hover:border-outline bg-surface-container-lowest hover:text-on-surface"
              >
                <span className="material-symbols-outlined text-base">
                  {">>"}
                </span>
              </button>
            </div>
          </footer>
        </div>
      </main>
    </div>
  );
}

/* Local UI Building Blocks */
function StatCard({
  label,
  value,
  colorClass,
}: {
  icon?: unknown;
  label: string;
  value: string | number;
  colorClass: string;
}) {
  return (
    <div className="bg-surface-container-lowest p-5 rounded-2xl border border-outline-variant/20 flex items-center gap-4 shadow-xs">
      <div
        className={`w-12 h-12 rounded-xl flex items-center justify-center shrink-0 ${colorClass}`}
      >
        <span className="material-symbols-outlined text-xl">
          {label == "Total" ? (
            <MailOpen />
          ) : label == "Success" ? (
            <Clover />
          ) : (
            <Cable />
          )}
        </span>
      </div>
      <div>
        <p className="text-outline text-[10px] font-extrabold uppercase tracking-widest">
          {label}
        </p>
        <p className="text-2xl font-black font-headline text-on-surface mt-0.5 tracking-tight">
          {value}
        </p>
      </div>
    </div>
  );
}

function TabButton({
  label,
  count,
  active,
  onClick,
}: {
  label: string;
  count: number | null;
  active: boolean;
  onClick: () => void;
}) {
  return (
    <button
      onClick={onClick}
      className={`flex items-center justify-center border-b-2 px-4 pb-4 pt-5 gap-2 transition-all font-headline font-bold text-sm whitespace-nowrap outline-none ${
        active
          ? "border-primary text-primary"
          : "border-transparent text-on-surface-variant hover:text-on-surface"
      }`}
    >
      <span>{label}</span>
      {count !== null && (
        <span
          className={`text-[10px] font-extrabold px-2 py-0.5 rounded-full ${
            active
              ? "bg-primary/10 text-primary"
              : "bg-surface-container-highest text-on-surface-variant"
          }`}
        >
          {count}
        </span>
      )}
    </button>
  );
}
