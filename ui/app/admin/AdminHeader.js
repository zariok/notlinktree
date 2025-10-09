'use client';
import Link from 'next/link';
import { useState } from 'react';
import { usePathname } from 'next/navigation';
import { Bars3Icon, XMarkIcon } from '@heroicons/react/24/outline';

export default function AdminHeader({ onLogout }) {
  const [menuOpen, setMenuOpen] = useState(false);
  const pathname = usePathname();
  const navLinks = [
    { href: '/admin', label: 'Home' },
    { href: '/admin/config', label: 'Config' },
    { href: '/admin/upload-avatar', label: 'Avatar' },
  ];
  const isActive = (href) => pathname === href;

  return (
    <nav className="bg-white border-b border-gray-200">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="flex justify-between h-16 items-center">
          {/* Left: Logo or App Name (optional) */}
          <div className="flex items-center">
            <span className="font-bold text-xl text-blue-700 hidden sm:block">Admin</span>
          </div>
          {/* Center/Left: Nav Links */}
          <div className="hidden sm:flex space-x-6">
            {navLinks.map(link => (
              <Link
                key={link.href}
                href={link.href}
                className={`text-gray-700 hover:text-blue-600 font-medium ${isActive(link.href) ? 'underline text-blue-700' : ''}`}
              >
                {link.label}
              </Link>
            ))}
          </div>
          {/* Right: Logout */}
          <div className="hidden sm:flex items-center">
            <button
              onClick={onLogout}
              className="ml-4 px-4 py-2 bg-red-600 text-white rounded hover:bg-red-700 font-medium"
            >
              Logout
            </button>
          </div>
          {/* Mobile Hamburger */}
          <div className="sm:hidden flex items-center">
            <button onClick={() => setMenuOpen(!menuOpen)} className="p-2 rounded-md text-gray-700 hover:bg-gray-200 focus:outline-none">
              {menuOpen ? <XMarkIcon className="h-6 w-6" /> : <Bars3Icon className="h-6 w-6" />}
            </button>
          </div>
        </div>
      </div>
      {/* Mobile Menu */}
      {menuOpen && (
        <div className="sm:hidden bg-white border-t border-gray-200 px-4 pb-4">
          {navLinks.map(link => (
            <Link
              key={link.href}
              href={link.href}
              className={`block py-2 text-gray-700 hover:text-blue-600 font-medium ${isActive(link.href) ? 'underline text-blue-700' : ''}`}
              onClick={() => setMenuOpen(false)}
            >
              {link.label}
            </Link>
          ))}
          <button
            onClick={() => { setMenuOpen(false); onLogout(); }}
            className="w-full text-left py-2 text-red-600 hover:bg-red-50 font-medium"
          >
            Logout
          </button>
        </div>
      )}
    </nav>
  );
} 