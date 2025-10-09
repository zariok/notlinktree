'use client';
import React, { useState, useEffect } from 'react';
import { PlusIcon, TrashIcon, PencilIcon } from '@heroicons/react/24/outline';
import { Menu } from '@headlessui/react';
import { Bars3Icon, XMarkIcon } from '@heroicons/react/24/outline';
import Link from 'next/link';
import AdminHeader from './AdminHeader';
import AdminFooter from './AdminFooter';

const LINK_TYPES = [
  { value: 'HomePage', label: 'Home Page' },
  { value: 'Twitter', label: 'Twitter' },
  { value: 'Instagram', label: 'Instagram' },
  { value: 'Facebook', label: 'Facebook' },
  { value: 'LinkedIn', label: 'LinkedIn' },
  { value: 'YouTube', label: 'YouTube' },
  { value: 'TikTok', label: 'TikTok' },
  { value: 'Twitch', label: 'Twitch' },
  { value: 'Discord', label: 'Discord' },
  { value: 'Spotify', label: 'Spotify' },
  { value: 'SoundCloud', label: 'SoundCloud' },
  { value: 'Patreon', label: 'Patreon' },
  { value: 'OnlyFans', label: 'OnlyFans' },
  { value: 'BlueSky', label: 'BlueSky' },
  { value: 'Mastodon', label: 'Mastodon' },
  { value: 'GitHub', label: 'GitHub' },
  { value: 'Medium', label: 'Medium' },
  { value: 'Substack', label: 'Substack' },
  { value: 'Newsletter', label: 'Newsletter' },
  { value: 'Podcast', label: 'Podcast' },
  { value: 'Portfolio', label: 'Portfolio' },
  { value: 'Shop', label: 'Shop' },
  { value: 'Other', label: 'Other' }
];

const LINK_TYPE_URL_EXAMPLES = {
  Twitter: 'https://x.com/username',
  Instagram: 'https://instagram.com/username',
  Facebook: 'https://facebook.com/username',
  LinkedIn: 'https://linkedin.com/in/username',
  YouTube: 'https://youtube.com/@username',
  TikTok: 'https://tiktok.com/@username',
  Twitch: 'https://twitch.tv/username',
  Discord: 'https://discord.gg/invitecode',
  Spotify: 'https://open.spotify.com/user/username',
  SoundCloud: 'https://soundcloud.com/username',
  Patreon: 'https://patreon.com/username',
  OnlyFans: 'https://onlyfans.com/username',
  BlueSky: 'https://bsky.app/profile/username',
  Mastodon: 'https://mastodon.social/@username',
  GitHub: 'https://github.com/username',
  Medium: 'https://medium.com/@username',
  Substack: 'https://username.substack.com',
  Newsletter: 'https://newsletter.com/username',
  Podcast: 'https://podcasts.com/username',
  Portfolio: 'https://yourdomain.com',
  Shop: 'https://shop.com/yourshop',
  HomePage: 'https://yourdomain.com',
  Other: '',
};

const LINK_TYPE_TITLE_EXAMPLES = {
  Twitter: 'Twitter',
  Instagram: 'Instagram',
  Facebook: 'Facebook',
  LinkedIn: 'LinkedIn',
  YouTube: 'YouTube',
  TikTok: 'TikTok',
  Twitch: 'Twitch',
  Discord: 'Discord',
  Spotify: 'Spotify',
  SoundCloud: 'SoundCloud',
  Patreon: 'Patreon',
  OnlyFans: 'OnlyFans',
  BlueSky: 'BlueSky',
  Mastodon: 'Mastodon',
  GitHub: 'GitHub',
  Medium: 'Medium',
  Substack: 'Substack',
  Newsletter: 'Newsletter',
  Podcast: 'Podcast',
  Portfolio: 'Portfolio',
  Shop: 'Shop',
  HomePage: 'Home Page',
  Other: '',
};

export default function AdminPage() {
  const [links, setLinks] = useState([]);
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [password, setPassword] = useState('');
  const [editingLink, setEditingLink] = useState(null);
  const [newLink, setNewLink] = useState({
    title: '',
    url: '',
    type: 'Other',
    description: ''
  });
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');
  const [loading, setLoading] = useState(false);
  const [menuOpen, setMenuOpen] = useState(false);
  const [uiConfig, setUiConfig] = useState({ username: '', title: '', primaryColor: '', secondaryColor: '', backgroundColor: '' });
  const [descActive, setDescActive] = useState(false);

  useEffect(() => {
    const token = localStorage.getItem('adminToken');
    if (token) {
      setIsAuthenticated(true);
      fetchConfigAndLinks();
    }
  }, []);

  const fetchConfigAndLinks = async () => {
    try {
      const response = await fetch('/api/admin/config', {
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('adminToken')}`
        }
      });
      if (!response.ok) {
        if (response.status === 401) {
          localStorage.removeItem('adminToken');
          setIsAuthenticated(false);
          setError('Session expired. Please log in again.');
          return;
        }
        throw new Error('Failed to fetch config');
      }
      const data = await response.json();
      if (data.success && data.data) {
        if (data.data.links) {
          setLinks(data.data.links);
        }
        if (data.data.ui) {
          setUiConfig(data.data.ui);
        }
      } else {
        setError(data.error ? data.error.message : 'Failed to load config');
      }
    } catch (err) {
      setError('Error loading config: ' + err.message);
    }
  };

  const handleLogin = async (e) => {
    e.preventDefault();
    setError('');
    setLoading(true);
    try {
      const response = await fetch('/api/admin/login', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ password }),
      });
      const data = await response.json();
      if (data.success && data.data && data.data.token) {
        localStorage.setItem('adminToken', data.data.token);
        setIsAuthenticated(true);
        fetchConfigAndLinks();
      } else {
        setError(data.error ? data.error.message : 'Invalid password');
      }
    } catch (err) {
      setError('Login error: ' + err.message);
    } finally {
      setLoading(false);
    }
  };

  const handleLogout = () => {
    localStorage.removeItem('adminToken');
    setIsAuthenticated(false);
    setLinks([]);
    setPassword('');
    setSuccess('Logged out successfully');
  };

  const handleRefreshConfig = async () => {
    setError("");
    setSuccess("");
    try {
      const response = await fetch("/api/admin/refresh-config", {
        method: "POST",
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('adminToken')}`
        }
      });
      const data = await response.json();
      if (!response.ok) {
        throw new Error(data.error ? data.error.message : 'Failed to refresh config');
      }
      if (data.success && data.data && data.data.status) {
        setSuccess(data.data.status);
      } else {
        setSuccess('Config refreshed from disk');
      }
      fetchConfigAndLinks();
    } catch (err) {
      setError('Error refreshing config: ' + err.message);
    }
  };

  const handleSaveLink = async (e) => {
    e.preventDefault();
    setError('');
    setSuccess('');

    try {
      const response = await fetch('/api/admin/links', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${localStorage.getItem('adminToken')}`
        },
        body: JSON.stringify(newLink),
      });

      if (!response.ok) {
        if (response.status === 401) {
          localStorage.removeItem('adminToken');
          setIsAuthenticated(false);
          setError('Session expired. Please log in again.');
          return;
        }
        throw new Error('Failed to save link');
      }

      const data = await response.json();
      if (!data.success || !data.data) {
        throw new Error(data.error ? data.error.message : 'Failed to save link');
      }
      const savedLink = data.data;
      const updatedLinks = editingLink
        ? links.map(link => link.id === editingLink.id ? savedLink : link)
        : [...links, savedLink];

      setLinks(updatedLinks);
      setEditingLink(null);
      setNewLink({ title: '', url: '', type: 'Other', description: '' });
      setSuccess(editingLink ? 'Link updated successfully' : 'Link added successfully');
    } catch (err) {
      setError('Error saving link: ' + err.message);
    }
  };

  const handleDeleteLink = async (linkId) => {
    setError('');
    setSuccess('');

    try {
      const response = await fetch(`/api/admin/links/${linkId}`, {
        method: 'DELETE',
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('adminToken')}`
        }
      });

      if (!response.ok) {
        if (response.status === 401) {
          localStorage.removeItem('adminToken');
          setIsAuthenticated(false);
          setError('Session expired. Please log in again.');
          return;
        }
        throw new Error('Failed to delete link');
      }

      const updatedLinks = links.filter(link => link.id !== linkId);
      setLinks(updatedLinks);
      setSuccess('Link deleted successfully');
    } catch (err) {
      setError('Error deleting link: ' + err.message);
    }
  };

  const handleTypeChange = (e) => {
    const type = e.target.value;
    let url = newLink.url;
    let title = newLink.title;
    // Only autofill if not editing or if url is empty or matches previous example
    if (!editingLink && (!url || LINK_TYPE_URL_EXAMPLES[newLink.type] === url)) {
      url = LINK_TYPE_URL_EXAMPLES[type] || '';
    }
    // Only autofill title if not editing or if title is empty or matches previous example
    if (!editingLink && (!title || LINK_TYPE_TITLE_EXAMPLES[newLink.type] === title)) {
      title = LINK_TYPE_TITLE_EXAMPLES[type] || '';
    }
    setNewLink({ ...newLink, type, url, title });
  };

  const handleUrlChange = (e) => {
    const url = e.target.value;
    let title = newLink.title;
    // If title is empty or matches the previous example, try to auto-generate from URL
    if (!editingLink && (!title || LINK_TYPE_TITLE_EXAMPLES[newLink.type] === title)) {
      try {
        const urlObj = new URL(url);
        // Use the hostname as a fallback title
        title = urlObj.hostname.replace('www.', '').split('.')[0].charAt(0).toUpperCase() + urlObj.hostname.replace('www.', '').split('.')[0].slice(1);
      } catch {
        title = '';
      }
    }
    setNewLink({ ...newLink, url, title });
  };

  if (!isAuthenticated) {
    return (
      <div className="min-h-screen bg-gray-100 flex items-center justify-center">
        <div className="max-w-md w-full space-y-8 p-8 bg-white rounded-lg shadow-lg">
          <div>
            <h2 className="text-center text-3xl font-extrabold text-gray-900">
              Admin Login
            </h2>
          </div>
          {success && (
            <div className="bg-green-50 border-l-4 border-green-400 p-4 mb-2">
              <div className="flex">
                <div className="flex-shrink-0">
                  <svg className="h-5 w-5 text-green-400" viewBox="0 0 20 20" fill="currentColor">
                    <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clipRule="evenodd" />
                  </svg>
                </div>
                <div className="ml-3">
                  <p className="text-sm text-green-700">{success}</p>
                </div>
              </div>
            </div>
          )}
          {error && (
            <div className="bg-red-50 border-l-4 border-red-400 p-4 mb-2">
              <div className="flex">
                <div className="flex-shrink-0">
                  <svg className="h-5 w-5 text-red-400" viewBox="0 0 20 20" fill="currentColor">
                    <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clipRule="evenodd" />
                  </svg>
                </div>
                <div className="ml-3">
                  <p className="text-sm text-red-700">{error}</p>
                </div>
              </div>
            </div>
          )}
          <form className="mt-8 space-y-6" onSubmit={handleLogin} method="post" autoComplete="off">
            <div>
              <label htmlFor="password" className="block text-sm font-medium text-gray-700">Password</label>
              <input
                id="password"
                name="password"
                type="password"
                required
                className="appearance-none rounded-md relative block w-full px-3 py-2 border border-gray-300 placeholder-gray-500 text-gray-900 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 focus:z-10 sm:text-sm"
                placeholder="Admin Password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                disabled={loading}
                autoFocus
              />
            </div>
            <div>
              <button
                type="submit"
                className="group relative w-full flex justify-center py-2 px-4 border border-transparent text-sm font-medium rounded-md text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 disabled:opacity-50"
                disabled={loading}
              >
                {loading ? 'Signing in...' : 'Sign in'}
              </button>
            </div>
          </form>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-100 flex flex-col">
      <AdminHeader onLogout={handleLogout} />
      <main className="flex-1">
        <div className="max-w-7xl mx-auto">
          <div className="mb-8">
            <h1 className="text-3xl font-bold text-gray-900">Manage Links</h1>
          </div>

          {error && (
            <div className="mb-4 bg-red-50 border-l-4 border-red-400 p-4">
              <div className="flex">
                <div className="flex-shrink-0">
                  <svg className="h-5 w-5 text-red-400" viewBox="0 0 20 20" fill="currentColor">
                    <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clipRule="evenodd" />
                  </svg>
                </div>
                <div className="ml-3">
                  <p className="text-sm text-red-700">{error}</p>
                </div>
              </div>
            </div>
          )}

          {success && (
            <div className="mb-4 bg-green-50 border-l-4 border-green-400 p-4">
              <div className="flex">
                <div className="flex-shrink-0">
                  <svg className="h-5 w-5 text-green-400" viewBox="0 0 20 20" fill="currentColor">
                    <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clipRule="evenodd" />
                  </svg>
                </div>
                <div className="ml-3">
                  <p className="text-sm text-green-700">{success}</p>
                </div>
              </div>
            </div>
          )}

          <div className="bg-white shadow overflow-hidden sm:rounded-lg">
            <form onSubmit={handleSaveLink} className="p-6 border-b border-gray-200">
              <div className="grid grid-cols-1 gap-6 sm:grid-cols-2">
                <div>
                  <label className="block text-sm font-medium text-gray-700">Type</label>
                  <select
                    required
                    value={newLink.type}
                    onChange={handleTypeChange}
                    className="mt-1 block w-full border border-gray-300 rounded-md shadow-sm py-2 px-3 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm"
                  >
                    {LINK_TYPES.map(type => (
                      <option key={type.value} value={type.value}>
                        {type.label}
                      </option>
                    ))}
                  </select>
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700">URL</label>
                  <input
                    type="url"
                    required
                    value={newLink.url}
                    onChange={handleUrlChange}
                    placeholder={LINK_TYPE_URL_EXAMPLES[newLink.type] || 'https://'}
                    className="mt-1 block w-full border border-gray-300 rounded-md shadow-sm py-2 px-3 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm"
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700">Title</label>
                  <input
                    type="text"
                    required
                    value={newLink.title}
                    onChange={(e) => setNewLink({ ...newLink, title: e.target.value })}
                    placeholder={LINK_TYPE_TITLE_EXAMPLES[newLink.type] || ''}
                    className="mt-1 block w-full border border-gray-300 rounded-md shadow-sm py-2 px-3 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm"
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700">Description</label>
                  {!descActive ? (
                    <input
                      type="text"
                      value={newLink.description}
                      onFocus={() => setDescActive(true)}
                      placeholder="(optional)"
                      className="mt-1 block w-full border border-gray-300 rounded-md shadow-sm py-2 px-3 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm"
                    />
                  ) : null}
                  {descActive ? (
                    <input
                      type="text"
                      value={newLink.description}
                      onChange={(e) => setNewLink({ ...newLink, description: e.target.value })}
                      className="mt-1 block w-full border border-gray-300 rounded-md shadow-sm py-2 px-3 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm"
                    />
                  ) : null}
                </div>
              </div>
              <div className="mt-6 flex justify-end space-x-3">
                {editingLink && (
                  <button
                    type="button"
                    onClick={() => {
                      setEditingLink(null);
                      setNewLink({ title: '', url: '', type: 'Other', description: '' });
                    }}
                    className="inline-flex items-center px-4 py-2 border border-gray-300 text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
                  >
                    Cancel
                  </button>
                )}
                <button
                  type="submit"
                  className="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
                >
                  {editingLink ? 'Update Link' : 'Add Link'}
                </button>
              </div>
            </form>

            <div className="divide-y divide-gray-200">
              {links.map((link) => (
                <div key={link.id} className="p-6 flex items-center justify-between">
                  <div className="flex-1">
                    <h3 className="text-lg font-medium text-gray-900">{link.title}</h3>
                    <p className="mt-1 text-sm text-gray-500">{link.description}</p>
                    <div className="mt-2 flex items-center">
                      <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-gray-100 text-gray-800">
                        {LINK_TYPES.find(t => t.value === link.type)?.label || link.type}
                      </span>
                      <a href={link.url} target="_blank" rel="noopener noreferrer" className="ml-4 text-sm text-indigo-600 hover:text-indigo-500">
                        {link.url}
                      </a>
                    </div>
                    <div className="mt-2 flex items-center">
                      <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-green-100 text-green-800">
                        {link.clicks || 0} clicks
                      </span>
                    </div>
                  </div>
                  <div className="ml-4 flex-shrink-0 flex space-x-2">
                    <button
                      onClick={() => {
                        setEditingLink(link);
                        setNewLink(link);
                      }}
                      className="inline-flex items-center p-2 border border-transparent rounded-full shadow-sm text-white bg-indigo-600 hover:bg-indigo-700"
                    >
                      <PencilIcon className="h-5 w-5" />
                    </button>
                    <button
                      onClick={() => handleDeleteLink(link.id)}
                      className="inline-flex items-center p-2 border border-transparent rounded-full shadow-sm text-white bg-red-600 hover:bg-red-700"
                    >
                      <TrashIcon className="h-5 w-5" />
                    </button>
                  </div>
                </div>
              ))}
            </div>
          </div>
        </div>
      </main>
      <AdminFooter />
    </div>
  );
} 