"use client";
import defaultAvatar from "@/assets/avatarsvg.png"; // Adjust alias if needed

import Image from "next/image";
import { useParams } from "next/navigation";
import { MapPin, Star, User } from "lucide-react";

import { useGetUserByIdQuery } from "@/api/userapi";
import {
  useGetUserPortfolioQuery,
  type PortfolioItem,
} from "@/api/portofolioapi";

import PortfolioCard from "@/components/Portofoliocard";
import { useMemo, useState } from "react";
import { useGetUserReviewsQuery } from "@/api/reviewsapi";

const PublicProfile = () => {
  const params = useParams();
  const userId = Number(params.id);

  const { data: userData, isLoading } = useGetUserByIdQuery(userId, {
    skip: !userId,
  });

  const { data: portfolioData } = useGetUserPortfolioQuery(userId, {
    skip: !userId,
  });

  const portfolioItems: PortfolioItem[] = portfolioData?.portfolio ?? [];
  const { data: reviewsData } = useGetUserReviewsQuery(userData?.id ?? 0, {
    skip: !userData?.id,
  });
  const [visibleReviewsCount, setVisibleReviewsCount] = useState(2);

  // 2. Safely sorts reviews so the most recent ones always come first
  const sortedReviews = useMemo(() => {
    if (!reviewsData?.reviews) return [];
    return [...reviewsData.reviews].sort(
      (a, b) =>
        new Date(b.created_at).getTime() - new Date(a.created_at).getTime(),
    );
  }, [reviewsData]);
  // 3. Cuts the array down to the visible window size
  const displayedReviews = sortedReviews.slice(0, visibleReviewsCount);
  const rating = reviewsData?.average_rating ?? 0;
  const reviewCount = reviewsData?.review_count ?? 0;

  const profileHighlights = [
    {
      label: "Projects Delivered",
      value: `${portfolioItems.length}`,
    },
    {
      label: "Clients",
      value: "1+",
    },
  ];

  const skillsArray: string[] = Array.isArray(userData?.skills)
    ? userData.skills.map(String)
    : typeof userData?.skills === "string"
      ? userData.skills.split(",")
      : [];

  if (isLoading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <span className="loading loading-spinner loading-lg" />
      </div>
    );
  }

  return (
    <div className="flex p-8 bg-[#ebedf1] min-h-screen">
      {/* LEFT SIDE */}
      <div className="flex flex-col gap-7 w-[30%]">
        <div className="flex flex-col items-center bg-white p-5 rounded-lg shadow-md">
          <div className="avatar">
            <div className="ring-primary ring-offset-base-100 flex h-24 w-24 items-center justify-center overflow-hidden rounded-full ring-2 ring-offset-2 bg-linear-to-br from-gray-100 to-gray-200 shadow-inner">
              {userData?.profile_picture_url ? (
                <Image
                  src={userData.profile_picture_url}
                  alt="Profile picture"
                  width={96}
                  height={96}
                  className="h-full w-full object-cover"
                />
              ) : (
                <div className="flex h-full w-full items-center justify-center bg-white/60">
                  <User className="h-10 w-10 text-gray-500" />
                </div>
              )}
            </div>
          </div>

          <h1 className="text-[22px] font-bold mt-4 text-center">
            {userData?.first_name} {userData?.last_name}
          </h1>

          <p className="text-jobBlue text-sm text-center">
            {userData?.headline || "Frontend Developer"}
          </p>

          <div className="mt-2 flex items-center gap-2">
            <div className="flex items-center gap-0.5 text-yellow-400">
              {Array.from({ length: 5 }, (_, index) => (
                <Star
                  key={index}
                  className={`h-4 w-4 ${
                    index < Math.floor(rating)
                      ? "fill-current"
                      : "text-gray-300"
                  }`}
                />
              ))}
            </div>

            <span className="text-xs font-semibold text-gray-600">
              {rating.toFixed(1)} ({reviewCount} reviews)
            </span>
          </div>

          <p className="flex items-center gap-1 text-gray-500 text-sm mt-2">
            <MapPin className="h-4 w-4" />
            {userData?.location || "Location not added"}
          </p>

          <div className="flex justify-between items-center w-full mt-4 bg-[#ebedf1] px-3 py-3 rounded-lg">
            <p className="text-gray-500 text-xs">Hourly Rate</p>

            <p className="text-jobBlue font-semibold text-sm">
              {userData?.hourly_rate
                ? `${userData.hourly_rate} Birr/hr`
                : "Not specified"}
            </p>
          </div>

          <div className="flex justify-between items-center w-full mt-3 bg-[#ebedf1] px-3 py-3 rounded-lg">
            <p className="text-gray-500 text-xs">Phone Number</p>

            <p className="text-jobBlue font-semibold text-sm">
              {userData?.phone_number || "Not available"}
            </p>
          </div>
        </div>

        {/* SKILLS */}
        <div className="bg-white p-6 rounded-lg shadow-md">
          <p className="font-semibold text-lg">Skills</p>

          <div className="flex flex-wrap gap-2 mt-4">
            {skillsArray.length > 0 ? (
              skillsArray.map((skill) => (
                <span
                  key={skill}
                  className="bg-[#d2e1ff] text-jobBlue px-3 py-1 rounded-full text-sm"
                >
                  {skill.trim()}
                </span>
              ))
            ) : (
              <p className="text-sm text-gray-500">No skills added yet.</p>
            )}
          </div>
        </div>
      </div>

      {/* RIGHT SIDE */}
      <div className="flex flex-col gap-6 w-[70%] ml-8">
        {/* ABOUT */}
        <div className="bg-white p-6 rounded-lg shadow-md">
          <p className="font-semibold text-lg">About Me</p>

          <p className="text-gray-500 mt-4 text-sm ">
            {userData?.bio || "No bio added yet."}
          </p>
        </div>

        {/* PORTFOLIO */}
        <div className="bg-white p-6 rounded-lg shadow-md">
          <div className="flex items-center justify-between mb-3">
            <p className="font-semibold text-lg">Portfolio & Stats</p>
          </div>

          <div className="grid grid-cols-1 sm:grid-cols-2 gap-3 mb-6">
            {profileHighlights.map((item) => (
              <div
                key={item.label}
                className="rounded-lg border border-gray-200 px-4 py-3 bg-[#fafbfc]"
              >
                <p className="text-xs text-gray-500">{item.label}</p>

                <p className="text-base font-semibold text-gray-900 mt-1">
                  {item.value}
                </p>
              </div>
            ))}
          </div>

          <div className="mt-6 border-t pt-6">
            <p className="text-sm font-semibold text-gray-700 mb-4">
              Featured Projects
            </p>

            {portfolioItems.length === 0 ? (
              <p className="text-sm text-gray-500">
                No portfolio projects added yet.
              </p>
            ) : (
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                {portfolioItems.map((project) => (
                  <PortfolioCard
                    key={project.id}
                    image={project.image_url}
                    title={project.title}
                    date={`${project.start_date} - ${project.end_date}`}
                    description={project.description}
                    tags={project.tech_stack}
                  />
                ))}
              </div>
            )}
          </div>
        </div>

        <div className="bg-white p-6 rounded-lg shadow-md">
          <p className="font-semibold text-lg mb-2">Work History</p>

          {displayedReviews.length > 0 ? (
            <>
              {/* Renders only the active sliced batch */}
              {displayedReviews.map((review) => (
                <ReviewItem key={review.id} review={review} />
              ))}

              {/* Conditionally displays the "Load More" action if there are more reviews left */}
              {sortedReviews.length > visibleReviewsCount && (
                <div className="mt-6 pt-3 border-t border-gray-100 flex justify-center">
                  <button
                    type="button"
                    onClick={() => setVisibleReviewsCount((prev) => prev + 2)}
                    className="text-blue-600 hover:text-blue-700 font-bold text-sm transition-colors cursor-pointer flex items-center gap-1"
                  >
                    Load more reviews
                    <span className="text-xs text-gray-400">
                      ({sortedReviews.length - visibleReviewsCount} remaining)
                    </span>
                  </button>
                </div>
              )}
            </>
          ) : (
            <p className="text-gray-500 mt-4 text-sm">No reviews yet.</p>
          )}
        </div>
      </div>
    </div>
  );
};

export default PublicProfile;

interface ReviewProps {
  review: {
    id: number;
    contract_id: number;
    client_id: number;
    freelancer_id: number;
    rating: number;
    note: string;
    created_at: string;
  };
}

// 2. The sub-component that maps keys correctly and handles isolated states
const ReviewItem = ({ review }: ReviewProps) => {
  const { data: clientData } = useGetUserByIdQuery(review.client_id);
  // TODO: We will plug in manual API fetches here in the next step using review.client_id or review.contract_id

  return (
    <div className="mt-6 border-b border-gray-100 pb-4 last:border-none">
      {/* Fallback title until we pull the contract info */}

      <div className="flex items-center gap-3">
        <Image
          src={clientData?.profile_picture_url || defaultAvatar}
          alt={`${clientData ? `${clientData.first_name} ${clientData.last_name}` : "Client"}'s profile picture`}
          width={32}
          height={32}
          className="h-8 w-8 rounded-full object-cover"
        />

        <div className="flex flex-col">
          <p className=" text-xs ">
            {clientData
              ? `${clientData.first_name} ${clientData.last_name}`
              : "Loading client..."}
          </p>

          <p className="text-gray-500 text-xs ">
            {new Date(review.created_at).toLocaleDateString()}
          </p>
        </div>
      </div>

      <div className="mt-2 flex items-center gap-2">
        <div className="flex items-center gap-0.5 text-yellow-400">
          {Array.from({ length: 5 }, (_, index) => (
            <Star
              key={index}
              className={`h-4 w-4 ${
                index < Math.floor(review.rating)
                  ? "fill-current"
                  : "text-gray-300"
              }`}
              aria-hidden="true"
            />
          ))}
        </div>
        <span className="text-xs font-semibold text-gray-600">
          {review.rating.toFixed(1)}
        </span>
      </div>

      {/* Changed review.comment to review.note to match your JSON */}
      <p className="text-gray-600 mt-2 text-sm italic">
        &quot;{review.note}&quot;
      </p>
    </div>
  );
};
