import './globals.css';

export const metadata = {
  title: 'NotLinkTree Admin',
  description: 'Admin panel for NotLinkTree',
};

export default function AdminLayout({ children }) {
  return (
    <div className="admin-layout">
      {children}
    </div>
  );
} 