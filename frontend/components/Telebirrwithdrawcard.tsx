const TeleBirrWithdrawCard = () => {
  return (
    <div className="flex flex-col  gap-5 rounded-xl border border-slate-200 bg-white p-6 shadow-sm">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <span className="inline-flex h-8 w-8 items-center justify-center rounded-md bg-yellow-100 text-yellow-700">
            <svg
              xmlns="http://www.w3.org/2000/svg"
              className="h-4 w-4"
              viewBox="0 0 24 24"
              fill="none"
              aria-hidden="true"
            >
              <path
                d="M4 9.5A2.5 2.5 0 0 1 6.5 7h11A2.5 2.5 0 0 1 20 9.5v7A2.5 2.5 0 0 1 17.5 19h-11A2.5 2.5 0 0 1 4 16.5v-7Z"
                stroke="currentColor"
                strokeWidth="1.8"
              />
              <path
                d="M16 13a1.5 1.5 0 1 0 0-3h4v3h-4Z"
                stroke="currentColor"
                strokeWidth="1.8"
                strokeLinejoin="round"
              />
              <path
                d="M7 7V6a1 1 0 0 1 1-1h8"
                stroke="currentColor"
                strokeWidth="1.8"
                strokeLinecap="round"
              />
            </svg>
          </span>
          <h2 className="text-lg font-semibold text-slate-800">
            Withdraw to TeleBirr
          </h2>
        </div>
        <span className="rounded-full bg-emerald-50 px-3 py-1 text-xs font-semibold text-emerald-700">
          Fast payout
        </span>
      </div>

      <div className="flex gap-6 grow ">
        <div className="flex flex-col gap-1.5 grow">
          <label
            htmlFor="telebirr-amount"
            className="text-xs font-semibold uppercase tracking-wide text-slate-500"
          >
            Amount
          </label>
          <div className="relative">
            <span className="pointer-events-none absolute inset-y-0 left-3 inline-flex items-center text-slate-400">
              <svg
                xmlns="http://www.w3.org/2000/svg"
                className="h-5 w-5"
                viewBox="0 0 24 24"
                fill="none"
                aria-hidden="true"
              >
                <path
                  d="M12 4v16M8.5 7.5h4a2.5 2.5 0 0 1 0 5h-1a2.5 2.5 0 0 0 0 5h4"
                  stroke="currentColor"
                  strokeWidth="1.8"
                  strokeLinecap="round"
                  strokeLinejoin="round"
                />
              </svg>
            </span>
            <input
              id="telebirr-amount"
              type="text"
              placeholder="$100.00"
              className="h-11 w-full rounded-lg border border-slate-300 bg-white pl-10 pr-3 text-sm text-slate-700 outline-none transition focus:border-emerald-500 focus:ring-2 focus:ring-emerald-100"
            />
          </div>
          <p className="text-xs text-slate-400">Min of 100 birr</p>
        </div>

        <div className="flex flex-col gap-1.5 grow">
          <label
            htmlFor="telebirr-phone"
            className="text-xs font-semibold uppercase tracking-wide text-slate-500"
          >
            TeleBirr Phone Number
          </label>
          <div className="relative">
            <span className="pointer-events-none absolute inset-y-0 left-3 inline-flex items-center text-slate-400">
              <svg
                xmlns="http://www.w3.org/2000/svg"
                className="h-5 w-5"
                viewBox="0 0 24 24"
                fill="none"
                aria-hidden="true"
              >
                <rect
                  x="7"
                  y="2.5"
                  width="10"
                  height="19"
                  rx="2"
                  stroke="currentColor"
                  strokeWidth="1.8"
                />
                <path
                  d="M10 5.5h4M11 18.5h2"
                  stroke="currentColor"
                  strokeWidth="1.8"
                  strokeLinecap="round"
                />
              </svg>
            </span>
            <input
              id="telebirr-phone"
              type="text"
              placeholder="09xxxxxx"
              className="h-11 w-full rounded-lg border border-slate-300 bg-white pl-10 pr-3 text-sm text-slate-700 outline-none transition focus:border-emerald-500 focus:ring-2 focus:ring-emerald-100"
            />
          </div>
          <p className="text-xs text-slate-400">
            Enter the phone number linked to your TeleBirr account
          </p>
        </div>
      </div>

      <div className="flex items-center justify-between rounded-lg bg-slate-50 px-3 py-2 text-xs text-slate-600">
        <span>Withdrawal fee: 1.5%</span>
        <span>Min: $10 | Max: $2,000</span>
      </div>

      <button
        type="button"
        className="inline-flex h-11 items-center justify-center gap-2 rounded-lg bg-emerald-600 px-4 text-sm font-semibold text-white transition hover:bg-emerald-700"
      >
        <svg
          xmlns="http://www.w3.org/2000/svg"
          className="h-4 w-4"
          viewBox="0 0 24 24"
          fill="none"
          aria-hidden="true"
        >
          <path
            d="M4 12h16M14 7l6 5-6 5"
            stroke="currentColor"
            strokeWidth="2"
            strokeLinecap="round"
            strokeLinejoin="round"
          />
        </svg>
        Withdraw Funds
      </button>
    </div>
  );
};

export default TeleBirrWithdrawCard;
