import type { NextConfig } from "next"

const nextConfig: NextConfig = {
  allowedDevOrigins: [
    "radiant-laughter-production-b8ac.up.railway.app",
    "*.up.railway.app",
    "localhost:3000",
    "localhost:3001"
  ]
}

export default nextConfig
