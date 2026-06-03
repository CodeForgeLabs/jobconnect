"use client";
import React, { useState } from "react";
import {
  type ExperienceLevel,
  type JobType,
  type WorkMode,
  useCreateJobMutation,
  useDeleteJobMutation,
  useGetMyJobsQuery,
} from "@/api/jobsapi";
import { useGetWalletBalanceQuery } from "@/api/walletapi";
import {
  AlertTriangle,
  Banknote,
  Clock3,
  CreditCard,
  Plus,
  Trash2,
} from "lucide-react";
import { useRouter } from "next/navigation";
import {
  validateOptionalPositiveDecimal,
  validateOptionalPositiveWholeNumber,
  validatePositiveDecimal,
  validatePositiveWholeNumber,
} from "@/lib/fieldValidation";

type MilestoneDraft = {
  amount: string;
  deadline: string;
  description: string;
};

type JobFormState = {
  title: string;
  category: string;
  company_name: string;
  description: string;
  experience_level: ExperienceLevel;
  hourly_rate: string;
  budget: string;
  is_private: boolean;
  job_type: JobType;
  location: string;
  max_weekly_hours: string;
  skills: string;
  work_mode: WorkMode;
};

type FormErrors = Partial<Record<string, string>>;

const defaultMilestone = (): MilestoneDraft => ({
  amount: "",
  deadline: "",
  description: "",
});

const defaultForm: JobFormState = {
  title: "",
  category: "",
  company_name: "",
  description: "",
  experience_level: "ENTRY",
  hourly_rate: "",
  budget: "",
  is_private: false,
  job_type: "FIXED",
  location: "",
  max_weekly_hours: "",
  skills: "",
  work_mode: "REMOTE",
};

const formatMoney = (currency: string, amountMinor: number) =>
  new Intl.NumberFormat("en-US", {
    style: "currency",
    currency,
    maximumFractionDigits: 2,
  }).format(amountMinor);

const toDateInputValue = (date: Date) => date.toISOString().slice(0, 10);

export default function MyPostingsView() {
  const [activeTab, setActiveTab] = useState("open");
  const [isCreateOpen, setIsCreateOpen] = useState(false);
  const [currentPage, setCurrentPage] = useState(1);
  const [formError, setFormError] = useState<string | null>(null);
  const [formErrors, setFormErrors] = useState<FormErrors>({});
  const [milestones, setMilestones] = useState<MilestoneDraft[]>([
    defaultMilestone(),
  ]);
  const pageSize = 4;
  const [form, setForm] = useState<JobFormState>(defaultForm);

  const {
    data: myJobs = [],
    isLoading,
    isError,
    refetch,
  } = useGetMyJobsQuery();

  const { data: wallet } = useGetWalletBalanceQuery();

  const [createJob, { isLoading: isCreating }] = useCreateJobMutation();
  const [deleteJob] = useDeleteJobMutation();
  const walletCurrency = wallet?.Currency ?? "ETB";
  const walletBalanceMinor = wallet?.BalanceMinor ?? 0;
  const parsedBudgetValue = Number(form.budget || 0);
  const budgetValue = Number.isFinite(parsedBudgetValue) ? parsedBudgetValue : 0;
  const budgetMinor = budgetValue;
  const milestoneTotal = milestones.reduce((total, milestone) => {
    const amount = Number(milestone.amount || 0);
    return total + (Number.isFinite(amount) ? amount : 0);
  }, 0);
  const milestoneTotalMinor = Math.round(milestoneTotal);
  const balanceShortfallMinor = Math.max(0, budgetMinor - walletBalanceMinor);
  const hasEnoughBalance = walletBalanceMinor >= budgetMinor;
  const budgetInputError = validateOptionalPositiveDecimal(
    form.budget,
    "Budget",
  );
  const hourlyRateInputError =
    form.job_type === "HOURLY"
      ? validateOptionalPositiveDecimal(form.hourly_rate, "Hourly rate")
      : null;
  const weeklyHoursInputError =
    form.job_type === "HOURLY"
      ? validateOptionalPositiveWholeNumber(
          form.max_weekly_hours,
          "Max weekly hours",
        )
      : null;
  const getMilestoneAmountError = (amount: string) =>
    validateOptionalPositiveDecimal(amount, "Amount");

  const normalizeStatus = (status?: string): "OPEN" | "CLOSED" | "PAUSED" => {
    const value = (status || "").toUpperCase();
    if (value === "CLOSED" || value === "PAUSED")
      return value as "CLOSED" | "PAUSED";
    return "OPEN";
  };

  const getPostedTime = (value: string | Date) => {
    if (!value) return "Posted recently";

    const postedDate = new Date(value);
    if (Number.isNaN(postedDate.getTime())) return "Posted recently";

    return `Posted ${postedDate.toLocaleDateString(undefined, {
      month: "short",
      day: "numeric",
      year: "numeric",
    })}`;
  };

  const filteredPostings = myJobs.filter((job) => {
    const status = normalizeStatus(job.status);
    if (activeTab === "open") return status === "OPEN";
    if (activeTab === "closed") return status === "CLOSED";
    return status !== "OPEN" && status !== "CLOSED";
  });

  const totalPages = Math.max(1, Math.ceil(filteredPostings.length / pageSize));
  const activePage = Math.min(currentPage, totalPages);

  const startIndex = (activePage - 1) * pageSize;
  const paginatedPostings = filteredPostings.slice(
    startIndex,
    startIndex + pageSize,
  );

  const openCount = myJobs.filter(
    (job) => normalizeStatus(job.status) === "OPEN",
  ).length;
  const closedCount = myJobs.filter(
    (job) => normalizeStatus(job.status) === "CLOSED",
  ).length;
  const draftCount = myJobs.filter(
    (job) => normalizeStatus(job.status) === "PAUSED",
  ).length;

  const resetForm = () => {
    setForm(defaultForm);
    setMilestones([defaultMilestone()]);
    setFormErrors({});
    setFormError(null);
  };

  const updateMilestone = (
    index: number,
    field: keyof MilestoneDraft,
    value: string,
  ) => {
    setMilestones((current) =>
      current.map((milestone, milestoneIndex) =>
        milestoneIndex === index ? { ...milestone, [field]: value } : milestone,
      ),
    );
  };

  const addMilestone = () => {
    setMilestones((current) => [...current, defaultMilestone()]);
  };

  const removeMilestone = (index: number) => {
    setMilestones((current) =>
      current.length === 1
        ? current
        : current.filter((_, currentIndex) => currentIndex !== index),
    );
  };

  const validateForm = () => {
    const errors: FormErrors = {};
    const trimmedTitle = form.title.trim();
    const trimmedCategory = form.category.trim();
    const trimmedCompany = form.company_name.trim();
    const trimmedDescription = form.description.trim();
    const trimmedLocation = form.location.trim();
    const trimmedSkills = form.skills.trim();

    if (!trimmedTitle) errors.title = "Job title is required.";
    if (!trimmedCategory) errors.category = "Category is required.";
    if (!trimmedCompany) errors.company_name = "Company name is required.";
    if (!trimmedDescription) errors.description = "Description is required.";
    if (!trimmedLocation) errors.location = "Location is required.";
    if (!trimmedSkills) errors.skills = "Add at least one skill.";

    const budgetError = validatePositiveDecimal(form.budget, "Budget");
    if (budgetError) errors.budget = budgetError;

    if (form.job_type === "HOURLY") {
      const hourlyRateError = validatePositiveDecimal(
        form.hourly_rate,
        "Hourly rate",
      );
      const weeklyHoursError = validatePositiveWholeNumber(
        form.max_weekly_hours,
        "Max weekly hours",
      );

      if (hourlyRateError) errors.hourly_rate = hourlyRateError;
      if (weeklyHoursError) errors.max_weekly_hours = weeklyHoursError;
    }

    let previousDeadline: number | null = null;
    const hasInvalidMilestone =
      form.job_type === "HOURLY"
        ? false
        : milestones.some((milestone, index) => {
            const parsedDeadline = new Date(milestone.deadline);
            const amountError = validatePositiveDecimal(
              milestone.amount,
              "Amount",
            );

            const valid =
              form.job_type === "HOURLY" ||
              (milestone.description.trim().length > 0 &&
                !amountError &&
                milestone.deadline.length > 0 &&
                !Number.isNaN(parsedDeadline.getTime()));

            if (!valid) {
              errors[`milestone-${index}`] =
                amountError ||
                "Each milestone needs a description, amount, and deadline.";
              return true;
            }

            const currentDeadline = parsedDeadline.getTime();
            if (
              previousDeadline !== null &&
              currentDeadline <= previousDeadline
            ) {
              errors[`milestone-${index}`] =
                "Milestones must be entered in chronological order by deadline.";
              return true;
            }

            previousDeadline = currentDeadline;
            return false;
          });

    if (form.job_type !== "HOURLY") {
      if (milestones.length === 0 || hasInvalidMilestone) {
        errors.milestones = "Add at least one milestone with a deadline.";
      }

      if (milestoneTotal <= 0) {
        errors.milestones = "Milestone amounts must total more than zero.";
      }

      if (Math.abs(milestoneTotal - budgetValue) > 0.01) {
        errors.milestones = "Milestone amounts must match the total budget.";
      }
    }

    if (!hasEnoughBalance) {
      errors.balance = `You need ${formatMoney(walletCurrency, balanceShortfallMinor)} more to post this job.`;
    }

    return errors;
  };

  const handleCreateJob = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();

    const validationErrors = validateForm();
    setFormErrors(validationErrors);

    if (Object.keys(validationErrors).length > 0) {
      setFormError("Fix the highlighted fields before posting the job.");
      return;
    }

    const skills = form.skills
      .split(",")
      .map((skill) => skill.trim())
      .filter(Boolean);
    const jobType: "FIXED" | "HOURLY" = form.job_type;

    const payloadMilestones =
      jobType === "HOURLY"
        ? []
        : milestones.map((milestone) => ({
            amount: Number(milestone.amount),
            description: milestone.description.trim(),
            deadline: new Date(milestone.deadline),
          }));

    try {
      await createJob({
        title: form.title.trim(),
        category: form.category.trim(),
        company_name: form.company_name.trim(),
        description: form.description.trim(),
        experience_level: form.experience_level,
        hourly_rate: Number(form.hourly_rate || 0),
        budget: budgetValue,
        is_private: form.is_private,
        job_type: jobType,
        location: form.location.trim(),
        max_weekly_hours: Number(form.max_weekly_hours || 0),
        skills,
        work_mode: form.work_mode,
        milestones: payloadMilestones,
      }).unwrap();

      setIsCreateOpen(false);
      resetForm();
      refetch();
    } catch (error) {
      console.error("Failed to create job", error);
      setFormError("Unable to create the job right now.");
    }
  };

  const pageNumbers = (() => {
    if (totalPages <= 5) {
      return Array.from({ length: totalPages }, (_, index) => index + 1);
    }

    const pages: Array<number | "ellipsis"> = [1];
    const start = Math.max(2, activePage - 1);
    const end = Math.min(totalPages - 1, activePage + 1);

    if (start > 2) pages.push("ellipsis");

    for (let page = start; page <= end; page += 1) {
      pages.push(page);
    }

    if (end < totalPages - 1) pages.push("ellipsis");

    pages.push(totalPages);
    return pages;
  })();

  const handleTabClick = (tab: string) => {
    setActiveTab(tab);
    setCurrentPage(1);
  };
  const router = useRouter();

  const handleJobDelete = (jobId: number) => {
    deleteJob(jobId);
    console.log(`Delete job with ID: ${jobId}`);
  };

  return (
    <div className="min-h-screen flex flex-col bg-surface text-on-surface transition-colors duration-200 selection:bg-primary-fixed selection:text-primary">
      {/* Primary Workspace Frame */}
      <main className="flex-1 w-full max-w-6xl mx-auto px-4 py-8 md:py-12 space-y-8">
        {/* Dynamic Context Header Block */}
        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
          <div>
            <h1 className="text-3xl font-black tracking-tight font-headline">
              My Postings
            </h1>
            <p className="text-on-surface-variant text-sm mt-1">
              Manage, update, and audit your active corporate requisitions.
            </p>
          </div>
          <button
            className="bg-primary text-white px-5 py-3 rounded-xl font-bold text-sm hover:shadow-lg hover:shadow-primary/20 active:scale-98 transition-all flex items-center justify-center gap-2 w-full sm:w-auto"
            onClick={() => {
              setIsCreateOpen((current) => !current);
              if (isCreateOpen) {
                resetForm();
              }
            }}
          >
            <Plus className="h-4 w-4" />
            Post a New Job
          </button>
        </div>

        {isCreateOpen ? (
          <form
            onSubmit={handleCreateJob}
            className="overflow-hidden rounded-3xl border border-outline-variant/20 bg-surface-container-low shadow-sm"
          >
            <div className="border-b border-outline-variant/15 bg-linear-to-r from-primary/10 via-transparent to-secondary/10 px-5 py-5 md:px-6">
              <div className="flex flex-col gap-4 md:flex-row md:items-center md:justify-between">
                <div>
                  <h2 className="text-xl font-black tracking-tight font-headline">
                    Create New Job
                  </h2>
                  <p className="mt-1 text-sm text-on-surface-variant">
                    Add milestone deadlines and make sure the budget is covered
                    before posting.
                  </p>
                </div>
                <div className="grid gap-3 sm:grid-cols-2">
                  <div className="rounded-2xl border border-outline-variant/20 bg-surface px-4 py-3">
                    <p className="text-[10px] font-extrabold uppercase tracking-[0.2em] text-on-surface-variant">
                      Wallet balance
                    </p>
                    <p className="mt-1 text-lg font-black text-on-surface">
                      {formatMoney(walletCurrency, walletBalanceMinor)}
                    </p>
                  </div>
                  <div className="rounded-2xl border border-outline-variant/20 bg-surface px-4 py-3">
                    <p className="text-[10px] font-extrabold uppercase tracking-[0.2em] text-on-surface-variant">
                      Job funding needed
                    </p>
                    <p className="mt-1 text-lg font-black text-on-surface">
                      {formatMoney(walletCurrency, budgetMinor)}
                    </p>
                  </div>
                </div>
              </div>
              {formError ? (
                <div className="mt-4 flex items-start gap-3 rounded-2xl border border-amber-200 bg-amber-50 px-4 py-3 text-amber-800">
                  <AlertTriangle className="mt-0.5 h-4 w-4 shrink-0" />
                  <p className="text-sm font-medium">{formError}</p>
                </div>
              ) : null}
              {formErrors.balance ? (
                <div className="mt-4 flex items-start gap-3 rounded-2xl border border-red-200 bg-red-50 px-4 py-3 text-red-700">
                  <CreditCard className="mt-0.5 h-4 w-4 shrink-0" />
                  <p className="text-sm font-medium">{formErrors.balance}</p>
                </div>
              ) : null}
            </div>

            <div className="grid gap-6 p-5 md:grid-cols-[minmax(0,1.6fr)_minmax(280px,0.9fr)] md:p-6">
              <div className="space-y-6">
                <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
                  <Field label="Title" error={formErrors.title}>
                    <input
                      className="input input-bordered w-full"
                      placeholder="Mobile app redesign"
                      value={form.title}
                      onChange={(event) =>
                        setForm((current) => ({
                          ...current,
                          title: event.target.value,
                        }))
                      }
                    />
                  </Field>
                  <Field label="Company name" error={formErrors.company_name}>
                    <input
                      className="input input-bordered w-full"
                      placeholder="Acme Studios"
                      value={form.company_name}
                      onChange={(event) =>
                        setForm((current) => ({
                          ...current,
                          company_name: event.target.value,
                        }))
                      }
                    />
                  </Field>
                  <Field label="Category" error={formErrors.category}>
                    <input
                      className="input input-bordered w-full"
                      placeholder="Design"
                      value={form.category}
                      onChange={(event) =>
                        setForm((current) => ({
                          ...current,
                          category: event.target.value,
                        }))
                      }
                    />
                  </Field>
                  <Field label="Location" error={formErrors.location}>
                    <input
                      className="input input-bordered w-full"
                      placeholder="Remote or Addis Ababa"
                      value={form.location}
                      onChange={(event) =>
                        setForm((current) => ({
                          ...current,
                          location: event.target.value,
                        }))
                      }
                    />
                  </Field>
                </div>

                <Field label="Description" error={formErrors.description}>
                  <textarea
                    className="textarea textarea-bordered min-h-36 w-full"
                    placeholder="Describe the scope, deliverables, and expectations."
                    value={form.description}
                    onChange={(event) =>
                      setForm((current) => ({
                        ...current,
                        description: event.target.value,
                      }))
                    }
                  />
                </Field>

                <div className="grid grid-cols-1 gap-4 md:grid-cols-3">
                  <Field label="Budget" error={budgetInputError || formErrors.budget}>
                    <input
                      className="input input-bordered w-full"
                      placeholder="1500"
                      type="number"
                      min="0"
                      step="0.01"
                      value={form.budget}
                      onChange={(event) =>
                        setForm((current) => ({
                          ...current,
                          budget: event.target.value,
                        }))
                      }
                    />
                  </Field>
                  <Field
                    label="Hourly rate"
                    error={hourlyRateInputError || formErrors.hourly_rate}
                  >
                    <input
                      className="input input-bordered w-full"
                      placeholder="35"
                      type="number"
                      min="0"
                      step="0.01"
                      value={form.hourly_rate}
                      onChange={(event) =>
                        setForm((current) => ({
                          ...current,
                          hourly_rate: event.target.value,
                        }))
                      }
                    />
                  </Field>
                  <Field
                    label="Max weekly hours"
                    error={
                      weeklyHoursInputError || formErrors.max_weekly_hours
                    }
                  >
                    <input
                      className="input input-bordered w-full"
                      placeholder="20"
                      type="number"
                      min="0"
                      step="1"
                      value={form.max_weekly_hours}
                      onChange={(event) =>
                        setForm((current) => ({
                          ...current,
                          max_weekly_hours: event.target.value,
                        }))
                      }
                    />
                  </Field>
                </div>

                <Field label="Skills" error={formErrors.skills}>
                  <input
                    className="input input-bordered w-full"
                    placeholder="React, Figma, REST APIs"
                    value={form.skills}
                    onChange={(event) =>
                      setForm((current) => ({
                        ...current,
                        skills: event.target.value,
                      }))
                    }
                  />
                </Field>

                <div className="grid grid-cols-1 gap-4 md:grid-cols-3">
                  <Field label="Job type">
                    <select
                      className="select select-bordered w-full"
                      value={form.job_type}
                      onChange={(event) => {
                        const value = event.target.value as JobType;
                        setForm((current) => ({
                          ...current,
                          job_type: value,
                        }));
                        // When switching to hourly, remove milestones entirely
                        if (value === "HOURLY") {
                          setMilestones([]);
                        }
                      }}
                    >
                      <option value="FIXED">Fixed</option>
                      <option value="HOURLY">Hourly</option>
                    </select>
                  </Field>
                  <Field label="Experience level">
                    <select
                      className="select select-bordered w-full"
                      value={form.experience_level}
                      onChange={(event) =>
                        setForm((current) => ({
                          ...current,
                          experience_level: event.target
                            .value as ExperienceLevel,
                        }))
                      }
                    >
                      <option value="ENTRY">Entry</option>
                      <option value="INTERMEDIATE">Intermediate</option>
                      <option value="EXPERT">Expert</option>
                    </select>
                  </Field>
                  <Field label="Work mode">
                    <select
                      className="select select-bordered w-full"
                      value={form.work_mode}
                      onChange={(event) =>
                        setForm((current) => ({
                          ...current,
                          work_mode: event.target.value as WorkMode,
                        }))
                      }
                    >
                      <option value="REMOTE">Remote</option>
                      <option value="ONSITE">Onsite</option>
                      <option value="HYBRID">Hybrid</option>
                    </select>
                  </Field>
                </div>

                <label className="flex cursor-pointer items-center gap-3 rounded-2xl border border-outline-variant/30 bg-surface px-4 py-3">
                  <input
                    type="checkbox"
                    className="checkbox checkbox-sm"
                    checked={form.is_private}
                    onChange={(event) =>
                      setForm((current) => ({
                        ...current,
                        is_private: event.target.checked,
                      }))
                    }
                  />
                  <span>
                    <span className="block text-sm font-bold">Private job</span>
                    <span className="block text-xs text-on-surface-variant">
                      Hide the posting from public search results.
                    </span>
                  </span>
                </label>
              </div>

              {form.job_type !== "HOURLY" && (
                <div className="space-y-4 rounded-2xl border border-outline-variant/20 bg-surface px-4 py-4 md:px-5">
                  <div className="flex items-center justify-between gap-3">
                    <div>
                      <h3 className="text-base font-black tracking-tight">
                        Milestones
                      </h3>
                      <p className="text-xs text-on-surface-variant">
                        Each milestone needs a deadline and amount.
                      </p>
                    </div>
                    <button
                      type="button"
                      onClick={addMilestone}
                      className="inline-flex items-center gap-2 rounded-xl border border-outline-variant/20 px-3 py-2 text-xs font-bold text-primary transition-colors hover:bg-primary/5"
                    >
                      <Plus className="h-4 w-4" />
                      Add milestone
                    </button>
                  </div>

                  {formErrors.milestones ? (
                    <p className="mt-3 rounded-xl border border-red-200 bg-red-50 px-3 py-2 text-xs font-medium text-red-700">
                      {formErrors.milestones}
                    </p>
                  ) : null}

                  <div className="mt-4 space-y-3">
                    {milestones.map((milestone, index) => (
                      <div
                        key={`${index}`}
                        className="rounded-2xl border border-outline-variant/20 bg-surface-container-low p-3"
                      >
                        <div className="mb-3 flex items-center justify-between gap-3">
                          <p className="text-sm font-bold text-on-surface">
                            Milestone {index + 1}
                          </p>
                          <button
                            type="button"
                            onClick={() => removeMilestone(index)}
                            disabled={milestones.length === 1}
                            className="inline-flex items-center gap-1 text-xs font-bold text-on-surface-variant transition-colors hover:text-red-600 disabled:cursor-not-allowed disabled:opacity-40"
                          >
                            <Trash2 className="h-3.5 w-3.5" />
                            Remove
                          </button>
                        </div>

                        <div className="space-y-3">
                          {(() => {
                            const milestoneAmountError =
                              getMilestoneAmountError(milestone.amount);
                            const milestoneError =
                              milestoneAmountError ||
                              formErrors[`milestone-${index}`];

                            return (
                              <>
                          <input
                            className="input input-bordered w-full"
                            placeholder="Milestone description"
                            value={milestone.description}
                            onChange={(event) =>
                              updateMilestone(
                                index,
                                "description",
                                event.target.value,
                              )
                            }
                          />
                          <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
                            <input
                              className={`input input-bordered w-full ${
                                milestoneAmountError ? "input-error" : ""
                              }`}
                              placeholder="Amount"
                              type="number"
                              min="0"
                              step="0.01"
                              value={milestone.amount}
                              onChange={(event) =>
                                updateMilestone(
                                  index,
                                  "amount",
                                  event.target.value,
                                )
                              }
                            />
                            <input
                              className="input input-bordered w-full"
                              type="date"
                              min={toDateInputValue(new Date())}
                              value={milestone.deadline}
                              onChange={(event) =>
                                updateMilestone(
                                  index,
                                  "deadline",
                                  event.target.value,
                                )
                              }
                            />
                          </div>
                          {milestoneError ? (
                            <p className="text-xs font-medium text-red-600">
                              {milestoneError}
                            </p>
                          ) : null}
                              </>
                            );
                          })()}
                        </div>
                      </div>
                    ))}
                  </div>
                </div>
              )}

              {(form.job_type as JobType) === "HOURLY" && (
                <div className="rounded-2xl border border-outline-variant/20 bg-surface px-4 py-4 md:px-5">
                  <p className="text-sm font-medium text-on-surface-variant">
                    Hourly jobs do not use milestones — payment is tracked by
                    time.
                  </p>
                </div>
              )}

              <div className="rounded-2xl border border-outline-variant/20 bg-linear-to-br from-primary/10 via-surface to-secondary/10 p-4">
                <p className="text-[10px] font-extrabold uppercase tracking-[0.2em] text-on-surface-variant">
                  Posting summary
                </p>
                <div className="mt-3 space-y-2 text-sm">
                  <SummaryRow
                    label="Budget"
                    value={formatMoney(walletCurrency, budgetMinor)}
                  />
                  <SummaryRow
                    label="Milestone total"
                    value={formatMoney(walletCurrency, milestoneTotalMinor)}
                  />
                  <SummaryRow
                    label="Balance status"
                    value={
                      hasEnoughBalance
                        ? "Ready to post"
                        : `Need ${formatMoney(walletCurrency, balanceShortfallMinor)} more`
                    }
                    valueClassName={
                      hasEnoughBalance ? "text-emerald-600" : "text-red-600"
                    }
                  />
                </div>
                <p className="mt-3 text-xs text-on-surface-variant">
                  The milestone total must match the budget before the job can
                  be posted.
                </p>
              </div>

              <div className="flex items-center justify-end gap-3 pt-2">
                <button
                  type="button"
                  className="rounded-xl border border-outline-variant/20 px-4 py-3 text-sm font-bold text-on-surface-variant transition-colors hover:bg-surface-container"
                  onClick={() => {
                    setIsCreateOpen(false);
                    resetForm();
                  }}
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  className="inline-flex items-center gap-2 rounded-xl bg-primary px-5 py-3 text-sm font-bold text-white transition-all hover:shadow-lg hover:shadow-primary/20 disabled:cursor-not-allowed disabled:opacity-60"
                  disabled={isCreating}
                >
                  {isCreating ? "Creating..." : "Create Job"}
                </button>
              </div>
            </div>
          </form>
        ) : null}

        {/* Tab Selection Filter System */}
        <div className="border-b border-outline-variant/20 px-2 flex overflow-x-auto scrollbar-hide gap-2">
          <TabButton
            label="Open"
            count={openCount}
            active={activeTab === "open"}
            onClick={() => handleTabClick("open")}
          />
          <TabButton
            label="Closed"
            count={closedCount}
            active={activeTab === "closed"}
            onClick={() => handleTabClick("closed")}
          />
        </div>

        {/* Main Postings Content Stack */}
        <div className="flex flex-col gap-4">
          {isLoading ? (
            <div className="p-5 rounded-2xl border border-outline-variant/30 bg-surface-container-low text-on-surface-variant font-medium">
              Loading your jobs...
            </div>
          ) : isError ? (
            <div className="p-5 rounded-2xl border border-red-200 bg-red-50 text-red-700 font-medium">
              Failed to load your jobs.
            </div>
          ) : filteredPostings.length === 0 ? (
            <div className="p-5 rounded-2xl border border-outline-variant/30 bg-surface-container-low text-on-surface-variant font-medium">
              No jobs found for this tab.
            </div>
          ) : (
            paginatedPostings.map((job) => {
              const closed = normalizeStatus(job.status) === "CLOSED";
              const JobIcon = job.job_type === "HOURLY" ? Clock3 : Banknote;

              return (
                <div
                  key={job.id}
                  className={`p-5 rounded-2xl border flex flex-col md:flex-row md:items-center justify-between gap-6 transition-all group ${
                    closed
                      ? "bg-surface-container-low/40 border-dashed border-outline-variant/60"
                      : "bg-surface-container-lowest border-outline-variant/30 hover:shadow-md hover:border-outline-variant/80"
                  }`}
                >
                  {/* Left Side: Structural Identity Block */}
                  <div className="flex items-start gap-4 min-w-0">
                    <div
                      className={`w-12 h-12 rounded-xl flex items-center justify-center shrink-0 transition-colors ${
                        closed
                          ? "bg-surface-container-highest text-outline"
                          : "bg-primary/10 text-primary group-hover:bg-primary group-hover:text-white"
                      }`}
                    >
                      <JobIcon className="h-5 w-5" aria-hidden="true" />
                    </div>

                    <div
                      className="space-y-1 min-w-0 cursor-pointer"
                      onClick={() => {
                        router.push(`/client/mypostings/proposals/${job.id}`);
                      }}
                    >
                      <h3
                        className={`font-bold text-lg font-headline truncate transition-colors ${
                          closed
                            ? "text-on-surface-variant"
                            : "text-on-surface group-hover:text-primary"
                        }`}
                      >
                        {job.title}
                      </h3>
                      <div className="flex flex-wrap items-center gap-x-4 gap-y-1 text-xs font-medium text-on-surface-variant">
                        <span className="flex items-center gap-1">
                          <span className="material-symbols-outlined text-sm text-outline">
                            location_on
                          </span>{" "}
                          {job.location}
                        </span>
                        <span className="flex items-center gap-1">
                          <span className="material-symbols-outlined text-sm text-outline">
                            schedule
                          </span>{" "}
                          {getPostedTime(job.created_at)}
                        </span>
                      </div>
                    </div>
                  </div>

                  {/* Right Side: Requisition Status & Contextual Controls */}
                  <div className="flex items-center justify-between md:justify-end gap-6 lg:gap-10 border-t md:border-t-0 pt-4 md:pt-0 border-outline-variant/10">
                    {/* Numeric Metric Display */}
                    <div className="flex flex-col items-center min-w-16">
                      <span className="font-black text-xl font-headline tracking-tight text-on-surface">
                        {job.applications_count}
                      </span>
                      <span className="text-outline text-[10px] font-extrabold uppercase tracking-widest mt-0.5">
                        Applicants
                      </span>
                    </div>

                    {/* Badge Context */}
                    <div className="flex items-center gap-2 min-w-20 justify-center">
                      <span
                        className={`w-2 h-2 rounded-full ${closed ? "bg-outline" : "bg-emerald-500"}`}
                      />
                      <span
                        className={`font-extrabold text-xs uppercase tracking-wider ${closed ? "text-on-surface-variant" : "text-emerald-500 dark:text-emerald-400"}`}
                      >
                        {normalizeStatus(job.status)}
                      </span>
                    </div>

                    {/* Inline Interaction Hub */}
                    <div className="flex items-center gap-1">
                      {closed ? (
                        <button className="px-4 py-2 bg-surface-container-highest hover:bg-surface-container text-on-surface font-bold text-xs rounded-xl border border-outline-variant/20 transition-all">
                          Archived
                        </button>
                      ) : (
                        <>
                          <button
                            className="px-4 py-2 bg-surface text-red-500 font-bold text-xs rounded-xl border border-red-100 transition-all"
                            onClick={() => handleJobDelete(job.id)}
                          >
                            Delete job
                          </button>
                        </>
                      )}
                    </div>
                  </div>
                </div>
              );
            })
          )}
        </div>

        {/* Modular Segmented Pagination Bar */}
        <div className="mt-12 flex justify-center">
          <nav className="flex items-center gap-1.5 bg-surface-container-low p-1.5 rounded-2xl border border-outline-variant/20 shadow-xs">
            <button
              type="button"
              disabled={activePage === 1}
              onClick={() => setCurrentPage((page) => Math.max(1, page - 1))}
              className="p-2 rounded-xl text-outline hover:bg-surface-container hover:text-on-surface transition-all flex items-center disabled:opacity-40 disabled:hover:bg-transparent disabled:hover:text-outline"
            >
              <span className="material-symbols-outlined text-base">
                {"<<"}
              </span>
            </button>

            {pageNumbers.map((page, index) =>
              page === "ellipsis" ? (
                <span
                  key={`ellipsis-${index}`}
                  className="text-outline text-xs px-1 select-none"
                >
                  ...
                </span>
              ) : (
                <button
                  key={page}
                  type="button"
                  onClick={() => setCurrentPage(page)}
                  className={`w-9 h-9 rounded-xl font-bold text-xs transition-all ${
                    activePage === page
                      ? "bg-primary text-white shadow-xs"
                      : "text-on-surface-variant hover:bg-surface-container"
                  }`}
                >
                  {page}
                </button>
              ),
            )}

            <button
              type="button"
              disabled={activePage === totalPages}
              onClick={() =>
                setCurrentPage((page) => Math.min(totalPages, page + 1))
              }
              className="p-2 rounded-xl text-outline hover:bg-surface-container hover:text-on-surface transition-all flex items-center disabled:opacity-40 disabled:hover:bg-transparent disabled:hover:text-outline"
            >
              <span className="material-symbols-outlined text-base">
                {">>"}
              </span>
            </button>
          </nav>
        </div>
      </main>
    </div>
  );
}

function Field({
  label,
  error,
  children,
}: {
  label: string;
  error?: string;
  children: React.ReactNode;
}) {
  return (
    <label className="space-y-2">
      <span className="block text-sm font-bold text-on-surface">{label}</span>
      {children}
      {error ? (
        <span className="block text-xs font-medium text-red-600">{error}</span>
      ) : null}
    </label>
  );
}

function SummaryRow({
  label,
  value,
  valueClassName = "text-on-surface",
}: {
  label: string;
  value: string;
  valueClassName?: string;
}) {
  return (
    <div className="flex items-center justify-between gap-3 rounded-xl bg-surface-container-low px-3 py-2">
      <span className="text-xs font-semibold text-on-surface-variant">
        {label}
      </span>
      <span className={`text-sm font-black ${valueClassName}`}>{value}</span>
    </div>
  );
}

/* Local Utility Interface Subcomponents */
type TabButtonProps = {
  label: string;
  count: number;
  active: boolean;
  onClick: () => void;
};

function TabButton({ label, count, active, onClick }: TabButtonProps) {
  return (
    <button
      onClick={onClick}
      className={`flex items-center gap-2 px-5 py-3 border-b-2 transition-all font-headline font-bold text-sm outline-none whitespace-nowrap ${
        active
          ? "border-primary text-primary"
          : "border-transparent text-on-surface-variant hover:text-on-surface"
      }`}
    >
      <span>{label}</span>
      <span
        className={`text-[10px] font-extrabold px-2 py-0.5 rounded-md ${
          active
            ? "bg-primary/10 text-primary"
            : "bg-surface-container-highest text-on-surface-variant"
        }`}
      >
        {count}
      </span>
    </button>
  );
}
