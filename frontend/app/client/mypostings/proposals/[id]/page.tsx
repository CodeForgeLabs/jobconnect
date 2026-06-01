"use client";

import Image from "next/image";
import { useCallback, useEffect, useMemo, useState } from "react";
import { useParams, useRouter } from "next/navigation";
import { useGetJobByIdQuery } from "@/api/jobsapi";
import { useSendMessageMutation } from "@/api/messageapi";
import defualtAvatar from "@/assets/avatarsvg.png";
import {
  ProposalApplicant,
  useGetJobProposalsMutation,
  useUpdateProposalMutation,
} from "@/api/proposalapi";
import { useCreateContractMutation } from "@/api/contractapi";
import { User, useGetMeQuery, useGetUserByIdQuery } from "@/api/userapi";

const DEFAULT_AVATAR_URL =
  "https://img.daisyui.com/images/stock/photo-1534528741775-53994a69daeb.webp";

const parseSkills = (skills?: string) =>
  (skills || "")
    .split(",")
    .map((skill) => skill.trim())
    .filter(Boolean);

const formatMoney = (amount?: number) => {
  if (typeof amount !== "number") return "N/A";

  return amount.toLocaleString(undefined, {
    style: "currency",
    currency: "USD",
    maximumFractionDigits: 0,
  });
};

const formatPostedDate = (value: string | Date) => {
  if (!value) return "Posted recently";

  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return "Posted recently";

  return `Posted ${date.toLocaleDateString(undefined, {
    month: "short",
    day: "numeric",
    year: "numeric",
  })}`;
};

const buildInvitationMessage = (jobTitle: string, companyName: string) =>
  `Hello,

We reviewed your proposal for ${jobTitle} at ${companyName} and were impressed by your background and approach.

We would like to invite you for an interview so we can discuss the role, expectations, and next steps in more detail.

Please reply with a few times that work well for you this week. We look forward to speaking with you.

Best regards,
${companyName} Hiring Team`;

export default function JobProposalsPage() {
  const params = useParams<{ id: string }>();
  const jobId = Number(params.id);
  const isValidJobId = Number.isFinite(jobId) && jobId > 0;

  const {
    data: jobResponse,
    isLoading: isLoadingJob,
    isError: isJobError,
  } = useGetJobByIdQuery(jobId, {
    skip: !isValidJobId,
  });
  const { data: me } = useGetMeQuery();

  const [getJobProposals, { isLoading: isLoadingProposals }] =
    useGetJobProposalsMutation();
  const [sendMessage, { isLoading: isSendingMessage }] =
    useSendMessageMutation();

  const [proposals, setProposals] = useState<ProposalApplicant[]>([]);
  const [proposalError, setProposalError] = useState<string | null>(null);

  const job = jobResponse?.job;
  const requiredSkills = useMemo(() => parseSkills(job?.skills), [job?.skills]);

  const refreshProposals = useCallback(async () => {
    if (!isValidJobId) return;

    try {
      const result = await getJobProposals({ job_id: jobId }).unwrap();
      setProposals(result);
      setProposalError(null);
    } catch (error) {
      console.error("Failed to load job proposals", error);
      setProposalError("Unable to load job proposals.");
    }
  }, [getJobProposals, isValidJobId, jobId]);

  useEffect(() => {
    if (!isValidJobId) return;

    let isMounted = true;

    (async () => {
      try {
        const result = await getJobProposals({ job_id: jobId }).unwrap();
        if (!isMounted) return;
        setProposals(result);
        setProposalError(null);
      } catch (error) {
        if (!isMounted) return;
        console.error("Failed to load job proposals", error);
        setProposalError("Unable to load job proposals.");
      }
    })();

    return () => {
      isMounted = false;
    };
  }, [getJobProposals, isValidJobId, jobId]);

  if (!isValidJobId) {
    return (
      <div className="min-h-screen bg-surface text-on-surface">
        <main className="mx-auto max-w-6xl px-4 py-16 md:px-8">
          <h1 className="text-3xl font-extrabold tracking-tight text-primary">
            Invalid job id
          </h1>
        </main>
      </div>
    );
  }

  if (isLoadingJob) {
    return (
      <div className="min-h-screen bg-surface text-on-surface">
        <main className="mx-auto max-w-6xl px-4 py-16 md:px-8">
          <h1 className="text-3xl font-extrabold tracking-tight text-primary">
            Loading job details...
          </h1>
        </main>
      </div>
    );
  }

  if (isJobError || !job) {
    return (
      <div className="min-h-screen bg-surface text-on-surface">
        <main className="mx-auto max-w-6xl px-4 py-16 md:px-8">
          <h1 className="text-3xl font-extrabold tracking-tight text-primary">
            Job not found
          </h1>
        </main>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-surface text-on-surface selection:bg-primary-fixed selection:text-primary">
      <main className="mx-auto max-w-7xl px-4 py-10 md:px-8 lg:py-14">
        <header className="mb-10 space-y-4">
          <div className="flex flex-wrap items-center gap-3 text-sm">
            <span className="rounded-full bg-tertiary-fixed px-3 py-1 text-[10px] font-bold uppercase tracking-[0.2em] text-on-tertiary-fixed-variant">
              {job.status || "Open"}
            </span>
            <span className="text-on-surface-variant">Job ID #{job.id}</span>
            <span className="text-on-surface-variant">
              {formatPostedDate(job.created_at)}
            </span>
          </div>

          <div className="flex flex-col gap-3 md:flex-row md:items-end md:justify-between">
            <div className="max-w-3xl space-y-2">
              <h1 className="text-3xl font-black tracking-tight text-primary md:text-5xl">
                {job.title}
              </h1>
              <p className="text-on-surface-variant md:text-lg">
                Review and manage the incoming proposals for this posting.
              </p>
            </div>

            <div className="grid grid-cols-2 gap-3 text-sm md:grid-cols-4">
              <HeaderStat
                label="Proposals"
                value={String(proposals?.length ?? 0)}
              />
              <HeaderStat
                label={job.job_type === "HOURLY" ? "Hourly" : "Budget"}
                value={
                  job.job_type === "HOURLY"
                    ? `${formatMoney(job.hourly_rate)}/hr`
                    : formatMoney(job.budget)
                }
              />
              <HeaderStat label="Work mode" value={job.work_mode || "N/A"} />
              <HeaderStat label="Level" value={job.experience_level || "N/A"} />
            </div>
          </div>
        </header>

        <section className="grid gap-8 lg:grid-cols-[1.2fr_0.8fr]">
          <div className="space-y-6">
            <div className="rounded-3xl border border-outline-variant/20 bg-surface-container-low p-6 shadow-sm md:p-8">
              <div className="grid gap-4 md:grid-cols-3">
                <InfoCard
                  label="Company"
                  value={job.company_name || "Company"}
                />
                <InfoCard
                  label="Location"
                  value={job.location || "Location not specified"}
                />
                <InfoCard
                  label="Weekly hours"
                  value={
                    typeof job.max_weekly_hours === "number" &&
                    job.max_weekly_hours > 0
                      ? `${job.max_weekly_hours} hrs/week`
                      : "Not specified"
                  }
                />
              </div>

              <div className="mt-8 space-y-4">
                <h2 className="text-xl font-bold text-primary">
                  About the role
                </h2>
                <p className="leading-relaxed text-on-surface-variant">
                  {job.description}
                </p>
              </div>

              {requiredSkills.length > 0 && (
                <div className="mt-8 space-y-3">
                  <h3 className="text-sm font-bold uppercase tracking-[0.2em] text-on-surface-variant">
                    Required skills
                  </h3>
                  <div className="flex flex-wrap gap-2">
                    {requiredSkills.map((skill) => (
                      <span
                        key={skill}
                        className="rounded-full bg-surface-container px-3 py-1 text-xs font-semibold text-on-surface-variant"
                      >
                        {skill}
                      </span>
                    ))}
                  </div>
                </div>
              )}
            </div>
          </div>

          <aside className="space-y-6">
            <div className="rounded-3xl border border-outline-variant/20 bg-surface-container-low p-6 shadow-sm md:p-8">
              <h2 className="text-lg font-bold text-primary">Summary</h2>
              <div className="mt-5 space-y-4 text-sm">
                <SummaryRow label="Status" value={job.status || "OPEN"} />
                <SummaryRow
                  label="Company"
                  value={job.company_name || "Company"}
                />
                <SummaryRow label="Location" value={job.location || "N/A"} />
                <SummaryRow label="Type" value={job.job_type || "N/A"} />
                <SummaryRow label="Mode" value={job.work_mode || "N/A"} />
              </div>
            </div>
          </aside>
        </section>

        <section className="mt-8 space-y-5">
          <div className="flex items-center justify-between">
            <div>
              <h2 className="text-2xl font-bold tracking-tight text-on-surface">
                Job proposals
              </h2>
            </div>
          </div>

          {isLoadingProposals ? (
            <div className="rounded-3xl border border-dashed border-outline-variant/30 bg-surface-container-low p-8 text-center text-sm text-on-surface-variant">
              Loading proposals...
            </div>
          ) : proposalError ? (
            <div className="rounded-3xl border border-dashed border-error/40 bg-error-container/20 p-8 text-center text-sm text-error">
              {proposalError}
            </div>
          ) : !proposals ? (
            <div className="rounded-3xl border border-dashed border-outline-variant/30 bg-surface-container-low p-8 text-center text-sm text-on-surface-variant">
              No proposals yet.
            </div>
          ) : (
            <div className="space-y-5 w-full">
              {proposals.map((proposal, index) => (
                <ProposalCard
                  key={
                    proposal.proposal_id ??
                    proposal.user_id ??
                    proposal.sender_id ??
                    index
                  }
                  proposal={proposal}
                  jobId={jobId}
                  senderId={me?.id ?? 0}
                  onActionComplete={refreshProposals}
                  onSendMessage={sendMessage}
                  isSendingMessage={isSendingMessage}
                  invitationMessage={buildInvitationMessage(
                    job?.title || "the role",
                    job?.company_name || "our company",
                  )}
                />
              ))}
            </div>
          )}
        </section>
      </main>
    </div>
  );
}

function HeaderStat({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-2xl border border-outline-variant/20 bg-white/70 px-4 py-3 shadow-sm">
      <div className="text-[10px] font-bold uppercase tracking-[0.2em] text-on-surface-variant">
        {label}
      </div>
      <div className="mt-1 text-sm font-bold text-on-surface">{value}</div>
    </div>
  );
}

function InfoCard({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-2xl bg-white p-4 shadow-sm ring-1 ring-outline-variant/10">
      <div className="text-[10px] font-bold uppercase tracking-[0.2em] text-on-surface-variant">
        {label}
      </div>
      <div className="mt-1 text-sm font-semibold text-on-surface">{value}</div>
    </div>
  );
}

function SummaryRow({ label, value }: { label: string; value: string }) {
  return (
    <div className="flex items-center justify-between gap-4 border-b border-outline-variant/10 pb-3 last:border-b-0 last:pb-0">
      <span className="text-on-surface-variant">{label}</span>
      <span className="font-semibold text-on-surface">{value}</span>
    </div>
  );
}

function ProposalCard({
  proposal,
  jobId,
  senderId,
  onActionComplete,
  onSendMessage,
  isSendingMessage,
  invitationMessage,
}: {
  proposal: ProposalApplicant;
  jobId: number;
  senderId: number;
  onActionComplete: () => Promise<void>;
  onSendMessage: ReturnType<typeof useSendMessageMutation>[0];
  isSendingMessage: boolean;
  invitationMessage: string;
}) {
  const router = useRouter();
  const resolvedUserId = proposal.user_id ?? proposal.sender_id ?? 0;
  const { data: applicantUser } = useGetUserByIdQuery(resolvedUserId, {
    skip: !resolvedUserId,
  });
  const [updateProposal, { isLoading: isUpdatingProposal }] =
    useUpdateProposalMutation();
  const [createContract, { isLoading: isCreatingContract }] =
    useCreateContractMutation();
  const [actionError, setActionError] = useState<string | null>(null);
  const [activeAction, setActiveAction] = useState<"invite" | "hire" | null>(
    null,
  );

  const applicant: User | undefined = applicantUser;
  const displayName =
    `${applicant?.first_name || proposal.first_name || ""} ${applicant?.last_name || proposal.last_name || ""}`.trim() ||
    applicant?.email ||
    proposal.email ||
    "Applicant";
  const avatarUrl = applicant?.profile_picture_url || defualtAvatar;
  const headline = applicant?.headline || proposal.headline || "Freelancer";
  const skills = parseSkills(
    applicant?.skills ? String(applicant.skills) : proposal.skills,
  );
  const proposalText = proposal.description || "No cover letter provided.";

  const handleInviteToTalk = async () => {
    if (!proposal.proposal_id || !resolvedUserId || !senderId) {
      setActionError("This proposal cannot be updated right now.");
      return;
    }

    setActiveAction("invite");
    setActionError(null);

    try {
      await onSendMessage({
        caption: "",
        image_url: "",
        receiver_id: resolvedUserId,
        sender_id: senderId,
        text: invitationMessage,
        type: "text",
        video_url: "",
      }).unwrap();

      await updateProposal({
        id: proposal.proposal_id,
        payload: { status: "INVITED" },
      }).unwrap();
      await onActionComplete();
    } catch (error) {
      console.error("Failed to invite applicant", error);
      setActionError("Unable to invite this applicant right now.");
    } finally {
      setActiveAction(null);
    }
  };

  const handleHireApplicant = async () => {
    if (!proposal.proposal_id || !resolvedUserId) {
      setActionError("This applicant cannot be hired right now.");
      return;
    }

    setActiveAction("hire");
    setActionError(null);

    try {
      await updateProposal({
        id: proposal.proposal_id,
        payload: { status: "HIRED" },
      }).unwrap();

      await createContract({
        freelancer_id: String(resolvedUserId),
        job_id: String(jobId),
      }).unwrap();

      await onActionComplete();
    } catch (error) {
      console.error("Failed to hire applicant", error);
      setActionError("Unable to hire this applicant right now.");
    } finally {
      setActiveAction(null);
    }
  };

  return (
    <article className="w-full  rounded-3xl border border-outline-variant/20 bg-surface-container-lowest p-6 shadow-sm transition-all hover:shadow-md md:p-7">
      <div className="flex flex-col gap-6 lg:flex-row lg:items-start lg:justify-between">
        <div
          onClick={() => {
            router.push(`/freelancer/profile/${resolvedUserId}`);
          }}
          className="flex gap-4"
        >
          <div className="h-16 w-16 overflow-hidden rounded-2xl bg-surface-container">
            <Image
              src={avatarUrl}
              alt={displayName}
              width={64}
              height={64}
              className="h-full w-full object-cover"
            />
          </div>

          <div className="min-w-0 space-y-2">
            <div>
              <h3 className="text-xl font-bold text-on-surface">
                {displayName}
              </h3>
              <p className="text-sm text-on-surface-variant">{headline}</p>
            </div>

            <div className="flex flex-wrap gap-2 text-xs">
              {applicant?.location && <Badge>{applicant.location}</Badge>}
              {typeof applicant?.hourly_rate === "number" && (
                <Badge>{formatMoney(applicant.hourly_rate)}/hr</Badge>
              )}
            </div>
          </div>
        </div>

        <div className="flex flex-wrap gap-3">
          <button
            type="button"
            onClick={handleInviteToTalk}
            disabled={
              !resolvedUserId ||
              activeAction !== null ||
              isUpdatingProposal ||
              isCreatingContract ||
              isSendingMessage ||
              !senderId ||
              proposal.status === "INVITED" ||
              proposal.status === "HIRED"
            }
            className="rounded-xl border border-primary/20 px-4 py-2.5 text-sm font-bold text-primary transition-all hover:bg-primary/5 disabled:cursor-not-allowed disabled:opacity-50"
          >
            {activeAction === "invite"
              ? "Inviting..."
              : proposal.status === "PENDING"
                ? "Invite to talk"
                : "Invited"}
          </button>
          <button
            type="button"
            onClick={handleHireApplicant}
            disabled={
              !resolvedUserId ||
              activeAction !== null ||
              isUpdatingProposal ||
              isCreatingContract ||
              proposal.status === "HIRED" ||
              proposal.status === "PENDING" ||
              proposal.status === "REJECTED"
            }
            className="rounded-xl bg-primary px-4 py-2.5 text-sm font-bold text-white transition-all hover:opacity-90 disabled:cursor-not-allowed disabled:opacity-50"
          >
            {proposal.status === "HIRED"
              ? "Hired"
              : activeAction === "hire"
                ? "Hiring..."
                : proposal.status === "PENDING"
                  ? "Hire"
                  : "Rejected"}
          </button>
        </div>

        {actionError ? (
          <p className="text-sm font-medium text-red-600">{actionError}</p>
        ) : null}
      </div>

      <div className="mt-6 space-y-4">
        <div>
          <h4 className="text-xs font-bold uppercase tracking-[0.2em] text-on-surface-variant">
            Cover letter
          </h4>
          <p className="mt-2 whitespace-pre-wrap leading-relaxed text-on-surface-variant">
            {proposalText}
          </p>
        </div>

        <div className="flex"></div>
        {skills.length > 0 && (
          <div>
            <h4 className="text-xs font-bold uppercase tracking-[0.2em] text-on-surface-variant">
              Skills
            </h4>
            <div className="mt-3 flex flex-wrap gap-2">
              {skills.map((skill) => (
                <span
                  key={skill}
                  className="rounded-full bg-surface-container px-3 py-1 text-xs font-semibold text-on-surface-variant"
                >
                  {skill}
                </span>
              ))}
            </div>
          </div>
        )}
      </div>
    </article>
  );
}

function Badge({ children }: { children: React.ReactNode }) {
  return (
    <span className="rounded-full bg-surface-container px-3 py-1 font-semibold text-on-surface-variant">
      {children}
    </span>
  );
}
