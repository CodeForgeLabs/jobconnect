"use client"
import React, { useState } from 'react'
import { Search, X } from 'lucide-react';
import { useSelector } from 'react-redux';
import { selectIsLoggedIn } from '../features/login/loginSlice';
import { usePathname } from 'next/navigation';
import Image from 'next/image';
import logo from '@/assets/Background.svg'

const Navbar = () => {

  const [searchQuery, setSearchQuery] = useState('');
  const isLoggedIn = useSelector(selectIsLoggedIn);
  const pathname = usePathname();

  const handleClear = () => {
    setSearchQuery('');
  };


    return (
        <div className="navbar min-h-7 h-12 py-0 px-6 bg-base-100 shadow-sm">

  <div className="flex flex-1 items-center">
    <Image src={logo} alt="logo" className="w-8 h-8 mx-3" />
    <a className="btn btn-ghost text-xl p-0">JobConnect</a>

    {isLoggedIn &&(<div className="relative flex items-center w-full max-w-md">
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
    </div>) }
    






  </div>

      <div className='flex gap-4'>
         <a
           href="/find-work"
           className={`btn btn-sm bg-transparent border-none ${
             pathname === "/"
               ? " border-blue-500 text-blue-600"
               : ""
           } hover:text-black `}
         >
           Find Work
         </a>

         <a
           href="/find-talent"
           className={`btn btn-sm bg-transparent border-none ${
             pathname === "/find-talent"
               ? "border-b-2 border-blue-500 text-blue-600"
               : ""
           } hover:text-black `}
         >
           Find Talent
         </a>

         <a 
            href= "/login"
           className="btn btn-sm bg-transparent border-none hover:text-black"
         >
           Login
         </a>

         <a href="/signup" 
         className="btn btn-sm bg-jobBlue text-white border-none hover:text-blue-600">
           Sign Up
         </a>
      </div>
     
  
    

  <div className="flex gap-2">
    {isLoggedIn && (
    <div className="dropdown dropdown-end">
      <div tabIndex={0} role="button" className="btn btn-ghost btn-circle avatar">
        <div className="w-10 rounded-full">
          <img
            alt="prof"
            src="https://img.daisyui.com/images/stock/photo-1534528741775-53994a69daeb.webp" />
        </div>
      </div>



      


      
      <ul
        tabIndex={-1}
        className="menu menu-sm dropdown-content bg-base-100 rounded-box z-1 mt-3 w-52 p-2 shadow">
        <li>
          <a className="justify-between">
            Profile
            <span className="badge">New</span>
          </a>
        </li>
        <li><a>Settings</a></li>
        <li><a>Logout</a></li>
      </ul>?



    </div>)}
  </div>
</div>
    )
}

export default Navbar




