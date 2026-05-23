import React from 'react';

export default function FindTalent() {
  return (
    <div className="min-h-screen bg-[#f8f9ff] text-[#0d1c2e] font-sans selection:bg-[#2e3192] selection:text-white">
      

     

      {/* Main Content Layout */}
      <main className="pt-32 pb-20 px-8 max-w-[1440px] mx-auto flex gap-12">
        {/* Side Filter Panel */}
        <aside className="hidden lg:block w-72 flex-shrink-0">
          <div className="sticky top-32 space-y-10">
            <div>
              <h3 className="text-xs font-bold uppercase tracking-[0.15em] text-[#464652] mb-6 font-headline">Filter Results</h3>
              
              {/* Category */}
              <div className="space-y-4 mb-8">
                <label className="block text-sm font-semibold text-[#0d1c2e]">Category</label>
                <select className="w-full bg-white border border-[#c7c5d4]/40 rounded-lg p-3 text-sm focus:ring-2 focus:ring-[#00435f] outline-none transition-all">
                  <option>UI/UX Design</option>
                  <option>Development</option>
                  <option>Marketing</option>
                  <option>Data Science</option>
                </select>
              </div>
              
              {/* Hourly Rate */}
              <div className="space-y-4 mb-8">
                <div className="flex justify-between items-center">
                  <label className="text-sm font-semibold text-[#0d1c2e]">Hourly Rate</label>
                  <span className="text-xs text-[#15157d] font-bold">$40 - $150+</span>
                </div>
                <input className="w-full h-1.5 bg-[#e6eeff] rounded-lg appearance-none cursor-pointer accent-[#15157d]" type="range" />
              </div>
              
              {/* Location */}
              <div className="space-y-4 mb-8">
                <label className="block text-sm font-semibold text-[#0d1c2e]">Location</label>
                <div className="relative">
                  <span className="material-symbols-outlined absolute left-3 top-1/2 -translate-y-1/2 text-slate-400 text-sm">location_on</span>
                  <input className="w-full pl-10 pr-4 py-3 bg-white border border-[#c7c5d4]/40 rounded-lg text-sm focus:ring-2 focus:ring-[#00435f] outline-none transition-all" placeholder="Country or City" type="text" />
                </div>
              </div>
              
              {/* Job Success */}
              <div className="space-y-4">
                <label className="block text-sm font-semibold text-[#0d1c2e]">Job Success Score</label>
                <div className="space-y-3">
                  <label className="flex items-center gap-3 cursor-pointer group">
                    <input className="w-5 h-5 rounded border-[#c7c5d4]/50 text-[#15157d] focus:ring-[#15157d]/20 accent-[#15157d]" type="checkbox" />
                    <span className="text-sm text-[#464652] group-hover:text-[#0d1c2e] transition-colors">90% & up</span>
                  </label>
                  <label className="flex items-center gap-3 cursor-pointer group">
                    <input className="w-5 h-5 rounded border-[#c7c5d4]/50 text-[#15157d] focus:ring-[#15157d]/20 accent-[#15157d]" type="checkbox" />
                    <span className="text-sm text-[#464652] group-hover:text-[#0d1c2e] transition-colors">80% & up</span>
                  </label>
                  <label className="flex items-center gap-3 cursor-pointer group">
                    <input className="w-5 h-5 rounded border-[#c7c5d4]/50 text-[#15157d] focus:ring-[#15157d]/20 accent-[#15157d]" type="checkbox" />
                    <span className="text-sm text-[#464652] group-hover:text-[#0d1c2e] transition-colors">Any success rate</span>
                  </label>
                </div>
              </div>
            </div>
            
            {/* Promo Card */}
            <div className="p-6 bg-[#2e3192] rounded-lg text-white space-y-4 relative overflow-hidden">
              <div className="relative z-10">
                <h4 className="font-bold text-lg leading-tight font-headline">Need help finding the right fit?</h4>
                <p className="text-xs text-[#9da1ff] leading-relaxed">Our talent specialists can curate a shortlist of top-tier freelancers for your project.</p>
                <button className="mt-4 bg-[#c6e7ff] text-[#001e2e] px-4 py-2 rounded-full text-xs font-bold hover:bg-white transition-colors">Contact Expert</button>
              </div>
              <div className="absolute -right-4 -bottom-4 w-24 h-24 bg-white/10 rounded-full blur-2xl"></div>
            </div>
          </div>
        </aside>

        {/* Search & Results Area */}
        <section className="flex-grow">
          {/* Search Header */}
          <div className="mb-12 space-y-6">
            <h1 className="text-4xl font-extrabold tracking-tight text-[#15157d] font-headline">Discover Top Talent</h1>
            <div className="flex gap-4 p-2 bg-[#eff4ff] rounded-2xl">
              <div className="flex-grow relative">
                <span className="material-symbols-outlined absolute left-5 top-1/2 -translate-y-1/2 text-slate-400">search</span>
                <input className="w-full pl-14 pr-6 py-4 bg-transparent border-none text-lg focus:ring-0 outline-none placeholder:text-slate-400" placeholder="Search by skills, roles, or keywords..." type="text" />
              </div>
              <button className="bg-[#15157d] px-10 py-4 rounded-xl text-white font-bold hover:shadow-xl transition-all active:scale-95">Search</button>
            </div>
            <div className="flex gap-3 items-center">
              <span className="text-xs font-bold text-[#464652] uppercase tracking-widest">Popular:</span>
              <div className="flex gap-2">
                <span className="px-3 py-1 bg-[#dce9ff] text-[#15157d] text-[10px] font-bold rounded-full cursor-pointer hover:bg-[#15157d] hover:text-white transition-all uppercase">Figma</span>
                <span className="px-3 py-1 bg-[#dce9ff] text-[#15157d] text-[10px] font-bold rounded-full cursor-pointer hover:bg-[#15157d] hover:text-white transition-all uppercase">React.js</span>
                <span className="px-3 py-1 bg-[#dce9ff] text-[#15157d] text-[10px] font-bold rounded-full cursor-pointer hover:bg-[#15157d] hover:text-white transition-all uppercase">Brand Identity</span>
              </div>
            </div>
          </div>

          {/* Freelancer List */}
          <div className="space-y-8">
            
            {/* Result Card 1 */}
            <div className="bg-white p-8 rounded-lg transition-all duration-300 hover:shadow-[0_32px_64px_-16px_rgba(13,28,46,0.08)] hover:-translate-y-1 flex gap-8 relative overflow-hidden group">
              <div className="flex-shrink-0">
                <div className="relative">
                  <div className="w-24 h-24 rounded-2xl overflow-hidden shadow-lg">
                    <img alt="Freelancer Avatar" className="w-full h-full object-cover" src="https://lh3.googleusercontent.com/aida-public/AB6AXuCRVYWNqtALDrpIgoNtB_LFpGYxkvmR6_-d8-NJRskzvrP4U7fDqPHgce3fUzBTxGcfuQqJOw2iEaxQgw77XmD67NoDeoYNyCc0tG_owegZrZbR2fM2xiQiJ5ILOk01Fv5xGA-405cC42JfVhdR3GP5Gl-_FQ3WeXEOzjVZa_-CZ-8prM_I4k6qeXlmiz78o4H8Uq_Czq7vaTwiKfEL0iAePwRQ-6rgSwzyleYPiQQlyEWrByPZ6fpQqp-QvtapUZRI7ARRqkMFcQp-" />
                  </div>
                  <div className="absolute -bottom-2 -right-2 bg-[#c6e7ff] text-[#004c6c] px-2 py-0.5 rounded-full text-[10px] font-bold border-2 border-white">Top Rated</div>
                </div>
              </div>
              <div className="flex-grow space-y-4">
                <div className="flex justify-between items-start">
                  <div>
                    <h2 className="text-2xl font-bold text-[#0d1c2e] leading-tight font-headline">Sarah Chen</h2>
                    <p className="text-[#15157d] font-medium">Senior Product & Brand Strategist</p>
                  </div>
                  <div className="text-right">
                    <div className="flex items-center gap-1 justify-end text-orange-400">
                      <span className="material-symbols-outlined text-sm" style={{ fontVariationSettings: "'FILL' 1" }}>star</span>
                      <span className="text-[#0d1c2e] font-bold">4.9</span>
                      <span className="text-[#464652] text-xs">(128 reviews)</span>
                    </div>
                    <p className="text-sm font-bold text-[#0d1c2e]">$85/hr</p>
                  </div>
                </div>
                <p className="text-[#464652] text-sm leading-relaxed max-w-2xl">
                  Specializing in zero-to-one product launches and architectural design systems. I bridge the gap between complex engineering and human-centric design for high-growth SaaS startups...
                </p>
                <div className="flex flex-wrap gap-2">
                  <span className="px-3 py-1 bg-[#e6eeff] text-[#57587f] text-xs rounded-full">SaaS Design</span>
                  <span className="px-3 py-1 bg-[#e6eeff] text-[#57587f] text-xs rounded-full">React Integration</span>
                  <span className="px-3 py-1 bg-[#e6eeff] text-[#57587f] text-xs rounded-full">User Research</span>
                  <span className="px-3 py-1 bg-[#e6eeff] text-[#57587f] text-xs rounded-full">+4 more</span>
                </div>
                <div className="flex gap-8 pt-4 border-t border-[#c7c5d4]/20">
                  <div>
                    <p className="text-[10px] uppercase tracking-widest text-[#464652] mb-1">Total Earned</p>
                    <p className="text-sm font-bold text-[#0d1c2e]">$200k+</p>
                  </div>
                  <div>
                    <p className="text-[10px] uppercase tracking-widest text-[#464652] mb-1">Success Rate</p>
                    <p className="text-sm font-bold text-[#0d1c2e]">98%</p>
                  </div>
                  <div>
                    <p className="text-[10px] uppercase tracking-widest text-[#464652] mb-1">Location</p>
                    <p className="text-sm font-bold text-[#0d1c2e]">Singapore</p>
                  </div>
                </div>
              </div>
              <div className="flex flex-col gap-3 justify-center">
                <button className="bg-[#15157d] text-white px-6 py-3 rounded-xl font-bold whitespace-nowrap hover:shadow-lg transition-all active:scale-95">Invite to Job</button>
                <button className="bg-[#d5e3fc] text-[#15157d] px-6 py-3 rounded-xl font-bold whitespace-nowrap hover:bg-[#dce9ff] transition-all active:scale-95">View Profile</button>
              </div>
            </div>

            {/* Result Card 2 */}
            <div className="bg-white p-8 rounded-lg transition-all duration-300 hover:shadow-[0_32px_64px_-16px_rgba(13,28,46,0.08)] hover:-translate-y-1 flex gap-8 relative overflow-hidden group">
              <div className="flex-shrink-0">
                <div className="relative">
                  <div className="w-24 h-24 rounded-2xl overflow-hidden shadow-lg">
                    <img alt="Freelancer Avatar" className="w-full h-full object-cover" src="https://lh3.googleusercontent.com/aida-public/AB6AXuBxCMfgzg0w9_MzxUkflkn1MyvHKaEppAdhJSzke4xysFP7Rw9zNJbORlW15gYnPZgMEiHNuRBvhxBsjBssXISAEEKEbx4PTcg60M5HtbGzZnwCz6nAlOeKvh2fr0_nWdYs0l9J-6LpIxIH4WvYJ_v2FlSu2Rbh6QezHGrqyicybOVG1OQYjviv6bilOiRENJBvkbK7lP5wgLBEmFpXAQRlDqXqNU73hmVec8iOBaZ0YAGX5IccrQMJ5KmsZkSMhqS12y2c3pXwEtF3" />
                  </div>
                </div>
              </div>
              <div className="flex-grow space-y-4">
                <div className="flex justify-between items-start">
                  <div>
                    <h2 className="text-2xl font-bold text-[#0d1c2e] leading-tight font-headline">Marcus Thorne</h2>
                    <p className="text-[#15157d] font-medium">Senior Full-Stack Engineer & AWS Architect</p>
                  </div>
                  <div className="text-right">
                    <div className="flex items-center gap-1 justify-end text-orange-400">
                      <span className="material-symbols-outlined text-sm" style={{ fontVariationSettings: "'FILL' 1" }}>star</span>
                      <span className="text-[#0d1c2e] font-bold">5.0</span>
                      <span className="text-[#464652] text-xs">(84 reviews)</span>
                    </div>
                    <p className="text-sm font-bold text-[#0d1c2e]">$110/hr</p>
                  </div>
                </div>
                <p className="text-[#464652] text-sm leading-relaxed max-w-2xl">
                  Specializing in building scalable, enterprise-grade cloud infrastructures and high-performance web applications. Expert in Node.js, Python, and modern cloud ecosystems...
                </p>
                <div className="flex flex-wrap gap-2">
                  <span className="px-3 py-1 bg-[#e6eeff] text-[#57587f] text-xs rounded-full">AWS Architect</span>
                  <span className="px-3 py-1 bg-[#e6eeff] text-[#57587f] text-xs rounded-full">Node.js</span>
                  <span className="px-3 py-1 bg-[#e6eeff] text-[#57587f] text-xs rounded-full">PostgreSQL</span>
                  <span className="px-3 py-1 bg-[#e6eeff] text-[#57587f] text-xs rounded-full">+6 more</span>
                </div>
                <div className="flex gap-8 pt-4 border-t border-[#c7c5d4]/20">
                  <div>
                    <p className="text-[10px] uppercase tracking-widest text-[#464652] mb-1">Total Earned</p>
                    <p className="text-sm font-bold text-[#0d1c2e]">$500k+</p>
                  </div>
                  <div>
                    <p className="text-[10px] uppercase tracking-widest text-[#464652] mb-1">Success Rate</p>
                    <p className="text-sm font-bold text-[#0d1c2e]">100%</p>
                  </div>
                  <div>
                    <p className="text-[10px] uppercase tracking-widest text-[#464652] mb-1">Location</p>
                    <p className="text-sm font-bold text-[#0d1c2e]">Berlin, DE</p>
                  </div>
                </div>
              </div>
              <div className="flex flex-col gap-3 justify-center">
                <button className="bg-[#15157d] text-white px-6 py-3 rounded-xl font-bold whitespace-nowrap hover:shadow-lg transition-all active:scale-95">Invite to Job</button>
                <button className="bg-[#d5e3fc] text-[#15157d] px-6 py-3 rounded-xl font-bold whitespace-nowrap hover:bg-[#dce9ff] transition-all active:scale-95">View Profile</button>
              </div>
            </div>

            {/* Pagination/Load More */}
            <div className="pt-8 flex justify-center">
              <button className="group flex items-center gap-3 px-8 py-4 bg-white shadow-xl shadow-slate-200/50 rounded-2xl text-[#15157d] font-bold hover:bg-[#dce9ff] transition-all active:scale-95">
                <span>Load more talent</span>
                <span className="material-symbols-outlined group-hover:translate-y-1 transition-transform">expand_more</span>
              </button>
            </div>
            
          </div>
        </section>
      </main>
    </div>
  );
}