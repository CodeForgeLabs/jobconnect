const TransactionHistoryCard = () => {
  return (
    <div className="rounded-lg border border-gray-200 bg-white">
      <div className="flex justify-between border-b border-gray-200 p-6">
        <h2 className="text-lg font-semibold text-gray-800">
          Transaction History
        </h2>
        <p className="text-sm text-jobBlue">View all</p>
      </div>

      <div className="overflow-hidden rounded-b-lg">
        <div className="grid grid-cols-4 gap-4 border-b border-gray-100 bg-gray-50 px-6 py-3 text-xs font-semibold uppercase tracking-wide text-gray-500">
          <p>Transaction / Type</p>
          <p>Date</p>
          <p>Status</p>
          <p className="text-right">Amount</p>
        </div>

        <div className="divide-y divide-gray-100 px-6 text-sm text-gray-700">
          <div className="grid grid-cols-4 items-center gap-4 py-4">
            <div className="flex items-center gap-3">
              <span className="inline-flex h-9 w-9 items-center justify-center rounded-full bg-green-100 text-green-700">
                <svg
                  className="h-5 w-5"
                  aria-hidden="true"
                  xmlns="http://www.w3.org/2000/svg"
                  fill="none"
                  viewBox="0 0 24 24"
                >
                  <path
                    stroke="currentColor"
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth="1.8"
                    d="M4 20h16M5 10h14M6 10v8M10 10v8M14 10v8M18 10v8M4 10 12 4l8 6"
                  />
                </svg>
              </span>
              <div>
                <p className="font-medium text-gray-900">TeleBirr Withdraw</p>
                <p className="text-xs text-gray-500">Withdrawal</p>
              </div>
            </div>
            <p>Apr 05, 2026</p>
            <span className="inline-flex items-center w-fit rounded-full bg-emerald-50 p-1 text-xs font-semibold text-emerald-700">
              Completed
            </span>
            <p className="self-center text-right font-semibold text-gray-500">- 500 birr</p>
          </div>

          <div className="grid grid-cols-4 items-center gap-4 py-4">
            <div className="flex items-center gap-3">
              <span className="inline-flex h-9 w-9 items-center justify-center rounded-full bg-jobBlue text-white">
                <svg
                  className="h-4 w-4"
                  aria-hidden="true"
                  xmlns="http://www.w3.org/2000/svg"
                  fill="none"
                  viewBox="0 0 24 24"
                >
                  <path
                    stroke="currentColor"
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth="1.8"
                    d="M3 6h2l.4 2M7 13h10l3-7H6.4M7 13 6 8m1 5-1.3 5.2A1 1 0 0 0 6.7 19h10.6a1 1 0 0 0 1-.8L19 13M9 21h1m4 0h1"
                  />
                </svg>
              </span>
              <div>
                <p className="font-medium text-gray-900">Buy Connects</p>
                <p className="text-xs text-gray-500">Purchase</p>
              </div>
            </div>
            <p>Apr 03, 2026</p>
            <span className="inline-flex items-center w-fit rounded-full bg-yellow-50 px-2.5 py-1 text-xs font-semibold text-yellow-700">
              Pending
            </span>
            <p className="self-center text-right font-semibold text-yellow-700">- 400 birr</p>
          </div>

          <div className="grid grid-cols-4 items-center gap-4 py-4">
            <div className="flex items-center gap-3">
              <span className="inline-flex h-9 w-9 items-center justify-center rounded-full bg-yellow-100 text-yellow-700">
                <svg
                  className="h-5 w-5"
                  aria-hidden="true"
                  xmlns="http://www.w3.org/2000/svg"
                  fill="none"
                  viewBox="0 0 24 24"
                >
                  <path
                    stroke="currentColor"
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth="1.8"
                    d="M12 5v12m0 0-4-4m4 4 4-4M5 19h14"
                  />
                </svg>
              </span>
              <div>
                <p className="font-medium text-gray-900">Job Payment</p>
                <p className="text-xs text-gray-500">Earning</p>
              </div>
            </div>
            <p>Apr 01, 2026</p>
            <span className=" w-fit rounded-full bg-emerald-50 p-1 text-xs font-semibold text-emerald-700 items-center inline-flex">
              Completed
            </span>
            <p className="self-center text-right font-semibold text-emerald-700 ">
              + 1,250 birr
            </p>
          </div>
        </div>
      </div>
    </div>
  );
};

export default TransactionHistoryCard;
