// components/submit/MilestoneCard.tsx

export function MilestoneCard() {
  return (
    <div className="bg-white rounded-xl p-6 md:p-8 flex flex-col md:flex-row justify-between gap-6 shadow-sm">
      
      <div className="flex items-center gap-4">
        <div className="h-14 w-14 bg-indigo-600 text-white flex items-center justify-center rounded-full text-xl">
          ✓
        </div>

        <div>
          <p className="text-xs font-bold uppercase text-gray-400">
            Phase 3 of 5
          </p>
          <h3 className="font-bold text-lg">
            Fintech App Redesign
          </h3>
        </div>
      </div>

      <div className="flex gap-6 md:gap-10 text-sm md:text-base">
        <div>
          <p className="text-xs text-gray-400">Due Date</p>
          <p className="font-semibold">Oct 24, 2023</p>
        </div>

        <div>
          <p className="text-xs text-gray-400">Value</p>
          <p className="font-semibold text-indigo-600">$1,250</p>
        </div>
      </div>
    </div>
  )
}