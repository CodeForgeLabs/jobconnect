const BuyconnectsCard = () => {
  return (
    <div className="flex flex-col gap-5 rounded-xl bg-white p-8 shadow-sm">
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
        <div className="rounded-lg border border-gray-200 p-4">
          <span className="flex justify-between text-sm ">
            <p>10 Connects</p>
            <p className="text-jobBlue">100 birr</p>
          </span>
          <span className="flex justify-between">
            <p className="text-xs text-gray-500">For just starting</p>
            <p className="text-xs text-gray-500">select</p>
          </span>
        </div>

        <div className="relative rounded-lg border-2 border-jobBlue p-4 ">
          <span className="pointer-events-none absolute right-2 top-[-10] rounded-full bg-jobBlue px-2.5 py-1 text-[8px] font-semibold uppercase tracking-wide text-white">
            Best value
          </span>
          <span className="flex justify-between text-sm">
            <p>50 Connects</p>
            <p className="text-jobBlue">400 birr</p>
          </span>
          <span className="flex justify-between">
            <p className="text-xs text-gray-500">Most popular option</p>
            <p className="text-xs text-gray-500">select</p>
          </span>
        </div>

        <div className="rounded-lg border border-gray-200 p-4">
          <span className="flex justify-between text-sm">
            <p>100 Connects</p>
            <p className="text-jobBlue">700 birr</p>
          </span>
          <span className="flex justify-between">
            <p className="text-xs text-gray-500">For active users</p>
            <p className="text-xs text-gray-500">select</p>
          </span>
        </div>
      </div>

      <button
        type="button"
        className="mt-4 w-full text-sm font-semibold text-white bg-jobBlue hover:opacity-80 py-3 px-6 rounded-lg"
      >
        Purchase Now
      </button>
    </div>
  );
};

export default BuyconnectsCard;
