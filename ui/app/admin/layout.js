import './globals.css';

export const metadata = {
  title: 'NotLinkTree Admin',
  description: 'Admin panel for NotLinkTree',
};

export default function RootLayout({ children }) {
  return (
    <html lang="en">
      <body>
        {children}
      </body>
    </html>
  );
} 