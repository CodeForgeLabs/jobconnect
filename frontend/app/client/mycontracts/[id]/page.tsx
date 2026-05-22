"use client"
import React, { useState } from 'react';

export default function ContractManagement() {
  // Interactive UI State Engines
  const [activeTab, setActiveTab] = useState('Active');
  const [chatOpen, setChatOpen] = useState(false);
  const [chatMessage, setChatMessage] = useState('');

  // Sample Chat Data Layer state
  const [messages, setMessages] = useState([
    {
      id: 1,
      isUser: false,
      text: "Hi! I've just submitted the CI/CD documentation for Milestone 3. Let me know if you need any adjustments."
    },
    {
      id: 2,
      isUser: true,
      text: "Thanks Alex, reviewing it now with the DevOps team."
    }
  ]);

  // Handler for sending new messages
  const handleSendMessage = (e) => {
    e.preventDefault();
    if (!chatMessage.trim()) return;
    
    setMessages([
      ...messages,
      {
        id: Date.now(),
        isUser: true,
        text: chatMessage
      }
    ]);
    setChatMessage('');
  };

  return (
    <div className="bg-surface text-on-surface min-h-screen  selection:bg-primary-fixed selection:text-primary">
      
     

      {/* Main Content Canvas Layout */}
      <main className="pt-12 pb-16 px-4 md:px-8 max-w-7xl mx-auto">
        
        {/* Header Breadcrumb Stack & Primary Controls */}
        <div className="flex flex-col lg:flex-row lg:items-end justify-between gap-6 mb-12">
          <div>
            <nav className="flex items-center gap-2 text-xs md:text-sm text-on-secondary-container mb-4 font-label">
              <span>Contracts</span>
              <span className="material-symbols-outlined text-[10px]">chevron_right</span>
              <span className="font-medium text-primary">Senior Systems Architect</span>
            </nav>
            <h1 className="text-2xl md:text-4xl font-headline font-extrabold tracking-tight text-on-background">
              Senior Systems Architect - <span className="text-primary-container font-bold">Alex Rivera</span>
            </h1>
            <p className="mt-2 text-on-secondary-container text-sm max-w-2xl leading-relaxed">
              Full-scale infrastructure migration and CI/CD pipeline optimization for the Q3 Digital Transformation initiative.
            </p>
          </div>
          
          <div className="flex items-center gap-3 w-full lg:w-auto">
            <button className="flex-1 lg:flex-none px-6 py-2.5 bg-surface-container-low text-primary font-semibold rounded-full hover:bg-surface-container-high transition-colors text-xs md:text-sm">
              Pause Contract
            </button>
            <button className="flex-1 lg:flex-none px-6 py-2.5 bg-error-container text-on-error-container font-semibold rounded-full hover:opacity-90 transition-colors text-xs md:text-sm">
              End Contract
            </button>
          </div>
        </div>

        {/* Dashboard Bento Data Grid */}
        <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mb-12">
          
          {/* Freelancer Profile Summary Block */}
          <div className="bg-surface-container-lowest p-6 md:p-8 rounded-lg shadow-sm border border-outline-variant/20 flex flex-col items-center text-center">
            <div className="relative mb-4">
              <div className="w-24 h-24 rounded-full overflow-hidden border-4 border-surface">
                <img 
                  alt="Alex Rivera Profile" 
                  className="w-full h-full object-cover" 
                  src="https://lh3.googleusercontent.com/aida-public/AB6AXuDP3tGC-HxvufGWUXWSwjbuGwPvGC2OMfmS_v0jLB1FTXldBhr-2E3fGcJ2-SSfVjMEE1OLa0wbXQ_sUG3mcgdwe1Qu5A6ECRh-ZpWhgxYInwff_xUcn2VO1X8jQFkMzvQPsuVigL_9nR4Bz9cYF1HWzCSyFZbu9fP_psmmxcwOFmlwstmndflBiW9PEYwz-HSnt60IMcgOLGv2NIGGaAzFagKECCMNkNCPdzE70oArYuHjaJhUlkBkxdvTIHPf9ppFEqFXVHKAiNz9"
                />
              </div>
              <div className="absolute bottom-1 right-1 w-5 h-5 bg-tertiary-fixed-dim border-2 border-surface rounded-full"></div>
            </div>
            <h2 className="text-xl font-headline font-bold text-on-background">Alex Rivera</h2>
            <p className="text-xs md:text-sm text-on-secondary-container mb-4 font-body">Senior Systems Architect</p>
            
            <div className="flex gap-2 mb-6">
              <span className="bg-tertiary-fixed text-on-tertiary-fixed-variant px-3 py-1 rounded-full text-[10px] font-bold uppercase tracking-wider">Top Rated</span>
              <span className="bg-surface-container text-on-secondary-container px-3 py-1 rounded-full text-[10px] font-medium uppercase tracking-wider">100% Success</span>
            </div>
            
            <button 
              onClick={() => setChatOpen(true)}
              className="w-full flex items-center justify-center gap-2 py-3 bg-secondary-container text-on-secondary-fixed font-bold rounded-full hover:bg-secondary-fixed transition-all text-xs md:text-sm"
            >
              <span className="material-symbols-outlined text-sm">chat</span>
              Message Alex
            </button>
          </div>

          {/* Financial Breakdown Container Card */}
          <div className="md:col-span-2 bg-surface-container-low p-6 md:p-8 rounded-lg flex flex-col justify-between">
            <div>
              <h3 className="text-xs font-label uppercase tracking-widest text-on-secondary-container mb-8 font-bold">Financial Overview</h3>
              <div className="grid grid-cols-1 sm:grid-cols-3 gap-6 sm:gap-8">
                <div>
                  <p className="text-xs text-on-secondary-container mb-1 font-body">Total Budget</p>
                  <p className="text-2xl md:text-3xl font-headline font-extrabold text-primary">$24,500.00</p>
                </div>
                <div>
                  <p className="text-xs text-on-secondary-container mb-1 font-body">Amount Paid</p>
                  <p className="text-2xl md:text-3xl font-headline font-extrabold text-on-background">$12,000.00</p>
                </div>
                <div>
                  <p className="text-xs text-on-secondary-container mb-1 font-body">Remaining</p>
                  <p className="text-2xl md:text-3xl font-headline font-extrabold text-on-tertiary-container">$12,500.00</p>
                </div>
              </div>
            </div>

            {/* Completion Strategy Progress Tracking Bar */}
            <div className="mt-8">
              <div className="flex justify-between items-center mb-3">
                <span className="text-xs font-medium text-on-secondary-container font-label">Project Progress</span>
                <span className="text-xs font-bold text-primary font-label">48% Complete</span>
              </div>
              <div className="w-full h-3 bg-surface-container-high rounded-full overflow-hidden">
                <div className="h-full bg-primary rounded-full transition-all duration-500" style={{ width: '48.9%' }}></div>
              </div>
            </div>
          </div>

        </div>

        {/* Milestones & Work Item Iteration Section */}
        <section className="mb-12">
          <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 mb-8">
            <h2 className="text-xl md:text-2xl font-headline font-bold text-on-background">Milestones &amp; Payments</h2>
            <div className="flex bg-surface-container-low p-1 rounded-full border border-outline-variant/10 self-start sm:self-auto">
              <button 
                onClick={() => setActiveTab('Active')}
                className={`px-6 py-2 text-xs font-bold rounded-full transition-all ${activeTab === 'Active' ? 'bg-surface-container-lowest shadow-xs text-primary' : 'text-on-secondary-container hover:text-primary'}`}
              >
                Active
              </button>
              <button 
                onClick={() => setActiveTab('Past')}
                className={`px-6 py-2 text-xs font-bold rounded-full transition-all ${activeTab === 'Past' ? 'bg-surface-container-lowest shadow-xs text-primary' : 'text-on-secondary-container hover:text-primary'}`}
              >
                Past
              </button>
            </div>
          </div>

          {/* Dynamic Filter Dependent Viewport Segment */}
          {activeTab === 'Active' ? (
            <div className="space-y-6">
              
              {/* Active Pending Milestone Card */}
              <div className="bg-surface-container-lowest rounded-xl shadow-xs border border-primary-container/10 overflow-hidden">
                <div className="p-6 md:p-8 flex flex-col lg:flex-row lg:items-center justify-between gap-6 md:gap-8">
                  <div className="flex items-start gap-4 md:gap-6">
                    <div className="w-12 h-12 md:w-14 md:h-14 bg-tertiary-fixed rounded-2xl flex items-center justify-center shrink-0 shadow-xs">
                      <span className="material-symbols-outlined text-on-tertiary-fixed-variant text-2xl md:text-3xl" style={{ fontVariationSettings: "'FILL' 1" }}>pending_actions</span>
                    </div>
                    <div className="space-y-2">
                      <div className="flex flex-wrap items-center gap-3">
                        <h4 className="font-headline font-bold text-lg md:text-xl text-on-background">Milestone 3: CI/CD Pipeline AWS Migration</h4>
                        <span className="bg-tertiary-container text-on-tertiary-container px-3 py-1 rounded-full text-[10px] font-extrabold uppercase tracking-widest">Pending Approval</span>
                      </div>
                      <p className="text-xs md:text-sm text-on-secondary-container max-w-2xl leading-relaxed font-body">
                        Verification of automated deployment stages in staging. Submission includes full architectural diagrams and security sign-off for the new pipeline infrastructure.
                      </p>
                      <div className="flex items-center gap-6 pt-2 text-xs text-on-secondary-container font-label">
                        <span className="flex items-center gap-1.5 font-medium">
                          <span className="material-symbols-outlined text-base md:text-lg">calendar_today</span> Due Sep 15
                        </span>
                        <span className="flex items-center gap-1.5 font-bold text-primary">
                          <span className="material-symbols-outlined text-base md:text-lg">payments</span> $6,500.00
                        </span>
                      </div>
                    </div>
                  </div>
                  
                  <div className="flex flex-row lg:flex-col sm:flex-row gap-3 items-center border-t lg:border-t-0 pt-6 lg:pt-0 w-full lg:w-auto justify-end">
                    <button className="flex-1 lg:flex-none px-5 py-2.5 text-on-secondary-container font-semibold hover:bg-surface-container-low rounded-full transition-all text-xs md:text-sm whitespace-nowrap">
                      Request Changes
                    </button>
                    <button className="flex-1 lg:flex-none px-8 py-3 bg-primary text-white font-bold rounded-full shadow-md shadow-primary/10 hover:scale-[1.01] active:scale-95 transition-all text-xs md:text-sm whitespace-nowrap">
                      Approve and Pay
                    </button>
                  </div>
                </div>
              </div>

              {/* Upcoming Milestone Row Segment */}
              <div className="bg-surface-container-lowest p-6 md:p-8 rounded-xl border border-outline-variant/20 flex flex-col lg:flex-row lg:items-center justify-between gap-6 md:gap-8 opacity-75">
                <div className="flex items-start gap-4 md:gap-6">
                  <div className="w-12 h-12 md:w-14 md:h-14 bg-surface-container rounded-2xl flex items-center justify-center shrink-0">
                    <span className="material-symbols-outlined text-outline text-2xl md:text-3xl">hourglass_top</span>
                  </div>
                  <div className="space-y-2">
                    <div className="flex flex-wrap items-center gap-3">
                      <h4 className="font-headline font-bold text-lg md:text-xl text-on-background">Milestone 4: Security Hardening &amp; Pentesting</h4>
                      <span className="bg-surface-container text-on-secondary-container px-3 py-1 rounded-full text-[10px] font-bold uppercase tracking-widest">In Progress</span>
                    </div>
                    <p className="text-xs md:text-sm text-on-secondary-container max-w-2xl leading-relaxed font-body">
                      Implementation of Zero-Trust architecture and final stress testing protocols. Preparing environment for external security audit.
                    </p>
                    <div className="flex items-center gap-6 pt-2 text-xs text-on-secondary-container font-medium font-label">
                      <span className="flex items-center gap-1.5"><span class="material-symbols-outlined text-base md:text-lg">calendar_today</span> Due Oct 12</span>
                      <span className="flex items-center gap-1.5"><span class="material-symbols-outlined text-base md:text-lg">payments</span> $6,000.00</span>
                    </div>
                  </div>
                </div>
                <div className="text-right lg:w-auto w-full">
                  <span className="text-xs md:text-sm font-semibold text-on-secondary-container/60 block italic px-4 py-2">
                    Awaiting submission
                  </span>
                </div>
              </div>

            </div>
          ) : (
            <div className="bg-surface-container-lowest p-8 rounded-xl border border-outline-variant/20 text-center py-12">
              <span className="material-symbols-outlined text-3xl text-outline mb-2">history</span>
              <p className="text-sm font-medium text-on-secondary-container font-body">Historical or past approved milestones appear here.</p>
            </div>
          )}
        </section>

        {/* Message Shortcut Component Block Frame */}
        <div className="fixed bottom-6 right-6 md:bottom-10 md:right-10 z-40 group">
          <button 
            onClick={() => setChatOpen(!chatOpen)}
            className="w-14 h-14 md:w-16 md:h-16 bg-primary-container text-white rounded-full shadow-xl flex items-center justify-center hover:scale-105 active:scale-95 transition-all duration-300 relative"
          >
            <span className="material-symbols-outlined text-2xl" style={{ fontVariationSettings: "'FILL' 1" }}>chat</span>
            {!chatOpen && (
              <span className="absolute -top-1 -right-1 w-5 h-5 bg-error text-white text-[10px] flex items-center justify-center rounded-full font-bold">2</span>
            )}
          </button>

          {/* Contextual Real-time Mini Chat Module */}
          <div className={`absolute bottom-20 right-0 w-72 md:w-80 bg-surface-container-lowest shadow-2xl rounded-lg border border-outline-variant/20 overflow-hidden transition-all duration-300 origin-bottom-right ${
            chatOpen ? 'scale-100 opacity-100 visible' : 'scale-90 opacity-0 invisible'
          }`}>
            <div className="bg-primary-container p-4 text-white flex items-center justify-between">
              <div className="flex items-center gap-2.5">
                <div className="w-8 h-8 rounded-full overflow-hidden border border-white/20">
                  <img 
                    alt="Alex Rivera Small Profile" 
                    className="w-full h-full object-cover" 
                    src="https://lh3.googleusercontent.com/aida-public/AB6AXuBw7pLfXPiMDCwXm3SWJrDbzRnwLdhoKit-xmT1-z8u22qNdW-eFsLqfXNr3gaPvr9CCWwm-PGifnusJ_C2nTD9_DKvYzab3bCP3briK6iF5eHUr__28FMSje5ZM6yvxdGJGSNrFqmGBZHV9FSDci6HTSloy7R3wklOgKY2TDCiKMyyYiTVa_EaQJxwY4mPgBKZfkOGdlDd3wBRuA41Fq2yX4huQYWXw1ctcQXdbFti5UkIPyCqXVSW9GO5t1Nsh7yAPaLiRVHHl42v"
                  />
                </div>
                <div>
                  <p className="text-xs font-bold leading-none font-headline">Alex Rivera</p>
                  <p className="text-[10px] opacity-70 mt-0.5 font-label">Online</p>
                </div>
              </div>
              <button 
                onClick={() => setChatOpen(false)}
                className="material-symbols-outlined text-lg hover:bg-white/10 rounded-full p-1 transition-colors flex items-center justify-center"
              >
                close
              </button>
            </div>
            
            {/* Dynamic Interactive Feed Frame */}
            <div className="h-64 p-4 overflow-y-auto space-y-4 bg-surface">
              {messages.map((msg) => (
                <div key={msg.id} className={`flex ${msg.isUser ? 'flex-row-reverse' : ''} gap-2`}>
                  <div className={`max-w-[80%] p-3 rounded-lg text-xs font-body ${
                    msg.isUser 
                      ? 'bg-primary-container text-white rounded-tr-none' 
                      : 'bg-surface-container-low text-on-surface rounded-tl-none border border-outline-variant/10'
                  }`}>
                    {msg.text}
                  </div>
                </div>
              ))}
            </div>

            {/* Form Messaging Submission Context */}
            <form onSubmit={handleSendMessage} className="p-3 border-t border-outline-variant/20 bg-white">
              <div className="relative flex items-center">
                <input 
                  value={chatMessage}
                  onChange={(e) => setChatMessage(e.target.value)}
                  className="w-full pl-4 pr-10 py-2.5 bg-surface-container-low border-none rounded-full text-xs focus:ring-2 focus:ring-primary-container/30 text-on-surface focus:outline-none" 
                  placeholder="Type a message..." 
                  type="text"
                />
                <button 
                  type="submit" 
                  className="absolute right-2 material-symbols-outlined text-primary text-xl p-1 hover:bg-surface-container rounded-full flex items-center justify-center transition-colors"
                >
                  send
                </button>
              </div>
            </form>
          </div>
        </div>

      </main>
    </div>
  );
}