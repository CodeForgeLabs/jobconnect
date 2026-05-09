"use client";

import Jobcard from "@/components/Jobcard";
import { useEffect } from "react";
import { useRouter } from "next/navigation";
import { useSelector } from "react-redux";
import {
  selectIsAuthenticated,
  selectIsHydrated,
} from "@/features/login/loginSlice";

const dummyJobs = [
  {
    title: "Redesign a SaaS landing page",
    pay: "900",
    type: "fixed" as const,
    rating: 4.5,
    description:
      "We are looking for a frontend developer to redesign a B2B SaaS landing page with a stronger hero section, pricing cards, customer testimonials, feature highlights, and a clear call-to-action flow. The final result should feel modern, conversion-focused, and fully responsive across desktop and mobile.",
    postTime: "2 hours ago",
    tags: ["React", "Tailwind", "Landing Page", "Responsive UI"],
  },
  {
    title: "Fix bugs in a freelancer dashboard",
    pay: "28",
    type: "hourly" as const,
    rating: 4,
    description:
      "Need a Next.js developer to clean up several UI issues across a freelancer dashboard, including card spacing, table alignment, mobile responsiveness, and a few inconsistent button states. This is a short-term task, but there may be follow-up work if the quality is strong.",
    postTime: "5 hours ago",
    tags: ["Next.js", "CSS", "Bug Fixes", "Dashboard"],
  },
  {
    title: "Build a Figma to React component set",
    pay: "650",
    type: "fixed" as const,
    rating: 5,
    description:
      "Convert a full set of Figma designs into reusable React components for a job marketplace web app. The components should be clean, modular, and easy to reuse across multiple pages, with strong attention to spacing, typography, and pixel accuracy.",
    postTime: "1 day ago",
    tags: ["Figma", "React", "UI Components", "Design System"],
  },
];

const FreelancerDashboard = () => {
  const router = useRouter();
  const isHydrated = useSelector(selectIsHydrated);
  const isAuthenticated = useSelector(selectIsAuthenticated);

  useEffect(() => {
    if (!isHydrated) return;
    if (!isAuthenticated) {
      router.replace("/login");
    }
  }, [isHydrated, isAuthenticated, router]);

  if (!isHydrated || !isAuthenticated) {
    return (
      <div className="p-8 bg-[#eff1f5] min-h-screen">
        <p className="text-gray-600">Preparing your workspace...</p>
      </div>
    );
  }

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
        <div className="mt-3 flex gap-3">
          <a
            href="/onboarding"
            className="inline-flex items-center rounded-full border border-[#ccd6c4] bg-white px-4 py-2 text-xs font-semibold text-[#1f1f1f] hover:bg-[#f7fbf5]"
          >
            Continue onboarding
          </a>
          <a
            href="/account"
            className="inline-flex items-center rounded-full bg-[#108a00] px-4 py-2 text-xs font-semibold text-white hover:bg-[#0d7300]"
          >
            Open account hub
          </a>
        </div>
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
            <p className="text-xs text-jobBlue">View all</p>
          </div>

          <div className="flex flex-col gap-4">
            {dummyJobs.map((job) => (
              <Jobcard
                key={job.title}
                title={job.title}
                pay={job.pay}
                type={job.type}
                rating={job.rating}
                description={job.description}
                postTime={job.postTime}
                tags={job.tags}
              />
            ))}
          </div>
        </div>

        <div className="w-[32%]">
          <div className=" bg-white border border-gray-200 rounded-lg ">
            <div className="flex justify-between p-4">
              <h2 className="text-sm font-semibold text-gray-800">
              Active Contracts
            </h2>
            <p className="text-xs text-gray-500 mt-1">
              10
            </p>
            </div>

            <div className="border-t border-gray-200 h-10 p-4">

            </div>
            
            
          </div>
        </div>
      </div>
    </div>
  );
};

export default FreelancerDashboard;
