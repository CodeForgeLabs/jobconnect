"use client";

import { useState } from "react";
import { useBuyConnectMutation } from "@/api/walletapi";

interface ConnectPackage {
  id: number;
  connects: number;
  price: number;
  description: string;
  bestValue?: boolean;
}

const packages: ConnectPackage[] = [
  {
    id: 1,
    connects: 10,
    price: 100,
    description: "For just starting",
  },
  {
    id: 2,
    connects: 50,
    price: 400,
    description: "Most popular option",
    bestValue: true,
  },
  {
    id: 3,
    connects: 100,
    price: 700,
    description: "For active users",
  },
];

const BuyconnectsCard = () => {
  const [selectedPackage, setSelectedPackage] = useState<number>(2);
  const [buyConnect, { isLoading }] = useBuyConnectMutation();
  const [statusMessage, setStatusMessage] = useState<string | null>(null);

  const handlePurchase = async () => {
    const buyPackage = packages.find((pkg) => pkg.id === selectedPackage);
    if (buyPackage) {
      try {
        const response = await buyConnect({
          amount: buyPackage.connects,
        }).unwrap();
        setStatusMessage(response.message);
      } catch {
        setStatusMessage("Unable to buy connects");
      }
    }
  };

  return (
    <div className="flex grow h-fit flex-col gap-5 rounded-xl bg-white p-8 shadow-sm">
      <div className="flex items-center gap-3">
        <span className="inline-flex h-9 w-9 items-center justify-center rounded-md bg-jobBlue text-white shadow-sm">
          <svg
            xmlns="http://www.w3.org/2000/svg"
            className="h-4 w-4"
            viewBox="0 0 24 24"
            fill="none"
            aria-hidden="true"
          >
            <path
              d="M13 2 4 14h6l-1 8 11-14h-6l-1-6Z"
              stroke="currentColor"
              strokeWidth="1.8"
              strokeLinejoin="round"
            />
          </svg>
        </span>

        <p className="text-lg font-semibold text-slate-800">Buy connects</p>
      </div>

      <div className="flex flex-col gap-2">
        {packages.map((pkg) => {
          const isSelected = selectedPackage === pkg.id;

          return (
            <button
              key={pkg.id}
              type="button"
              onClick={() => setSelectedPackage(pkg.id)}
              className={`relative rounded-lg p-4 text-left transition-all duration-200 ${
                isSelected
                  ? "border-2 border-jobBlue bg-blue-50"
                  : "border border-gray-200 hover:border-jobBlue/50"
              }`}
            >
              {pkg.bestValue && (
                <span className="pointer-events-none absolute right-2 -top-2 rounded-full bg-jobBlue px-2.5 py-1 text-[8px] font-semibold uppercase tracking-wide text-white">
                  Best value
                </span>
              )}

              <span className="flex justify-between text-sm">
                <p>{pkg.connects} Connects</p>

                <p className="text-jobBlue">{pkg.price} birr</p>
              </span>

              <span className="mt-1 flex justify-between">
                <p className="text-xs text-gray-500">{pkg.description}</p>

                <p
                  className={`text-xs font-medium ${
                    isSelected ? "text-jobBlue" : "text-gray-500"
                  }`}
                >
                  {isSelected ? "Selected" : "Select"}
                </p>
              </span>
            </button>
          );
        })}
      </div>
      <span>
        {statusMessage && (
          <p className="text-sm text-red-500">{statusMessage}</p>
        )}
      </span>

      <button
        type="button"
        disabled={isLoading}
        className="mt-4 w-full rounded-lg bg-jobBlue px-6 py-3 text-sm font-semibold text-white hover:opacity-80 disabled:cursor-not-allowed disabled:opacity-60"
        onClick={handlePurchase}
      >
        {isLoading ? "Processing..." : "Purchase Now"}
      </button>
    </div>
  );
};

export default BuyconnectsCard;
