import ContractJobCard from "@/components/Contractjobcard";
import { MilestoneCard } from "@/components/milestoncard";
const MyContractsPage = () => {
  return (
    <div className="flex flex-col gap p-8 bg-[#eff1f5] min-h-screen">
      <div>
        <div>
          <h1 className="text-4xl text-jobBlue">My Contracts</h1>
          <p className="text-sm text-gray-500">
            View and manage your active contracts
          </p>
        </div>
      </div>
      <div className="flex gap-4 justify-between mt-6">
        <div className="flex flex-col gap-6 w-[70%]">
          <div className="flex items-center gap-8">
            <ul className="flex items-center gap-4 rounded-full bg-[#f3f4f6] p-1.25 shadow-sm">
              <li>
                <button
                  type="button"
                  className="rounded-full bg-white px-8 py-1.5 text-sm font-semibold text-jobBlue shadow-sm"
                >
                  Active
                </button>
              </li>
              <li>
                <button
                  type="button"
                  className="rounded-full px-8 py-1.5 text-sm font-semibold text-gray-400 hover:bg-white/80"
                >
                  Completed
                </button>
              </li>
              <li>
                <button
                  type="button"
                  className="rounded-full px-8 py-1.5 text-sm font-semibold text-gray-400 hover:bg-white/80"
                >
                  Pending
                </button>
              </li>
            </ul>

            <div className="relative">
              <svg
                xmlns="http://www.w3.org/2000/svg"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                strokeWidth="2"
                className="pointer-events-none absolute left-4 top-1/2 h-4 w-4 -translate-y-1/2 text-[#9aa8bd]"
              >
                <circle cx="11" cy="11" r="7" />
                <path d="m20 20-3.5-3.5" />
              </svg>
              <input
                type="text"
                placeholder="Search contracts..."
                className="w-64 rounded-full border border-[#dbe1ea] bg-white py-2 pl-11 pr-5 text-sm text-jobBlue placeholder:text-[#9aa8bd] shadow-sm outline-none focus:border-jobBlue"
              />
            </div>
          </div>
          <div className="flex flex-wrap items-center gap-4">
            <ContractJobCard
              milestoneCurrent={2}
              milestoneTotal={5}
              employer="Aster Technologies"
              projectName="JobConnect Frontend Revamp"
              milestoneDescription="Build and deliver the dashboard UI with fully responsive contract filtering and search experience."
              dueDate="Apr 20, 2026"
              amountToBePaid="$1,250"
              ContractValue="$5,000"
              status="Active"
              clientId="CL-20417"
            />
            <ContractJobCard
              milestoneCurrent={2}
              milestoneTotal={5}
              employer="Aster Technologies"
              projectName="JobConnect Frontend Revamp"
              milestoneDescription="Build and deliver the dashboard UI with fully responsive contract filtering and search experience."
              dueDate="Apr 20, 2026"
              amountToBePaid="$1,250"
              ContractValue="$5,000"
              status="Active"
              clientId="CL-20417"
            />
          </div>
          <MilestoneCard/>
        </div>

        {/* //contracts stats on the right*/}
        <div className=" bg-white">fdgdfg</div>
      </div>
    </div>
  );
};

export default MyContractsPage;
