import { ShoppingBag, Sparkles, Wallet } from "lucide-react";

interface JobcardProps {
  title: string;
  pay: string;
  type: "FIXED" | "HOURLY";
  description: string;
  postTime: string;
  tags: string[];
  companyName?: string;
  status?: string;
  skills?: string;
  jobType?: "FIXED" | "HOURLY";
  hourlyRate?: string;
  budget?: string;
  experienceLevel?: string;
  createdAt?: string;
  index?: number;
  onApply?: () => void;
}

const Jobcard = ({
  title,
  pay,
  type,
  description,
  postTime,
  tags,
  companyName,
  status,
  skills,
  jobType,
  hourlyRate,
  budget,
  experienceLevel,
  createdAt,
  index = 0,
  onApply,
}: JobcardProps) => {
  const displayCompany = companyName || "Client";
  const displayPostedAt = createdAt || postTime;
  const displayStatus = status === "OPEN" ? "Best Match" : status || type;
  const displayJobType = jobType || type;
  const displayPay =
    displayJobType === "HOURLY" ? `${hourlyRate || pay} / hr` : budget || pay;
  const displaySkills = skills
    ? skills
        .split(",")
        .map((skill) => skill.trim())
        .filter(Boolean)
    : tags;

  const icon =
    index === 0 ? (
      <Sparkles className="h-5 w-5 text-primary-container" aria-hidden="true" />
    ) : index === 1 ? (
      <Wallet className="h-5 w-5 text-primary-container" aria-hidden="true" />
    ) : (
      <ShoppingBag
        className="h-5 w-5 text-primary-container"
        aria-hidden="true"
      />
    );

  return (
    <div className="space-y-5 rounded-lg border border-outline-variant/10 bg-surface-container-lowest p-6 transition-all hover:border-primary/20">
      <div className="mb-4 flex flex-col items-start justify-between gap-4 sm:flex-row">
        <div className="flex gap-4">
          <div className="flex h-11 w-11 shrink-0 items-center justify-center rounded-xl bg-surface-container">
            {icon}
          </div>
          <div>
            <h4 className="text-lg font-bold text-on-surface">{title}</h4>
            <p className="text-sm text-on-surface-variant">
              {displayCompany} • Posted {displayPostedAt}
            </p>
          </div>
        </div>
        <span className="whitespace-nowrap rounded-full bg-tertiary-fixed px-4 py-1 text-xs font-bold uppercase tracking-wider text-on-tertiary-fixed-variant">
          {displayStatus}
        </span>
      </div>

      <p className="mb-5 line-clamp-2 text-sm leading-relaxed text-on-surface-variant">
        {description}
      </p>

      <div className="mb-5 flex flex-wrap gap-2">
        {displaySkills.slice(0, 3).map((skill) => (
          <span
            key={skill}
            className="rounded-md bg-surface-container px-3 py-1 text-xs font-medium text-on-surface-variant"
          >
            {skill}
          </span>
        ))}
      </div>

      <div className="flex flex-col gap-4 border-t border-outline-variant/10 pt-5 sm:flex-row sm:items-center sm:justify-between">
        <div className="flex flex-wrap gap-4 text-sm font-semibold text-on-surface">
          <span>{displayPay}</span>
          <span className="font-normal text-on-surface-variant">
            {experienceLevel || "Any level"}
          </span>
        </div>
        <button
          type="button"
          onClick={onApply}
          className="rounded-full bg-primary px-5 py-2.5 text-sm font-bold text-white transition-all hover:bg-primary-container active:scale-95"
        >
          Apply Now
        </button>
      </div>
    </div>
  );
};

export default Jobcard;
