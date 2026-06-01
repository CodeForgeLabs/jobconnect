"use client";

import { useRouter } from "next/navigation";
import { BarChart3, Brain, ChevronDown, Clock3, Search, X } from "lucide-react";
import { KeyboardEvent, useMemo, useState } from "react";
import Jobcard from "@/components/Jobcard";
import {
  useGetJobsQuery,
  useGetMyJobsQuery,
  type JobFilterParams,
} from "@/api/jobsapi";

const formatPostedTime = (createdAt?: string) => {
  if (!createdAt) return "recently";

  const parsedDate = new Date(createdAt);
  if (Number.isNaN(parsedDate.getTime())) return "recently";

  const diffInSeconds = Math.floor((Date.now() - parsedDate.getTime()) / 1000);

  if (diffInSeconds < 60) return "just now";
  if (diffInSeconds < 3600) return `${Math.floor(diffInSeconds / 60)} min ago`;
  if (diffInSeconds < 86400)
    return `${Math.floor(diffInSeconds / 3600)} hrs ago`;
  return `${Math.floor(diffInSeconds / 86400)} days ago`;
};

const parseSkills = (skills: string) =>
  skills
    .split(",")
    .map((skill) => skill.trim())
    .filter(Boolean);

const Jobsearch = () => {
  const router = useRouter();
  const [skills, setSkills] = useState<string[]>([]);
  const [skillInput, setSkillInput] = useState("");
  const [searchText, setSearchText] = useState("");
  const [appliedSearchText, setAppliedSearchText] = useState("");
  const [jobType, setJobType] = useState<string | undefined>(undefined);
  const [experienceLevel, setExperienceLevel] = useState<string | undefined>(
    undefined,
  );
  const [budgetMin, setBudgetMin] = useState("");
  const [showMyJobs, setShowMyJobs] = useState(false);

  const addSkill = (value: string) => {
    const trimmedValue = value.trim();

    if (!trimmedValue) return;

    const isDuplicate = skills.some(
      (skill) => skill.toLowerCase() === trimmedValue.toLowerCase(),
    );

    if (!isDuplicate) {
      setSkills((prevSkills) => [...prevSkills, trimmedValue]);
    }

    setSkillInput("");
  };

  const removeSkill = (skillToRemove: string) => {
    setSkills((prevSkills) =>
      prevSkills.filter((skill) => skill !== skillToRemove),
    );
  };

  const handleSkillKeyDown = (event: KeyboardEvent<HTMLInputElement>) => {
    if (event.key === "Enter" || event.key === ",") {
      event.preventDefault();
      addSkill(skillInput);
    }
  };

  const handleApplySearch = () => {
    setAppliedSearchText(searchText.trim());
  };

  const handleClearAllFilters = () => {
    setSkills([]);
    setSkillInput("");
    setSearchText("");
    setAppliedSearchText("");
    setJobType(undefined);
    setExperienceLevel(undefined);
    setBudgetMin("");
  };

  const filters = useMemo<JobFilterParams>(() => {
    const nextFilters: JobFilterParams = {};

    if (appliedSearchText) nextFilters.title = appliedSearchText;
    if (jobType) nextFilters.job_type = jobType;
    if (experienceLevel) nextFilters.experience_level = experienceLevel;
    if (budgetMin && !Number.isNaN(Number(budgetMin))) {
      nextFilters.budget_min = Number(budgetMin);
    }

    return nextFilters;
  }, [appliedSearchText, jobType, experienceLevel, budgetMin]);

  const {
    data: allJobs = [],
    isLoading: isLoadingAllJobs,
    isFetching: isFetchingAllJobs,
    isError: isAllJobsError,
  } = useGetJobsQuery(filters);

  const {
    data: myJobs = [],
    isLoading: isLoadingMyJobs,
    isFetching: isFetchingMyJobs,
    isError: isMyJobsError,
  } = useGetMyJobsQuery(undefined, { skip: !showMyJobs });

  const visibleJobs = useMemo(() => {
    const baseJobs = showMyJobs ? myJobs : allJobs;

    if (skills.length === 0) return baseJobs;

    const loweredSkills = skills.map((skill) => skill.toLowerCase());

    return baseJobs.filter((job) => {
      const tags = parseSkills(job.skills).map((skill) => skill.toLowerCase());
      return loweredSkills.every((skill) =>
        tags.some((tag) => tag.includes(skill)),
      );
    });
  }, [allJobs, myJobs, showMyJobs, skills]);

  const isLoading = showMyJobs ? isLoadingMyJobs : isLoadingAllJobs;
  const isFetching = showMyJobs ? isFetchingMyJobs : isFetchingAllJobs;
  const isError = showMyJobs ? isMyJobsError : isAllJobsError;

  return (
    <div className=" px-10 py-10 flex flex-col lg:flex-row gap-8 min-h-screen bg-surface">
      {/* Sidebar Filters */}
      <div className="w-full lg:w-80 shrink-0 h-fit bg-white border border-gray-100 p-6 rounded-2xl shadow-xs flex flex-col gap-6">
        <div className="flex justify-between items-center pb-2 border-b border-gray-100">
          <p className="text-xs font-bold uppercase tracking-wider text-primary">
            Advanced Filters
          </p>
          <button
            type="button"
            onClick={handleClearAllFilters}
            className="text-jobBlue hover:text-jobBlue/80 text-xs font-semibold transition-colors"
          >
            Clear all
          </button>
        </div>

        {/* Job Type Filter */}
        <div className="flex flex-col gap-2.5">
          <p className="flex items-center gap-2 font-bold text-xs uppercase tracking-wider text-gray-400">
            <Clock3 className="h-3.5 w-3.5 text-gray-400" />
            Job Type
          </p>
          <div className="flex flex-col gap-2">
            <label className="flex items-center gap-3 cursor-pointer py-1 group select-none">
              <input
                type="radio"
                name="job-type"
                checked={jobType === undefined}
                onChange={() => setJobType(undefined)}
                className="radio radio-sm radio-primary transition-all"
              />
              <span className="text-sm font-medium text-gray-600 group-hover:text-gray-900 transition-colors">
                Any
              </span>
            </label>
            <label className="flex items-center gap-3 cursor-pointer py-1 group select-none">
              <input
                type="radio"
                name="job-type"
                checked={jobType === "FIXED"}
                onChange={() => setJobType("FIXED")}
                className="radio radio-sm radio-primary transition-all"
              />
              <span className="text-sm font-medium text-gray-600 group-hover:text-gray-900 transition-colors">
                Fixed rate
              </span>
            </label>
            <label className="flex items-center gap-3 cursor-pointer py-1 group select-none">
              <input
                type="radio"
                name="job-type"
                checked={jobType === "HOURLY"}
                onChange={() => setJobType("HOURLY")}
                className="radio radio-sm radio-primary transition-all"
              />
              <span className="text-sm font-medium text-gray-600 group-hover:text-gray-900 transition-colors">
                Hourly rate
              </span>
            </label>
          </div>
        </div>

        <hr className="border-gray-100" />

        {/* Experience Level Filter */}
        <div className="flex flex-col gap-2.5">
          <p className="flex items-center gap-2 font-bold text-xs uppercase tracking-wider text-gray-400">
            <BarChart3 className="h-3.5 w-3.5 text-gray-400" />
            Experience
          </p>
          <div className="flex flex-col gap-2">
            <label className="flex items-center gap-3 cursor-pointer py-1 group select-none">
              <input
                type="radio"
                name="experience"
                checked={experienceLevel === undefined}
                onChange={() => setExperienceLevel(undefined)}
                className="radio radio-sm radio-primary transition-all"
              />
              <span className="text-sm font-medium text-gray-600 group-hover:text-gray-900 transition-colors">
                Any
              </span>
            </label>
            <label className="flex items-center gap-3 cursor-pointer py-1 group select-none">
              <input
                type="radio"
                name="experience"
                checked={experienceLevel === "ENTRY"}
                onChange={() => setExperienceLevel("ENTRY")}
                className="radio radio-sm radio-primary transition-all"
              />
              <span className="text-sm font-medium text-gray-600 group-hover:text-gray-900 transition-colors">
                Entry Level
              </span>
            </label>
            <label className="flex items-center gap-3 cursor-pointer py-1 group select-none">
              <input
                type="radio"
                name="experience"
                checked={experienceLevel === "INTERMEDIATE"}
                onChange={() => setExperienceLevel("INTERMEDIATE")}
                className="radio radio-sm radio-primary transition-all"
              />
              <span className="text-sm font-medium text-gray-600 group-hover:text-gray-900 transition-colors">
                Intermediate
              </span>
            </label>
            <label className="flex items-center gap-3 cursor-pointer py-1 group select-none">
              <input
                type="radio"
                name="experience"
                checked={experienceLevel === "EXPERT"}
                onChange={() => setExperienceLevel("EXPERT")}
                className="radio radio-sm radio-primary transition-all"
              />
              <span className="text-sm font-medium text-gray-600 group-hover:text-gray-900 transition-colors">
                Expert
              </span>
            </label>
          </div>
        </div>

        <hr className="border-gray-100" />

        {/* Budget Filter */}
        <div className="flex flex-col gap-2.5">
          <p className="flex items-center gap-2 font-bold text-xs uppercase tracking-wider text-gray-400">
            <BarChart3 className="h-3.5 w-3.5 text-gray-400" />
            Budget
          </p>
          <div className="flex items-center gap-3">
            <div className="relative w-full">
              <input
                type="number"
                placeholder="Min"
                value={budgetMin}
                onChange={(event) => setBudgetMin(event.target.value)}
                className="input input-sm input-bordered w-full rounded-xl bg-surface focus:outline-hidden transition-all"
              />
            </div>
            <div className="relative w-full">
              <input
                type="number"
                placeholder="Max"
                disabled
                className="input input-sm input-bordered w-full rounded-xl bg-gray-50 text-gray-400 opacity-60 cursor-not-allowed select-none"
              />
            </div>
          </div>
          <div className="p-3 rounded-xl bg-surface-container-low border border-outline-variant/10">
            <p className="text-[11px] text-gray-400 leading-normal">
              Max budget constraints are currently un-supported by the active
              API route layer.
            </p>
          </div>
        </div>

        <hr className="border-gray-100" />

        {/* Skills Filter */}
        <div className="flex flex-col gap-2.5">
          <p className="flex items-center gap-2 font-bold text-xs uppercase tracking-wider text-gray-400">
            <Brain className="h-3.5 w-3.5 text-gray-400" />
            Skills Filter
          </p>

          {skills.length > 0 && (
            <div className="flex flex-wrap gap-1.5 p-2 rounded-xl bg-surface max-h-32 overflow-y-auto border border-gray-100">
              {skills.map((skill) => (
                <span
                  key={skill}
                  className="inline-flex items-center gap-1.5 rounded-lg bg-primary/10 pl-2.5 pr-1.5 py-1 text-xs font-semibold text-primary transition-all animate-fadeIn"
                >
                  {skill}
                  <button
                    type="button"
                    onClick={() => removeSkill(skill)}
                    className="rounded-md p-0.5 hover:bg-primary/20 text-primary transition-colors"
                    aria-label={`Remove ${skill}`}
                  >
                    <X className="h-3 w-3" />
                  </button>
                </span>
              ))}
            </div>
          )}

          <input
            type="text"
            value={skillInput}
            onChange={(event) => setSkillInput(event.target.value)}
            onKeyDown={handleSkillKeyDown}
            onBlur={() => addSkill(skillInput)}
            placeholder="Type skill & hit Enter"
            className="input input-sm input-bordered w-full rounded-xl bg-surface focus:outline-hidden transition-all"
          />
        </div>
      </div>

      {/* Main Content Area */}
      <div className="flex-1 flex flex-col gap-6">
        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
          <h1 className="text-3xl font-extrabold tracking-tight text-on-surface">
            Find Work
          </h1>

          {/* Track Switch Toggle */}
          <div className="bg-gray-100 p-1 rounded-xl flex items-center gap-1 w-fit border border-gray-200/40">
            <button
              type="button"
              onClick={() => setShowMyJobs(false)}
              className={`px-4 py-1.5 rounded-lg text-xs font-bold tracking-wide transition-all ${
                !showMyJobs
                  ? "bg-white text-primary shadow-xs"
                  : "text-gray-500 hover:text-gray-900 bg-transparent"
              }`}
            >
              Search All Jobs
            </button>
            <button
              type="button"
              onClick={() => setShowMyJobs(true)}
              className={`px-4 py-1.5 rounded-lg text-xs font-bold tracking-wide transition-all ${
                showMyJobs
                  ? "bg-white text-primary shadow-xs"
                  : "text-gray-500 hover:text-gray-900 bg-transparent"
              }`}
            >
              My Jobs
            </button>
          </div>
        </div>

        {/* Search Input Bar Control */}
        <div className="flex w-full items-center rounded-2xl border border-gray-200 bg-white p-1.5 shadow-xs focus-within:ring-2 focus-within:ring-primary/20 focus-within:border-primary transition-all">
          <div className="pl-3.5 flex items-center justify-center shrink-0">
            <Search className="h-4 w-4 text-gray-400" />
          </div>
          <input
            type="text"
            value={searchText}
            onChange={(event) => setSearchText(event.target.value)}
            onKeyDown={(event) => {
              if (event.key === "Enter") {
                event.preventDefault();
                handleApplySearch();
              }
            }}
            placeholder="Search jobs, required stack components, or workflow tags..."
            className="ml-3 w-full bg-transparent text-sm text-gray-800 outline-none py-2.5 placeholder:text-gray-400"
          />
          <button
            type="button"
            onClick={handleApplySearch}
            className="shrink-0 text-sm font-bold text-white bg-primary hover:bg-primary-container py-2.5 px-6 rounded-xl transition-all shadow-xs"
          >
            Search
          </button>
        </div>

        {/* Summary Metric Header Block */}
        <div className="flex justify-between items-center bg-white/40 px-2 py-1 rounded-xl">
          <p className="text-gray-500 text-sm font-medium">
            {isLoading ? (
              <span className="inline-flex items-center gap-2 text-primary font-semibold">
                <span className="loading loading-spinner loading-xs"></span>{" "}
                Loading matches...
              </span>
            ) : (
              <span>
                Available items:{" "}
                <span className="font-bold text-on-surface">
                  {visibleJobs.length} matches found
                </span>
              </span>
            )}
            {isFetching && !isLoading ? (
              <span className="text-xs text-primary ml-1 animate-pulse">
                (updating pipeline...)
              </span>
            ) : (
              ""
            )}
          </p>

          <div className="dropdown dropdown-bottom dropdown-end">
            <div
              tabIndex={0}
              role="button"
              className="btn btn-sm bg-white hover:bg-gray-50 border border-gray-200 text-xs font-bold text-gray-700 rounded-xl gap-2 shadow-xs normal-case"
            >
              Sort by: Newest{" "}
              <ChevronDown className="h-3.5 w-3.5 text-gray-500" />
            </div>
            <ul
              tabIndex={-1}
              className="dropdown-content menu bg-base-100 rounded-2xl z-10 w-48 p-2 mt-1 border border-gray-100 shadow-md"
            >
              <li>
                <a className="text-xs font-semibold py-2">Newest first</a>
              </li>
            </ul>
          </div>
        </div>

        {/* Dynamic Job List Container */}
        <div className="flex flex-col gap-4">
          {isError ? (
            <div className="rounded-2xl border border-red-200 bg-red-50 p-4 text-sm font-medium text-red-700 shadow-xs animate-fadeIn">
              System communication bottleneck failed to parse remote collection.
              Please re-query structural view.
            </div>
          ) : null}

          {!isLoading && !isError && visibleJobs.length === 0 ? (
            <div className="rounded-2xl border border-gray-100 bg-white p-12 text-center shadow-xs">
              <p className="text-sm font-semibold text-gray-700">
                No matching open scopes located
              </p>
              <p className="text-xs text-gray-400 mt-1 max-w-xs mx-auto">
                Modify your current tag selection parameters or criteria filters
                to explore wider directory pipelines.
              </p>
            </div>
          ) : null}

          {!isError &&
            visibleJobs.map((job) => (
              <div
                key={job.id}
                onClick={() => router.push(`/freelancer/job/${job.id}`)}
                className="group block rounded-2xl bg-white border border-gray-100 hover:border-gray-200 shadow-xs hover:shadow-sm transition-all duration-200 cursor-pointer overflow-hidden relative"
              >
                <Jobcard
                  title={job.title}
                  pay={String(
                    job.job_type === "HOURLY" ? job.hourly_rate : job.budget,
                  )}
                  type={job.job_type === "HOURLY" ? "HOURLY" : "FIXED"}
                  description={job.description}
                  postTime={job.created_at}
                  tags={parseSkills(job.skills)}
                  companyName={job.company_name || "Unknown client"}
                  status={job.status}
                  skills={job.skills}
                  jobType={job.job_type === "HOURLY" ? "HOURLY" : "FIXED"}
                  hourlyRate={String(job.hourly_rate)}
                  budget={String(job.budget)}
                  experienceLevel={job.experience_level}
                  createdAt={job.created_at}
                  index={visibleJobs.indexOf(job)}
                  onApply={() => router.push(`/freelancer/job/${job.id}`)}
                />
              </div>
            ))}
        </div>
      </div>
    </div>
  );
};

export default Jobsearch;
