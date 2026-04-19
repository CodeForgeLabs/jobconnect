"use client";

import { BarChart3, Brain, ChevronDown, Clock3, Search, X } from "lucide-react";
import { KeyboardEvent, useState } from "react";
import Jobcard from "@/components/Jobcard";

const Jobsearch = () => {
  const [skills, setSkills] = useState<string[]>([]);
  const [skillInput, setSkillInput] = useState("");

  const addSkill = (value: string) => {
    const trimmedValue = value.trim();

    if (!trimmedValue) return;

    const isDuplicate = skills.some(
      (skill) => skill.toLowerCase() === trimmedValue.toLowerCase(),
    );

    if (!isDuplicate) {
      setSkills((prevSkills) => [...prevSkills, trimmedValue]);
    }

    setSkillInput("");
  };

  const removeSkill = (skillToRemove: string) => {
    setSkills((prevSkills) =>
      prevSkills.filter((skill) => skill !== skillToRemove),
    );
  };

  const handleSkillKeyDown = (event: KeyboardEvent<HTMLInputElement>) => {
    if (event.key === "Enter" || event.key === ",") {
      event.preventDefault();
      addSkill(skillInput);
    }
  };

  return (
    <div className="flex p-14 bg-[#ebedf1] gap-8">
      <div className=" flex flex-col gap-4 w-[30%] h-fit bg-white p-5 rounded-lg shadow-md   ">
        <div className="flex justify-between items-center">
          <p> Advanced Filters</p>
          <span className="text-jobBlue text-xs"> clear all</span>
        </div>

        <div className="flex flex-col">
          <p className="flex  items-center gap-2 font-medium text-sm mb-2">
            <Clock3 className="h-4 w-4" />
            job type
          </p>
          <span>
            <input
              type="radio"
              name="job-type"
              defaultChecked
              className="radio radio-sm"
            />
            <label className="ml-2 text-[12px] text-gray-500">Fixed rate</label>
          </span>
          <span>
            <input
              type="radio"
              name="job-type"
              className="radio radio-sm radio-primary"
            />
            <label className="ml-2 text-[12px] text-gray-500">
              Hourly rate
            </label>
          </span>
        </div>

        <div className="flex flex-col">
          <p className="flex  items-center gap-2 font-medium text-sm mb-2">
            <BarChart3 className="h-4 w-4" />
            Experience
          </p>
          <span>
            <input
              type="radio"
              name="experience"
              defaultChecked
              className="radio radio-sm"
            />
            <label className="ml-2 text-[12px] text-gray-500">
              Entry Level
            </label>
          </span>
          <span>
            <input
              type="radio"
              name="experience"
              className="radio radio-sm radio-primary"
            />
            <label className="ml-2 text-[12px] text-gray-500">
              Intermediate{" "}
            </label>
          </span>
          <span>
            <input
              type="radio"
              name="experience"
              className="radio radio-sm radio-primary"
            />
            <label className="ml-2 text-[12px] text-gray-500">Expert</label>
          </span>
        </div>

        <div>
          <p className="flex  items-center gap-2 font-medium text-sm mb-2">
            <BarChart3 className="h-4 w-4" />
            Budget
          </p>
          <div className="flex items-center gap-2">
            <input
              type="number"
              placeholder="Min"
              className="input input-sm input-bordered w-full max-w-xs"
            />
            <input
              type="number"
              placeholder="Max"
              className="input input-sm input-bordered w-full max-w-xs"
            />
          </div>
        </div>

        <div>
          <p className="flex  items-center gap-2 font-medium text-sm mb-2">
            <Brain className="h-4 w-4" />
            Skills
          </p>

          <div className="mb-2 flex flex-wrap gap-2">
            {skills.map((skill) => (
              <span
                key={skill}
                className="inline-flex items-center gap-1 rounded-full bg-blue-100 px-2 py-1 text-xs text-blue-700"
              >
                {skill}
                <button
                  type="button"
                  onClick={() => removeSkill(skill)}
                  className="rounded-full p-0.5 hover:bg-blue-200"
                  aria-label={`Remove ${skill}`}
                >
                  <X className="h-3 w-3" />
                </button>
              </span>
            ))}
          </div>

          <input
            type="text"
            value={skillInput}
            onChange={(event) => setSkillInput(event.target.value)}
            onKeyDown={handleSkillKeyDown}
            onBlur={() => addSkill(skillInput)}
            placeholder="Type a skill and press Enter"
            className="input input-sm input-bordered w-full"
          />
        </div>
      </div>



      <div className=" w-full ">
        <h1 className="text-3xl font-bold">
          Find Work
        </h1>

       

              <div className=" flex justify-between">
                    <p className="text-gray-500 text-[16px]"> found this many jobs tailored to your skill</p>
                    <div className="dropdown dropdown-bottom dropdown-end">
                    <div tabIndex={0} role="button" className="btn m-1 text-sm font-medium gap-2">Sort by: Newest <ChevronDown className="h-4 w-4" /></div>
                    <ul tabIndex={-1} className="dropdown-content menu bg-base-100 rounded-box z-1 w-52 p-2 shadow-sm">
                      <li><a>Newest first</a></li>
                      <li><a>Most relevant</a></li>
                    </ul>
                      </div>
              </div>

              
               <div className="mt-5 flex w-full items-center rounded-lg border border-gray-200 bg-white pl-4 p-0.5  shadow-sm">
          <Search className="h-4 w-4 text-gray-400" />
          <input
            type="text"
            placeholder="Search jobs, skills, or keywords"
            className="ml-3 w-full bg-transparent text-sm text-gray-700 outline-none py-3"
          />
          <button
            type="button"
            className="h-full ml-4 text-sm font-semibold text-white bg-jobBlue hover:opacity-80 py-3 px-6 rounded-lg"
          >
            Search
          </button>
        </div>
        
        <div className="flex flex-col gap-4 py-9 ">
          <Jobcard
            title="Build a responsive website"
            pay="$500"
            type="fixed"
            
            description="Looking for a skilled web developer to create a responsive website for my business. The website should be modern, user-friendly, and optimized for both desktop and mobile devices."
            postTime="2 hours ago"
            tags={["HTML", "CSS", "JavaScript", "React"]}
          />
            <Jobcard
            title="Design a logo for my startup"
            pay="$200"
            type="hourly"
            
            description="Seeking a creative graphic designer to design a logo for my new startup. The logo should be unique, memorable, and reflect the essence of my brand."
            postTime="5 hours ago"
            tags={["Graphic Design", "Logo Design", "Adobe Illustrator"]}
          />

          <Jobcard
            title="Develop a mobile app"
            pay="$1000"
            type="fixed"
            
            description="I need a talented mobile app developer to create a cross-platform mobile application for my business. The app should have a sleek design, smooth performance, and be compatible with both iOS and Android devices."
            postTime="1 day ago"
            tags={["Mobile App Development", "React Native", "iOS", "Android"]}
          />
        </div>

      </div>
    </div>
  );
};

export default Jobsearch;
