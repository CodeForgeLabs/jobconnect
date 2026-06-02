"use client";

import Image from "next/image";

export interface ProjectCardProps {
  image: string;
  title: string;
  date: string;
  description: string;
  tags: string[];
}

const ProjectCard = ({
  image,
  title,
  date,
  description,
  tags,
}: ProjectCardProps) => {
  return (
    <div className="group flex flex-col">
      {/* Image */}
      <div className="relative aspect-16/10 rounded-2xl overflow-hidden shadow-md border border-gray-200">
        <Image
          src={image}
          alt={title}
          fill
          sizes="(max-width: 768px) 100vw, 50vw"
          className="object-cover group-hover:scale-105 transition-transform duration-500"
        />

        <div className="absolute inset-0 bg-black/5 opacity-0 group-hover:opacity-100 transition-opacity"></div>
      </div>

      {/* Content */}
      <div className="mt-5 flex-1">
        <div className="flex justify-between items-start gap-2 mb-2">
          <h4 className="text-lg font-semibold group-hover:text-blue-600 transition-colors leading-tight">
            {title}
          </h4>

          <span className="text-[10px] font-semibold uppercase tracking-widest text-gray-500 bg-gray-100 px-2 py-0.5 rounded">
            {date}
          </span>
        </div>

        <p className="text-sm text-gray-600 mb-4 line-clamp-2">
          {description}
        </p>

        <div className="flex flex-wrap gap-1.5">
          {tags.map((tag) => (
            <span
              key={tag}
              className="px-2 py-0.5 bg-blue-100 text-blue-700 text-[10px] font-semibold rounded"
            >
              {tag}
            </span>
          ))}
        </div>
      </div>
    </div>
  );
};

export default ProjectCard;