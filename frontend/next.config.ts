import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  turbopack: {
    root: __dirname,
  },

  images: {
    domains: ["res.cloudinary.com" , "img.daisyui.com",],
  },
};

export default nextConfig;