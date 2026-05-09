import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  turbopack: {
    root: __dirname,
  },
  async rewrites() {
    const gateway = process.env.GATEWAY_ORIGIN || "http://localhost:8080";
    return [
      {
        source: "/api/v1/:path*",
        destination: `${gateway}/api/v1/:path*`,
      },
      {
        source: "/healthz",
        destination: `${gateway}/healthz`,
      },
    ];
  },
};

export default nextConfig;
