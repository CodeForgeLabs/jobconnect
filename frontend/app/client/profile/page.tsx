"use client";

import Image from "next/image";
import Link from "next/link";
import { useEffect, useMemo, useRef, useState } from "react";
import {
  BriefcaseBusiness,
  Building2,
  CalendarDays,
  Edit3,
  Loader2,
  Mail,
  MapPin,
  Phone,
  Plus,
  Save,
  Upload,
  UserCircle2,
  X,
} from "lucide-react";

import { type Job, useGetMyJobsQuery } from "@/api/jobsapi";
import {
  useGetMeQuery,
  useUpdateMeMutation,
  useUploadImageMutation,
} from "@/api/userapi";

type ProfileForm = {
  first_name: string;
  last_name: string;
  company_name: string;
  headline: string;
  phone_number: string;
  location: string;
  bio: string;
};

const emptyForm: ProfileForm = {
  first_name: "",
  last_name: "",
  company_name: "",
  headline: "",
  phone_number: "",
  location: "",
  bio: "",
};

const formatDate = (value?: Date | string): string => {
  if (!value) return "Recently";

  const parsed = new Date(value);
  if (Number.isNaN(parsed.getTime())) return "Recently";

  return parsed.toLocaleDateString("en-US", {
    month: "short",
    day: "numeric",
    year: "numeric",
  });
};

const formatJobBudget = (job: Job) => {
  if (job.job_type === "HOURLY") {
    return `${Number(job.hourly_rate || 0).toLocaleString()} birr/hr`;
  }

  return `${Number(job.budget || 0).toLocaleString()} birr`;
};

const parseSkills = (skills: string) =>
  skills
    .split(",")
    .map((skill) => skill.trim())
    .filter(Boolean)
    .slice(0, 4);

export default function ClientProfile() {
  const fileInputRef = useRef<HTMLInputElement | null>(null);
  const { data: me, isLoading, refetch } = useGetMeQuery();
  const { data: jobs = [], isLoading: isLoadingJobs } = useGetMyJobsQuery();
  const [updateMe, { isLoading: isUpdating }] = useUpdateMeMutation();
  const [uploadImage, { isLoading: isUploading }] = useUploadImageMutation();
  const [isEditing, setIsEditing] = useState(false);
  const [form, setForm] = useState<ProfileForm>(emptyForm);
  const [message, setMessage] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!me) return;

    setForm({
      first_name: me.first_name || "",
      last_name: me.last_name || "",
      company_name: me.company_name || "",
      headline: me.headline || "",
      phone_number: me.phone_number || "",
      location: me.location || "",
      bio: me.bio || "",
    });
  }, [me]);

  const displayName =
    [me?.first_name, me?.last_name].filter(Boolean).join(" ") ||
    "Client Profile";
  const initials = displayName
    .split(" ")
    .map((part) => part[0])
    .slice(0, 2)
    .join("");

  const openJobCount = useMemo(
    () =>
      jobs.filter((job) => String(job.status).toUpperCase() === "OPEN").length,
    [jobs],
  );

  const privateJobCount = useMemo(
    () => jobs.filter((job) => job.is_private).length,
    [jobs],
  );

  const totalProposals = useMemo(
    () => jobs.reduce((total, job) => total + (job.applications_count ?? 0), 0),
    [jobs],
  );

  const recentJobs = jobs.slice(0, 6);

  const updateForm = (field: keyof ProfileForm, value: string) => {
    setForm((current) => ({ ...current, [field]: value }));
  };

  const handleCancel = () => {
    if (me) {
      setForm({
        first_name: me.first_name || "",
        last_name: me.last_name || "",
        company_name: me.company_name || "",
        headline: me.headline || "",
        phone_number: me.phone_number || "",
        location: me.location || "",
        bio: me.bio || "",
      });
    }

    setIsEditing(false);
    setError(null);
    setMessage(null);
  };

  const handleSave = async () => {
    setError(null);
    setMessage(null);

    if (!form.first_name.trim() || !form.last_name.trim()) {
      setError("First name and last name are required.");
      return;
    }

    try {
      await updateMe({
        first_name: form.first_name.trim(),
        last_name: form.last_name.trim(),
        company_name: form.company_name.trim(),
        headline: form.headline.trim(),
        phone_number: form.phone_number.trim(),
        location: form.location.trim(),
        bio: form.bio.trim(),
      }).unwrap();

      await refetch();
      setIsEditing(false);
      setMessage("Profile updated successfully.");
    } catch {
      setError("Unable to update your profile right now.");
    }
  };

  const handleFileChange = async (
    event: React.ChangeEvent<HTMLInputElement>,
  ) => {
    const file = event.target.files?.[0];
    if (!file) return;

    setError(null);
    setMessage(null);

    try {
      const uploadResult = await uploadImage(file).unwrap();
      await updateMe({
        profile_picture_url: uploadResult.secure_url,
      }).unwrap();
      await refetch();
      setMessage("Profile photo updated.");
    } catch {
      setError("Unable to upload your profile photo right now.");
    } finally {
      if (fileInputRef.current) fileInputRef.current.value = "";
    }
  };

  if (isLoading) {
    return (
      <main className="min-h-screen bg-surface px-6 py-12">
        <div className="mx-auto max-w-7xl space-y-6">
          <div className="h-56 animate-pulse rounded-3xl bg-slate-200/70" />
          <div className="grid gap-6 lg:grid-cols-[0.85fr_1.15fr]">
            <div className="h-96 animate-pulse rounded-3xl bg-slate-200/70" />
            <div className="h-96 animate-pulse rounded-3xl bg-slate-200/70" />
          </div>
        </div>
      </main>
    );
  }

  return (
    <main className="min-h-screen bg-surface px-5 py-12 text-on-surface md:px-8">
      <div className="mx-auto max-w-7xl space-y-8">
        <header className="rounded-3xl border border-outline-variant/20 bg-surface-container-lowest p-6 shadow-sm md:p-8">
          <div className="flex flex-col gap-6 lg:flex-row lg:items-end lg:justify-between">
            <div className="flex flex-col gap-5 md:flex-row md:items-center">
              <div className="relative h-28 w-28 shrink-0 overflow-hidden rounded-3xl bg-surface-container-high ring-4 ring-white">
                {me?.profile_picture_url ? (
                  <Image
                    src={me.profile_picture_url}
                    alt={`${displayName} profile photo`}
                    width={112}
                    height={112}
                    className="h-full w-full object-cover"
                  />
                ) : (
                  <div className="grid h-full w-full place-items-center bg-linear-to-br from-primary to-primary-container text-3xl font-black text-on-primary">
                    {initials || <UserCircle2 className="h-12 w-12" />}
                  </div>
                )}
                <button
                  type="button"
                  onClick={() => fileInputRef.current?.click()}
                  disabled={isUploading || isUpdating}
                  className="absolute bottom-2 right-2 grid h-9 w-9 place-items-center rounded-full bg-white text-primary shadow-md transition hover:bg-surface-container-low disabled:cursor-not-allowed disabled:opacity-60"
                  aria-label="Upload profile photo"
                >
                  {isUploading ? (
                    <Loader2 className="h-4 w-4 animate-spin" />
                  ) : (
                    <Upload className="h-4 w-4" />
                  )}
                </button>
                <input
                  ref={fileInputRef}
                  type="file"
                  accept="image/*"
                  className="hidden"
                  onChange={handleFileChange}
                />
              </div>

              <div>
                <p className="text-xs font-bold uppercase tracking-[0.24em] text-primary">
                  Client Profile
                </p>
                <h1 className="mt-2 text-4xl font-extrabold tracking-tight text-on-surface md:text-5xl">
                  {displayName}
                </h1>
                <p className="mt-2 max-w-2xl text-sm leading-7 text-on-surface-variant md:text-base">
                  {me?.headline ||
                    "Add a short client headline so freelancers understand what kind of work you hire for."}
                </p>
              </div>
            </div>

            <div className="flex flex-wrap gap-3">
              {isEditing ? (
                <>
                  <button
                    type="button"
                    onClick={handleCancel}
                    disabled={isUpdating}
                    className="inline-flex items-center gap-2 rounded-xl border border-outline-variant/30 bg-white px-5 py-3 text-sm font-bold text-on-surface-variant transition hover:bg-surface-container-low disabled:cursor-not-allowed disabled:opacity-60"
                  >
                    <X className="h-4 w-4" />
                    Cancel
                  </button>
                  <button
                    type="button"
                    onClick={() => void handleSave()}
                    disabled={isUpdating}
                    className="inline-flex items-center gap-2 rounded-xl bg-primary px-5 py-3 text-sm font-bold text-on-primary transition hover:opacity-90 disabled:cursor-not-allowed disabled:opacity-60"
                  >
                    {isUpdating ? (
                      <Loader2 className="h-4 w-4 animate-spin" />
                    ) : (
                      <Save className="h-4 w-4" />
                    )}
                    Save profile
                  </button>
                </>
              ) : (
                <button
                  type="button"
                  onClick={() => setIsEditing(true)}
                  className="inline-flex items-center gap-2 rounded-xl bg-primary px-5 py-3 text-sm font-bold text-on-primary transition hover:opacity-90"
                >
                  <Edit3 className="h-4 w-4" />
                  Edit profile
                </button>
              )}
              <Link
                href="/client/mypostings"
                className="inline-flex items-center gap-2 rounded-xl bg-secondary-container px-5 py-3 text-sm font-bold text-on-secondary-container transition hover:opacity-90"
              >
                <Plus className="h-4 w-4" />
                Post job
              </Link>
            </div>
          </div>
        </header>

        {(message || error) && (
          <div
            className={`rounded-2xl border px-4 py-3 text-sm font-semibold ${
              error
                ? "border-rose-200 bg-rose-50 text-rose-700"
                : "border-emerald-200 bg-emerald-50 text-emerald-700"
            }`}
            role="status"
          >
            {error || message}
          </div>
        )}

        <section className="grid gap-6 md:grid-cols-3">
          <div className="rounded-3xl border border-outline-variant/20 bg-surface-container-lowest p-6 shadow-sm">
            <p className="text-xs font-bold uppercase tracking-[0.2em] text-on-surface-variant">
              Jobs Posted
            </p>
            <p className="mt-3 text-4xl font-black text-on-surface">
              {jobs.length}
            </p>
          </div>
          <div className="rounded-3xl border border-outline-variant/20 bg-surface-container-lowest p-6 shadow-sm">
            <p className="text-xs font-bold uppercase tracking-[0.2em] text-on-surface-variant">
              Open Jobs
            </p>
            <p className="mt-3 text-4xl font-black text-on-surface">
              {openJobCount}
            </p>
          </div>
          <div className="rounded-3xl border border-outline-variant/20 bg-surface-container-lowest p-6 shadow-sm">
            <p className="text-xs font-bold uppercase tracking-[0.2em] text-on-surface-variant">
              Total Proposals
            </p>
            <p className="mt-3 text-4xl font-black text-on-surface">
              {totalProposals}
            </p>
          </div>
        </section>

        <div className="grid gap-8 lg:grid-cols-[0.85fr_1.15fr]">
          <section className="space-y-6">
            <div className="rounded-3xl border border-outline-variant/20 bg-surface-container-lowest p-6 shadow-sm">
              <div className="flex items-center justify-between gap-4">
                <h2 className="text-xl font-bold text-on-surface">
                  Profile information
                </h2>
                <span className="rounded-full bg-primary/10 px-3 py-1 text-xs font-bold text-primary">
                  {privateJobCount} private jobs
                </span>
              </div>

              <div className="mt-6 space-y-4">
                <div className="grid gap-4 sm:grid-cols-2">
                  <EditableField
                    label="First name"
                    value={form.first_name}
                    readValue={me?.first_name || "N/A"}
                    isEditing={isEditing}
                    onChange={(value) => updateForm("first_name", value)}
                  />
                  <EditableField
                    label="Last name"
                    value={form.last_name}
                    readValue={me?.last_name || "N/A"}
                    isEditing={isEditing}
                    onChange={(value) => updateForm("last_name", value)}
                  />
                </div>
                <EditableField
                  label="Company"
                  value={form.company_name}
                  readValue={me?.company_name || "Add company name"}
                  isEditing={isEditing}
                  onChange={(value) => updateForm("company_name", value)}
                  icon={Building2}
                />
                <EditableField
                  label="Headline"
                  value={form.headline}
                  readValue={me?.headline || "Add client headline"}
                  isEditing={isEditing}
                  onChange={(value) => updateForm("headline", value)}
                  icon={BriefcaseBusiness}
                />
                <EditableField
                  label="Email"
                  value={me?.email || ""}
                  readValue={me?.email || "N/A"}
                  isEditing={false}
                  onChange={() => undefined}
                  icon={Mail}
                />
                <EditableField
                  label="Phone"
                  value={form.phone_number}
                  readValue={me?.phone_number || "Add phone number"}
                  isEditing={isEditing}
                  onChange={(value) => updateForm("phone_number", value)}
                  icon={Phone}
                />
                <EditableField
                  label="Location"
                  value={form.location}
                  readValue={me?.location || "Add location"}
                  isEditing={isEditing}
                  onChange={(value) => updateForm("location", value)}
                  icon={MapPin}
                />
              </div>
            </div>

            <div className="rounded-3xl border border-outline-variant/20 bg-surface-container-lowest p-6 shadow-sm">
              <h2 className="text-xl font-bold text-on-surface">Profile bio</h2>
              {isEditing ? (
                <textarea
                  value={form.bio}
                  onChange={(event) => updateForm("bio", event.target.value)}
                  rows={7}
                  className="mt-4 w-full resize-none rounded-2xl border border-outline-variant/30 bg-white px-4 py-3 text-sm leading-7 text-on-surface outline-none transition focus:border-primary focus:ring-2 focus:ring-primary/15"
                  placeholder="Describe your company, hiring style, project needs, and what freelancers can expect when working with you."
                />
              ) : (
                <p className="mt-4 whitespace-pre-wrap text-sm leading-7 text-on-surface-variant">
                  {me?.bio ||
                    "Add a client bio to introduce your company, hiring needs, communication style, and project expectations."}
                </p>
              )}
            </div>
          </section>

          <section className="rounded-3xl border border-outline-variant/20 bg-surface-container-lowest p-6 shadow-sm">
            <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
              <div>
                <p className="text-xs font-bold uppercase tracking-[0.2em] text-primary">
                  Jobs Posted
                </p>
                <h2 className="mt-2 text-2xl font-bold text-on-surface">
                  Your recent postings
                </h2>
              </div>
              <Link
                href="/client/mypostings"
                className="inline-flex items-center justify-center rounded-xl border border-outline-variant/30 px-4 py-2.5 text-sm font-bold text-on-surface-variant transition hover:bg-surface-container-low"
              >
                View all jobs
              </Link>
            </div>

            {isLoadingJobs ? (
              <div className="mt-6 rounded-2xl border border-dashed border-outline-variant/30 bg-surface-container-low p-8 text-center">
                <Loader2 className="mx-auto h-6 w-6 animate-spin text-primary" />
                <p className="mt-3 text-sm font-semibold text-on-surface-variant">
                  Loading posted jobs...
                </p>
              </div>
            ) : recentJobs.length > 0 ? (
              <div className="mt-6 space-y-4">
                {recentJobs.map((job) => (
                  <article
                    key={job.id}
                    className="rounded-2xl border border-outline-variant/20 bg-surface-container-low p-5"
                  >
                    <div className="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
                      <div className="min-w-0">
                        <div className="flex flex-wrap items-center gap-2">
                          <h3 className="text-lg font-bold text-on-surface">
                            {job.title}
                          </h3>
                          <span className="rounded-full bg-white px-3 py-1 text-[10px] font-black uppercase tracking-widest text-primary">
                            {job.is_private ? "Private" : "Public"}
                          </span>
                        </div>
                        <p className="mt-2 line-clamp-2 text-sm leading-6 text-on-surface-variant">
                          {job.description}
                        </p>
                      </div>
                      <span className="shrink-0 rounded-full bg-primary px-3 py-1 text-xs font-bold uppercase tracking-wider text-on-primary">
                        {job.status}
                      </span>
                    </div>

                    <div className="mt-4 flex flex-wrap gap-2 text-xs font-semibold text-on-surface-variant">
                      <span className="rounded-full bg-white px-3 py-1">
                        {job.job_type}
                      </span>
                      <span className="rounded-full bg-white px-3 py-1">
                        {formatJobBudget(job)}
                      </span>
                      <span className="inline-flex items-center gap-1 rounded-full bg-white px-3 py-1">
                        <CalendarDays className="h-3.5 w-3.5" />
                        {formatDate(job.created_at)}
                      </span>
                      <span className="rounded-full bg-white px-3 py-1">
                        {job.applications_count ?? 0} proposals
                      </span>
                    </div>

                    {job.skills ? (
                      <div className="mt-3 flex flex-wrap gap-2">
                        {parseSkills(job.skills).map((skill) => (
                          <span
                            key={`${job.id}-${skill}`}
                            className="rounded-full bg-primary/10 px-3 py-1 text-[11px] font-bold text-primary"
                          >
                            {skill}
                          </span>
                        ))}
                      </div>
                    ) : null}
                  </article>
                ))}
              </div>
            ) : (
              <div className="mt-6 rounded-2xl border border-dashed border-outline-variant/30 bg-surface-container-low p-8 text-center">
                <BriefcaseBusiness className="mx-auto h-10 w-10 text-primary" />
                <h3 className="mt-4 text-lg font-bold text-on-surface">
                  No jobs posted yet
                </h3>
                <p className="mx-auto mt-2 max-w-md text-sm leading-6 text-on-surface-variant">
                  Post your first job to start receiving proposals and build out
                  your client profile.
                </p>
                <Link
                  href="/client/mypostings"
                  className="mt-5 inline-flex items-center gap-2 rounded-xl bg-primary px-5 py-3 text-sm font-bold text-on-primary transition hover:opacity-90"
                >
                  <Plus className="h-4 w-4" />
                  Post a job
                </Link>
              </div>
            )}
          </section>
        </div>
      </div>
    </main>
  );
}

function EditableField({
  label,
  value,
  readValue,
  isEditing,
  onChange,
  icon: Icon,
}: {
  label: string;
  value: string;
  readValue: string;
  isEditing: boolean;
  onChange: (value: string) => void;
  icon?: typeof UserCircle2;
}) {
  return (
    <div className="rounded-2xl border border-outline-variant/20 bg-surface-container-low p-4">
      <div className="flex items-center gap-2">
        {Icon ? <Icon className="h-4 w-4 text-primary" /> : null}
        <p className="text-xs font-bold uppercase tracking-[0.18em] text-on-surface-variant">
          {label}
        </p>
      </div>
      {isEditing ? (
        <input
          value={value}
          onChange={(event) => onChange(event.target.value)}
          className="mt-3 w-full rounded-xl border border-outline-variant/30 bg-white px-3 py-2 text-sm font-semibold text-on-surface outline-none transition focus:border-primary focus:ring-2 focus:ring-primary/15"
        />
      ) : (
        <p className="mt-3 break-words text-sm font-semibold text-on-surface">
          {readValue}
        </p>
      )}
    </div>
  );
}
