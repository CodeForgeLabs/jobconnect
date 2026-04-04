import Image from "next/image";
import hero from "@/assets/macbook.jpg";
import { Handshake, MapPin, Star } from "lucide-react";
const profile = () => {
  const compensationType = "Hourly";
  const compensationAmount = "$50.00/hr";
  const rating = 4.8;
  const reviewCount = 128;
  const clientRating = 4.9;
  const clientReviewCount = 42;
  const profileHighlights = [
    { label: "Projects Delivered", value: "62+" },
    
    { label: "Clients", value: "31" },
  ];
  const featuredProjects = [
    {
      name: "SaaS Analytics Dashboard",
      summary:
        "Built a modern analytics dashboard with role-based access, chart widgets, and export features for enterprise teams.",
      stack: ["Next.js", "TypeScript", "Tailwind", "Supabase"],
      outcome: "+38% user retention",
    },
    {
      name: "E-commerce Storefront Redesign",
      summary:
        "Redesigned checkout and product pages with performance-first UI patterns and reusable component architecture.",
      stack: ["React", "Redux", "Node.js", "Stripe"],
      outcome: "+24% conversion rate",
    },
    {
      name: "Recruitment Platform MVP",
      summary:
        "Shipped end-to-end freelancer matching flows including onboarding, search filtering, and interview scheduling.",
      stack: ["Next.js", "PostgreSQL", "Prisma", "Docker"],
      outcome: "From 0 to 12k users",
    },
  ];
  const testimonials = [
    {
      client: "Meklit G.",
      note: "Excellent communication and very clean code delivery. The project was completed ahead of schedule.",
    },
    {
      client: "Dawit T.",
      note: "Great eye for product details and UX. Turned our rough idea into a polished platform.",
    },
  ];

  return (
    <div className="flex p-8 bg-[#ebedf1] ">
      <div className="flex flex-col gap-7 w-[30%]">
        <div className="flex flex-col items-center  bg-white p-5 rounded-lg shadow-md">
          <div className="avatar">
            <div className="ring-primary ring-offset-base-100 w-20 rounded-full ring-2 ring-offset-2">
              <Image src={hero} alt="Profile picture" />
            </div>
          </div>

          <h1 className="text-[20px] font-bold mt-4">John Doe</h1>
          <p className="text-jobBlue text-sm">Web Developer</p>
          <div className="mt-2 flex items-center gap-2">
            <div className="flex items-center gap-0.5 text-yellow-400">
              {Array.from({ length: 5 }, (_, index) => (
                <Star
                  key={index}
                  className={`h-4 w-4 ${index < Math.floor(rating) ? "fill-current" : "text-gray-300"}`}
                  aria-hidden="true"
                />
              ))}
            </div>
            <span className="text-xs font-semibold text-gray-600">
              {rating.toFixed(1)} ({reviewCount} reviews)
            </span>
          </div>
          <p className="flex items-center gap-1 text-gray-500 text-sm">
            <MapPin className="h-4 w-4" />
            Addis Ababa, Ethiopia
          </p>

          <div className="flex justify-between items-center w-full mt-3 bg-[#ebedf1] px-3 py-2 rounded-lg">
            <p className="text-gray-500 text-xs"> {compensationType} rate</p>
            <p className="text-jobBlue font-semibold text-sm">
              {compensationAmount}
            </p>
          </div>

          <div className="flex justify-between items-center w-full mt-3 bg-[#ebedf1] px-3 py-2 rounded-lg">
            <p className="text-gray-500 text-xs"> Job Success</p>
            <p className="text-jobBlue font-semibold text-sm">95%</p>
          </div>

          <button className="btn btn-primary w-full mt-5">
            <Handshake className="h-4 w-4" />
            Hire Me
          </button>
          <button className="btn bg-[#ebedf1] w-full mt-2">Send Message</button>
        </div>

        <div className="bg-white p-6 rounded-lg shadow-md">
          <p className="font-semibold text-lg">Skills</p>
          <div className="flex flex-wrap gap-2 mt-3">
            <span className="bg-[#d2e1ff] text-jobBlue px-3 py-1 rounded-full text-sm">
              JavaScript
            </span>
            <span className="bg-[#d2e1ff] text-jobBlue px-3 py-1 rounded-full text-sm">
              React
            </span>
            <span className="bg-[#d2e1ff] text-jobBlue px-3 py-1 rounded-full text-sm">
              Node.js
            </span>
            <span className="bg-[#d2e1ff] text-jobBlue px-3 py-1 rounded-full text-sm">
              CSS
            </span>
            <span className="bg-[#d2e1ff] text-jobBlue px-3 py-1 rounded-full text-sm">
              HTML
            </span>
          </div>
        </div>
      </div>

      <div className="flex flex-col gap-6 w-[70%] ml-8">
        <div className="bg-white p-6 rounded-lg shadow-md">
          <p className="font-semibold text-lg">About Me</p>
          <p className="text-gray-500 mt-3 text-sm">
            I am a passionate web developer with over 5 years of experience in
            building responsive and user-friendly websites. I specialize in
            JavaScript, React, and Node.js, and have a strong background in
            front-end development. I am dedicated to delivering high-quality
            work and ensuring client satisfaction. I am a passionate web
            developer with over 5 years of experience in building responsive and
            user-friendly websites. I specialize in JavaScript, React, and
            Node.js, and have a strong background in front-end development. I am
            dedicated to delivering high-quality work and ensuring client
            satisfaction.
          </p>
        </div>

        <div className="bg-white p-6 rounded-lg shadow-md">
          <div className="flex flex-wrap items-center justify-between gap-3">
            <p className="font-semibold text-lg">Portfolio</p>
            <div className="flex items-center gap-2">
              <button className="btn btn-sm bg-[#ebedf1] border-none">
                View CV
              </button>
              <button className="btn btn-sm btn-primary">Download CV</button>
            </div>
          </div>

          <div className="grid grid-cols-1 sm:grid-cols-3 gap-3 mt-4">
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

          <div className="mt-5">
            <p className="text-sm font-semibold text-gray-700">
              Featured Projects
            </p>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4 mt-3">
              {featuredProjects.map((project) => (
                <div
                  key={project.name}
                  className="rounded-xl border border-gray-200 p-4 bg-linear-to-br from-white to-[#f6f9ff]"
                >
                  <div className="flex items-start justify-between gap-3">
                    <p className="font-semibold text-gray-900">
                      {project.name}
                    </p>
                    <span className="text-[11px] px-2 py-1 rounded-full bg-[#d2e1ff] text-jobBlue font-medium">
                      {project.outcome}
                    </span>
                  </div>
                  <p className="text-sm text-gray-500 mt-2">
                    {project.summary}
                  </p>
                  <div className="flex flex-wrap gap-2 mt-3">
                    {project.stack.map((tech) => (
                      <span
                        key={tech}
                        className="text-[11px] px-2 py-1 rounded-md bg-white border border-gray-200 text-gray-600"
                      >
                        {tech}
                      </span>
                    ))}
                  </div>
                </div>
              ))}
            </div>
          </div>

         
        </div>

        <div className="bg-white p-6 rounded-lg shadow-md">
          <p className="font-semibold text-lg">Work History</p>
          <div className="mt-4">
            <p className="font-semibold text-sm">Senior Web Developer</p>
            <p className="text-gray-500 text-xs">
              Tech Company - Jan 2020 to Present
            </p>
            <div className="mt-2 flex items-center gap-2">
              <div className="flex items-center gap-0.5 text-yellow-400">
                {Array.from({ length: 5 }, (_, index) => (
                  <Star
                    key={index}
                    className={`h-4 w-4 ${index < Math.floor(clientRating) ? "fill-current" : "text-gray-300"}`}
                    aria-hidden="true"
                  />
                ))}
              </div>
              <span className="text-xs font-semibold text-gray-600">
                Client rating {clientRating.toFixed(1)} ({clientReviewCount}{" "}
                reviews)
              </span>
            </div>
            <p className="text-gray-500 mt-2 text-sm">
              &quot; He is responsible for leading a team of developers in
              creating and maintaining the company&apos;s main web application.
              He has successfully implemented new features, optimized
              performance, and ensured the application is scalable and secure.
              He has also collaborated closely with designers and product
              managers to deliver a seamless user experience.&quot;
            </p>
          </div>
        </div>
      </div>
    </div>
  );
};

export default profile;
