"use client";
import React, { useState } from "react";
import Typingeffect from "@/components/Typingeffect";
import { useLoginMutation } from "@/api/userapi";
import { logIn } from "@/features/login/loginSlice";
import { setIsClient, setIsFreelancer } from "@/features/login/loginSlice";
import { useDispatch } from "react-redux";
import { useSelector } from "react-redux";
import { selectIsLoggedIn } from "@/features/login/loginSlice";
// import Signup from './Signup'

import { useRouter } from "next/navigation";

const Login = () => {
  const dispatch = useDispatch();
  const router = useRouter();
  const [email, setEmail] = useState<string>("nb@gmail.com");
  const [password, setPassword] = useState<string>("nb");
  const [err, setErr] = useState("");
  const [login, { isLoading }] = useLoginMutation();
  const isLoggedIn = useSelector(selectIsLoggedIn);

  const handleLogin = async () => {
    try {
      const res = (await login({ email, password }).unwrap()) as {
        role?: string;
      };
      if (res) {
        logIn();
        dispatch(logIn());

        // ✅ set role (IMPORTANT FIX)

        if (res.role === "client") {
          dispatch(setIsClient(true));

          dispatch(setIsFreelancer(false));
        } else if (res.role === "freelancer") {
          dispatch(setIsFreelancer(true));
          console.log("User is a freelancer");
          console.log(res.role);
          console.log(isLoggedIn);
          dispatch(setIsClient(false));
        }

        if (res.role === "FREELANCER") {
          router.push("/freelancer/dashboard");
        } else {
          router.push("/client/dashboard");
        }
      }
    } catch (error) {
      console.log(error);
      setErr("Invalid email or password");
    }
  };

  return (
    <div className="hero bg-base-200 min-h-screen">
      <div className="hero-content flex-col lg:flex-row-reverse">
        <div className="text-center lg:text-left">
          <Typingeffect />

          <p className="py-6">
            Find the right talent. Or become the talent. Log in to connect with
            clients, showcase your skills, and build your freelance career. Or
            hire talented professionals to bring your ideas to life.
          </p>
        </div>
        <div className="card bg-base-100 w-full max-w-sm shrink-0 shadow-2xl">
          <form
            className="card-body"
            onSubmit={(e) => {
              e.preventDefault(); // Prevent default form submission
              handleLogin(); // Call your login function
            }}
          >
            <div className="card bg-base-100 w-full max-w-sm shrink-0 shadow-2xl">
              <div className="card-body">
                <fieldset className="fieldset">
                  <label className="label">Email</label>
                  <input
                    type="email"
                    className="input"
                    placeholder="Email"
                    value={email}
                    onChange={(e) => setEmail(e.target.value)}
                  />

                  <label className="label">Password</label>
                  <input
                    type="password"
                    className="input"
                    placeholder="Password"
                    value={password}
                    onChange={(e) => setPassword(e.target.value)}
                  />

                  <div className="text-red-500 mt-2">{err}</div>
                  <button className="btn btn-neutral mt-4" disabled={isLoading}>
                    {isLoading ? "Logging in..." : "Login"}
                  </button>
                </fieldset>
              </div>
            </div>

            <p>
              Don&#39;t have an account?{" "}
              <span
                onClick={() => {
                  router.push("/signup");
                }}
                className="text-secondary underline"
              >
                Sign up
              </span>
            </p>
          </form>
        </div>
      </div>
    </div>
  );
};

export default Login;
