import Image from "next/image";
type ContractJobCardProps = {
  milestoneCurrent: number;
  milestoneTotal: number;
  employer: string;
  projectName: string;
  milestoneDescription: string;
  dueDate: string;
  amountToBePaid: string;
  status: string;
  clientId: string;
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
  status,
  clientId,
    profilePictureUrl,
}: ContractJobCardProps) => {
  return (
    <div>
        <div>
            {/* card title */}
            <div className="flex">
                <Image src={profilePictureUrl || "/default-profile.png"} alt={`${employer}'s profile picture`} className="w-10 h-10 rounded-full mr-4" />
                <div>
                    <h3 className="text-lg font-semibold">{projectName}</h3>
                    <p className="text-sm text-gray-500">{employer}</p>
                </div>
                <p className="text-sm font-medium text-green-500">{status}</p>
            </div>
            <div>
                <span>
                    <p> milestone {milestoneCurrent} of {milestoneTotal}</p> 
                    <p>{milestoneCurrent/milestoneTotal * 100}%</p>
                 </span>
                 <progress className="progress w-56" value={milestoneCurrent/milestoneTotal * 100} max="100"></progress>
                 <p>{milestoneDescription}</p>
              
            </div>
        </div>

        {/* the buttons */}
        <div>

        </div>




    </div>
    >
    );
};

export default ContractJobCard;
