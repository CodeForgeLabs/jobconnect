"use client";

import React, { useEffect, useRef, useState } from "react";
import { useRouter } from "next/navigation";
import type { FetchBaseQueryError } from "@reduxjs/toolkit/query";
import {
  useRegisterMutation,
  useSendOtpMutation,
  useVerifyOtpMutation,
} from "@/api/userapi";

export default function CareerArchSignUp() {
  const router = useRouter();
  const [showPassword, setShowPassword] = useState(false);
  const [firstName, setFirstName] = useState("");
  const [lastName, setLastName] = useState("");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [role, setRole] = useState<"CLIENT" | "FREELANCER">("FREELANCER");
  const [companyName, setCompanyName] = useState("");
  const [termsAccepted, setTermsAccepted] = useState(false);
  const [policyDialog, setPolicyDialog] = useState<"terms" | "privacy" | null>(
    null,
  );
  const [step, setStep] = useState<"details" | "otp">("details");
  const [otpDigits, setOtpDigits] = useState<string[]>(Array(4).fill(""));
  const [error, setError] = useState("");
  const [success, setSuccess] = useState("");
  const [otpNotice, setOtpNotice] = useState("");
  const otpInputs = useRef<Array<HTMLInputElement | null>>([]);
  const [register, { isLoading: isRegistering }] = useRegisterMutation();
  const [sendOtp, { isLoading: isSendingOtp }] = useSendOtpMutation();
  const [verifyOtp, { isLoading: isVerifyingOtp }] = useVerifyOtpMutation();
  const passwordRequirements =
    /^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)(?=.*[^A-Za-z\d]).{8,}$/;
  const termsSections = [
    {
      title: "Professional Conduct",
      body: "All users must act honestly and respectfully. Users must provide real information, communicate clearly, and avoid fraud, abuse, discrimination, or illegal activities on the platform.",
    },
    {
      title: "Account Responsibility",
      body: "You are responsible for keeping your account secure. Any activity done through your account is your responsibility, so you must not share your login details with others.",
    },
    {
      title: "Payments and Job Structure",
      body: "JobConnect supports hourly and fixed-price jobs with milestones. For fixed jobs, client funds are held when the job is created. For milestone-based jobs, each milestone has its own payment, which is released to the freelancer only after client approval.",
    },
    {
      title: "Milestone Rules and Refunds",
      body: "When a milestone is submitted and approved, payment is fully released to the freelancer. If a milestone is submitted but not yet approved and the contract is terminated, payment is split equally between client and freelancer. For unsubmitted milestones, the client receives a full refund when the contract is cancelled.",
    },
    {
      title: "Platform Use",
      body: "Users must not bypass the payment system, post fake jobs, upload harmful content, impersonate others, or attempt to break or misuse the platform in any way.",
    },
  ];
  const privacySections = [
    {
      title: "Information We Collect",
      body: "We collect account details, contact information, role information, profile content, project activity, payment-related records, and verification information needed to operate the service.",
    },
    {
      title: "How Information Is Used",
      body: "Information is used to create and secure accounts, match clients with freelancers, support contracts, process payments, prevent fraud, resolve disputes, and improve platform reliability.",
    },
    {
      title: "Data Sharing",
      body: "Relevant profile, contract, and communication details may be shared between clients and freelancers. We do not sell personal information to third parties.",
    },
    {
      title: "Security and Retention",
      body: "We apply reasonable safeguards to protect user information and retain records only as needed for platform operations, legal obligations, payment history, and dispute resolution.",
    },
    {
      title: "User Choices",
      body: "Users may update account information, control what they publish on their profile, and request assistance with account or privacy concerns through platform support.",
    },
  ];

  const getErrorMessage = (err: unknown, fallback: string) => {
    if (typeof err === "object" && err && "data" in err) {
      const errorData = (err as FetchBaseQueryError).data;

      if (typeof errorData === "string") {
        return errorData;
      }

      if (
        typeof errorData === "object" &&
        errorData !== null &&
        "message" in errorData &&
        typeof (errorData as { message?: unknown }).message === "string"
      ) {
        return (errorData as { message: string }).message;
      }
    }

    if (err instanceof Error && err.message) {
      return err.message;
    }

    return fallback;
  };

  useEffect(() => {
    if (step === "otp") {
      otpInputs.current[0]?.focus();
    }
  }, [step]);

  const otpCode = otpDigits.join("");

  const buildRegisterPayload = () => ({
    email,
    first_name: firstName,
    last_name: lastName,
    password,
    company_name: role === "CLIENT" ? companyName : undefined,
    role: role === "CLIENT" ? ("CLIENT" as const) : ("FREELANCER" as const),
  });

  const setOtpDigit = (index: number, value: string) => {
    const nextValue = value.replace(/\D/g, "").slice(0, 1);

    setOtpDigits((current) => {
      const nextDigits = [...current];
      nextDigits[index] = nextValue;
      return nextDigits;
    });

    if (nextValue && index < otpInputs.current.length - 1) {
      otpInputs.current[index + 1]?.focus();
    }
  };

  const handleOtpKeyDown = (
    event: React.KeyboardEvent<HTMLInputElement>,
    index: number,
  ) => {
    if (event.key === "Backspace" && !otpDigits[index] && index > 0) {
      otpInputs.current[index - 1]?.focus();
    }
  };

  const handleOtpPaste = (event: React.ClipboardEvent<HTMLInputElement>) => {
    event.preventDefault();
    const pastedCode = event.clipboardData.getData("text").replace(/\D/g, "");

    if (!pastedCode) {
      return;
    }

    const nextDigits = Array(4)
      .fill("")
      .map((_, index) => pastedCode[index] ?? "");

    setOtpDigits(nextDigits);
    const nextIndex = Math.min(pastedCode.length, otpInputs.current.length - 1);
    otpInputs.current[nextIndex]?.focus();
  };

  const sendVerificationCode = async (emailAddress: string) => {
    try {
      await sendOtp({ email: emailAddress }).unwrap();
      setOtpNotice(`We sent a 4-digit code to ${emailAddress}.`);
    } catch (sendError) {
      console.error("OTP send failed:", sendError);
      setOtpNotice(
        `Account created. We could not send the code automatically. Use resend to try again for ${emailAddress}.`,
      );
    }
  };

  const handleSubmit = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setError("");
    setSuccess("");
    setOtpNotice("");

    if (!firstName.trim() || !lastName.trim() || !email.trim() || !password) {
      setError("Please fill in all required fields before continuing.");
      return;
    }
    if (!passwordRequirements.test(password)) {
      setError(
        "Password must be at least 8 characters long and include an uppercase letter, a lowercase letter, a number, and a special character.",
      );
      return;
    }

    if (role === "CLIENT" && !companyName.trim()) {
      setError("Please add your company name to continue as a client.");
      return;
    }

    if (!termsAccepted) {
      setError("You must accept the terms before creating an account.");
      return;
    }

    setSuccess("Enter the verification code to continue.");
    setStep("otp");
    await sendVerificationCode(email);
  };

  const handleVerifyOtp = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setError("");
    setSuccess("");

    if (otpCode.length !== 4) {
      setError("Enter the full 4-digit verification code.");
      return;
    }

    try {
      await verifyOtp({ email, otp: otpCode }).unwrap();
    } catch (verifyError) {
      console.error("OTP verification failed:", verifyError);
      setError("That code did not match. Try again or resend a new one.");
      return;
    }

    try {
      await register(buildRegisterPayload()).unwrap();
      setSuccess("Email verified and account created. You can now log in.");
      if (role === "CLIENT") {
        router.push("/client/dashboard");
      } else {
        router.push("/freelancer/dashboard");
      }
    } catch (verifyError) {
      console.log("Registration after OTP verification failed:", verifyError);
      setError(
        getErrorMessage(
          verifyError,
          "Your code was verified, but we could not create the account. Please try again.",
        ),
      );
    }
  };

  const handleResendOtp = async () => {
    setError("");
    setOtpNotice("");
    setOtpDigits(Array(4).fill(""));
    await sendVerificationCode(email);
    otpInputs.current[0]?.focus();
  };

  const handleEditDetails = () => {
    setStep("details");
    setError("");
    setSuccess("");
    setOtpNotice("");
  };
  const passwordRules = {
    hasLength: password.length >= 8,
    hasUppercase: /[A-Z]/.test(password),
    hasLowercase: /[a-z]/.test(password),
    hasNumber: /\d/.test(password),
    hasSpecial: /[^A-Za-z\d]/.test(password),
  };
  return (
    <div className="bg-surface text-on-surface flex flex-col min-h-screen">
      {/* Suppression Notice: BottomNavBar and SideNav are suppressed for this transactional Signup flow */}

      <main className="grow pt-24 pb-12 px-6 flex items-center justify-center relative overflow-hidden">
        {/* Subtle Architectural Background Elements */}
        <div className="absolute top-0 right-0 w-125 h-125 bg-surface-container rounded-full blur-[120px] -z-10 opacity-60"></div>
        <div className="absolute bottom-0 left-0 w-100 h-100 bg-tertiary-fixed rounded-full blur-[100px] -z-10 opacity-40"></div>

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
              Join the world&apos;s most curated network of talent and visionary
              clients. Precision in every match.
            </p>
            <div className="grid grid-cols-2 gap-6 mt-4">
              <div className="p-6 bg-surface-container-lowest rounded-lg shadow-sm border border-outline-variant/10">
                <div className="text-3xl font-bold text-primary mb-1">
                  Verified
                </div>
                <div className="text-sm text-on-surface-variant font-medium">
                  Vetted Professionals
                </div>
              </div>
              <div className="p-6 bg-surface-container-lowest rounded-lg shadow-sm border border-outline-variant/10">
                <div className="text-3xl font-bold text-primary mb-1">
                  Fast-Track
                </div>
                <div className="text-sm text-on-surface-variant font-medium">
                  Under 24h Project Turnaround
                </div>
              </div>
            </div>
          </div>

          {/* Right Side: Signup Form */}
          <div className="bg-surface-container-lowest/80 backdrop-blur-md p-8 md:p-12 rounded-lg shadow-xl shadow-on-surface/5 border border-outline-variant/20">
            <div className="mb-10">
              <h2 className="text-3xl font-bold text-on-surface tracking-tight mb-2">
                {step === "details" ? "Create Account" : "Verify your email"}
              </h2>
              <p className="text-on-surface-variant">
                {step === "details"
                  ? "Start your journey with Jobconnect today."
                  : "Enter the code we sent to your inbox to activate the account."}
              </p>
            </div>
            <form
              className="flex flex-col gap-6"
              onSubmit={step === "details" ? handleSubmit : handleVerifyOtp}
            >
              <div className="flex items-center gap-3 rounded-2xl border border-outline-variant/20 bg-surface-container-low px-4 py-3">
                <div
                  className={`flex h-9 w-9 items-center justify-center rounded-full text-sm font-bold ${
                    step === "details"
                      ? "bg-primary text-on-primary"
                      : "bg-primary-container text-on-primary-container"
                  }`}
                >
                  1
                </div>
                <div className="flex-1">
                  <div className="text-sm font-semibold text-on-surface">
                    Account details
                  </div>
                  <div className="text-xs text-on-surface-variant">
                    {step === "details" ? "Current step" : "Completed"}
                  </div>
                </div>
                <div
                  className={`flex h-9 w-9 items-center justify-center rounded-full text-sm font-bold ${
                    step === "otp"
                      ? "bg-primary text-on-primary"
                      : "bg-surface-container-high text-on-surface-variant"
                  }`}
                >
                  2
                </div>
                <div className="flex-1 text-right">
                  <div className="text-sm font-semibold text-on-surface">
                    Verification
                  </div>
                  <div className="text-xs text-on-surface-variant">
                    {step === "otp" ? "Current step" : "Pending"}
                  </div>
                </div>
              </div>

              {step === "details" ? (
                <>
                  <div className="flex flex-col gap-3">
                    <span className="text-sm font-semibold uppercase tracking-wider text-on-surface-variant font-label">
                      I am joining as a...
                    </span>
                    <div className="grid grid-cols-2 gap-4">
                      <label className="relative flex flex-col items-center justify-center p-4 cursor-pointer rounded-lg border-2 border-surface-container-high bg-surface-container-low hover:bg-surface-container-highest transition-all group has-checked:border-primary has-checked:bg-primary-fixed">
                        <input
                          checked={role === "CLIENT"}
                          className="sr-only"
                          onChange={() => setRole("CLIENT")}
                          name="role"
                          type="radio"
                          value="CLIENT"
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
                        <div className="absolute top-2 right-2 opacity-0 group-has-checked:opacity-100 transition-opacity">
                          <svg
                            aria-hidden="true"
                            className="h-5 w-5 text-primary"
                            fill="none"
                            viewBox="0 0 24 24"
                            xmlns="http://www.w3.org/2000/svg"
                          >
                            <circle
                              cx="12"
                              cy="12"
                              fill="currentColor"
                              r="10"
                            />
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
                      <label className="relative flex flex-col items-center justify-center p-4 cursor-pointer rounded-lg border-2 border-surface-container-high bg-surface-container-low hover:bg-surface-container-highest transition-all group has-checked:border-primary has-checked:bg-primary-fixed">
                        <input
                          checked={role === "FREELANCER"}
                          className="sr-only"
                          onChange={() => setRole("FREELANCER")}
                          name="role"
                          type="radio"
                          value="FREELANCER"
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
                        <span className="font-bold text-primary">
                          Freelancer
                        </span>
                        <div className="absolute top-2 right-2 opacity-0 group-has-checked:opacity-100 transition-opacity">
                          <svg
                            aria-hidden="true"
                            className="h-5 w-5 text-primary"
                            fill="none"
                            viewBox="0 0 24 24"
                            xmlns="http://www.w3.org/2000/svg"
                          >
                            <circle
                              cx="12"
                              cy="12"
                              fill="currentColor"
                              r="10"
                            />
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

                  <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                    <div className="flex flex-col gap-2">
                      <label className="text-xs font-bold uppercase tracking-widest text-on-surface-variant font-label ml-1">
                        First Name
                      </label>
                      <input
                        autoComplete="given-name"
                        className="w-full px-6 py-3 rounded-full bg-surface-container-lowest border border-outline-variant/30 focus:ring-2 focus:ring-tertiary-container focus:border-transparent transition-all outline-none"
                        value={firstName}
                        onChange={(event) => setFirstName(event.target.value)}
                        placeholder="Jane"
                        required
                        type="text"
                      />
                    </div>
                    <div className="flex flex-col gap-2">
                      <label className="text-xs font-bold uppercase tracking-widest text-on-surface-variant font-label ml-1">
                        Last Name
                      </label>
                      <input
                        autoComplete="family-name"
                        className="w-full px-6 py-3 rounded-full bg-surface-container-lowest border border-outline-variant/30 focus:ring-2 focus:ring-tertiary-container focus:border-transparent transition-all outline-none"
                        value={lastName}
                        onChange={(event) => setLastName(event.target.value)}
                        placeholder="Doe"
                        required
                        type="text"
                      />
                    </div>
                  </div>

                  <div className="flex flex-col gap-2">
                    {role === "CLIENT" && (
                      <>
                        <label className="text-xs font-bold uppercase tracking-widest text-on-surface-variant font-label ml-1">
                          Company Name
                        </label>
                        <input
                          className="w-full px-6 py-3 rounded-full bg-surface-container-lowest border border-outline-variant/30 focus:ring-2 focus:ring-tertiary-container focus:border-transparent transition-all outline-none"
                          value={companyName}
                          onChange={(event) =>
                            setCompanyName(event.target.value)
                          }
                          placeholder="Acme Corp"
                          required={role === "CLIENT"}
                          type="text"
                        />
                      </>
                    )}

                    <label className="text-xs font-bold uppercase tracking-widest text-on-surface-variant font-label ml-1">
                      Email Address
                    </label>
                    <input
                      autoComplete="email"
                      className="w-full px-6 py-3 rounded-full bg-surface-container-lowest border border-outline-variant/30 focus:ring-2 focus:ring-tertiary-container focus:border-transparent transition-all outline-none"
                      value={email}
                      onChange={(event) => setEmail(event.target.value)}
                      placeholder="jane@example.com"
                      required
                      type="email"
                    />
                  </div>

                  <div className="flex flex-col gap-2">
                    <label className="text-xs font-bold uppercase tracking-widest text-on-surface-variant font-label ml-1">
                      Password
                    </label>
                    <div className="relative">
                      <input
                        autoComplete="new-password"
                        minLength={8}
                        className="w-full px-6 py-3 rounded-full bg-surface-container-lowest border border-outline-variant/30 focus:ring-2 focus:ring-tertiary-container focus:border-transparent transition-all outline-none"
                        value={password}
                        onChange={(event) => setPassword(event.target.value)}
                        placeholder="••••••••"
                        required
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
                  {password.length > 0 && (
                    <div className="grid grid-cols-1 sm:grid-cols-2 gap-2 ml-1 p-3 rounded-2xl bg-surface-container-low border border-outline-variant/10 text-xs transition-all animate-fadeIn">
                      <div
                        className={`flex items-center gap-2 ${passwordRules.hasLength ? "text-success font-medium" : "text-error font-medium"}`}
                      >
                        <span className="text-sm">
                          {passwordRules.hasLength ? "✓" : "✕"}
                        </span>{" "}
                        8+ Characters
                      </div>

                      <div
                        className={`flex items-center gap-2 ${passwordRules.hasUppercase ? "text-success font-medium" : "text-error font-medium"}`}
                      >
                        <span className="text-sm">
                          {passwordRules.hasUppercase ? "✓" : "✕"}
                        </span>{" "}
                        Uppercase letter
                      </div>

                      <div
                        className={`flex items-center gap-2 ${passwordRules.hasLowercase ? "text-success font-medium" : "text-error font-medium"}`}
                      >
                        <span className="text-sm">
                          {passwordRules.hasLowercase ? "✓" : "✕"}
                        </span>{" "}
                        Lowercase letter
                      </div>

                      <div
                        className={`flex items-center gap-2 ${passwordRules.hasNumber ? "text-success font-medium" : "text-error font-medium"}`}
                      >
                        <span className="text-sm">
                          {passwordRules.hasNumber ? "✓" : "✕"}
                        </span>{" "}
                        At least one number
                      </div>

                      <div
                        className={`flex items-center gap-2 ${passwordRules.hasSpecial ? "text-success font-medium" : "text-error font-medium"}`}
                      >
                        <span className="text-sm">
                          {passwordRules.hasSpecial ? "✓" : "✕"}
                        </span>{" "}
                        Special character (!@#$...)
                      </div>
                    </div>
                  )}

                  <div className="flex items-center gap-3 px-1">
                    <input
                      className="w-5 h-5 rounded border-outline-variant text-primary focus:ring-primary"
                      id="terms"
                      checked={termsAccepted}
                      onChange={(event) =>
                        setTermsAccepted(event.target.checked)
                      }
                      aria-describedby="terms-description"
                      type="checkbox"
                    />
                    <p
                      id="terms-description"
                      className="text-sm text-on-surface-variant leading-tight"
                    >
                      I agree to the{" "}
                      <button
                        type="button"
                        className="text-primary font-semibold hover:underline"
                        onClick={() => setPolicyDialog("terms")}
                      >
                        Terms of Service
                      </button>{" "}
                      and{" "}
                      <button
                        type="button"
                        className="text-primary font-semibold hover:underline"
                        onClick={() => setPolicyDialog("privacy")}
                      >
                        Privacy Policy
                      </button>
                      .
                    </p>
                  </div>

                  <button
                    className="w-full bg-linear-to-br from-primary to-primary-container text-on-primary py-4 rounded-full font-bold text-lg shadow-lg hover:shadow-primary/20 hover:scale-[1.01] active:scale-[0.99] transition-all mt-2 disabled:cursor-not-allowed disabled:opacity-70"
                    disabled={isRegistering || isSendingOtp}
                    type="submit"
                  >
                    {isRegistering || isSendingOtp
                      ? "Creating Account..."
                      : "Create Account"}
                  </button>

                  <p className="text-center text-on-surface-variant text-sm mt-2">
                    Already have an account?{" "}
                    <a
                      className="text-primary font-bold hover:underline"
                      href="/login"
                    >
                      Log in
                    </a>
                  </p>
                </>
              ) : (
                <div className="space-y-6 rounded-3xl border border-outline-variant/20 bg-surface-container-low p-6 shadow-sm">
                  <div className="flex items-start justify-between gap-4">
                    <div>
                      <p className="text-xs font-semibold uppercase tracking-[0.25em] text-on-surface-variant">
                        Verification code
                      </p>
                      <h3 className="mt-2 text-xl font-bold text-on-surface">
                        Check your inbox
                      </h3>
                      <p className="mt-2 text-sm text-on-surface-variant leading-relaxed">
                        We sent a 4-digit code to{" "}
                        <span className="font-semibold text-on-surface">
                          {email}
                        </span>
                        . Enter it below to finish creating your account.
                      </p>
                    </div>
                    <button
                      className="rounded-full border border-outline-variant/20 px-4 py-2 text-sm font-semibold text-on-surface-variant transition-colors hover:bg-surface-container-high hover:text-on-surface"
                      onClick={handleEditDetails}
                      type="button"
                    >
                      Edit details
                    </button>
                  </div>

                  <div className="grid grid-cols-4 gap-3">
                    {otpDigits.map((digit, index) => (
                      <input
                        key={index}
                        ref={(element) => {
                          otpInputs.current[index] = element;
                        }}
                        className="h-14 rounded-2xl border border-outline-variant/30 bg-surface-container-lowest text-center text-xl font-bold tracking-[0.3em] text-on-surface outline-none transition-all placeholder:text-on-surface-variant/30 focus:border-primary focus:ring-2 focus:ring-primary/20"
                        inputMode="numeric"
                        maxLength={1}
                        onChange={(event) =>
                          setOtpDigit(index, event.target.value)
                        }
                        onKeyDown={(event) => handleOtpKeyDown(event, index)}
                        onPaste={handleOtpPaste}
                        placeholder="•"
                        type="text"
                        value={digit}
                      />
                    ))}
                  </div>

                  <div className="flex flex-wrap items-center justify-between gap-3 rounded-2xl bg-primary/5 px-4 py-3 text-sm text-on-surface-variant">
                    <span>
                      {otpNotice ||
                        "If the code does not arrive, check spam or resend it."}
                    </span>
                    <button
                      className="font-semibold text-primary hover:underline disabled:cursor-not-allowed disabled:opacity-60"
                      onClick={handleResendOtp}
                      disabled={isSendingOtp}
                      type="button"
                    >
                      {isSendingOtp ? "Resending..." : "Resend code"}
                    </button>
                  </div>

                  <button
                    className="w-full bg-linear-to-br from-primary to-primary-container text-on-primary py-4 rounded-full font-bold text-lg shadow-lg hover:shadow-primary/20 hover:scale-[1.01] active:scale-[0.99] transition-all disabled:cursor-not-allowed disabled:opacity-70"
                    disabled={isVerifyingOtp || otpCode.length !== 4}
                    type="submit"
                  >
                    {isVerifyingOtp ? "Verifying..." : "Verify & continue"}
                  </button>

                  <p className="text-center text-sm text-on-surface-variant">
                    Enter all four digits to unlock the next step.
                  </p>
                </div>
              )}

              {(error || success) && (
                <p
                  className={`text-center text-sm mt-1 ${error ? "text-red-500" : "text-green-600"}`}
                  role="status"
                >
                  {error || success}
                </p>
              )}
            </form>
          </div>
        </div>
      </main>

      {policyDialog ? (
        <div
          className="fixed inset-0 z-50 flex items-center justify-center bg-on-surface/55 px-4 py-8"
          role="dialog"
          aria-modal="true"
          aria-labelledby="policy-dialog-title"
        >
          <div className="max-h-[90vh] w-full max-w-2xl overflow-hidden rounded-2xl border border-outline-variant/20 bg-surface-container-lowest shadow-2xl">
            <div className="flex items-start justify-between gap-4 border-b border-outline-variant/20 px-6 py-5">
              <div>
                <p className="text-xs font-bold uppercase tracking-[0.24em] text-on-surface-variant">
                  Jobconnect policy
                </p>
                <h3
                  id="policy-dialog-title"
                  className="mt-2 text-2xl font-bold text-on-surface"
                >
                  {policyDialog === "terms"
                    ? "Terms of Service"
                    : "Privacy Policy"}
                </h3>
              </div>
              <button
                type="button"
                className="rounded-full border border-outline-variant/30 px-3 py-1 text-sm font-bold text-on-surface-variant transition hover:bg-surface-container-high hover:text-on-surface"
                onClick={() => setPolicyDialog(null)}
                aria-label="Close policy dialog"
              >
                Close
              </button>
            </div>

            <div className="max-h-[65vh] overflow-y-auto px-6 py-5">
              <p className="text-sm leading-7 text-on-surface-variant">
                {policyDialog === "terms"
                  ? "By creating an account, you agree to follow these rules and professional guidelines while using Jobconnect."
                  : "This policy explains how Jobconnect handles account, project, and platform information in a responsible manner."}
              </p>

              <div className="mt-5 space-y-4">
                {(policyDialog === "terms"
                  ? termsSections
                  : privacySections
                ).map((section, index) => (
                  <section
                    key={section.title}
                    className="rounded-2xl border border-outline-variant/20 bg-surface-container-low p-4"
                  >
                    <div className="flex items-start gap-3">
                      <span className="flex h-7 w-7 shrink-0 items-center justify-center rounded-full bg-primary text-xs font-bold text-on-primary">
                        {index + 1}
                      </span>
                      <div>
                        <h4 className="text-sm font-bold text-on-surface">
                          {section.title}
                        </h4>
                        <p className="mt-2 text-sm leading-6 text-on-surface-variant">
                          {section.body}
                        </p>
                      </div>
                    </div>
                  </section>
                ))}
              </div>

              <div className="mt-5 rounded-2xl border border-primary/20 bg-primary/5 p-4 text-sm leading-6 text-on-surface-variant">
                These terms are intended to support fair work, clear
                communication, secure payments, and a trustworthy professional
                marketplace.
              </div>
            </div>

            <div className="flex justify-end gap-3 border-t border-outline-variant/20 px-6 py-4">
              <button
                type="button"
                className="rounded-full border border-outline-variant/30 px-5 py-2 text-sm font-bold text-on-surface-variant transition hover:bg-surface-container-high hover:text-on-surface"
                onClick={() => setPolicyDialog(null)}
              >
                Continue reviewing
              </button>
              <button
                type="button"
                className="rounded-full bg-primary px-5 py-2 text-sm font-bold text-on-primary transition hover:opacity-90"
                onClick={() => {
                  setTermsAccepted(true);
                  setPolicyDialog(null);
                }}
              >
                Accept and close
              </button>
            </div>
          </div>
        </div>
      ) : null}
    </div>
  );
}
