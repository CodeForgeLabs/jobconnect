import Image from "next/image";
const submitMilestonePage = () => {
  return (
    <div className="bg-surface text-on-surface selection:bg-primary-fixed selection:text-on-primary-fixed antialiased min-h-screen">
      {/* TopNavBar */}
      

      {/* Main Content */}
      <main className="pt-32 pb-24 px-8 max-w-[1440px] mx-auto min-h-screen">
        <div className="grid grid-cols-1 lg:grid-cols-12 gap-12">
          {/* Content Area */}
          <div className="lg:col-span-8">
            {/* Header */}
            <header className="mb-12">
              <a
                className="flex items-center gap-2 text-on-surface-variant hover:text-primary mb-4 group transition-colors"
                href="#"
              >
                <span
                  className="material-symbols-outlined text-sm"
                  data-icon="arrow_back"
                >
                  arrow_back
                </span>
                <span className="text-sm font-medium">
                  Back to Contract Details
                </span>
              </a>
              <h1 className="text-4xl md:text-5xl font-extrabold tracking-tight text-on-surface mb-2">
                Submit Milestone Work
              </h1>
              <p className="text-xl text-on-surface-variant font-medium">
                Milestone 3: API Integration for Transaction History
              </p>
            </header>

            {/* Milestone Summary Card */}
            <div className="bg-surface-container-low rounded-lg p-8 mb-12 flex flex-col md:flex-row justify-between items-start md:items-center gap-6">
              <div className="flex items-center gap-6">
                <div className="h-16 w-16 bg-primary rounded-full flex items-center justify-center text-white shadow-xl shadow-primary/20">
                  <span
                    className="material-symbols-outlined text-3xl"
                    data-icon="task_alt"
                  >
                    task_alt
                  </span>
                </div>
                <div>
                  <div className="flex gap-2 mb-1">
                    <span className="bg-tertiary-fixed text-on-tertiary-fixed-variant text-[10px] uppercase tracking-widest font-bold px-3 py-1 rounded-full">
                      Phase 3 of 5
                    </span>
                    <span className="bg-surface-container-highest text-on-surface-variant text-[10px] uppercase tracking-widest font-bold px-3 py-1 rounded-full">
                      Contract ID: #FJ-2849
                    </span>
                  </div>
                  <h3 className="font-bold text-lg">Fintech App Redesign</h3>
                </div>
              </div>
              <div className="flex gap-8">
                <div className="text-right">
                  <p className="text-[10px] uppercase tracking-widest font-bold text-on-surface-variant mb-1">
                    Due Date
                  </p>
                  <p className="font-bold text-lg">Oct 24, 2023</p>
                </div>
                <div className="text-right">
                  <p className="text-[10px] uppercase tracking-widest font-bold text-on-surface-variant mb-1">
                    Value
                  </p>
                  <p className="font-bold text-lg text-primary">$1,250.00</p>
                </div>
              </div>
            </div>

            {/* Submission Form */}
            <form className="space-y-10">
              {/* Text Area */}
              <div className="space-y-3">
                <label className="block text-[10px] uppercase tracking-widest font-bold text-on-surface-variant ml-2">
                  Message to Client
                </label>
                <textarea
                  className="w-full bg-surface-container-lowest border-none ring-1 ring-outline-variant focus:ring-2 focus:ring-tertiary-container focus:shadow-[0_0_8px_rgba(0,67,95,0.2)] rounded-md p-6 text-on-surface placeholder:text-outline/50 transition-all outline-none"
                  placeholder="Describe the work completed, mention any specific files to look at, or provide instructions for testing..."
                  rows={5}
                ></textarea>
              </div>

              {/* File Upload Section */}
              <div className="space-y-3">
                <label className="block text-[10px] uppercase tracking-widest font-bold text-on-surface-variant ml-2">
                  Deliverables & Assets
                </label>
                <div className="group relative bg-surface-container-low border-2 border-dashed border-outline-variant rounded-lg p-12 flex flex-col items-center justify-center transition-all hover:bg-surface-container hover:border-primary-container cursor-pointer overflow-hidden">
                  <div className="bg-white p-4 rounded-full shadow-sm mb-4 group-hover:scale-110 transition-transform">
                    <span
                      className="material-symbols-outlined text-4xl text-primary"
                      data-icon="cloud_upload"
                    >
                      cloud_upload
                    </span>
                  </div>
                  <h4 className="text-lg font-bold text-on-surface">
                    Drop files here or click to browse
                  </h4>
                  <p className="text-on-surface-variant text-sm mt-1">
                    Maximum file size: 50MB. Zip, PDF, PNG, JPG, or DOCX.
                  </p>
                  <input
                    className="absolute inset-0 opacity-0 cursor-pointer"
                    type="file"
                  />
                </div>

                {/* Mock Uploaded Files */}
                <div className="grid grid-cols-1 md:grid-cols-2 gap-4 mt-4">
                  <div className="flex items-center justify-between bg-white p-4 rounded-md border border-outline-variant/30">
                    <div className="flex items-center gap-3">
                      <span
                        className="material-symbols-outlined text-primary"
                        data-icon="folder_zip"
                      >
                        folder_zip
                      </span>
                      <span className="text-sm font-medium">
                        api_integration_v1.zip
                      </span>
                    </div>
                    <button
                      className="text-error hover:bg-error-container/20 p-1 rounded-full transition-all"
                      type="button"
                    >
                      <span
                        className="material-symbols-outlined text-lg"
                        data-icon="close"
                      >
                        close
                      </span>
                    </button>
                  </div>
                </div>
              </div>

              {/* Links Input */}
              <div className="space-y-3">
                <label className="block text-[10px] uppercase tracking-widest font-bold text-on-surface-variant ml-2">
                  External Links
                </label>
                <div className="flex gap-3">
                  <div className="relative flex-1">
                    <span
                      className="material-symbols-outlined absolute left-4 top-1/2 -translate-y-1/2 text-outline"
                      data-icon="link"
                    >
                      link
                    </span>
                    <input
                      className="w-full bg-surface-container-lowest border-none ring-1 ring-outline-variant focus:ring-2 focus:ring-tertiary-container rounded-md pl-12 pr-4 py-4 text-on-surface outline-none"
                      placeholder="https://github.com/username/project"
                      type="url"
                    />
                  </div>
                  <button
                    className="bg-secondary-container text-on-secondary-container font-bold px-6 rounded-md hover:bg-surface-container-highest transition-all scale-95 active:scale-90"
                    type="button"
                  >
                    Add Link
                  </button>
                </div>
              </div>

              {/* Actions */}
              <div className="flex flex-col sm:flex-row items-center gap-6 pt-6">
                <button
                  className="w-full sm:w-auto bg-gradient-to-br from-primary to-primary-container text-white px-12 py-5 rounded-full font-bold text-lg shadow-xl shadow-primary/25 hover:shadow-2xl hover:shadow-primary/40 transition-all scale-95 active:scale-90"
                  type="submit"
                >
                  Submit for Review
                </button>
                <button
                  className="w-full sm:w-auto text-primary font-bold px-8 py-4 hover:bg-surface-container-highest rounded-full transition-all scale-95 active:scale-90"
                  type="button"
                >
                  Save as Draft
                </button>
              </div>
            </form>
          </div>

          {/* Right Sidebar Info */}
          <aside className="lg:col-span-4 space-y-8">
            <div className="bg-surface-container-highest/30 rounded-lg p-8 border border-white">
              <h3 className="text-xl font-extrabold tracking-tight mb-6 flex items-center gap-3">
                <span
                  className="material-symbols-outlined text-primary"
                  data-icon="tips_and_updates"
                >
                  tips_and_updates
                </span>
                Submission Tips
              </h3>
              <ul className="space-y-6">
                <li className="flex gap-4">
                  <span className="flex-shrink-0 h-6 w-6 bg-primary/10 text-primary rounded-full flex items-center justify-center text-[10px] font-bold">
                    1
                  </span>
                  <div>
                    <p className="text-sm font-bold text-on-surface mb-1">
                      Be Detailed
                    </p>
                    <p className="text-sm text-on-surface-variant leading-relaxed">
                      Clearly explain what you've accomplished to help the
                      client review faster.
                    </p>
                  </div>
                </li>
                <li className="flex gap-4">
                  <span className="flex-shrink-0 h-6 w-6 bg-primary/10 text-primary rounded-full flex items-center justify-center text-[10px] font-bold">
                    2
                  </span>
                  <div>
                    <p className="text-sm font-bold text-on-surface mb-1">
                      Verify Links
                    </p>
                    <p className="text-sm text-on-surface-variant leading-relaxed">
                      Ensure all external repositories or design files have
                      public access enabled.
                    </p>
                  </div>
                </li>
                <li className="flex gap-4">
                  <span className="flex-shrink-0 h-6 w-6 bg-primary/10 text-primary rounded-full flex items-center justify-center text-[10px] font-bold">
                    3
                  </span>
                  <div>
                    <p className="text-sm font-bold text-on-surface mb-1">
                      Review Files
                    </p>
                    <p className="text-sm text-on-surface-variant leading-relaxed">
                      Double-check that you're uploading the final versions of
                      your work.
                    </p>
                  </div>
                </li>
              </ul>
            </div>

            {/* Next Milestone Preview */}
            <div className="bg-indigo-900 rounded-lg p-8 text-white relative overflow-hidden group">
              <div className="relative z-10">
                <p className="text-[10px] uppercase tracking-widest font-bold opacity-70 mb-2">
                  Next Milestone
                </p>
                <h4 className="text-lg font-bold mb-4">
                  Milestone 4: Dashboard UI Implementation
                </h4>
                <div className="h-1.5 w-full bg-white/20 rounded-full mb-6">
                  <div className="h-full w-3/5 bg-tertiary-fixed rounded-full"></div>
                </div>
                <button className="flex items-center gap-2 text-sm font-bold text-tertiary-fixed hover:text-white transition-colors">
                  View Roadmap{" "}
                  <span
                    className="material-symbols-outlined text-sm"
                    data-icon="arrow_forward"
                  >
                    arrow_forward
                  </span>
                </button>
              </div>
              {/* Decorative background element */}
              <div className="absolute -right-4 -bottom-4 h-32 w-32 bg-white/5 rounded-full blur-2xl group-hover:bg-white/10 transition-all"></div>
            </div>

            <div className="p-4 border-l-4 border-primary/20 bg-surface-container-low rounded-r-lg">
              <p className="text-xs font-bold text-on-surface-variant italic leading-relaxed">
                "Your submission will be reviewed by the client. Once approved,
                the funds held in escrow ($1,250.00) will be released to your
                wallet."
              </p>
            </div>
          </aside>
        </div>
      </main>

      {/* BottomNavBar (Mobile Only) */}
      <nav className="md:hidden fixed bottom-0 left-0 w-full flex justify-around items-center px-6 pb-6 pt-3 bg-white/80 dark:bg-slate-900/80 backdrop-blur-2xl z-50 rounded-t-[2.5rem] shadow-[0_-10px_40px_rgba(0,0,0,0.04)]">
        <a
          className="flex flex-col items-center justify-center text-slate-400 dark:text-slate-500 p-2 hover:text-indigo-500 transition-all active:scale-90 duration-300"
          href="#"
        >
          <span className="material-symbols-outlined" data-icon="home">
            home
          </span>
          <span className="text-[10px] uppercase tracking-widest font-['Inter'] mt-1">
            Home
          </span>
        </a>
        <a
          className="flex flex-col items-center justify-center bg-indigo-600 text-white rounded-full p-3 mb-2 transform -translate-y-2 shadow-lg shadow-indigo-500/40 active:scale-90 duration-300"
          href="#"
        >
          <span className="material-symbols-outlined" data-icon="work">
            work
          </span>
          <span className="text-[10px] uppercase tracking-widest font-['Inter'] mt-1">
            Projects
          </span>
        </a>
        <a
          className="flex flex-col items-center justify-center text-slate-400 dark:text-slate-500 p-2 hover:text-indigo-500 transition-all active:scale-90 duration-300"
          href="#"
        >
          <span className="material-symbols-outlined" data-icon="chat_bubble">
            chat_bubble
          </span>
          <span className="text-[10px] uppercase tracking-widest font-['Inter'] mt-1">
            Chat
          </span>
        </a>
        <a
          className="flex flex-col items-center justify-center text-slate-400 dark:text-slate-500 p-2 hover:text-indigo-500 transition-all active:scale-90 duration-300"
          href="#"
        >
          <span
            className="material-symbols-outlined"
            data-icon="account_balance_wallet"
          >
            account_balance_wallet
          </span>
          <span className="text-[10px] uppercase tracking-widest font-['Inter'] mt-1">
            Wallet
          </span>
        </a>
      </nav>
    </div>
  );
};

export default submitMilestonePage;
