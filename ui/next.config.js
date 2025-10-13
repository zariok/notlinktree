/** @type {import('next').NextConfig} */
const nextConfig = {
  output: 'export',
  experimental: {
    inlineCss: true,
  },
  trailingSlash: true,
  // Disable image optimization for static export
  images: {
    unoptimized: true,
  },
};

module.exports = nextConfig; 