"use client";

import React, { useState } from "react";

export default function CareerArchSignUp() {
  const [showPassword, setShowPassword] = useState(false);

  return (
    <div className="bg-surface text-on-surface flex flex-col min-h-screen">
      {/* Suppression Notice: BottomNavBar and SideNav are suppressed for this transactional Signup flow */}

      <main className="flex-grow pt-24 pb-12 px-6 flex items-center justify-center relative overflow-hidden">
        {/* Subtle Architectural Background Elements */}
        <div className="absolute top-0 right-0 w-[500px] h-[500px] bg-surface-container rounded-full blur-[120px] -z-10 opacity-60"></div>
        <div className="absolute bottom-0 left-0 w-[400px] h-[400px] bg-tertiary-fixed rounded-full blur-[100px] -z-10 opacity-40"></div>

        <div className="max-w-6xl w-full grid grid-cols-1 lg:grid-cols-2 gap-12 ">
          {/* Left Side: Editorial Content */}
          <div className="hidden lg:flex flex-col gap-8 pr-12">
            <div className="inline-flex items-center gap-2 bg-tertiary-fixed text-on-tertiary-fixed-variant px-4 py-1 rounded-full w-fit">
              <svg
                aria-hidden="true"
                className="h-4 w-4"
                fill="none"
                viewBox="0 0 24 24"
                xmlns="http://www.w3.org/2000/svg"
              >
                <path
                  d="M12 3l1.65 3.35L17 8l-3.35 1.65L12 13l-1.65-3.35L7 8l3.35-1.65L12 3z"
                  fill="currentColor"
                />
                <path
                  d="M18.5 13.5l.95 1.95 1.95.95-1.95.95-.95 1.95-.95-1.95-1.95-.95 1.95-.95.95-1.95z"
                  fill="currentColor"
                />
                <path
                  d="M5.5 14.5l.8 1.65 1.65.8-1.65.8-.8 1.65-.8-1.65-1.65-.8 1.65-.8.8-1.65z"
                  fill="currentColor"
                />
              </svg>
              <span className="text-xs font-bold uppercase tracking-widest">
                Architectural Precision
              </span>
            </div>
            <h1 className="text-6xl font-extrabold tracking-tighter text-primary leading-[1.1]">
              Build your future with{" "}
              <span className="text-on-tertiary-container">intent.</span>
            </h1>
            <p className="text-xl text-on-surface-variant leading-relaxed max-w-md">
              Join the world&apos;s most curated network of architectural talent
              and visionary clients. Precision in every match.
            </p>
            <div className="grid grid-cols-2 gap-6 mt-4">
              <div className="p-6 bg-surface-container-lowest rounded-lg shadow-sm border border-outline-variant/10">
                <div className="text-3xl font-bold text-primary mb-1">12k+</div>
                <div className="text-sm text-on-surface-variant font-medium">
                  Vetted Architects
                </div>
              </div>
              <div className="p-6 bg-surface-container-lowest rounded-lg shadow-sm border border-outline-variant/10">
                <div className="text-3xl font-bold text-primary mb-1">98%</div>
                <div className="text-sm text-on-surface-variant font-medium">
                  Match Success Rate
                </div>
              </div>
            </div>
          </div>

          {/* Right Side: Signup Form */}
          <div className="bg-surface-container-lowest/80 backdrop-blur-md p-8 md:p-12 rounded-lg shadow-xl shadow-on-surface/5 border border-outline-variant/20">
            <div className="mb-10">
              <h2 className="text-3xl font-bold text-on-surface tracking-tight mb-2">
                Create Account
              </h2>
              <p className="text-on-surface-variant">
                Start your journey with Jobconnect today.
              </p>
            </div>
            <form className="flex flex-col gap-6">
              {/* Role Selection */}
              <div className="flex flex-col gap-3">
                <span className="text-sm font-semibold uppercase tracking-wider text-on-surface-variant font-label">
                  I am joining as a...
                </span>
                <div className="grid grid-cols-2 gap-4">
                  <label className="relative flex flex-col items-center justify-center p-4 cursor-pointer rounded-lg border-2 border-surface-container-high bg-surface-container-low hover:bg-surface-container-highest transition-all group has-[:checked]:border-primary has-[:checked]:bg-primary-fixed">
                    <input
                      defaultChecked
                      className="sr-only"
                      name="role"
                      type="radio"
                      value="client"
                    />
                    <svg
                      aria-hidden="true"
                      className="mb-2 h-8 w-8 text-primary"
                      fill="none"
                      viewBox="0 0 24 24"
                      xmlns="http://www.w3.org/2000/svg"
                    >
                      <path
                        d="M9 6V4.5A1.5 1.5 0 0 1 10.5 3h3A1.5 1.5 0 0 1 15 4.5V6"
                        stroke="currentColor"
                        strokeLinecap="round"
                        strokeLinejoin="round"
                        strokeWidth="1.8"
                      />
                      <rect
                        height="12"
                        rx="2"
                        stroke="currentColor"
                        strokeWidth="1.8"
                        width="18"
                        x="3"
                        y="6"
                      />
                      <path
                        d="M3 11h18"
                        stroke="currentColor"
                        strokeLinecap="round"
                        strokeWidth="1.8"
                      />
                    </svg>
                    <span className="font-bold text-primary">Client</span>
                    <div className="absolute top-2 right-2 opacity-0 group-has-[:checked]:opacity-100 transition-opacity">
                      <svg
                        aria-hidden="true"
                        className="h-5 w-5 text-primary"
                        fill="none"
                        viewBox="0 0 24 24"
                        xmlns="http://www.w3.org/2000/svg"
                      >
                        <circle cx="12" cy="12" fill="currentColor" r="10" />
                        <path
                          d="M8 12.5l2.5 2.5L16 9.5"
                          stroke="white"
                          strokeLinecap="round"
                          strokeLinejoin="round"
                          strokeWidth="2"
                        />
                      </svg>
                    </div>
                  </label>
                  <label className="relative flex flex-col items-center justify-center p-4 cursor-pointer rounded-lg border-2 border-surface-container-high bg-surface-container-low hover:bg-surface-container-highest transition-all group has-[:checked]:border-primary has-[:checked]:bg-primary-fixed">
                    <input
                      className="sr-only"
                      name="role"
                      type="radio"
                      value="freelancer"
                    />
                    <svg
                      aria-hidden="true"
                      className="mb-2 h-8 w-8 text-primary"
                      fill="none"
                      viewBox="0 0 24 24"
                      xmlns="http://www.w3.org/2000/svg"
                    >
                      <path
                        d="M4 19h16"
                        stroke="currentColor"
                        strokeLinecap="round"
                        strokeWidth="1.8"
                      />
                      <path
                        d="M6 19v-5l6-4 6 4v5"
                        stroke="currentColor"
                        strokeLinecap="round"
                        strokeLinejoin="round"
                        strokeWidth="1.8"
                      />
                      <path
                        d="M10 19v-3h4v3"
                        stroke="currentColor"
                        strokeLinecap="round"
                        strokeLinejoin="round"
                        strokeWidth="1.8"
                      />
                      <path
                        d="M12 5l1.25 2.5L16 8l-2 1.95.5 2.8L12 11.5l-2.5 1.25.5-2.8L8 8l2.75-.5L12 5z"
                        stroke="currentColor"
                        strokeLinecap="round"
                        strokeLinejoin="round"
                        strokeWidth="1.6"
                      />
                    </svg>
                    <span className="font-bold text-primary">Freelancer</span>
                    <div className="absolute top-2 right-2 opacity-0 group-has-[:checked]:opacity-100 transition-opacity">
                      <svg
                        aria-hidden="true"
                        className="h-5 w-5 text-primary"
                        fill="none"
                        viewBox="0 0 24 24"
                        xmlns="http://www.w3.org/2000/svg"
                      >
                        <circle cx="12" cy="12" fill="currentColor" r="10" />
                        <path
                          d="M8 12.5l2.5 2.5L16 9.5"
                          stroke="white"
                          strokeLinecap="round"
                          strokeLinejoin="round"
                          strokeWidth="2"
                        />
                      </svg>
                    </div>
                  </label>
                </div>
              </div>

              {/* Names */}
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div className="flex flex-col gap-2">
                  <label className="text-xs font-bold uppercase tracking-widest text-on-surface-variant font-label ml-1">
                    First Name
                  </label>
                  <input
                    className="w-full px-6 py-3 rounded-full bg-surface-container-lowest border border-outline-variant/30 focus:ring-2 focus:ring-tertiary-container focus:border-transparent transition-all outline-none"
                    placeholder="Jane"
                    type="text"
                  />
                </div>
                <div className="flex flex-col gap-2">
                  <label className="text-xs font-bold uppercase tracking-widest text-on-surface-variant font-label ml-1">
                    Last Name
                  </label>
                  <input
                    className="w-full px-6 py-3 rounded-full bg-surface-container-lowest border border-outline-variant/30 focus:ring-2 focus:ring-tertiary-container focus:border-transparent transition-all outline-none"
                    placeholder="Doe"
                    type="text"
                  />
                </div>
              </div>

              {/* Email */}
              <div className="flex flex-col gap-2">
                <label className="text-xs font-bold uppercase tracking-widest text-on-surface-variant font-label ml-1">
                  Email Address
                </label>
                <input
                  className="w-full px-6 py-3 rounded-full bg-surface-container-lowest border border-outline-variant/30 focus:ring-2 focus:ring-tertiary-container focus:border-transparent transition-all outline-none"
                  placeholder="jane@example.com"
                  type="email"
                />
              </div>

              {/* Password */}
              <div className="flex flex-col gap-2">
                <label className="text-xs font-bold uppercase tracking-widest text-on-surface-variant font-label ml-1">
                  Password
                </label>
                <div className="relative">
                  <input
                    className="w-full px-6 py-3 rounded-full bg-surface-container-lowest border border-outline-variant/30 focus:ring-2 focus:ring-tertiary-container focus:border-transparent transition-all outline-none"
                    placeholder="••••••••"
                    type={showPassword ? "text" : "password"}
                  />
                  <button
                    className="absolute right-5 top-1/2 -translate-y-1/2 text-on-surface-variant hover:text-primary transition-colors"
                    onClick={() => setShowPassword((prev) => !prev)}
                    type="button"
                    aria-label={
                      showPassword ? "Hide password" : "Show password"
                    }
                  >
                    {showPassword ? (
                      <svg
                        aria-hidden="true"
                        className="h-5 w-5"
                        fill="none"
                        viewBox="0 0 24 24"
                        xmlns="http://www.w3.org/2000/svg"
                      >
                        <path
                          d="M3 3l18 18"
                          stroke="currentColor"
                          strokeLinecap="round"
                          strokeLinejoin="round"
                          strokeWidth="2"
                        />
                        <path
                          d="M10.58 10.58A2 2 0 0 0 13.42 13.42"
                          stroke="currentColor"
                          strokeLinecap="round"
                          strokeLinejoin="round"
                          strokeWidth="2"
                        />
                        <path
                          d="M9.88 5.08A10.94 10.94 0 0 1 12 4.9c5 0 8.27 3.11 9.5 6.1a11.45 11.45 0 0 1-2.18 3.27"
                          stroke="currentColor"
                          strokeLinecap="round"
                          strokeLinejoin="round"
                          strokeWidth="2"
                        />
                        <path
                          d="M6.12 6.12C4.31 7.22 3.1 8.9 2.5 11 3.73 13.99 7 17.1 12 17.1c1.5 0 2.85-.28 4.06-.78"
                          stroke="currentColor"
                          strokeLinecap="round"
                          strokeLinejoin="round"
                          strokeWidth="2"
                        />
                      </svg>
                    ) : (
                      <svg
                        aria-hidden="true"
                        className="h-5 w-5"
                        fill="none"
                        viewBox="0 0 24 24"
                        xmlns="http://www.w3.org/2000/svg"
                      >
                        <path
                          d="M2.5 12C3.73 9.01 7 5.9 12 5.9s8.27 3.11 9.5 6.1c-1.23 2.99-4.5 6.1-9.5 6.1S3.73 14.99 2.5 12z"
                          stroke="currentColor"
                          strokeLinecap="round"
                          strokeLinejoin="round"
                          strokeWidth="2"
                        />
                        <circle
                          cx="12"
                          cy="12"
                          r="2.7"
                          stroke="currentColor"
                          strokeWidth="2"
                        />
                      </svg>
                    )}
                  </button>
                </div>
              </div>

              {/* Terms */}
              <div className="flex items-center gap-3 px-1">
                <input
                  className="w-5 h-5 rounded border-outline-variant text-primary focus:ring-primary"
                  id="terms"
                  type="checkbox"
                />
                <label
                  className="text-sm text-on-surface-variant leading-tight"
                  htmlFor="terms"
                >
                  I agree to the{" "}
                  <a
                    className="text-primary font-semibold hover:underline"
                    href="#"
                  >
                    Terms of Service
                  </a>{" "}
                  and{" "}
                  <a
                    className="text-primary font-semibold hover:underline"
                    href="#"
                  >
                    Privacy Policy
                  </a>
                  .
                </label>
              </div>

              {/* CTA */}
              <button
                className="w-full bg-gradient-to-br from-primary to-primary-container text-on-primary py-4 rounded-full font-bold text-lg shadow-lg hover:shadow-primary/20 hover:scale-[1.01] active:scale-[0.99] transition-all mt-2"
                type="submit"
              >
                Create Account
              </button>

              {/* Login Link */}
              <p className="text-center text-on-surface-variant text-sm mt-2">
                Already have an account?{" "}
                <a className="text-primary font-bold hover:underline" href="#">
                  Log in
                </a>
              </p>
            </form>
          </div>
        </div>
      </main>
    </div>
  );
}
