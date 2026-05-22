"use client";
import Image from "next/image";
import { useState, useRef } from "react";

import { MapPin, Star, User, Edit, Trash2 } from "lucide-react";
import {
  useGetMeQuery,
  useUpdateMeMutation,
  useUploadImageMutation,
} from "@/api/userapi";
import {
  useGetUserPortfolioQuery,
  useCreatePortfolioMutation,
  useUpdatePortfolioMutation,
  useDeletePortfolioMutation,
  type PortfolioRequest,
  type PortfolioItem,
} from "@/api/portofolioapi";
import PortfolioCard from "@/components/Portofoliocard";

const Profile = () => {
  const { data: userData, refetch } = useGetMeQuery();
  const [updateMe, { isLoading: isUpdating }] = useUpdateMeMutation();
  const [uploadImage, { isLoading: isUploading }] = useUploadImageMutation();
  const fileInputRef = useRef<HTMLInputElement | null>(null);
  const portfolioImageInputRef = useRef<HTMLInputElement | null>(null);

  const { data: portfolioData, refetch: refetchPortfolio } =
    useGetUserPortfolioQuery(userData?.id ?? 0, { skip: !userData?.id });

  const portfolioItems: PortfolioItem[] = portfolioData?.portfolio ?? [];
  const [createPortfolio, { isLoading: isCreating }] =
    useCreatePortfolioMutation();
  const [updatePortfolio, { isLoading: isUpdatingPortfolio }] =
    useUpdatePortfolioMutation();
  const [deletePortfolio, { isLoading: isDeletingPortfolio }] =
    useDeletePortfolioMutation();

  const [isEditingPortfolio, setIsEditingPortfolio] = useState(false);
  const [editingPortfolioId, setEditingPortfolioId] = useState<number | null>(
    null,
  );
  const [portfolioForm, setPortfolioForm] = useState<PortfolioRequest>({
    title: "",
    description: "",
    image_url: "",
    start_date: "",
    end_date: "",
    tech_stack: [],
  });
  const [portfolioTechInput, setPortfolioTechInput] = useState("");

  const handleFileChange = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;
    try {
      // Cloudinary upload mutation expects a File (queryFn handles building FormData)
      type UploadRes = { secure_url?: string; url?: string };
      const uploadRes = (await uploadImage(file).unwrap()) as UploadRes;
      const uploadedUrl = uploadRes.secure_url ?? uploadRes.url;
      if (uploadedUrl) {
        await updateMe({ profile_picture_url: uploadedUrl }).unwrap();
        refetch();
      }
    } catch (err) {
      console.error("Upload failed", err);
    } finally {
      if (fileInputRef.current) fileInputRef.current.value = "";
    }
  };

  const handlePortfolioImageChange = async (
    e: React.ChangeEvent<HTMLInputElement>,
  ) => {
    const file = e.target.files?.[0];
    if (!file) return;
    try {
      type UploadRes = { secure_url?: string; url?: string };
      const uploadRes = (await uploadImage(file).unwrap()) as UploadRes;
      const uploadedUrl = uploadRes.secure_url ?? uploadRes.url;
      if (uploadedUrl) {
        setPortfolioForm((prev) => ({
          ...prev,
          image_url: uploadedUrl,
        }));
      }
    } catch (err) {
      console.error("Portfolio image upload failed", err);
    } finally {
      if (portfolioImageInputRef.current)
        portfolioImageInputRef.current.value = "";
    }
  };

  const [isEditingBio, setIsEditingBio] = useState(false);
  const [bio, setBio] = useState<string>("");
  const [isEditingSkills, setIsEditingSkills] = useState(false);
  const [skills, setSkills] = useState<string[]>([]);
  const [newSkill, setNewSkill] = useState("");

  const compensationType = "Hourly";
  const rating = 4.8;
  const reviewCount = 128;
  const clientRating = 4.9;
  const clientReviewCount = 42;
  const profileHighlights = [
    { label: "Projects Delivered", value: "62+" },

    { label: "Clients", value: "31" },
  ];
  const skillsArray: string[] = Array.isArray(userData?.skills)
    ? userData!.skills.map(String)
    : typeof userData?.skills === "string"
      ? userData.skills.split(",")
      : [];
  return (
    <div className="flex p-8 bg-[#ebedf1] ">
      <div className="flex flex-col gap-7 w-[30%]">
        <div className="flex flex-col items-center bg-white p-5 rounded-lg shadow-md">
          <div className="avatar relative">
            <div className="ring-primary ring-offset-base-100 flex h-20 w-20 items-center justify-center overflow-hidden rounded-full ring-2 ring-offset-2 bg-linear-to-br from-gray-100 to-gray-200 shadow-inner">
              {userData?.profile_picture_url ? (
                <Image
                  src={userData.profile_picture_url}
                  alt="Profile picture"
                  width={80}
                  height={80}
                  className="h-full w-full object-cover"
                />
              ) : (
                <div className="flex h-full w-full items-center justify-center bg-white/60 backdrop-blur-sm">
                  <User
                    className="h-10 w-10 text-gray-500"
                    aria-hidden="true"
                  />
                </div>
              )}
            </div>

            <input
              ref={fileInputRef}
              type="file"
              accept="image/*"
              className="hidden"
              onChange={handleFileChange}
            />

            <button
              type="button"
              aria-label="Upload profile picture"
              className="absolute -bottom-2 -right-2 flex h-8 w-8 items-center justify-center rounded-full bg-white text-jobBlue shadow-md"
              onClick={() => fileInputRef.current?.click()}
              disabled={isUploading || isUpdating}
            >
              {isUploading ? (
                <svg
                  className="h-4 w-4 animate-spin text-jobBlue"
                  viewBox="0 0 24 24"
                >
                  <circle
                    className="opacity-25"
                    cx="12"
                    cy="12"
                    r="10"
                    stroke="currentColor"
                    strokeWidth="4"
                    fill="none"
                  />
                  <path
                    className="opacity-75"
                    fill="currentColor"
                    d="M4 12a8 8 0 018-8v4a4 4 0 00-4 4H4z"
                  />
                </svg>
              ) : (
                <span className="text-lg font-bold">+</span>
              )}
            </button>
          </div>

          <h1 className="text-[20px] font-bold mt-4">
            {userData?.first_name} {userData?.last_name}
          </h1>
          <p className="text-jobBlue text-sm">{userData?.headline}</p>
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
            {userData?.location || "Add location"}
          </p>

          <div className="flex justify-between items-center w-full mt-3 bg-[#ebedf1] px-3 py-2 rounded-lg">
            <p className="text-gray-500 text-xs"> {compensationType} rate</p>
            <p className="text-jobBlue font-semibold text-sm">
              {userData?.hourly_rate
                ? `$${userData.hourly_rate}/hr`
                : "Set your rate"}
            </p>
          </div>

          <div className="flex justify-between items-center w-full mt-3 bg-[#ebedf1] px-3 py-2 rounded-lg">
            <p className="text-gray-500 text-xs"> Job Success</p>
            <p className="text-jobBlue font-semibold text-sm">95%</p>
          </div>

          {/* <button className="btn btn-primary w-full mt-5">
            <Handshake className="h-4 w-4" />
            Hire Me
          </button>
          <button className="btn bg-[#ebedf1] w-full mt-2">Send Message</button> */}
        </div>

        <div className="bg-white p-6 rounded-lg shadow-md">
          <div className="flex items-start justify-between">
            <p className="font-semibold text-lg">Skills</p>
            {!isEditingSkills && (
              <button
                className="btn btn-ghost btn-sm"
                onClick={() => {
                  setSkills(skillsArray);
                  setIsEditingSkills(true);
                }}
              >
                Edit
              </button>
            )}
          </div>

          {isEditingSkills ? (
            <div className="mt-3">
              <div className="flex gap-2">
                <input
                  value={newSkill}
                  onChange={(e) => setNewSkill(e.target.value)}
                  onKeyDown={(e) => {
                    if (e.key === "Enter") {
                      e.preventDefault();
                      const s = newSkill.trim();
                      if (s && !skills.includes(s)) {
                        setSkills((prev) => [...prev, s]);
                        setNewSkill("");
                      }
                    }
                  }}
                  className="w-full rounded-md border border-gray-200 p-2 text-sm"
                  placeholder="Add a skill and press Enter or click Add"
                />
                <button
                  className="btn btn-sm btn-primary"
                  onClick={() => {
                    const s = newSkill.trim();
                    if (s && !skills.includes(s)) {
                      setSkills((prev) => [...prev, s]);
                      setNewSkill("");
                    }
                  }}
                >
                  Add
                </button>
              </div>

              <div className="flex flex-wrap gap-2 mt-3">
                {skills.map((s) => (
                  <span
                    key={s}
                    className="inline-flex items-center gap-2 rounded-full bg-[#d2e1ff] px-3 py-1 text-sm text-jobBlue"
                  >
                    <span>{s}</span>
                    <button
                      aria-label={`Remove ${s}`}
                      onClick={() =>
                        setSkills((prev) => prev.filter((x) => x !== s))
                      }
                      className="ml-1 inline-flex h-5 w-5 items-center justify-center rounded-full bg-white text-xs text-gray-500"
                    >
                      ×
                    </button>
                  </span>
                ))}
              </div>

              <div className="mt-3 flex gap-2">
                <button
                  className="btn btn-primary btn-sm"
                  onClick={async () => {
                    try {
                      await updateMe({ skills }).unwrap();
                      setIsEditingSkills(false);
                      setNewSkill("");
                      refetch();
                    } catch (err) {
                      console.error(err);
                    }
                  }}
                  disabled={isUpdating}
                >
                  Save
                </button>
                <button
                  className="btn btn-sm"
                  onClick={() => {
                    setIsEditingSkills(false);
                    setSkills([]);
                    setNewSkill("");
                  }}
                >
                  Cancel
                </button>
              </div>
            </div>
          ) : (
            <div className="flex flex-wrap gap-2 mt-3">
              {skillsArray.map((skill) => (
                <span
                  key={String(skill)}
                  className="bg-[#d2e1ff] text-jobBlue px-3 py-1 rounded-full text-sm"
                >
                  {String(skill).trim()}
                </span>
              ))}
            </div>
          )}
        </div>
      </div>

      <div className="flex flex-col gap-6 w-[70%] ml-8">
        <div className="bg-white p-6 rounded-lg shadow-md">
          <div className="flex items-start justify-between">
            <p className="font-semibold text-lg">About Me</p>
            {!isEditingBio && (
              <button
                className="btn btn-ghost btn-sm flex items-center gap-2"
                onClick={() => setIsEditingBio(true)}
                aria-label="Edit bio"
              >
                <Edit className="h-4 w-4" />
                Edit
              </button>
            )}
          </div>

          {isEditingBio ? (
            <div className="mt-3">
              <textarea
                value={bio}
                onChange={(e) => setBio(e.target.value)}
                className="w-full rounded-md border border-gray-200 p-3 text-sm resize-none"
                rows={5}
                placeholder="Write a short bio to introduce yourself to clients"
              />

              <div className="mt-3 flex gap-2">
                <button
                  className="btn btn-primary btn-sm"
                  onClick={async () => {
                    try {
                      await updateMe({ bio }).unwrap();
                      setIsEditingBio(false);
                      refetch();
                    } catch (err) {
                      console.error(err);
                    }
                  }}
                  disabled={isUpdating}
                >
                  Save
                </button>
                <button
                  className="btn btn-sm"
                  onClick={() => {
                    setBio(userData?.bio || "");
                    setIsEditingBio(false);
                  }}
                >
                  Cancel
                </button>
              </div>
            </div>
          ) : (
            <p className="text-gray-500 mt-3 text-sm">
              {userData?.bio ||
                "Add a bio to tell clients about your experience, skills, and what makes you unique as a freelancer."}
            </p>
          )}
        </div>

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
            <div className="flex items-center justify-between mb-3">
              <p className="text-sm font-semibold text-gray-700">
                Featured Projects
              </p>
              {!isEditingPortfolio && (
                <button
                  className="btn btn-ghost btn-sm"
                  onClick={() => {
                    setEditingPortfolioId(null);
                    setPortfolioForm({
                      title: "",
                      description: "",
                      image_url: "",
                      start_date: "",
                      end_date: "",
                      tech_stack: [],
                    });
                    setIsEditingPortfolio(true);
                  }}
                >
                  + Add Project
                </button>
              )}
            </div>

            {isEditingPortfolio && (
              <div className="mb-6 p-4 border border-gray-200 rounded-lg bg-gray-50">
                <h4 className="font-semibold mb-3">
                  {editingPortfolioId ? "Edit Project" : "Add New Project"}
                </h4>

                <div className="space-y-3">
                  <input
                    type="text"
                    placeholder="Project Title"
                    value={portfolioForm.title}
                    onChange={(e) =>
                      setPortfolioForm((prev) => ({
                        ...prev,
                        title: e.target.value,
                      }))
                    }
                    className="w-full rounded-md border border-gray-200 p-2 text-sm"
                  />

                  <textarea
                    placeholder="Project Description"
                    value={portfolioForm.description}
                    onChange={(e) =>
                      setPortfolioForm((prev) => ({
                        ...prev,
                        description: e.target.value,
                      }))
                    }
                    className="w-full rounded-md border border-gray-200 p-2 text-sm resize-none"
                    rows={3}
                  />

                  <div className="border-2 border-dashed border-gray-300 rounded-lg p-4 flex flex-col items-center justify-center">
                    {portfolioForm.image_url ? (
                      <div className="w-full flex flex-col items-center gap-2">
                        <Image
                          src={portfolioForm.image_url}
                          alt="Portfolio preview"
                          width={200}
                          height={150}
                          className="rounded object-cover max-h-40"
                        />
                        <button
                          type="button"
                          onClick={() =>
                            portfolioImageInputRef.current?.click()
                          }
                          disabled={isUploading}
                          className="btn btn-sm btn-outline"
                        >
                          {isUploading ? "Uploading..." : "Change Image"}
                        </button>
                      </div>
                    ) : (
                      <div className="flex flex-col items-center gap-2">
                        <p className="text-sm text-gray-500">
                          Upload project image
                        </p>
                        <button
                          type="button"
                          onClick={() =>
                            portfolioImageInputRef.current?.click()
                          }
                          disabled={isUploading}
                          className="btn btn-sm btn-primary"
                        >
                          {isUploading ? "Uploading..." : "Choose Image"}
                        </button>
                      </div>
                    )}
                    <input
                      ref={portfolioImageInputRef}
                      type="file"
                      accept="image/*"
                      className="hidden"
                      onChange={handlePortfolioImageChange}
                    />
                  </div>

                  <div className="grid grid-cols-2 gap-2">
                    <input
                      type="date"
                      value={portfolioForm.start_date}
                      onChange={(e) =>
                        setPortfolioForm((prev) => ({
                          ...prev,
                          start_date: e.target.value,
                        }))
                      }
                      className="rounded-md border border-gray-200 p-2 text-sm"
                    />
                    <input
                      type="date"
                      value={portfolioForm.end_date}
                      onChange={(e) =>
                        setPortfolioForm((prev) => ({
                          ...prev,
                          end_date: e.target.value,
                        }))
                      }
                      className="rounded-md border border-gray-200 p-2 text-sm"
                    />
                  </div>

                  <div className="flex gap-2">
                    <input
                      type="text"
                      placeholder="Add technology (press Enter)"
                      value={portfolioTechInput}
                      onChange={(e) => setPortfolioTechInput(e.target.value)}
                      onKeyDown={(e) => {
                        if (e.key === "Enter") {
                          e.preventDefault();
                          const tech = portfolioTechInput.trim();
                          if (
                            tech &&
                            !portfolioForm.tech_stack.includes(tech)
                          ) {
                            setPortfolioForm((prev) => ({
                              ...prev,
                              tech_stack: [...prev.tech_stack, tech],
                            }));
                            setPortfolioTechInput("");
                          }
                        }
                      }}
                      className="flex-1 rounded-md border border-gray-200 p-2 text-sm"
                    />
                    <button
                      className="btn btn-sm btn-primary"
                      onClick={() => {
                        const tech = portfolioTechInput.trim();
                        if (tech && !portfolioForm.tech_stack.includes(tech)) {
                          setPortfolioForm((prev) => ({
                            ...prev,
                            tech_stack: [...prev.tech_stack, tech],
                          }));
                          setPortfolioTechInput("");
                        }
                      }}
                    >
                      Add
                    </button>
                  </div>

                  <div className="flex flex-wrap gap-2">
                    {portfolioForm.tech_stack.map((tech) => (
                      <span
                        key={tech}
                        className="inline-flex items-center gap-1 rounded-full bg-blue-100 px-2 py-1 text-xs text-jobBlue"
                      >
                        {tech}
                        <button
                          type="button"
                          onClick={() =>
                            setPortfolioForm((prev) => ({
                              ...prev,
                              tech_stack: prev.tech_stack.filter(
                                (t) => t !== tech,
                              ),
                            }))
                          }
                          className="text-xs font-bold"
                        >
                          ×
                        </button>
                      </span>
                    ))}
                  </div>

                  <div className="flex gap-2 pt-2">
                    <button
                      className="btn btn-primary btn-sm flex-1"
                      onClick={async () => {
                        try {
                          if (editingPortfolioId) {
                            await updatePortfolio({
                              id: editingPortfolioId,
                              data: portfolioForm,
                            }).unwrap();
                          } else {
                            await createPortfolio(portfolioForm).unwrap();
                          }
                          setIsEditingPortfolio(false);
                          setEditingPortfolioId(null);
                          setPortfolioTechInput("");
                          setPortfolioForm({
                            title: "",
                            description: "",
                            image_url: "",
                            start_date: "",
                            end_date: "",
                            tech_stack: [],
                          });
                          // refresh portfolio list
                          refetchPortfolio?.();
                        } catch (err) {
                          let errorMsg = "Unknown error";
                          if (err instanceof Error) {
                            errorMsg = err.message;
                          } else if (err && typeof err === "object") {
                            const errObj = err as Record<string, unknown>;
                            if (
                              errObj.data &&
                              typeof errObj.data === "object"
                            ) {
                              const dataObj = errObj.data as Record<
                                string,
                                unknown
                              >;
                              errorMsg = String(
                                dataObj.message || dataObj.error || err,
                              );
                            } else if (errObj.message) {
                              errorMsg = String(errObj.message);
                            }
                          }
                          console.error(
                            "Portfolio operation failed:",
                            errorMsg,
                          );
                        }
                      }}
                      disabled={isCreating || isUpdatingPortfolio}
                    >
                      {editingPortfolioId ? "Update" : "Create"}
                    </button>
                    <button
                      className="btn btn-sm flex-1"
                      onClick={() => {
                        setIsEditingPortfolio(false);
                        setEditingPortfolioId(null);
                        setPortfolioTechInput("");
                      }}
                    >
                      Cancel
                    </button>
                  </div>
                </div>
              </div>
            )}

            <div className="grid grid-cols-1 md:grid-cols-2 gap-4 mt-3">
              {portfolioItems.map((project) => (
                <div key={project.id} className="relative group">
                  <PortfolioCard
                    image={project.image_url}
                    title={project.title}
                    date={`${project.start_date} - ${project.end_date}`}
                    description={project.description}
                    tags={project.tech_stack}
                  />
                  {!isEditingPortfolio && (
                    <div className="absolute top-2 right-2 flex gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
                      <button
                        className="btn btn-sm btn-primary"
                        onClick={() => {
                          setEditingPortfolioId(project.id);
                          setPortfolioForm({
                            title: project.title,
                            description: project.description,
                            image_url: project.image_url,
                            start_date: project.start_date,
                            end_date: project.end_date,
                            tech_stack: project.tech_stack,
                          });
                          setIsEditingPortfolio(true);
                        }}
                      >
                        Edit
                      </button>
                      <button
                        className="btn btn-sm btn-error"
                        onClick={async () => {
                          try {
                            await deletePortfolio(project.id).unwrap();
                            refetchPortfolio?.();
                          } catch (err) {
                            let errorMsg = "Unknown error";
                            if (err instanceof Error) {
                              errorMsg = err.message;
                            } else if (err && typeof err === "object") {
                              const errObj = err as Record<string, unknown>;
                              if (
                                errObj.data &&
                                typeof errObj.data === "object"
                              ) {
                                const dataObj = errObj.data as Record<
                                  string,
                                  unknown
                                >;
                                errorMsg = String(
                                  dataObj.message || dataObj.error || err,
                                );
                              } else if (errObj.message) {
                                errorMsg = String(errObj.message);
                              }
                            }
                            console.error("Delete failed:", errorMsg);
                          }
                        }}
                        disabled={isDeletingPortfolio}
                      >
                        <Trash2 className="h-4 w-4" />
                      </button>
                    </div>
                  )}
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

export default Profile;
