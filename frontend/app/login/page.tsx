"use client"
import React, { useState } from 'react'
import Typingeffect from '@/components/Typingeffect'
// import Signup from './Signup'


import { useRouter } from 'next/navigation'


const Login = () => {
  
  const router = useRouter()
  const [username , setUsername] = useState<string>('')
  const [password , setPassword] = useState<string>('')
  const [err , setErr] = useState('')
  const handleLogin = async () => {
  const res = await fetch("http://localhost:5000/auth/login", {
    method: "POST",
    credentials: "include", // important for cookies
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({
      username,
      password,
    }),
  })

  if (res.ok) {
    router.push("/")
  } else {
    const data = await res.json()
    setErr(data.message)
  }
}
  return (
    <div className="hero bg-base-200 min-h-screen">
  <div className="hero-content flex-col lg:flex-row-reverse">
    <div className="text-center lg:text-left">
 <Typingeffect />

 
      <p className="py-6">
        Find the right talent.
Or become the talent.
      Log in to connect with clients, showcase your skills, and build your freelance career.
Or hire talented professionals to bring your ideas to life.
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
          <input type="email" className="input" placeholder="Email" />
          <label className="label">Password</label>
          <input type="password" className="input" placeholder="Password" />
          <div><a className="link link-hover">Forgot password?</a></div>
          <button className="btn btn-neutral mt-4">Login</button>
        </fieldset>
      </div>
    </div>

  <p>
    Don&#39;t have an account?{' '}
    <span
      onClick={() => {
   
        const modal = document.getElementById('my_modal_3') as HTMLDialogElement;
        if (modal) {
          modal.showModal();
        }
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
  )
}

export default Login