"use client";
import React, { useState } from "react";
import { Search } from "lucide-react";
import Link from "next/link";
import { useGetMyContractsQuery, Contract } from "../../../api/contractapi";
import { ArrowBigDownDashIcon } from "lucide-react";

export default function MyContracts() {
  // Application Interactive States
  const [activeFilter, setActiveFilter] = useState("All Active");
  const [searchQuery, setSearchQuery] = useState("");
  
  // Pagination State (Shows 6 items initially since it's a 2-column grid)
  const [visibleCount, setVisibleCount] = useState(4);

  // Fetch user's contracts from API
  const { data: apiContracts, isLoading, isError } = useGetMyContractsQuery();

  // Map API contract to UI-friendly shape
  const mapContract = (c: Contract) => {
    const freelancerName =
      `${c.freelancer_first_name ?? ""} ${c.freelancer_last_name ?? ""}`.trim() ||
      c.freelancer_email ||
      "Freelancer";
    const role =
      c.freelancer_headline || c.title || c.job_title || "Contractor";
    const avatar =
      c.freelancer_profile_picture_url ||
      `https://ui-avatars.com/api/?name=${encodeURIComponent(freelancerName)}`;
    const scope =
      c.job_title || c.description  || "No scope provided";
    const budget =
      c.type === "HOURLY" ? (c.hourly_rate ?? 0) : (c.total_budget ?? 0);
    const statusType = (c.status || "").toLowerCase();
    const status = c.status || "Active";
    return {
      id: c.contract_id,
      freelancerName,
      role,
      avatar,
      status,
      statusType,
      scope,
      budget,
      raw: c,
    };
  };

  // Filter Buttons
  const filterOptions = [
    "All Active",
    "In Review",
    "Milestones Pending",
    "Recently Completed",
  ];

  // Filter and search logic
  const uiContracts = (apiContracts ?? []).map(mapContract);

  const filteredContracts = uiContracts.filter((contract) => {
    const matchesSearch =
      contract.freelancerName
        .toLowerCase()
        .includes(searchQuery.toLowerCase()) ||
      contract.scope.toLowerCase().includes(searchQuery.toLowerCase()) ||
      contract.role.toLowerCase().includes(searchQuery.toLowerCase());

    if (!matchesSearch) return false;
    if (activeFilter === "All Active") return true;
    if (activeFilter === "In Review") return contract.statusType === "review";
    if (activeFilter === "Milestones Pending"){
      for (const milestone of contract.raw.milestones ?? []) {
        if (milestone.Status === "SUBMITTED") {
          return true;
        }
      }
    }
    return contract.statusType === "pending";
  });

  // Chunk array into current visible window
  const displayedContracts = filteredContracts.slice(0, visibleCount);

  return (
    <div className="min-h-screen bg-surface text-on-background selection:bg-primary-fixed selection:text-primary">
      {/* Main Framework Content Grid */}
      <main className="pt-16 pb-24 px-4 md:px-12 lg:px-24 max-w-7xl mx-auto">
        {/* Dynamic App Dashboard Header */}
        <div className="mb-12">
          <div className="flex flex-col sm:flex-row sm:items-end justify-between gap-6">
            <div>
              <span className="text-xs font-bold uppercase tracking-widest text-primary mb-2 block font-label">
                Management Center
              </span>
              <h1 className="text-3xl md:text-5xl font-display font-extrabold tracking-tight text-on-surface">
                Active Contracts
              </h1>
            </div>
            <div>
              <div className="bg-surface-container-lowest border border-outline-variant/30 px-6 py-3 rounded-2xl flex flex-col gap-0.5 shadow-xs">
                <span className="text-[10px] font-bold text-outline uppercase tracking-wider">
                  Total Active Spend
                </span>
                <span className="text-xl md:text-2xl text-primary font-black font-display">
                 {
                    apiContracts?.reduce((sum, c) => {
                      const contractBudget =
                        c.type === "HOURLY"
                          ? (c.hourly_rate ?? 0) * (c.weekly_hour_limit ?? 0)
                          : c.total_budget ?? 0;
                      return sum + contractBudget;
                    }, 0) || 0
                  }   
                </span>
              </div>
            </div>
          </div>
        </div>

        {/* Controls Layout Stack (Filters + Context Search Engine) */}
        <div className="mb-10 flex flex-col lg:flex-row gap-4 justify-between items-start lg:items-center">
          <div className="flex flex-wrap gap-2 w-full lg:w-auto">
            {filterOptions.map((option) => (
              <button
                key={option}
                onClick={() => {
                  setActiveFilter(option);
                  setVisibleCount(6); // Reset pagination window on switch
                }}
                className={`px-5 py-2.5 rounded-full text-xs font-bold tracking-wide transition-all ${
                  activeFilter === option
                    ? "bg-primary text-on-primary shadow-md shadow-primary/10"
                    : "bg-surface-container-lowest text-on-surface-variant border border-outline-variant/30 hover:bg-surface-container"
                }`}
              >
                {option}
              </button>
            ))}
          </div>

          <div className="relative w-full lg:max-w-xs">
            <span className="material-symbols-outlined absolute left-4 top-1/2 -translate-y-1/2 text-outline text-lg">
              <Search />{" "}
            </span>
            <input
              value={searchQuery}
              onChange={(e) => {
                setSearchQuery(e.target.value);
                setVisibleCount(6); // Reset pagination window on typing
              }}
              className="pl-11 pr-4 py-2.5 bg-surface-container-lowest border border-outline-variant/30 focus:outline-none focus:ring-2 focus:ring-primary/20 rounded-full text-xs w-full transition-all text-on-surface"
              placeholder="Search contracts or talent..."
              type="text"
            />
          </div>
        </div>

        {/* Interactive Responsive Contract Grid */}
        {isLoading ? (
          <div className="text-center py-16">Loading contracts…</div>
        ) : isError ? (
          <div className="text-center py-16">Unable to load contracts.</div>
        ) : displayedContracts.length > 0 ? (
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
            {displayedContracts.map((contract) => (
              <div
                key={contract.id}
                className="bg-surface-container-lowest border border-outline-variant/20 p-6 md:p-8 rounded-2xl hover:shadow-xl hover:border-outline-variant/40 transition-all duration-300 flex flex-col justify-between"
              >
                <div>
                  {/* Card Profile Section Block */}
                  <div className="flex flex-col sm:flex-row justify-between items-start gap-4 mb-6">
                    <div className="flex items-center gap-4">
                      <div className="w-14 h-14 rounded-full overflow-hidden border border-outline-variant/30 flex-shrink-0">
                        <img
                          alt={contract.freelancerName}
                          src={contract.avatar}
                          className="w-full h-full object-cover"
                        />
                      </div>
                      <div>
                        <h3 className="text-lg font-bold text-on-surface font-display">
                          {contract.freelancerName}
                        </h3>
                        <p className="text-on-surface-variant text-xs font-medium">
                          {contract.role}
                        </p>
                      </div>
                    </div>

                    <span
                      className={`px-3 py-1 rounded-full text-[10px] font-bold uppercase tracking-wider border ${
                        contract.statusType === "review"
                          ? "bg-amber-50 text-amber-700 border-amber-200"
                          : contract.statusType === "pending"
                            ? "bg-rose-50 text-rose-700 border-rose-200"
                            : "bg-emerald-50 text-emerald-700 border-emerald-200"
                      }`}
                    >
                      {contract.status}
                    </span>
                  </div>

                  {/* Context Scope Section */}
                  <div className="mb-6">
                    <h4 className="text-[10px] font-bold uppercase tracking-widest text-outline mb-1">
                      Project Scope
                    </h4>
                    <p className="text-md font-bold text-on-surface leading-snug font-headline">
                      {contract.scope}
                    </p>
                  </div>
                </div>

                {/* Footer Dynamic Action Frame Layout */}
                <div className="pt-4 border-t border-outline-variant/20 flex items-center justify-between gap-4 mt-auto">
                  <div className="flex flex-col">
                    {contract.raw?.type === "HOURLY" ? (
                      <>
                        <span className="text-[10px] font-bold uppercase tracking-wider text-outline">
                          Hourly Rate
                        </span>
                        <p className="text-xl font-black text-primary font-display">
                          ETB {" "} 
                          {(contract.raw.hourly_rate ?? 0).toLocaleString()}
                        </p>
                        <span className="text-[10px] text-on-surface-variant font-bold">
                          Weekly limit: {contract.raw.weekly_hour_limit ?? "—"}{" "}
                          hrs
                        </span>
                      </>
                    ) : (
                      <>
                        <span className="text-[10px] font-bold uppercase tracking-wider text-outline">
                          Total Budget
                        </span>
                        <p className="text-xl font-black text-primary font-display">
                          ETB {" "} 
                          {(
                            contract.raw.total_budget ?? contract.budget
                          ).toLocaleString()}
                        </p>
                      </>
                    )}
                  </div>
                  <Link
                    className="bg-primary text-white px-6 py-3 rounded-xl font-bold text-xs hover:bg-primary/90 active:scale-98 transition-all"
                    href={`/client/mycontracts/${contract.id}`}
                  >
                    View Details
                  </Link>
                </div>
              </div>
            ))}
          </div>
        ) : (
          <div className="text-center py-16 bg-surface-container-low rounded-2xl border border-dashed border-outline-variant/50">
            <span className="material-symbols-outlined text-4xl text-outline mb-2">
              folder_open
            </span>
            <p className="text-sm font-semibold text-on-surface-variant">
              No active contracts found matching criteria.
            </p>
          </div>
        )}

        {/* Load More/Pagination Element Wrapper */}
        {filteredContracts.length > visibleCount && (
          <div className="mt-16 flex flex-col items-center gap-3">
            <button 
              type="button"
              onClick={() => setVisibleCount((prev) => prev + 6)} // Increment chunk limits dynamically
              className="group flex items-center gap-2 text-primary font-bold text-sm hover:gap-3 transition-all cursor-pointer"
            >
              Load more active contracts
              <ArrowBigDownDashIcon className="transition-transform group-hover:translate-y-0.5" />
            </button>
            <p className="text-outline text-[10px] font-bold uppercase tracking-widest">
              Showing {displayedContracts.length} of {filteredContracts.length}{" "}
              active contracts
            </p>
          </div>
        )}
      </main>
    </div>
  );
}