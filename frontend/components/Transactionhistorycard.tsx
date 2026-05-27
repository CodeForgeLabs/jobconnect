"use client";

import { useGetWalletTransactionsQuery } from "@/api/walletapi";

const formatDate = (value: string) =>
  new Intl.DateTimeFormat("en-US", {
    month: "short",
    day: "2-digit",
    year: "numeric",
  }).format(new Date(value));

const TransactionHistoryCard = () => {
  const { data: transactions = [] } = useGetWalletTransactionsQuery();

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
          {transactions.length === 0 ? (
            <div className="grid grid-cols-4 items-center gap-4 py-4">
              <p className="col-span-4 text-sm text-gray-500">
                No transactions yet.
              </p>
            </div>
          ) : (
            transactions.map((transaction) => {
              const isPositive = transaction.Status === "SUCCESS";
              const amountPrefix = isPositive ? "+" : "-";
              const statusStyles =
                transaction.Status === "SUCCESS"
                  ? "bg-emerald-50 text-emerald-700"
                  : transaction.Status === "PENDING"
                    ? "bg-yellow-50 text-yellow-700"
                    : "bg-gray-100 text-gray-600";

              return (
                <div
                  key={transaction.ID}
                  className="grid grid-cols-4 items-center gap-4 py-4"
                >
                  <div className="flex items-center gap-3">
                    <span
                      className={`inline-flex h-9 w-9 items-center justify-center rounded-full ${
                        isPositive
                          ? "bg-emerald-100 text-emerald-700"
                          : "bg-jobBlue text-white"
                      }`}
                    >
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
                      <p className="font-medium text-gray-900">
                        {transaction.Description ||
                          transaction.Type ||
                          transaction.TxRef}
                      </p>
                      <p className="text-xs text-gray-500">
                        {transaction.Type || transaction.TxRef}
                      </p>
                    </div>
                  </div>
                  <p>{formatDate(transaction.CreatedAt)}</p>
                  <span
                    className={`inline-flex items-center w-fit rounded-full px-2.5 py-1 text-xs font-semibold ${statusStyles}`}
                  >
                    {transaction.Status || "Unknown"}
                  </span>
                  <p
                    className={`self-center text-right font-semibold ${
                      isPositive ? "text-emerald-700" : "text-gray-500"
                    }`}
                  >
                    {amountPrefix} {transaction.AmountMinor}
                    birr
                  </p>
                </div>
              );
            })
          )}
        </div>
      </div>
    </div>
  );
};

export default TransactionHistoryCard;
