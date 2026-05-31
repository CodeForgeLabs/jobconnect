"use client";

import { useGetJobsQuery } from "@/api/jobsapi";
import { useGetMeQuery } from "@/api/userapi";
import { useGetMyContractsQuery } from "@/api/contractapi";
import { useGetMyProposalsQuery } from "@/api/proposalapi";
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

export default function FreelancerDashboard() {
  const router = useRouter();
  const { data: me } = useGetMeQuery();
  const { data: jobs = [], isLoading } = useGetJobsQuery();
  const { data: proposals = [] } = useGetMyProposalsQuery();
  const { data: contracts = [] } = useGetMyContractsQuery();

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

  const spotlightContract = activeContracts[0] ?? contracts[0];
  const deadlineSource = (spotlightContract?.milestones ?? []).slice(
    0,
    3,
  ) as Array<{
    Description?: string;
    Amount?: number;
  }>;

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
              onClick = {() => router.push("/freelancer/wallet")}
              className="text-xs font-bold text-primary hover:underline font-label">
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
              className="text-sm font-bold text-primary flex items-center gap-2 group font-label">
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
                    <div
                      key={job.id}
                      className="bg-surface-container-lowest p-6 rounded-lg border border-outline-variant/10 hover:border-primary/20 transition-all cursor-pointer"
                    >
                      <div className="flex flex-col sm:flex-row justify-between items-start gap-4 mb-4">
                        <div className="flex gap-4">
                          <div className="w-11 h-11 bg-surface-container rounded-xl flex items-center justify-center shrink-0">
                            {index === 0 ? (
                              <Sparkles
                                className="h-5 w-5 text-primary-container"
                                aria-hidden="true"
                              />
                            ) : index === 1 ? (
                              <Wallet
                                className="h-5 w-5 text-primary-container"
                                aria-hidden="true"
                              />
                            ) : (
                              <ShoppingBag
                                className="h-5 w-5 text-primary-container"
                                aria-hidden="true"
                              />
                            )}
                          </div>
                          <div>
                            <h4 className="text-lg font-bold text-on-surface">
                              {job.title}
                            </h4>
                            <p className="text-sm text-on-surface-variant">
                              {job.company_name || "Client"} • Posted{" "}
                              {formatDate(job.created_at)}
                            </p>
                          </div>
                        </div>
                        <span className="bg-tertiary-fixed text-on-tertiary-fixed-variant px-4 py-1 rounded-full text-xs font-bold uppercase tracking-wider font-label whitespace-nowrap">
                          {job.status === "OPEN" ? "Best Match" : job.status}
                        </span>
                      </div>
                      <p className="text-on-surface-variant leading-relaxed mb-5 line-clamp-2 text-sm">
                        {job.description}
                      </p>
                      <div className="flex flex-wrap gap-2 mb-5">
                        {job.skills
                          .split(",")
                          .map((skill) => skill.trim())
                          .filter(Boolean)
                          .slice(0, 3)
                          .map((skill) => (
                            <span
                              key={skill}
                              className="px-3 py-1 bg-surface-container text-on-surface-variant text-xs font-medium rounded-md font-label"
                            >
                              {skill}
                            </span>
                          ))}
                      </div>
                      <div className="flex flex-col sm:flex-row justify-between items-stretch sm:items-center gap-4 pt-5 border-t border-outline-variant/10">
                        <div className="flex flex-wrap gap-4 text-sm font-semibold text-on-surface">
                          <span>
                            {job.job_type === "HOURLY"
                              ? `${formatMoney(job.hourly_rate)} / hr`
                              : formatMoney(job.budget)}
                          </span>
                          <span className="text-on-surface-variant font-normal">
                            {job.experience_level}
                          </span>
                        </div>
                        <button className="bg-primary text-white px-5 py-2.5 rounded-full font-bold hover:bg-primary-container transition-all active:scale-95 text-sm">
                          Apply Now
                        </button>
                      </div>
                    </div>
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
                Upcoming Deadlines
              </h3>

              <div className="space-y-6 relative after:absolute after:top-2 after:bottom-2 after:left-1.75 after:w-0.5 after:bg-outline-variant/20">
                {deadlineSource.length > 0 ? (
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
                            <div className="h-full bg-primary w-4/5"></div>
                          </div>
                          <p className="text-[10px] text-right mt-1 font-bold text-primary font-label">
                            80% Done
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
