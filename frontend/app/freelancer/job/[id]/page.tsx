"use client";
import React, { useState } from "react";
import { useParams } from "next/navigation";
import { useGetJobByIdQuery } from "@/api/jobsapi";
import {
  useGetMyProposalsQuery,
  useCreateProposalMutation,
} from "@/api/proposalapi";
const parseSkills = (skills: string) =>
  skills
    .split(",")
    .map((skill) => skill.trim())
    .filter(Boolean);

const formatPostedDate = (value?: string) => {
  if (!value) return "Posted recently";

  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return "Posted recently";

  return `Posted ${date.toLocaleDateString(undefined, {
    month: "short",
    day: "numeric",
    year: "numeric",
  })}`;
};

const formatMoney = (amount?: number) => {
  if (typeof amount !== "number") return "N/A";

  return amount.toLocaleString(undefined, {
    style: "currency",
    currency: "USD",
    maximumFractionDigits: 0,
  });
};

export default function JobDetailView() {
  const params = useParams<{ id: string }>();
  const jobId = Number(params.id);
  console.log("Job ID from URL:", jobId);
  const isValidJobId = Number.isFinite(jobId) && jobId > 0;
  const [coverLetter, setCoverLetter] = useState("");

  const {
    data: jobdata,
    isLoading,
    isError,
  } = useGetJobByIdQuery(jobId, {
    skip: !isValidJobId,
  });
  const [createProposal, { isLoading: isSubmittingProposal }] =
    useCreateProposalMutation();

  const job = jobdata?.job;
  const requiredSkills = parseSkills(job?.skills ?? "");
  const primarySkills = requiredSkills.slice(0, 3);
  const secondarySkills = requiredSkills.slice(3);

  const { data: proposalsData } = useGetMyProposalsQuery(undefined, {
    skip: !isValidJobId,
  });

  const myProposals = proposalsData;
  const hasApplied = myProposals?.some((proposal) => proposal.job_id === jobId);

  const roleHighlights =
    job?.milestones
      ?.map((milestone) => milestone.description)
      .filter(Boolean) ?? [];

  const fallbackHighlights = [
    "Review project details and align on expected delivery outcomes.",
    "Collaborate clearly with stakeholders throughout execution.",
    "Deliver high-quality work that matches the posted requirements.",
  ];

  const responsibilities =
    roleHighlights.length > 0 ? roleHighlights : fallbackHighlights;

  const handleSubmitProposal = async (
    event: React.FormEvent<HTMLFormElement>,
  ) => {
    event.preventDefault();

    if (!coverLetter.trim()) return;

    const result = await createProposal({
      job_id: jobId,
      cover_letter: coverLetter.trim(),
    });
    //refetch the job
      setCoverLetter(result.data?.description ?? "");
  };

  if (!isValidJobId) {
    return (
      <div className="bg-surface text-on-surface selection:bg-primary-fixed selection:text-primary min-h-screen">
        <main className="max-w-screen-2xl mx-auto px-6 md:px-8 pt-8 md:pt-12 mb-24">
          <header className="mb-12">
            <h1 className="text-4xl md:text-6xl font-extrabold text-primary tracking-tighter leading-tight mb-4">
              Invalid Job ID
            </h1>
          </header>
        </main>
      </div>
    );
  }

  if (isLoading) {
    return (
      <div className="bg-surface text-on-surface selection:bg-primary-fixed selection:text-primary min-h-screen">
        <main className="max-w-screen-2xl mx-auto px-6 md:px-8 pt-8 md:pt-12 mb-24">
          <header className="mb-12">
            <h1 className="text-4xl md:text-6xl font-extrabold text-primary tracking-tighter leading-tight mb-4">
              Loading job details...
            </h1>
          </header>
        </main>
      </div>
    );
  }

  if (isError || !job) {
    return (
      <div className="bg-surface text-on-surface selection:bg-primary-fixed selection:text-primary min-h-screen">
        <main className="max-w-screen-2xl mx-auto px-6 md:px-8 pt-8 md:pt-12 mb-24">
          <header className="mb-12">
            <h1 className="text-4xl md:text-6xl font-extrabold text-primary tracking-tighter leading-tight mb-4">
              Job not found
            </h1>
          </header>
        </main>
      </div>
    );
  }

  return (
    <div className="bg-surface text-on-surface selection:bg-primary-fixed selection:text-primary min-h-screen">
      <main className="max-w-screen-2xl mx-auto px-6 md:px-8 pt-8 md:pt-12 mb-24">
        {/* Hero Section */}
        <header className="mb-12">
          <div className="flex flex-wrap items-center gap-3 mb-4">
            <span className="bg-tertiary-fixed text-on-tertiary-fixed-variant px-3 py-1 rounded-full text-[10px] md:text-xs font-bold tracking-wide uppercase">
              {job.status || "Open Position"}
            </span>
            <span className="text-on-surface-variant text-sm font-medium">
              {formatPostedDate(job.created_at)}
            </span>
          </div>
          <h1 className="text-4xl md:text-6xl font-extrabold text-primary tracking-tighter leading-tight mb-4">
            {job.title}
          </h1>
          <div className="flex flex-col md:flex-row md:items-center gap-4 md:gap-6 text-on-surface-variant font-medium">
            <div className="flex items-center gap-2">
              <svg
                aria-hidden="true"
                className="h-5 w-5 text-primary"
                fill="none"
                viewBox="0 0 24 24"
                xmlns="http://www.w3.org/2000/svg"
              >
                <path
                  d="M12 21s7-6.1 7-11a7 7 0 1 0-14 0c0 4.9 7 11 7 11z"
                  stroke="currentColor"
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth="1.9"
                />
                <circle cx="12" cy="10" fill="currentColor" r="2.4" />
              </svg>
              <span>
                {job.location || "Location not specified"}
                {job.work_mode ? ` (${job.work_mode})` : ""}
              </span>
            </div>
            <div className="flex items-center gap-2">
              <svg
                aria-hidden="true"
                className="h-5 w-5 text-primary"
                fill="none"
                viewBox="0 0 24 24"
                xmlns="http://www.w3.org/2000/svg"
              >
                <path
                  d="M9 6V4.5A1.5 1.5 0 0 1 10.5 3h3A1.5 1.5 0 0 1 15 4.5V6"
                  stroke="currentColor"
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth="1.8"
                />
                <rect
                  height="12"
                  rx="2"
                  stroke="currentColor"
                  strokeWidth="1.8"
                  width="18"
                  x="3"
                  y="6"
                />
                <path
                  d="M3 11h18"
                  stroke="currentColor"
                  strokeLinecap="round"
                  strokeWidth="1.8"
                />
              </svg>
              <span>{job.company_name || "Company"}</span>
            </div>
          </div>
        </header>

        <div className="grid grid-cols-1 lg:grid-cols-12 gap-8 lg:gap-16">
          {/* Main Content Area */}
          <div className="lg:col-span-8 space-y-12 md:space-y-16">
            {/* Bento Grid Stats */}
            <section className="grid grid-cols-2 md:grid-cols-4 gap-4 md:gap-6 p-6 md:p-8 bg-surface-container-low rounded-xl">
              <StatItem
                label={job.job_type === "HOURLY" ? "Hourly Rate" : "Budget"}
                value={
                  job.job_type === "HOURLY"
                    ? `${formatMoney(job.hourly_rate)}/hr`
                    : formatMoney(job.budget)
                }
              />
              <StatItem
                label="Duration"
                value={
                  typeof job.max_weekly_hours === "number" &&
                  job.max_weekly_hours > 0
                    ? `${job.max_weekly_hours} hrs/week`
                    : "Not specified"
                }
              />
              <StatItem label="Level" value={job.experience_level || "N/A"} />
              <StatItem label="Work Type" value={job.work_mode || "N/A"} />
            </section>

            {/* Description */}
            <article className="prose prose-slate max-w-none">
              <h2 className="text-2xl md:text-3xl font-bold text-primary mb-6">
                About the Role
              </h2>
              <p className="text-on-surface-variant leading-relaxed text-base md:text-lg mb-6">
                {job.description}
              </p>
              <h3 className="text-xl md:text-2xl font-bold text-primary mb-4">
                Responsibilities
              </h3>
              <ul className="space-y-4 list-none p-0 text-on-surface-variant text-base md:text-lg">
                {responsibilities.map((item) => (
                  <ResponsibilityItem key={item} text={item} />
                ))}
              </ul>
            </article>

            {/* Skills */}
            <section>
              <h2 className="text-2xl font-bold text-primary mb-6">
                Required Skills
              </h2>
              <div className="flex flex-wrap gap-2 md:gap-3">
                {primarySkills.map((skill) => (
                  <span
                    key={skill}
                    className="px-4 md:px-6 py-2 md:py-2.5 bg-primary text-white rounded-full font-semibold text-xs md:text-sm shadow-lg shadow-primary/20"
                  >
                    {skill}
                  </span>
                ))}
                {secondarySkills.map((skill) => (
                  <span
                    key={skill}
                    className="px-4 md:px-6 py-2 md:py-2.5 bg-surface-container-highest text-primary rounded-full font-semibold text-xs md:text-sm"
                  >
                    {skill}
                  </span>
                ))}
                {requiredSkills.length === 0 ? (
                  <span className="px-4 md:px-6 py-2 md:py-2.5 bg-surface-container-highest text-primary rounded-full font-semibold text-xs md:text-sm">
                    No skills listed
                  </span>
                ) : null}
              </div>
            </section>
          </div>

          {/* Sidebar */}
          <aside className="lg:col-span-4 space-y-6 md:space-y-8">
            <div className="bg-primary p-6 md:p-8 rounded-xl text-white shadow-2xl shadow-primary/30">
              <button className="w-full py-4 bg-white text-primary font-extrabold rounded-lg text-lg mb-4 hover:scale-[1.02] active:scale-[0.98] transition-all">
                Apply Now
              </button>
              <button className="w-full py-4 border-2 border-white/30 text-white font-bold rounded-lg text-lg hover:bg-white/10 transition-colors flex items-center justify-center gap-2">
                <span className="material-symbols-outlined">favorite</span> Save
                Job
              </button>
            </div>

            <div className="bg-surface-container-low p-6 md:p-8 rounded-xl">
              <div className="flex items-center gap-4 mb-8">
                <div className="w-14 h-14 bg-white rounded-lg flex items-center justify-center shadow-sm">
                  <svg
                    aria-hidden="true"
                    className="h-8 w-8 text-primary"
                    fill="none"
                    viewBox="0 0 24 24"
                    xmlns="http://www.w3.org/2000/svg"
                  >
                    <path
                      d="M14.5 4.5c2.8.2 5 2.4 5 5.2 0 4.4-3.8 8.2-8.8 8.8l-2.2.2.2-2.2c.6-5 4.4-8.8 8.8-8.8z"
                      stroke="currentColor"
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth="1.8"
                    />
                    <circle cx="14.5" cy="9.5" r="1.2" fill="currentColor" />
                    <path
                      d="M8.2 14.8l-2.7.3-.3-2.7 2.1-2.1 2.9 2.9-2 1.6z"
                      stroke="currentColor"
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth="1.8"
                    />
                    <path
                      d="M5 19l3-1"
                      stroke="currentColor"
                      strokeLinecap="round"
                      strokeWidth="1.8"
                    />
                  </svg>
                </div>
                <div>
                  <div className="flex items-center gap-2">
                    <h3 className="text-lg font-bold text-primary leading-tight">
                      {job.company_name || "Company"}
                    </h3>
                    <svg
                      aria-hidden="true"
                      className="h-4 w-4 text-blue-500"
                      fill="none"
                      viewBox="0 0 24 24"
                      xmlns="http://www.w3.org/2000/svg"
                    >
                      <circle cx="12" cy="12" fill="currentColor" r="10" />
                      <path
                        d="M8 12.5l2.5 2.5L16 9.5"
                        stroke="white"
                        strokeLinecap="round"
                        strokeLinejoin="round"
                        strokeWidth="2"
                      />
                    </svg>
                  </div>
                  <p className="text-on-surface-variant text-xs">
                    {job.category || "Category not specified"}
                  </p>
                </div>
              </div>
              <div className="space-y-4">
                <ClientStat
                  label="Member since"
                  value={
                    job.created_at
                      ? String(new Date(job.created_at).getFullYear())
                      : "N/A"
                  }
                />
                <ClientStat
                  label="Total Spent"
                  value={formatMoney(job.budget)}
                />
                <ClientStat
                  label="Hire Rate"
                  value={`${job.applications_count ?? 0} applicants`}
                />
              </div>
            </div>
          </aside>
        </div>
        <div>
          {hasApplied ? (
            <p className="text-green-600 font-semibold mt-6">
              You have already applied to this job.
            </p>
          ) : (
            <form className="flex flex-col" onSubmit={handleSubmitProposal}>
              <textarea
                placeholder="Write your cover letter here..."
                className=" mt-8 min-w-24 py-6 px-2  bg-gray-100 border border-primary tablet:w-[60%] max-tablet:w-full"
                value={coverLetter}
                onChange={(event) => setCoverLetter(event.target.value)}
              />
              <button
                type="submit"
                className="mt-4 px-6 py-3 bg-primary text-white font-bold rounded-lg hover:bg-primary-dark transition-colors disabled:opacity-60 disabled:cursor-not-allowed"
                disabled={isSubmittingProposal}
              >
                Submit Proposal
              </button>
            </form>
          )}
        </div>
      </main>
    </div>
  );
}

// Sub-components for cleaner code
type LabelValueProps = {
  label: string;
  value: string;
};

type ResponsibilityItemProps = {
  text: string;
};

const StatItem = ({ label, value }: LabelValueProps) => (
  <div className="flex flex-col gap-1">
    <span className="text-on-surface-variant text-[10px] md:text-xs uppercase tracking-widest font-bold">
      {label}
    </span>
    <span className="text-lg md:text-2xl font-bold text-primary">{value}</span>
  </div>
);

const ResponsibilityItem = ({ text }: ResponsibilityItemProps) => (
  <li className="flex items-start gap-3">
    <svg
      aria-hidden="true"
      className="mt-1 h-5 w-5 shrink-0 text-primary"
      fill="none"
      viewBox="0 0 24 24"
      xmlns="http://www.w3.org/2000/svg"
    >
      <circle cx="12" cy="12" fill="currentColor" r="10" />
      <path
        d="M8 12.5l2.5 2.5L16 9.5"
        stroke="white"
        strokeLinecap="round"
        strokeLinejoin="round"
        strokeWidth="2"
      />
    </svg>
    <span>{text}</span>
  </li>
);

const ClientStat = ({ label, value }: LabelValueProps) => (
  <div className="flex justify-between items-center border-b border-outline-variant/30 pb-3 last:border-0 last:pb-0">
    <span className="text-on-surface-variant text-sm">{label}</span>
    <span className="font-bold text-primary">{value}</span>
  </div>
);
