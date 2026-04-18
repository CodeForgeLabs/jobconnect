import Image from "next/image";
type ContractJobCardProps = {
  milestoneCurrent: number;
  milestoneTotal: number;
  employer: string;
  projectName: string;
  milestoneDescription: string;
  dueDate: string;
  amountToBePaid: string;
  ContractValue: string;
  status: string;
  profilePictureUrl?: string;
};

const ContractJobCard = ({
  milestoneCurrent,
  milestoneTotal,
  employer,
  projectName,
  milestoneDescription,
  dueDate,
  amountToBePaid,
  ContractValue,
  status,
  profilePictureUrl,
}: ContractJobCardProps) => {
  return (
    <div className="flex pc:flex-wrap gap-5 bg-white rounded-lg shadow-md p-6  pc:w-[48%]">
      <div className="flex flex-col gap-5">
        {/* card title */}
        <div className="flex   gap-1">
          <Image
            src={profilePictureUrl || "/default-profile-gray.svg"}
            alt={`${employer}'s profile picture`}
            width={40}
            height={40}
            className="w-10 h-10 rounded-full mr-1"
          />
          <div className="flex-1 min-w-0">
            <h3 className="text-sm font-semibold text-gray-800 truncate">
              {projectName}
            </h3>
            <p className="text-[11px] text-gray-500 mt-0.5">{employer}</p>
          </div>
          <p className="text-[10px] font-medium text-green-700 bg-green-50 px-2 py-1 rounded-full whitespace-nowrap">
            {status}
          </p>
        </div>
        <div className="flex flex-col ">
          <span className="flex justify-between">
            <p className="text-[11px] text-gray-500">
              {" "}
              milestone {milestoneCurrent} of {milestoneTotal}
            </p>
            <p className="text-[11px] font-medium text-gray-700">
              {(milestoneCurrent / milestoneTotal) * 100}%
            </p>
          </span>
          <div
            className="h-2 w-full overflow-hidden rounded-full bg-blue-100"
            role="progressbar"
            aria-valuemin={0}
            aria-valuemax={100}
            aria-valuenow={(milestoneCurrent / milestoneTotal) * 100}
          >
            <div
              className="h-full rounded-full bg-linear-to-r from-sky-200 via-blue-300 to-blue-500"
              style={{
                width: `${(milestoneCurrent / milestoneTotal) * 100}%`,
              }}
            />
          </div>
          <p className="text-xs text-gray-600  mt-4 ">
            {milestoneDescription}
          </p>
        </div>
        <div className="flex gap-9">
          <span className="flex flex-col ">
            <p className="text-[11px] text-gray-400">Contract Value</p>
            <p className="text-xs font-semibold text-gray-700">
              {ContractValue}
            </p>
          </span>
          <span className="flex flex-col ">
            <p className="text-[11px] text-gray-400">Next Payment</p>
            <p className="text-xs font-semibold text-gray-700">
              {amountToBePaid}
            </p>
          </span>
          <span className="flex flex-col ">
            <p className="text-[11px] text-gray-400">Due Date</p>
            <p className="text-xs font-semibold text-gray-700">{dueDate}</p>
          </span>
        </div>
      </div>

      {/* the buttons */}
      <div className="flex flex-col gap-3 max-pc:justify-end  pc:w-full pc:flex-row ">
        <button
          type="button"
          className=" pc:grow rounded-full bg-gray-200  px-4 py-2 text-[11px] font-semibold text-gray-700 shadow-sm hover:bg-blue-600 pc:items-stretch"
        >
          Submit work
        </button>
        <button
          type="button"
          aria-label="Message"
          className="inline-flex items-center justify-center gap-2 rounded-full bg-jobBlue px-4 py-2 text-[11px] font-semibold text-white shadow-sm hover:bg-gray-300"
        >
          <span className="max-pc:inline pc:hidden">Message</span>
          <svg
            aria-hidden="true"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            strokeWidth="2"
            className="hidden h-4 w-4 pc:inline-block"
          >
            <path d="M21 15a4 4 0 0 1-4 4H8l-5 3V7a4 4 0 0 1 4-4h10a4 4 0 0 1 4 4z" />
          </svg>
        </button>
      </div>
    </div>
  );
};

export default ContractJobCard;
