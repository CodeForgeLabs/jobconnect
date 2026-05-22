"use client";

import Image from "next/image";
import Link from "next/link";
import { usePathname } from "next/navigation";
import { useState } from "react";
import { MessageCircle, Search, User, X } from "lucide-react";

import logo from "@/assets/Background.svg";
import { useGetMeQuery } from "@/api/userapi";

const Navbar = () => {
  const [searchQuery, setSearchQuery] = useState("");
  const pathname = usePathname();
  const { data: userData } = useGetMeQuery();

  const isLoggedIn = !!userData;
  const isFreelancer = userData?.role === "FREELANCER";
  const isClient = userData?.role === "CLIENT";

  const handleClear = () => {
    setSearchQuery("");
  };

  return (
    <div className="navbar sticky min-h-7 h-12 py-0 px-6 bg-base-100 shadow-sm">
      <div className="flex flex-1 items-center gap-4">
        <Link href="/" className="flex items-center gap-2">
          <Image src={logo} alt="logo" className="w-8 h-8" />
          <span className="btn btn-ghost text-xl p-0">JobConnect</span>
        </Link>

        {isLoggedIn && (
          <div className="relative flex items-center w-full max-w-md">
            <div className="absolute inset-y-0 left-0 flex items-center pl-3 pointer-events-none">
              <Search className="w-5 h-5 text-gray-400" />
            </div>
            <input
              type="text"
              className="block w-full py-2 pl-10 pr-10 text-sm text-gray-900 border border-gray-300 rounded-lg bg-gray-50 focus:ring-blue-500 focus:border-blue-500 focus:outline-none transition-all duration-200"
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
        )}
      </div>

      <div className="flex items-center gap-1">
        {isLoggedIn ? (
          <>
            <Link
              href="/messages"
              className="btn btn-sm bg-transparent border-none hover:text-black flex items-center gap-1"
            >
              <MessageCircle className="h-4 w-4" />
              Messages
            </Link>

            {isFreelancer && (
              <>
                <Link
                  href="/freelancer/jobsearch"
                  className={`btn btn-sm bg-transparent border-none hover:text-black ${
                    pathname === "/freelancer/jobsearch" ? "text-blue-600" : ""
                  }`}
                >
                  Find Work
                </Link>
                <Link
                  href="/freelancer/myproposals"
                  className={`btn btn-sm bg-transparent border-none hover:text-black ${
                    pathname === "/freelancer/myproposals"
                      ? "text-blue-600"
                      : ""
                  }`}
                >
                  Proposals
                </Link>

                <Link
                  href="/freelancer/mycontracts"
                  className={`btn btn-sm bg-transparent border-none hover:text-black ${
                    pathname === "/freelancer/mycontracts"
                      ? "text-blue-600"
                      : ""
                  }`}
                >
                  My Contracts
                </Link>
              </>
            )}

            {isClient && (
              <>
                <Link
                  href="/find-talent"
                  className={`btn btn-sm bg-transparent border-none hover:text-black ${
                    pathname === "/find-talent" ? "text-blue-600" : ""
                  }`}
                >
                  Find Talent
                </Link>
                <Link
                  href="/postings"
                  className={`btn btn-sm bg-transparent border-none hover:text-black ${
                    pathname === "/postings" ? "text-blue-600" : ""
                  }`}
                >
                  Postings
                </Link>
              </>
            )}

            <div className="dropdown dropdown-end">
              <div
                tabIndex={0}
                role="button"
                className="btn btn-ghost btn-circle avatar"
              >
                <div className="w-8 h-8 rounded-full overflow-hidden bg-gray-100 flex items-center justify-center">
                  {userData?.profile_picture_url ? (
                    <Image
                      src={userData.profile_picture_url}
                      alt="Profile picture"
                      width={32}
                      height={32}
                      className="h-full w-full object-cover"
                    />
                  ) : (
                    <User className="h-4 w-4 text-gray-500" />
                  )}
                </div>
              </div>

              <ul className="menu menu-sm dropdown-content bg-base-100 rounded-box z-1 mt-3 w-52 p-2 shadow">
                <li>
                  <Link href="/freelancer/profile">Profile</Link>
                </li>
                <li>
                  <Link href="/messages">Messages</Link>
                </li>
              </ul>
            </div>
          </>
        ) : (
          <>
            <Link
              href="/login"
              className="btn btn-sm bg-transparent border-none hover:text-black"
            >
              Login
            </Link>
            <Link
              href="/signup"
              className="btn btn-sm bg-jobBlue text-white border-none hover:text-blue-600"
            >
              Sign Up
            </Link>
          </>
        )}
      </div>
    </div>
  );
};

export default Navbar;
