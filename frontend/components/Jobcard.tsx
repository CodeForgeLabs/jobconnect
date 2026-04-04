import { Banknote, Clock3 } from "lucide-react";
import { useId } from "react";

interface JobcardProps {
  title: string;
  pay: string;
  type: "fixed" | "hourly";
  rating?: number;

  description: string;
  postTime: string;
  tags: string[];
}

const Jobcard = ({
  title,
  pay,
  type,
  rating = 5,

  description,
  postTime,
  tags,
}: JobcardProps) => {
  const ratingGroupName = useId();
  const normalizedRating = Math.max(0, Math.min(5, Math.round(rating * 2) / 2));

  return (
    <div className="flex flex-col gap-1 rounded-lg bg-white py-4 px-6 ">
      <div className="mb-3 flex items-start justify-between gap-4">
        <h2 className="text-xl font-bold text-gray-900">{title}</h2>

      </div>

      <div className="mb-3 flex flex-wrap items-center gap-3 text-sm text-gray-700">
        {type === "hourly" ? (
          <span className="flex items-center gap-1">
            
            <span className="font-semibold text-jobBlue">
              $ Hourly : {pay}/hr
            </span>
          </span>
        ) : (
          <span className="flex items-center gap-1">
            <Banknote className="h-3.5 w-3.5 text-jobBlue" />
            <span className="font-semibold text-jobBlue">$Fixed: {pay}</span>
          </span>
        )}

        <span className="text-gray-400">|</span>
        {/* <span>{rate}</span> */}
        <p className="flex items-center gap-1 text-sm text-gray-500">
          <Clock3 className="h-3.5 w-3.5" />
          Posted {postTime}
        </p>
      </div>

      <p className="mb-4 text-gray-500">{description}</p>

      {tags.length > 0 ? (
        <div className="mb-4 flex flex-wrap gap-2">
          {tags.map((tag) => (
            <span
              key={tag}
              className="rounded-sm bg-gray-100 px-3 py-1 text-xs font-semibold text-gray-600"
            >
              {tag}
            </span>
          ))}
        </div>
      ) : null}

      <div className="flex justify-between mt-2 border-t border-gray-200 pt-4">
        <div className="flex items-center rating rating-xs rating-half pointer-events-none">
            <input
              type="radio"
              name={ratingGroupName}
              className="rating-hidden"
            />
            <input
              type="radio"
              name={ratingGroupName}
              className="mask mask-star-2 mask-half-1 bg-yellow-400"
              aria-label="0.5 star"
              checked={normalizedRating === 0.5}
            />
            <input
              type="radio"
              name={ratingGroupName}
              className="mask mask-star-2 mask-half-2 bg-yellow-400"
              aria-label="1 star"
              checked={normalizedRating === 1}
            />
            <input
              type="radio"
              name={ratingGroupName}
              className="mask mask-star-2 mask-half-1 bg-yellow-400"
              aria-label="1.5 star"
              checked={normalizedRating === 1.5}
            />
            <input
              type="radio"
              name={ratingGroupName}
              className="mask mask-star-2 mask-half-2 bg-yellow-400"
              aria-label="2 star"
              checked={normalizedRating === 2}
            />
            <input
              type="radio"
              name={ratingGroupName}
              className="mask mask-star-2 mask-half-1 bg-yellow-400"
              aria-label="2.5 star"
              checked={normalizedRating === 2.5}
            />
            <input
              type="radio"
              name={ratingGroupName}
              className="mask mask-star-2 mask-half-2 bg-yellow-400"
              aria-label="3 star"
              checked={normalizedRating === 3}
            />
            <input
              type="radio"
              name={ratingGroupName}
              className="mask mask-star-2 mask-half-1 bg-yellow-400"
              aria-label="3.5 star"
              checked={normalizedRating === 3.5}
            />
            <input
              type="radio"
              name={ratingGroupName}
              className="mask mask-star-2 mask-half-2 bg-yellow-400"
              aria-label="4 star"
              checked={normalizedRating === 4}
            />
            <input
              type="radio"
              name={ratingGroupName}
              className="mask mask-star-2 mask-half-1 bg-yellow-400"
              aria-label="4.5 star"
              checked={normalizedRating === 4.5}
            />
            <input
              type="radio"
              name={ratingGroupName}
              className="mask mask-star-2 mask-half-2 bg-yellow-400"
              aria-label="5 star"
              checked={normalizedRating === 5}
            />
          </div>

         <button className="text-[12px] rounded-lg bg-jobBlue px-5 py-2 text-white hover:bg-blue-600">
        Apply Now
      </button>

      </div>

     
    </div>
  );
};

export default Jobcard;
