"use client";
import React from "react";

export default function JobDetailView() {
  return (
    <div className="bg-surface text-on-surface selection:bg-primary-fixed selection:text-primary min-h-screen">
      <main className="max-w-screen-2xl mx-auto px-6 md:px-8 pt-8 md:pt-12 mb-24">
        {/* Hero Section */}
        <header className="mb-12">
          <div className="flex flex-wrap items-center gap-3 mb-4">
            <span className="bg-tertiary-fixed text-on-tertiary-fixed-variant px-3 py-1 rounded-full text-[10px] md:text-xs font-bold tracking-wide uppercase">
              Open Position
            </span>
            <span className="text-on-surface-variant text-sm font-medium">
              Posted Oct 12, 2023
            </span>
          </div>
          <h1 className="text-4xl md:text-6xl font-extrabold text-primary tracking-tighter leading-tight mb-4">
            Senior Systems Architect
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
              <span>San Francisco, CA (Remote)</span>
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
              <span>AlphaCorp</span>
            </div>
          </div>
        </header>

        <div className="grid grid-cols-1 lg:grid-cols-12 gap-8 lg:gap-16">
          {/* Main Content Area */}
          <div className="lg:col-span-8 space-y-12 md:space-y-16">
            {/* Bento Grid Stats */}
            <section className="grid grid-cols-2 md:grid-cols-4 gap-4 md:gap-6 p-6 md:p-8 bg-surface-container-low rounded-xl">
              <StatItem label="Hourly Rate" value="$120/hr" />
              <StatItem label="Duration" value="6+ Months" />
              <StatItem label="Level" value="Expert" />
              <StatItem label="Work Type" value="Remote" />
            </section>

            {/* Description */}
            <article className="prose prose-slate max-w-none">
              <h2 className="text-2xl md:text-3xl font-bold text-primary mb-6">
                About the Role
              </h2>
              <p className="text-on-surface-variant leading-relaxed text-base md:text-lg mb-6">
                AlphaCorp is seeking a visionary Senior Systems Architect to
                lead the design and evolution of our next-generation distributed
                trading infrastructure. You will be at the helm of creating
                resilient, scalable, and high-performance systems that handle
                billions of transactions daily.
              </p>
              <h3 className="text-xl md:text-2xl font-bold text-primary mb-4">
                Responsibilities
              </h3>
              <ul className="space-y-4 list-none p-0 text-on-surface-variant text-base md:text-lg">
                <ResponsibilityItem text="Design end-to-end architectural frameworks for globally distributed cloud services." />
                <ResponsibilityItem text="Directly supervise the implementation of mission-critical Kubernetes clusters." />
                <ResponsibilityItem text="Perform deep-dive performance analysis and bottleneck identification." />
              </ul>
            </article>

            {/* Skills */}
            <section>
              <h2 className="text-2xl font-bold text-primary mb-6">
                Required Skills
              </h2>
              <div className="flex flex-wrap gap-2 md:gap-3">
                {[
                  "AWS Cloud Architecture",
                  "Kubernetes (EKS)",
                  "Distributed Systems",
                ].map((skill) => (
                  <span
                    key={skill}
                    className="px-4 md:px-6 py-2 md:py-2.5 bg-primary text-white rounded-full font-semibold text-xs md:text-sm shadow-lg shadow-primary/20"
                  >
                    {skill}
                  </span>
                ))}
                {["gRPC & Protobuf", "Terraform"].map((skill) => (
                  <span
                    key={skill}
                    className="px-4 md:px-6 py-2 md:py-2.5 bg-surface-container-highest text-primary rounded-full font-semibold text-xs md:text-sm"
                  >
                    {skill}
                  </span>
                ))}
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
                      AlphaCorp
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
                    Enterprise Tech Solutions
                  </p>
                </div>
              </div>
              <div className="space-y-4">
                <ClientStat label="Member since" value="2021" />
                <ClientStat label="Total Spent" value="$2.4M+" />
                <ClientStat label="Hire Rate" value="94%" />
              </div>
            </div>
          </aside>
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
