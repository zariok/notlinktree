/** @type {import('next').NextConfig} */
const nextConfig = {
  output: 'export',
  experimental: {
    inlineCss: true,
  },
  trailingSlash: true,
};

module.exports = nextConfig; 