import './globals.css';

export const metadata = {
  title: 'NotLinkTree',
  description: 'Your Social Media Links',
};

export default function RootLayout({ children }) {
  return (
    <html lang="en">
      <body>{children}</body>
    </html>
  );
} 