import Image, { type StaticImageData } from "next/image";
import { useId } from "react";

export interface TalentcardProps {
  profilePicture: string | StaticImageData;
  name: string;
  title : string;

  rating: number;
  description: string;
  specializationTags: string[];
  hourlyRate: number;
  profileLink?: string; // Optional link to the talent's profile
}

const Talentcard = ({
  profilePicture,
  name,
    title,
  rating,
  description,
  specializationTags,
  hourlyRate,
  profileLink,
}: TalentcardProps) => {
  const ratingGroupName = useId();
  const normalizedRating = Math.max(
    0.5,
    Math.min(5, Math.round(rating * 2) / 2),
  );

  return (
    <div className="card bg-base-100 shadow-xl p-6 gap-4 border border-gray-200 rounded-lg w-full">
      <div className="flex justify-between items-center gap-3">
        <div className="flex gap-4 justify-center items-center">
               <div className="avatar">
                <div className="ring-primary ring-offset-base-100 w-12 rounded-full ring-1 ">
                    <Image src={profilePicture} alt="Talent profile picture"  />
                </div>
                </div>

            <div>
                <h2 className="text font-semibold">{name}</h2>
                <p className="text-xs text-gray-400">{title}</p>
            </div>

        </div>
        
        
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
              readOnly
            />
            <input
              type="radio"
              name={ratingGroupName}
              className="mask mask-star-2 mask-half-2 bg-yellow-400"
              aria-label="1 star"
              checked={normalizedRating === 1}
              readOnly
            />
            <input
              type="radio"
              name={ratingGroupName}
              className="mask mask-star-2 mask-half-1 bg-yellow-400"
              aria-label="1.5 star"
              checked={normalizedRating === 1.5}
              readOnly
            />
            <input
              type="radio"
              name={ratingGroupName}
              className="mask mask-star-2 mask-half-2 bg-yellow-400"
              aria-label="2 star"
              checked={normalizedRating === 2}
              readOnly
            />
            <input
              type="radio"
              name={ratingGroupName}
              className="mask mask-star-2 mask-half-1 bg-yellow-400"
              aria-label="2.5 star"
              checked={normalizedRating === 2.5}
              readOnly
            />
            <input
              type="radio"
              name={ratingGroupName}
              className="mask mask-star-2 mask-half-2 bg-yellow-400"
              aria-label="3 star"
              checked={normalizedRating === 3}
              readOnly
            />
            <input
              type="radio"
              name={ratingGroupName}
              className="mask mask-star-2 mask-half-1 bg-yellow-400"
              aria-label="3.5 star"
              checked={normalizedRating === 3.5}
              readOnly
            />
            <input
              type="radio"
              name={ratingGroupName}
              className="mask mask-star-2 mask-half-2 bg-yellow-400"
              aria-label="4 star"
              checked={normalizedRating === 4}
              readOnly
            />
            <input
              type="radio"
              name={ratingGroupName}
              className="mask mask-star-2 mask-half-1 bg-yellow-400"
              aria-label="4.5 star"
              checked={normalizedRating === 4.5}
              readOnly
            />
            <input
              type="radio"
              name={ratingGroupName}
              className="mask mask-star-2 mask-half-2 bg-yellow-400"
              aria-label="5 star"
              checked={normalizedRating === 5}
              readOnly
            />
          </div>
          {/* <p className="text-sm text-yellow-400 font-semibold flex items-center gap-1">
            {rating.toFixed(1)} 
          </p> */}
        
      </div>

      <p className="text-sm text-gray-600">{description}</p>

      <div className="flex flex-wrap mb-2 ">
        {specializationTags.map((tag) => (
          <span key={tag} className="tag px-2 py-0.5 rounded-lg mr-2  shadow-sm text-xs font-medium bg-[#ebedf1] ">
            {tag}
          </span>
        ))}
      </div>
    
    <div className="flex justify-between items-center border-t border-gray-200 pt-4">

         <p className="font-semibold text-jobBlue">{hourlyRate} <span className="text-[14px]">ETB/hr</span>
         </p>
      {profileLink && (
        <a
          href={profileLink}
          className="text-sm font-bold border-none bg-none btn-sm mt-2 text-jobBlue"
          target="_blank"
          rel="noopener noreferrer"
        >
          View Profile
        </a>
      )}
        
    </div>
     
    </div>
  );
};

export default Talentcard;
