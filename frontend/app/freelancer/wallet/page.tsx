"use client";

import TeleBirrWithdrawCard from "@/components/Telebirrwithdrawcard";
import BuyconnectsCard from "@/components/Buyconnectscard";
import TransactionHistoryCard from "@/components/Transactionhistorycard";
import WithdrawCard from "@/components/Withdrawcard";
import {
  useGetWalletBalanceQuery,
  useGetWalletTransactionsQuery,
} from "@/api/walletapi";

import { useGetMeQuery } from "@/api/userapi";
const formatMoney = (currency: string, amountMinor: number) =>
  new Intl.NumberFormat("en-US", {
    style: "currency",
    currency,
    maximumFractionDigits: 2,
  }).format(amountMinor);

const WalletPage = () => {
  const { data: wallet } = useGetWalletBalanceQuery();
  const { data: transactions = [] } = useGetWalletTransactionsQuery();
  const { data: user } = useGetMeQuery();

  const currency = wallet?.Currency ?? "ETB";
  const currentBalance = wallet?.BalanceMinor ?? 0;

  const totalLifetimeEarnings = transactions.reduce((total, transaction) => {
    if (transaction.Status === "SUCCESS") {
      return total + transaction.AmountMinor;
    }

    return total;
  }, 0);

  const pendingClearance = transactions.reduce((total, transaction) => {
    if (transaction.Status === "PENDING") {
      return total + transaction.AmountMinor;
    }

    return total;
  }, 0);

  const totalConnects = transactions.reduce((total, transaction) => {
    if (
      transaction.Status === "SUCCESS" &&
      transaction.Description.toLowerCase().includes("connect")
    ) {
      return total + 1;
    }

    return total;
  }, 0);

  return (
    <div className="flex flex-col gap-6 p-7 bg-surface">
      <div>
        <h1>Wallet & Payments</h1>
        <p className="text-gray-500 text-sm">
          Manage your wallet and payment methods here.
        </p>
      </div>

      <div className="flex gap-6">
        <div className="relative flex grow flex-col overflow-hidden rounded-lg bg-jobBlue p-6 text-white shadow-lg">
          <h2 className=" text-[10px] font-medium">Current Balance</h2>
          <p className=" text-2xl font-bold leading-none">
            {formatMoney(currency, currentBalance)}
          </p>

          <span className="mt-4 inline-flex w-fit rounded-full bg-white/15 px-3 py-1 text-xs font-medium text-blue-50">
            +12.5% this month
          </span>

          <span
            className="pointer-events-none absolute bottom-0 right-0 translate-x-3 translate-y-3 text-white/20"
            aria-hidden="true"
          >
            <svg
              className="h-32 w-32"
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
                d="M13.6 16.733c.234.269.548.456.895.534a1.4 1.4 0 0 0 1.75-.762c.172-.615-.446-1.287-1.242-1.481-.796-.194-1.41-.861-1.241-1.481a1.4 1.4 0 0 1 1.75-.762c.343.077.654.26.888.524m-1.358 4.017v.617m0-5.939v.725M4 15v4m3-6v6M6 8.5 10.5 5 14 7.5 18 4m0 0h-3.5M18 4v3m2 8a5 5 0 1 1-10 0 5 5 0 0 1 10 0Z"
              />
            </svg>
          </span>
        </div>

        <div className="flex grow flex-col bg-white  rounded-lg  gap-5 p-6">
          <div className="flex items-start justify-between">
            <div className="flex flex-col ">
              <h2 className="text-[10px] font-medium text-gray-500">
                Total Lifetime Earnings{" "}
              </h2>
              <p className="text-2xl font-bold">
                {formatMoney(currency, totalLifetimeEarnings)}
              </p>
            </div>

            <span className="inline-flex h-9 w-9 items-center justify-center rounded-md bg-green-100 text-green-700">
              <svg
                className="w-6 h-6"
                aria-hidden="true"
                xmlns="http://www.w3.org/2000/svg"
                fill="none"
                viewBox="0 0 24 24"
              >
                <path
                  stroke="currentColor"
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth="2"
                  d="M4 20h16M5 10h14M6 10v8M10 10v8M14 10v8M18 10v8M4 10 12 4l8 6"
                />
              </svg>
            </span>
          </div>

          <div className="mt-1 flex gap-3  items-center px-9">
            <span className="flex flex-col grow border-l-2 border-gray-200 pl-3">
              <p className="text-[8px] font-semibold uppercase tracking-wide text-gray-500">
                Pending clearance
              </p>
              <p className="text-base font-bold ">
                {formatMoney(currency, pendingClearance)}
              </p>
            </span>

            <span className="flex flex-col grow border-l-2 border-gray-200 pl-3">
              <p className="text-[8px] font-semibold uppercase tracking-wide text-gray-500">
                Total connects
              </p>
              <p className="text-base font-bold">{user?.connect ?? 0}</p>
            </span>
          </div>
        </div>
      </div>

      <div className="flex gap-6">
        <div className="flex w-[70%] flex-col gap-5 ">
          <TeleBirrWithdrawCard />
          <WithdrawCard currency={currency} />
          <TransactionHistoryCard />
        </div>
        <div className=" w-[30%] ">
          <BuyconnectsCard />
        </div>
      </div>
    </div>
  );
};

export default WalletPage;
