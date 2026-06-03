"use client";

import { useState } from "react";
import { useGetMeQuery } from "@/api/userapi";
import { useWithdrawMutation } from "@/api/walletapi";
import {
  validatePhoneNumber,
  validatePositiveDecimal,
} from "@/lib/fieldValidation";

interface WithdrawCardProps {
  currency: string;
}

const WithdrawCard = ({ currency }: WithdrawCardProps) => {
  const { data: user } = useGetMeQuery();
  const [withdraw, { isLoading }] = useWithdrawMutation();

  const [amount, setAmount] = useState("");
  const [accountNumber, setAccountNumber] = useState("");
  const [bankCode, setBankCode] = useState("855"); // Defaulting to TeleBirr
  const [statusMessage, setStatusMessage] = useState<{
    text: string;
    isError: boolean;
  } | null>(null);
  const amountError = amount.trim()
    ? validatePositiveDecimal(amount, "Amount")
    : null;
  const accountNumberError = validatePhoneNumber(
    accountNumber,
    "Account / Phone Number",
  );

  const handleWithdraw = async () => {
    setStatusMessage(null);

    if (!amount || !accountNumber || !bankCode) {
      setStatusMessage({ text: "Please fill in all fields.", isError: true });
      return;
    }

    const submitAmountError = validatePositiveDecimal(amount, "Amount");
    if (submitAmountError) {
      setStatusMessage({ text: submitAmountError, isError: true });
      return;
    }

    if (accountNumberError) {
      setStatusMessage({ text: accountNumberError, isError: true });
      return;
    }

    try {
      const response = await withdraw({
        account_number: accountNumber,
        amount: amount, // API expects a string value
        bank_code: bankCode,
        currency: currency,
      }).unwrap();

      if (response.success) {
        setStatusMessage({
          text: response.message || "Withdrawal initiated successfully!",
          isError: false,
        });
        setAmount("");
        setAccountNumber("");
      } else {
        setStatusMessage({
          text: response.message || "Withdrawal failed.",
          isError: true,
        });
      }
    } catch {
      setStatusMessage({
        text: "Unable to process withdrawal request.",
        isError: true,
      });
    }
  };

  return (
    <div className="flex flex-col gap-5 rounded-xl border border-slate-200 bg-white p-6 shadow-sm">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <span className="inline-flex h-8 w-8 items-center justify-center rounded-md bg-indigo-100 text-indigo-700">
            <svg
              xmlns="http://www.w3.org/2000/svg"
              className="h-4 w-4"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
              strokeWidth="2"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                d="M12 19V5m0 14l-4-4m4 4l4-4"
              />
            </svg>
          </span>
          <h2 className="text-lg font-semibold text-slate-800">
            Withdraw Funds
          </h2>
        </div>
        <span className="rounded-full bg-indigo-50 px-3 py-1 text-xs font-semibold text-indigo-700">
          Secure Payout
        </span>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        {/* Amount */}
        <div className="flex flex-col gap-1.5">
          <label
            htmlFor="withdraw-amount"
            className="text-xs font-semibold uppercase tracking-wide text-slate-500"
          >
            Amount ({currency})
          </label>
          <input
            id="withdraw-amount"
            type="number"
            placeholder="0.00"
            value={amount}
            onChange={(e) => setAmount(e.target.value)}
            aria-invalid={Boolean(amountError)}
            className={`h-11 w-full rounded-lg border border-slate-300 bg-white px-3 text-sm text-slate-700 outline-none transition focus:border-indigo-500 focus:ring-2 focus:ring-indigo-100 ${
              amountError ? "border-rose-500 focus:border-rose-500 focus:ring-rose-100" : ""
            }`}
          />
          {amountError ? (
            <p className="text-xs font-medium text-rose-500">{amountError}</p>
          ) : null}
        </div>

        {/* Account Number */}
        <div className="flex flex-col gap-1.5">
          <label
            htmlFor="withdraw-account"
            className="text-xs font-semibold uppercase tracking-wide text-slate-500"
          >
            Account / Phone Number
          </label>
          <input
            id="withdraw-account"
            type="text"
            inputMode="tel"
            placeholder={user?.phone_number || "09xxxxxx"}
            value={accountNumber}
            onChange={(e) => setAccountNumber(e.target.value)}
            aria-invalid={Boolean(accountNumberError)}
            className={`h-11 w-full rounded-lg border border-slate-300 bg-white px-3 text-sm text-slate-700 outline-none transition focus:border-indigo-500 focus:ring-2 focus:ring-indigo-100 ${
              accountNumberError ? "border-rose-500 focus:border-rose-500 focus:ring-rose-100" : ""
            }`}
          />
          {accountNumberError ? (
            <p className="text-xs font-medium text-rose-500">
              {accountNumberError}
            </p>
          ) : null}
        </div>

        {/* Bank Code */}
        <div className="flex flex-col gap-1.5">
          <label
            htmlFor="withdraw-bank"
            className="text-xs font-semibold uppercase tracking-wide text-slate-500"
          >
            Payment Provider
          </label>
          <select
            id="withdraw-bank"
            value={bankCode}
            onChange={(e) => setBankCode(e.target.value)}
            className="h-11 w-full rounded-lg border border-slate-300 bg-white px-3 text-sm text-slate-700 outline-none transition focus:border-indigo-500 focus:ring-2 focus:ring-indigo-100"
          >
            <option value="855">TeleBirr</option>
            <option value="128">CBEBIRR</option>
          </select>
        </div>
      </div>

      <button
        type="button"
        onClick={handleWithdraw}
        disabled={isLoading}
        className="inline-flex h-11 items-center justify-center gap-2 rounded-lg bg-indigo-600 px-4 text-sm font-semibold text-white transition hover:bg-indigo-700 disabled:opacity-50"
      >
        {isLoading ? "Processing..." : "Request Withdrawal"}
      </button>

      {statusMessage && (
        <p
          className={`text-sm ${statusMessage.isError ? "text-rose-500" : "text-emerald-600"}`}
        >
          {statusMessage.text}
        </p>
      )}
    </div>
  );
};

export default WithdrawCard;
