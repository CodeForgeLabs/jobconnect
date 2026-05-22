"use client";
import React, { useState } from "react";

export default function ProposalsView() {
  const [darkMode, setDarkMode] = useState(false);
  const [activeTab, setActiveTab] = useState("active");

  const proposals = [
    {
      id: 1,
      title: "Senior React Developer for Fintech SaaS",
      company: "TechSolutions Inc.",
      time: "Submitted 2 hours ago",
      status: "Under Review",
      statusColor: "bg-amber-100 text-amber-800 dark:bg-amber-950/40 dark:text-amber-400",
      amount: "$4,500 - $6,000",
      type: "Fixed Price",
      icon: "code",
      iconBg: "bg-primary/10 text-primary"
    },
    {
      id: 2,
      title: "UI/UX Redesign for E-commerce App",
      company: "GreenMarket Co.",
      time: "Submitted yesterday",
      status: "Interviewing",
      statusColor: "bg-emerald-100 text-emerald-800 dark:bg-emerald-950/40 dark:text-emerald-400",
      amount: "$55/hr",
      type: "Hourly",
      icon: "brush",
      iconBg: "bg-indigo-100 dark:bg-indigo-950/40 text-indigo-600 dark:text-indigo-400"
    },
    {
      id: 3,
      title: "Social Media Marketing Strategy",
      company: "Vibe Agency",
      time: "Submitted 3 days ago",
      status: "Sent",
      statusColor: "bg-surface-container-highest text-on-surface-variant",
      amount: "$1,200",
      type: "Fixed Price",
      icon: "campaign",
      iconBg: "bg-orange-100 dark:bg-orange-950/40 text-orange-600 dark:text-orange-400"
    },
    {
      id: 4,
      title: "Python Automation Scripting",
      company: "DataFlow Systems",
      time: "Submitted 5 days ago",
      status: "Sent",
      statusColor: "bg-surface-container-highest text-on-surface-variant",
      amount: "$800",
      type: "Fixed Price",
      icon: "terminal",
      iconBg: "bg-pink-100 dark:bg-pink-950/40 text-pink-600 dark:text-pink-400"
    }
  ];

  return (
    <div className={`${darkMode ? "dark" : ""} min-h-screen flex flex-col bg-surface text-on-surface transition-colors duration-200 selection:bg-primary-fixed selection:text-primary`}>
      
      

      {/* Main Workspace Frame */}
      <main className="flex-1 max-w-6xl w-full mx-auto px-4 py-8 md:py-12 space-y-8">
        
        {/* Page Identity Dashboard Block */}
        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
          <div>
            <h1 className="text-3xl font-black tracking-tight font-headline">My Proposals</h1>
            <p className="text-on-surface-variant text-sm mt-1">Track and manage your active market listings.</p>
          </div>
          <button className="bg-primary text-white px-5 py-3 rounded-xl font-bold text-sm hover:shadow-lg hover:shadow-primary/20 active:scale-98 transition-all flex items-center justify-center gap-2 w-full sm:w-auto">
            <span className="material-symbols-outlined text-lg">+</span>
            Find New Work
          </button>
        </div>

        {/* Analytics Grid Block */}
        <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
          <StatCard icon="analytics" label="Proposal Views" value="142" colorClass="bg-primary/10 text-primary" />
          <StatCard icon="verified" label="Success Rate" value="92%" colorClass="bg-emerald-100 dark:bg-emerald-950/40 text-emerald-600 dark:text-emerald-400" />
          <StatCard icon="token" label="Available Connects" value="48" colorClass="bg-primary text-white shadow-sm" />
        </div>

        {/* Dynamic Interactive Document Manager Block */}
        <div className="bg-surface-container-lowest border border-outline-variant/30 rounded-2xl shadow-xs overflow-hidden">
          
          {/* Internal Tab Filter Ribbon */}
          <div className="flex border-b border-outline-variant/20 px-4 md:px-6 overflow-x-auto scrollbar-hide bg-surface-container-low">
            <TabButton label="Active" count={8} active={activeTab === "active"} onClick={() => setActiveTab("active")} />
            <TabButton label="Accepted" count={4} active={activeTab === "accepted"} onClick={() => setActiveTab("accepted")} />
            <TabButton label="Declined" count={null} active={activeTab === "declined"} onClick={() => setActiveTab("declined")} />
          </div>

          {/* Proposals List Segment */}
          <div className="divide-y divide-outline-variant/10">
            {proposals.map((proposal) => (
              <div 
                key={proposal.id}
                className="p-5 md:p-6 flex flex-col sm:flex-row sm:items-center justify-between gap-4 hover:bg-surface-container-low/40 transition-colors group"
              >
                {/* Media Meta Pair Layout */}
                <div className="flex gap-4 items-start min-w-0">
                  <div className={`w-12 h-12 rounded-xl flex items-center justify-center flex-shrink-0 ${proposal.iconBg}`}>
                    <span className="material-symbols-outlined text-xl">{proposal.icon}</span>
                  </div>
                  <div className="space-y-1 min-w-0">
                    <div className="flex flex-wrap items-center gap-2">
                      <h3 className="font-bold text-base text-on-surface font-headline truncate group-hover:text-primary transition-colors">
                        {proposal.title}
                      </h3>
                      <span className={`text-[10px] font-extrabold tracking-wide uppercase px-2 py-0.5 rounded-md ${proposal.statusColor}`}>
                        {proposal.status}
                      </span>
                    </div>
                    <p className="text-on-surface-variant text-xs font-medium">
                      {proposal.company} <span className="mx-1.5 text-outline">•</span> {proposal.time}
                    </p>
                  </div>
                </div>

                {/* Pricing Structure & Action Hub */}
                <div className="flex items-center justify-between sm:justify-end gap-6 sm:pl-0 pl-16">
                  <div className="sm:text-right">
                    <p className="font-extrabold text-sm font-headline text-on-surface">{proposal.amount}</p>
                    <p className="text-outline text-[11px] font-medium tracking-wide uppercase mt-0.5">{proposal.type}</p>
                  </div>
                  
                  <div className="flex items-center gap-2">
                    <button className="px-4 py-2 border border-outline-variant hover:border-outline hover:bg-surface text-on-surface-variant hover:text-on-surface rounded-xl text-xs font-bold transition-all">
                      View Proposal
                    </button>
                    <button className="p-2 text-outline hover:text-on-surface rounded-lg transition-colors">
                      <span className="material-symbols-outlined text-lg">more_vert</span>
                    </button>
                  </div>
                </div>
              </div>
            ))}
          </div>

          {/* Paginated Footer System */}
          <footer className="flex items-center justify-between px-6 py-4 border-t border-outline-variant/20 bg-surface-container-low">
            <span className="text-xs text-on-surface-variant font-medium">Showing 1 to 4 of 8 listings</span>
            <div className="flex gap-1.5">
              <button className="p-2 border border-outline-variant/40 rounded-xl text-outline opacity-40 cursor-not-allowed flex items-center">
                <span className="material-symbols-outlined text-base">chevron_left</span>
              </button>
              <button className="p-2 border border-outline-variant/40 hover:border-outline bg-surface-container-lowest text-on-surface-variant hover:text-on-surface rounded-xl shadow-xs transition-all flex items-center">
                <span className="material-symbols-outlined text-base">chevron_right</span>
              </button>
            </div>
          </footer>
        </div>
      </main>

      {/* Global Application Footer */}
      <footer className="border-t border-outline-variant/20 py-8 px-4 md:px-8 lg:px-16 bg-surface-container-low mt-auto">
        <div className="max-w-6xl w-full mx-auto flex flex-col sm:flex-row justify-between items-center gap-4 text-on-surface-variant text-xs font-medium">
          <div className="flex items-center gap-2">
            <span className="material-symbols-outlined text-primary text-sm">work</span>
            <p>© 2026 JobPulse Hub. All structural data encrypted.</p>
          </div>
          <div className="flex gap-6">
            <a className="hover:text-primary transition-colors" href="#">Help Center</a>
            <a className="hover:text-primary transition-colors" href="#">Terms of Service</a>
            <a className="hover:text-primary transition-colors" href="#">Privacy Policy</a>
          </div>
        </div>
      </footer>

    </div>
  );
}

/* Local UI Building Blocks */
function StatCard({ icon, label, value, colorClass }) {
  return (
    <div className="bg-surface-container-lowest p-5 rounded-2xl border border-outline-variant/20 flex items-center gap-4 shadow-xs">
      <div className={`w-12 h-12 rounded-xl flex items-center justify-center flex-shrink-0 ${colorClass}`}>
        <span className="material-symbols-outlined text-xl">{icon}</span>
      </div>
      <div>
        <p className="text-outline text-[10px] font-extrabold uppercase tracking-widest">{label}</p>
        <p className="text-2xl font-black font-headline text-on-surface mt-0.5 tracking-tight">{value}</p>
      </div>
    </div>
  );
}

function TabButton({ label, count, active, onClick }) {
  return (
    <button 
      onClick={onClick}
      className={`flex items-center justify-center border-b-2 px-4 pb-4 pt-5 gap-2 transition-all font-headline font-bold text-sm whitespace-nowrap outline-none ${
        active 
          ? "border-primary text-primary" 
          : "border-transparent text-on-surface-variant hover:text-on-surface"
      }`}
    >
      <span>{label}</span>
      {count !== null && (
        <span className={`text-[10px] font-extrabold px-2 py-0.5 rounded-full ${
          active ? "bg-primary/10 text-primary" : "bg-surface-container-highest text-on-surface-variant"
        }`}>
          {count}
        </span>
      )}
    </button>
  );
}