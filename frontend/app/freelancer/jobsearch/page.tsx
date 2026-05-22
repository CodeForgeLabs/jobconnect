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

  
  const visibleJobs = useMemo(() => {
    const baseJobs =  allJobs;

    if (skills.length === 0) return baseJobs;

    const loweredSkills = skills.map((skill) => skill.toLowerCase());

    return baseJobs.filter((job) => {
      const tags = parseSkills(job.skills).map((skill) => skill.toLowerCase());
      return loweredSkills.every((skill) =>
        tags.some((tag) => tag.includes(skill)),
      );
    });
  }, [ allJobs, skills]);

  const isLoading = isLoadingAllJobs;
  const isFetching = isFetchingAllJobs;
  const isError =  isAllJobsError;

  return (
    <div className="flex p-14 bg-[#ebedf1] gap-8">
      <div className=" flex flex-col gap-4 w-[30%] h-fit bg-white p-5 rounded-lg shadow-md   ">
        <div className="flex justify-between items-center">
          <p> Advanced Filters</p>
          <button
            type="button"
            onClick={handleClearAllFilters}
            className="text-jobBlue text-xs"
          >
            clear all
          </button>
        </div>

        <div className="flex flex-col">
          <p className="flex  items-center gap-2 font-medium text-sm mb-2">
            <Clock3 className="h-4 w-4" />
            job type
          </p>
          <span>
            <input
              type="radio"
              name="job-type"
              checked={jobType === undefined}
              onChange={() => setJobType(undefined)}
              className="radio radio-sm"
            />
            <label className="ml-2 text-[12px] text-gray-500">Any</label>
          </span>
          <span>
            <input
              type="radio"
              name="job-type"
              checked={jobType === "FIXED"}
              onChange={() => setJobType("FIXED")}
              className="radio radio-sm radio-primary"
            />
            <label className="ml-2 text-[12px] text-gray-500">Fixed rate</label>
          </span>
          <span>
            <input
              type="radio"
              name="job-type"
              checked={jobType === "HOURLY"}
              onChange={() => setJobType("HOURLY")}
              className="radio radio-sm radio-primary"
            />
            <label className="ml-2 text-[12px] text-gray-500">
              Hourly rate
            </label>
          </span>
        </div>

        <div className="flex flex-col">
          <p className="flex  items-center gap-2 font-medium text-sm mb-2">
            <BarChart3 className="h-4 w-4" />
            Experience
          </p>
          <span>
            <input
              type="radio"
              name="experience"
              checked={experienceLevel === undefined}
              onChange={() => setExperienceLevel(undefined)}
              className="radio radio-sm"
            />
            <label className="ml-2 text-[12px] text-gray-500">Any</label>
          </span>
          <span>
            <input
              type="radio"
              name="experience"
              checked={experienceLevel === "ENTRY"}
              onChange={() => setExperienceLevel("ENTRY")}
              className="radio radio-sm radio-primary"
            />
            <label className="ml-2 text-[12px] text-gray-500">
              Entry Level
            </label>
          </span>
          <span>
            <input
              type="radio"
              name="experience"
              checked={experienceLevel === "INTERMEDIATE"}
              onChange={() => setExperienceLevel("INTERMEDIATE")}
              className="radio radio-sm radio-primary"
            />
            <label className="ml-2 text-[12px] text-gray-500">
              Intermediate{" "}
            </label>
          </span>
          <span>
            <input
              type="radio"
              name="experience"
              checked={experienceLevel === "EXPERT"}
              onChange={() => setExperienceLevel("EXPERT")}
              className="radio radio-sm radio-primary"
            />
            <label className="ml-2 text-[12px] text-gray-500">Expert</label>
          </span>
        </div>

        <div>
          <p className="flex  items-center gap-2 font-medium text-sm mb-2">
            <BarChart3 className="h-4 w-4" />
            Budget
          </p>
          <div className="flex items-center gap-2">
            <input
              type="number"
              placeholder="Min"
              value={budgetMin}
              onChange={(event) => setBudgetMin(event.target.value)}
              className="input input-sm input-bordered w-full max-w-xs"
            />
            <input
              type="number"
              placeholder="Max"
              disabled
              className="input input-sm input-bordered w-full max-w-xs"
            />
          </div>
          <p className="mt-1 text-[11px] text-gray-400">
            Max budget is not supported by the current API.
          </p>
        </div>

        <div>
          <p className="flex  items-center gap-2 font-medium text-sm mb-2">
            <Brain className="h-4 w-4" />
            Skills
          </p>

          <div className="mb-2 flex flex-wrap gap-2">
            {skills.map((skill) => (
              <span
                key={skill}
                className="inline-flex items-center gap-1 rounded-full bg-blue-100 px-2 py-1 text-xs text-blue-700"
              >
                {skill}
                <button
                  type="button"
                  onClick={() => removeSkill(skill)}
                  className="rounded-full p-0.5 hover:bg-blue-200"
                  aria-label={`Remove ${skill}`}
                >
                  <X className="h-3 w-3" />
                </button>
              </span>
            ))}
          </div>

          <input
            type="text"
            value={skillInput}
            onChange={(event) => setSkillInput(event.target.value)}
            onKeyDown={handleSkillKeyDown}
            onBlur={() => addSkill(skillInput)}
            placeholder="Type a skill and press Enter"
            className="input input-sm input-bordered w-full"
          />
        </div>
      </div>

      <div className=" w-full ">
        <h1 className="text-3xl font-bold">Find Work</h1>

        <div className="mt-4 flex items-center gap-2">
          <button
            type="button"
            onClick={() => setShowMyJobs(false)}
            className={`btn btn-sm ${!showMyJobs ? "btn-primary" : "btn-ghost"}`}
          >
            Search All Jobs
          </button>
          <button
            type="button"
            onClick={() => setShowMyJobs(true)}
            className={`btn btn-sm ${showMyJobs ? "btn-primary" : "btn-ghost"}`}
          >
            My Jobs
          </button>
        </div>

        <div className=" flex justify-between">
          <p className="text-gray-500 text-[16px]">
            {isLoading ? "Loading jobs..." : `${visibleJobs.length} jobs found`}
            {isFetching && !isLoading ? " (updating...)" : ""}
          </p>
          <div className="dropdown dropdown-bottom dropdown-end">
            <div
              tabIndex={0}
              role="button"
              className="btn m-1 text-sm font-medium gap-2"
            >
              Sort by: Newest <ChevronDown className="h-4 w-4" />
            </div>
            <ul
              tabIndex={-1}
              className="dropdown-content menu bg-base-100 rounded-box z-1 w-52 p-2 shadow-sm"
            >
              <li>
                <a>Newest first</a>
              </li>
              <li>
                <a>Most relevant</a>
              </li>
            </ul>
          </div>
        </div>

        <div className="mt-5 flex w-full items-center rounded-lg border border-gray-200 bg-white pl-4 p-0.5  shadow-sm">
          <Search className="h-4 w-4 text-gray-400" />
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
            placeholder="Search jobs, skills, or keywords"
            className="ml-3 w-full bg-transparent text-sm text-gray-700 outline-none py-3"
          />
          <button
            type="button"
            onClick={handleApplySearch}
            className="h-full ml-4 text-sm font-semibold text-white bg-jobBlue hover:opacity-80 py-3 px-6 rounded-lg"
          >
            Search
          </button>
        </div>

        <div className="flex flex-col gap-4 py-9 ">
          {isError ? (
            <div className="rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700">
              Failed to load jobs. Please try again.
            </div>
          ) : null}

          {!isLoading && !isError && visibleJobs.length === 0 ? (
            <div className="rounded-lg border border-gray-200 bg-white px-4 py-8 text-center text-sm text-gray-500">
              No jobs found for the selected filters.
            </div>
          ) : null}

          {!isError &&
            visibleJobs.map((job) => (
             <div
                key={job.id}
                onClick={() => router.push(`job/${job.id}`)}
                className="cursor-pointer"
              >
              <Jobcard
                key={job.id}
                title={job.title}
                pay={String(
                  job.job_type === "HOURLY" ? job.hourly_rate : job.budget,
                )}
                type={job.job_type === "HOURLY" ? "hourly" : "fixed"}
                description={job.description}
                postTime={formatPostedTime(job.created_at)}
                tags={parseSkills(job.skills)}
              
              />

             </div>
              
            ))}
        </div>
      </div>
    </div>
  );
};

export default Jobsearch;
