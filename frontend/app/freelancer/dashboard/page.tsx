"use client"
import Jobcard from "@/components/Jobcard";
import BuyconnectsCard from "@/components/Buyconnectscard";
import { useGetJobsQuery } from "@/api/jobsapi";



const FreelancerDashboard = () => {
  const filters = {
    title: undefined,
    category: undefined,
    job_type: undefined,
    work_mode: undefined,
    experience_level: undefined,
    budget_min: undefined,
  };
  const {data : Jobs , isLoading} = useGetJobsQuery(filters);

  return (
    <div className="flex flex-col gap-8   p-8 bg-[#eff1f5]">
      <div>
        <h1 className="text-2xl font-bold text-gray-800">
          Welcome back, Nati!
        </h1>
        <p className=" text-xs text-gray-600">
          You have 3 tasks requiring your attention. Check your dashboard for
          details.
        </p>
      </div>

      <div className="flex gap-6 w-full">
        <div className="flex flex-col gap-3 w-1/3 border border-gray-200 bg-white rounded-lg p-4">
          <span className="flex justify-between">
            <p className="text-xs text-gray-500 text-center">Active Contracts</p>
            <span className="inline-flex h-8 w-8 items-center justify-center rounded-md bg-blue-50 text-jobBlue">
              <svg
                xmlns="http://www.w3.org/2000/svg"
                className="h-4 w-4"
                fill="none"
                viewBox="0 0 24 24"
                aria-hidden="true"
              >
                <path
                  d="M8 3h6l4 4v12a2 2 0 0 1-2 2H8a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2Z"
                  stroke="currentColor"
                  strokeWidth="1.8"
                  strokeLinecap="round"
                  strokeLinejoin="round"
                />
                <path
                  d="M14 3v4h4M9 12h6M9 16h6"
                  stroke="currentColor"
                  strokeWidth="1.8"
                  strokeLinecap="round"
                  strokeLinejoin="round"
                />
              </svg>
            </span>
          </span>
          <div className="flex items-center gap-2">
            <p className="text-3xl">12</p>
            <span className="inline-flex items-center gap-1 rounded-full bg-emerald-50 px-2 py-0.5 text-xs font-semibold text-emerald-600">
              <svg
                xmlns="http://www.w3.org/2000/svg"
                className="h-3.5 w-3.5 shrink-0"
                fill="none"
                viewBox="0 0 20 20"
                aria-hidden="true"
              >
                <path
                  d="M5 14 14 5M8 5h6v6"
                  stroke="currentColor"
                  strokeWidth="1.8"
                  strokeLinecap="round"
                  strokeLinejoin="round"
                />
              </svg>
              8%
            </span>
          </div>
          <p className="text-[10px] text-gray-400">vs Last month</p>
        </div>

        <div className="gap-4 flex flex-col w-1/3 border border-gray-200 bg-white rounded-lg p-4">
          <span className="flex justify-between">
            <p className="text-xs text-gray-500">Pending Proposals</p>
            <span className="inline-flex h-8 w-8 items-center justify-center rounded-md bg-yellow-100 text-yellow-700">
              <svg
                xmlns="http://www.w3.org/2000/svg"
                className="h-4 w-4"
                fill="none"
                viewBox="0 0 24 24"
                aria-hidden="true"
              >
                <path
                  d="m21 3-9.5 9.5"
                  stroke="currentColor"
                  strokeWidth="1.8"
                  strokeLinecap="round"
                  strokeLinejoin="round"
                />
                <path
                  d="m21 3-6.5 18-3.5-8-8-3.5L21 3Z"
                  stroke="currentColor"
                  strokeWidth="1.8"
                  strokeLinecap="round"
                  strokeLinejoin="round"
                />
              </svg>
            </span>
          </span>
          <div className="flex items-center gap-2">
            <p className="text-3xl">
              {" "}
              5 <span className="text-emerald-700 text-xs">+ 1</span>{" "}
            </p>
          </div>
          <p className="text-[10px] text-gray-400">Active bids in review</p>
        </div>

        <div className="flex flex-col gap-4 w-1/3 border border-gray-200 bg-white rounded-lg p-4">
          <span className="flex justify-between">
            <p className="text-xs text-gray-500">Total Earnings this Month</p>
            <span className="inline-flex h-8 w-8 items-center justify-center rounded-md bg-blue-50 text-jobBlue">
              <svg
                xmlns="http://www.w3.org/2000/svg"
                className="h-4 w-4"
                fill="none"
                viewBox="0 0 24 24"
                aria-hidden="true"
              >
                <path
                  d="M4 8a2 2 0 0 1 2-2h12a2 2 0 0 1 2 2v8a2 2 0 0 1-2 2H6a2 2 0 0 1-2-2V8Z"
                  stroke="currentColor"
                  strokeWidth="1.8"
                  strokeLinecap="round"
                  strokeLinejoin="round"
                />
                <path
                  d="M12 9.2v5.6M10.4 10.8h2.2a1.1 1.1 0 0 1 0 2.2h-1.2a1.1 1.1 0 0 0 0 2.2h2.2"
                  stroke="currentColor"
                  strokeWidth="1.8"
                  strokeLinecap="round"
                  strokeLinejoin="round"
                />
              </svg>
            </span>
          </span>
          <div className="flex items-center gap-2">
            <p className="text-3xl">4500 birr</p>
            <span className="inline-flex items-center gap-1 rounded-full bg-emerald-50 px-2 py-0.5 text-xs font-semibold text-emerald-600">
              <svg
                xmlns="http://www.w3.org/2000/svg"
                className="h-3.5 w-3.5 shrink-0"
                fill="none"
                viewBox="0 0 20 20"
                aria-hidden="true"
              >
                <path
                  d="M5 14 14 5M8 5h6v6"
                  stroke="currentColor"
                  strokeWidth="1.8"
                  strokeLinecap="round"
                  strokeLinejoin="round"
                />
              </svg>
              8%
            </span>
          </div>
          <p className="text-[10px] text-gray-400">Net income this month</p>
        </div>
      </div>

      <div className="flex gap-4 justify-between ">
                    <div className="flex w-[65%] flex-col gap-4">
              <div className="flex w-full items-center justify-between">
                <h2 className="text-lg font-semibold text-gray-800">
                  Recommended for you
                </h2>

                <p className="text-xs text-jobBlue">
                  View all
                </p>
              </div>

              {isLoading ? (
                <p className="text-sm text-gray-500">
                  Loading jobs...
                </p>
              ) : (
                <div className="flex flex-col gap-4">
                  {(Jobs ?? []).map((job) => (
                    <Jobcard
                      key={job.id}
                      title={job.title}
                      pay={String(job.budget)}
                      type={job.job_type}
                      rating={5}
                      description={job.description}
                      postTime={job.created_at}
                      tags={job.skills.split(",").map((tag) => tag.trim())}
                    />
                  ))}
                </div>
              )}
            </div>
            
        <BuyconnectsCard />
      </div>
    </div>
  );
};

export default FreelancerDashboard;
