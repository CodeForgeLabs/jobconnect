"use client";
import Image from "next/image";
import hero from "@/assets/hero.svg";
import postjob from "@/assets/postjob.svg";
import security from "@/assets/security.svg";
import person from "@/assets/person.svg";
import macbook from "@/assets/macbook.jpg";
import { Search, X } from "lucide-react";
import { useState } from "react";
import { useSelector } from "react-redux";
import { selectIsLoggedIn } from "@/features/login/loginSlice";
import Hireinfocard from "@/components/Hireinfocard";
import Browsejobcard from "@/components/Browsejobcard";
import Talentcard from "@/components/Talentcard";

export default function Home() {
  const [searchQuery, setSearchQuery] = useState("");
  const isLoggedIn = useSelector(selectIsLoggedIn);

  const handleClear = () => {
    setSearchQuery("");
  };
  return (
    <div>
      <div className="flex justify-between py-12 pl-10 pr-12 items-center bg-[#ebedf1]  w-full">
        <div>
          <span className="text-xs font-bold text-gray-500 uppercase tracking-wide bg-[#cde6fd] py-1 px-2 rounded-full mb-3 inline-block ">
            Future of Hiring in Ethiopia
          </span>
          <h1 className="flex flex-col text-4xl font-bold mb-4 ">
            Connect with Ethiopia’s
            <span className="text-jobBlue">top freelancers and clients</span>
          </h1>
          <p className="text-[16px] mb-6 text-gray-500">
            JobConnect is your go-to platform for finding freelance work or
            hiring talented professionals. Whether you&apos;re a freelancer
            looking for exciting projects or a client seeking skilled experts,
            JobConnect has you covered.
          </p>

          <div className="flex">
            <div className="relative flex items-center w-full max-w-md">
              <div className="absolute inset-y-0 left-0 flex items-center pl-3 pointer-events-none">
                <Search className="w-5 h-5 text-gray-400" />
              </div>
              <input
                type="text"
                className="block w-full py-2.25 pl-10 pr-10 text-sm text-gray-900 border border-gray-300 rounded-l-lg bg-gray-50 focus:ring-blue-500 focus:border-blue-500 focus:outline-none transition-all duration-200"
                placeholder="Search for jobs, skills, or talent..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
              />
              {searchQuery && (
                <button
                  onClick={handleClear}
                  className="absolute inset-y-0 right-0 flex items-center pr-3 text-gray-400 hover:text-gray-600"
                >
                  <X className="w-4 h-4" />
                </button>
              )}
            </div>

            <a
              href="/login"
              className="btn btn-primary rounded-r-lg rounedd-l-none "
            >
              Get Started
            </a>
          </div>
        </div>
        <Image src={hero} alt="hero" className="w-115 h-auto" />
      </div>

      <div className="flex flex-col py-12 px-10 mt-10">
        <h1 className="text-3xl font-bold  mb-4">How JobConnect Works</h1>
        <p className=" text-gray-500 mb-12">
          Post a job and hire a freelancer in Ethiopia in minutes. It&apos;s
          that simple.
        </p>

        <div className="flex  gap-6 max-tablet:flex-col ">
          <Hireinfocard
            icon={postjob}
            title="Post a Job"
            description="Describe your project and the skills you need. We'll help you find the perfect freelancer for the job."
          />
          <Hireinfocard
            icon={person}
            title="Choose a Freelancer"
            description="Review profiles and portfolios to find the right person for your project."
          />

          <Hireinfocard
            icon={security}
            title="Pay Safely"
            description="Use our secure payment system to pay for your project with confidence."
          />
        </div>
      </div>

      <div className="mt-10 flex flex-col py-12 pl-10 pr-12  bg-[#ebedf1]  w-full">
        <h1 className="text-3xl font-bold  mb-4">Browse by Category</h1>
        <p className=" text-gray-500 mb-12">
          Looking for work? Browse our categories to find the perfect job for
          your skills.
        </p>

        <div className="flex gap-6">
          <Browsejobcard
            image={macbook}
            title="Web Development"
            jobs="1200+ Jobs"
          />

          <Browsejobcard
            image={macbook}
            title="Design & Creative"
            jobs="1200+ Jobs"
          />

          <Browsejobcard
            image={macbook}
            title="Writing & Translation"
            jobs="1200+ Jobs"
          />

          <Browsejobcard
            image={macbook}
            title="Marketing & Sales"
            jobs="1200+ Jobs"
          />
        </div>


          



      </div>

      <div className="mt-10 py-12">
            <h1 className="text-3xl font-bold mb-4 text-center">Meet our top rated-talents</h1>
             <p className="text-center text-gray-500"> The highest rated freelancers on our platform </p>
             <div className="flex gap-6 px-8 py-8">
              <Talentcard
                profilePicture={macbook}
                name="Alex per"
                title="Web Developer"
                rating={3.5}
                description="Experienced web developer with a passion for creating beautiful and functional websites."
                specializationTags={["Web Development", "React", "Node.js"]}
                hourlyRate={50}
                profileLink="https://www.example.com/talent-profile"
              />


              <Talentcard
                profilePicture={macbook}
                name="Alex per"
                title="Web Developer"
                rating={3.5}
                description="Experienced web developer with a passion for creating beautiful and functional websites."
                specializationTags={["Web Development", "React", "Node.js"]}
                hourlyRate={50}
                profileLink="https://www.example.com/talent-profile"
              />


              <Talentcard
                profilePicture={macbook}
                name="Alex per"
                title="Web Developer"
                rating={3.5}
                description="Experienced web developer with a passion for creating beautiful and functional websites."
                specializationTags={["Web Development", "React", "Node.js"]}
                hourlyRate={50}
                profileLink="https://www.example.com/talent-profile"
              />


             </div>
             </div>
              


              <div className="bg-[#ebedf1] py-16 px-10 mt-10">
                <div className="flex flex-col justify-center bg-jobBlue py-12 h-80  rounded-2xl">
                <h1 className="text-4xl font-[Inter] font-bold mb-4 text-center text-white">Ready to get your work done?</h1>
                <p className="text-[18px] text-center text-gray-200 mb-6">Join thousands of companies using jobconnect to build their dream teams!</p>
                <div className="flex justify-center">
                  <a
                    href="/login"
                    className="btn btn-primary rounded-sm text-jobBlue bg-white p-6"
                  >
                    Hire a Freelancer
                  </a>

                  <a
                    href="/login"
                    className="btn btn-secondary rounded-lg text-white ml-4 bg-jobBlue border border-white p-6"
                  >
                    Find Work
                  </a>
                </div>

              </div>

            
          </div>



    </div>
  );
}
