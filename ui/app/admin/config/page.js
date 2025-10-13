"use client";
import { useState, useEffect } from "react";
import AdminHeader from '../AdminHeader';
import { usePathname } from 'next/navigation';

export default function AdminConfigPage() {
  const pathname = usePathname();
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [password, setPassword] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');
  
  const [uiConfig, setUiConfig] = useState({ username: "", title: "", primaryColor: "#6D28D9", secondaryColor: "#3B82F6" });
  const [uiLoading, setUiLoading] = useState(true);
  const [uiSuccess, setUiSuccess] = useState("");
  const [uiError, setUiError] = useState("");

  const [newPassword, setNewPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [pwSuccess, setPwSuccess] = useState("");
  const [pwError, setPwError] = useState("");
  const [pwLoading, setPwLoading] = useState(false);

  useEffect(() => {
    const token = localStorage.getItem('adminToken');
    if (token) {
      setIsAuthenticated(true);
      fetchConfig();
    }
  }, []); // Only run on mount

  const fetchConfig = async () => {
    try {
      const response = await fetch("/api/admin/config", {
        headers: { Authorization: `Bearer ${localStorage.getItem("adminToken")}` },
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
      if (data && data.data && data.data.ui) {
        setUiConfig({
          username: data.data.ui.username || "",
          title: data.data.ui.title || "",
          primaryColor: data.data.ui.primaryColor || "#6D28D9",
          secondaryColor: data.data.ui.secondaryColor || "#3B82F6",
        });
      }
    } catch (err) {
      setUiError("Failed to load config");
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
        fetchConfig();
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
    setPassword('');
    setSuccess('Logged out successfully');
  };

  const handleUiChange = (e) => {
    setUiConfig({ ...uiConfig, [e.target.name]: e.target.value });
  };

  const handleUiSave = async (e) => {
    e.preventDefault();
    setUiSuccess("");
    setUiError("");
    try {
      const res = await fetch("/api/admin/config", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          Authorization: `Bearer ${localStorage.getItem("adminToken")}`,
        },
        body: JSON.stringify({ ui: uiConfig }),
      });
      if (!res.ok) throw new Error("Failed to save config");
      setUiSuccess("UI config saved successfully!");
    } catch (err) {
      setUiError("Failed to save config");
    }
  };

  const handlePasswordSave = async (e) => {
    e.preventDefault();
    setPwSuccess("");
    setPwError("");
    if (newPassword !== confirmPassword) {
      setPwError("Passwords do not match");
      return;
    }
    setPwLoading(true);
    try {
      const res = await fetch("/api/admin/password", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          Authorization: `Bearer ${localStorage.getItem("adminToken")}`,
        },
        body: JSON.stringify({ password: newPassword }),
      });
      if (!res.ok) throw new Error("Failed to change password");
      setPwSuccess("Password changed successfully!");
      setNewPassword("");
      setConfirmPassword("");
    } catch (err) {
      setPwError("Failed to change password");
    } finally {
      setPwLoading(false);
    }
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
        <div className="max-w-2xl mx-auto py-8 px-4">
          <h1 className="text-2xl font-bold mb-6">Admin Config</h1>
          {/* UI Config Section */}
          <form onSubmit={handleUiSave} className="bg-white rounded-lg shadow p-6 mb-8">
            <h2 className="text-lg font-semibold mb-4">UI Config</h2>
            {uiError && <div className="text-red-600 mb-2">{uiError}</div>}
            {uiSuccess && <div className="text-green-600 mb-2">{uiSuccess}</div>}
            <div className="mb-4">
              <label className="block text-sm font-medium text-gray-700 mb-1">Username</label>
              <input type="text" name="username" value={uiConfig.username} onChange={handleUiChange} className="w-full border rounded px-3 py-2" />
            </div>
            <div className="mb-4">
              <label className="block text-sm font-medium text-gray-700 mb-1">Title</label>
              <input type="text" name="title" value={uiConfig.title} onChange={handleUiChange} className="w-full border rounded px-3 py-2" />
            </div>
            <div className="mb-4 flex gap-4">
              <div className="flex-1">
                <label className="block text-sm font-medium text-gray-700 mb-1">Primary Color</label>
                <div className="flex items-center gap-2">
                  <input type="color" name="primaryColor" value={uiConfig.primaryColor} onChange={handleUiChange} className="w-10 h-10 p-0 border-none" />
                  <input type="text" name="primaryColor" value={uiConfig.primaryColor} onChange={handleUiChange} className="w-24 border rounded px-2 py-1" />
                </div>
              </div>
              <div className="flex-1">
                <label className="block text-sm font-medium text-gray-700 mb-1">Secondary Color</label>
                <div className="flex items-center gap-2">
                  <input type="color" name="secondaryColor" value={uiConfig.secondaryColor} onChange={handleUiChange} className="w-10 h-10 p-0 border-none" />
                  <input type="text" name="secondaryColor" value={uiConfig.secondaryColor} onChange={handleUiChange} className="w-24 border rounded px-2 py-1" />
                </div>
              </div>
            </div>
            <button type="submit" className="mt-4 px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700">Save UI Config</button>
          </form>
          {/* Change Password Section */}
          <form onSubmit={handlePasswordSave} className="bg-white rounded-lg shadow p-6">
            <h2 className="text-lg font-semibold mb-4">Change Password</h2>
            {pwError && <div className="text-red-600 mb-2">{pwError}</div>}
            {pwSuccess && <div className="text-green-600 mb-2">{pwSuccess}</div>}
            <div className="mb-4">
              <label className="block text-sm font-medium text-gray-700 mb-1">New Password</label>
              <input type="password" name="newPassword" value={newPassword} onChange={e => setNewPassword(e.target.value)} className="w-full border rounded px-3 py-2" />
            </div>
            <div className="mb-4">
              <label className="block text-sm font-medium text-gray-700 mb-1">Confirm Password</label>
              <input type="password" name="confirmPassword" value={confirmPassword} onChange={e => setConfirmPassword(e.target.value)} className="w-full border rounded px-3 py-2" />
            </div>
            <button type="submit" className="mt-4 px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700" disabled={pwLoading}>{pwLoading ? "Saving..." : "Save New Password"}</button>
          </form>
        </div>
      </main>
    </div>
  );
} 