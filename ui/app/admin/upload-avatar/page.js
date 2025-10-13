"use client";
import React, { useState, useRef, useCallback, useEffect } from "react";
import Cropper from "react-easy-crop";
import getCroppedImg from "./utils/cropImage";
import AdminHeader from '../AdminHeader';
import AdminFooter from '../AdminFooter';
import Image from 'next/image';

export default function UploadAvatarPage() {
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [password, setPassword] = useState('');
  const [loading, setLoading] = useState(false);
  const [authError, setAuthError] = useState('');
  const [authSuccess, setAuthSuccess] = useState('');
  
  const [imageSrc, setImageSrc] = useState(null);
  const [crop, setCrop] = useState({ x: 0, y: 0 });
  const [zoom, setZoom] = useState(1);
  const [croppedAreaPixels, setCroppedAreaPixels] = useState(null);
  const [croppedImage, setCroppedImage] = useState(null);
  const [hasPreviewed, setHasPreviewed] = useState(false);
  const [uploading, setUploading] = useState(false);
  const [error, setError] = useState("");
  const [success, setSuccess] = useState("");
  const [avatarUrl, setAvatarUrl] = useState(null);
  const inputRef = useRef();

  useEffect(() => {
    const token = localStorage.getItem('adminToken');
    if (token) {
      setIsAuthenticated(true);
      fetchAvatar();
    }
  }, []); // Only run on mount

  const fetchAvatar = async () => {
    try {
      const bust = Date.now();
      const res = await fetch(`/api/avatar?ts=${bust}`);
      if (res.ok) {
        const blob = await res.blob();
        setAvatarUrl(URL.createObjectURL(blob));
      } else {
        setAvatarUrl(null);
      }
    } catch (err) {
      setAvatarUrl(null);
    }
  };

  const handleLogin = async (e) => {
    e.preventDefault();
    setAuthError('');
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
        fetchAvatar();
      } else {
        setAuthError(data.error ? data.error.message : 'Invalid password');
      }
    } catch (err) {
      setAuthError('Login error: ' + err.message);
    } finally {
      setLoading(false);
    }
  };

  const handleLogout = () => {
    localStorage.removeItem('adminToken');
    setIsAuthenticated(false);
    setPassword('');
    setAuthSuccess('Logged out successfully');
  };

  const onFileChange = async (e) => {
    setError("");
    setSuccess("");
    setHasPreviewed(false);
    setCroppedImage(null);
    const file = e.target.files[0];
    if (!file) return;
    if (!file.type.startsWith("image/")) {
      setError("Please select an image file.");
      return;
    }
    const reader = new FileReader();
    reader.onload = () => setImageSrc(reader.result);
    reader.readAsDataURL(file);
  };

  const onCropComplete = useCallback((_, croppedAreaPixels) => {
    setCroppedAreaPixels(croppedAreaPixels);
  }, []);

  // Reset preview state on crop/zoom change
  const handleCropChange = (newCrop) => {
    setCrop(newCrop);
    setHasPreviewed(false);
  };
  const handleZoomChange = (newZoom) => {
    setZoom(newZoom);
    setHasPreviewed(false);
  };

  const showCroppedImage = useCallback(async () => {
    try {
      const cropped = await getCroppedImg(imageSrc, croppedAreaPixels);
      setCroppedImage(cropped);
      setHasPreviewed(true);
    } catch (e) {
      setError("Failed to crop image.");
    }
  }, [imageSrc, croppedAreaPixels]);

  const handleUpload = async () => {
    setUploading(true);
    setError("");
    setSuccess("");
    try {
      const blob = await fetch(croppedImage).then((r) => r.blob());
      const formData = new FormData();
      formData.append("avatar", blob, "avatar.png");
      const token = localStorage.getItem("adminToken");
      const res = await fetch("/api/admin/avatar", {
        method: "POST",
        headers: {
          Authorization: `Bearer ${token}`,
        },
        body: formData,
      });
      const data = await res.json();
      if (!res.ok) throw new Error(data.error || "Upload failed");
      setSuccess("Avatar uploaded successfully!");
      setImageSrc(null);
      setCroppedImage(null);
      setHasPreviewed(false);
      // Refetch avatar with cache busting
      fetchAvatar();
    } catch (e) {
      setError(e.message);
    } finally {
      setUploading(false);
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
          {authSuccess && (
            <div className="bg-green-50 border-l-4 border-green-400 p-4 mb-2">
              <div className="flex">
                <div className="flex-shrink-0">
                  <svg className="h-5 w-5 text-green-400" viewBox="0 0 20 20" fill="currentColor">
                    <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clipRule="evenodd" />
                  </svg>
                </div>
                <div className="ml-3">
                  <p className="text-sm text-green-700">{authSuccess}</p>
                </div>
              </div>
            </div>
          )}
          {authError && (
            <div className="bg-red-50 border-l-4 border-red-400 p-4 mb-2">
              <div className="flex">
                <div className="flex-shrink-0">
                  <svg className="h-5 w-5 text-red-400" viewBox="0 0 20 20" fill="currentColor">
                    <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clipRule="evenodd" />
                  </svg>
                </div>
                <div className="ml-3">
                  <p className="text-sm text-red-700">{authError}</p>
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
        <div className="max-w-xl mx-auto py-8">
          <h1 className="text-2xl font-bold mb-4 text-center">Upload & Crop Avatar</h1>

          {/* Show existing avatar if any */}
          {avatarUrl && (
            <div className="flex flex-col items-center mb-6">
              <Image
                src={avatarUrl}
                alt="Current Avatar"
                width={128}
                height={128}
                className="w-32 h-32 rounded-full border mb-2"
                unoptimized
              />
              <span className="text-gray-500 text-sm">Current Avatar</span>
            </div>
          )}

          {/* Upload button */}
          <div className="flex justify-center mb-4">
            <input
              type="file"
              accept="image/*"
              onChange={onFileChange}
              ref={inputRef}
              className="hidden"
            />
            <button
              type="button"
              className="bg-blue-600 text-white px-4 py-2 rounded shadow"
              onClick={() => inputRef.current && inputRef.current.click()}
              disabled={uploading}
            >
              Upload Image
            </button>
          </div>

          {/* Cropper */}
          {imageSrc && (
            <div className="relative w-64 h-64 bg-gray-200 mx-auto mb-4">
              <Cropper
                image={imageSrc}
                crop={crop}
                zoom={zoom}
                aspect={1}
                cropShape="round"
                showGrid={false}
                onCropChange={handleCropChange}
                onZoomChange={handleZoomChange}
                onCropComplete={onCropComplete}
              />
            </div>
          )}

          {/* Preview and Upload buttons */}
          {imageSrc && (
            <div className="flex flex-col items-center mb-4">
              {!hasPreviewed ? (
                <button
                  className="bg-blue-600 text-white px-4 py-2 rounded mb-2"
                  onClick={showCroppedImage}
                  disabled={uploading}
                >
                  Preview Crop
                </button>
              ) : (
                <>
                  {croppedImage && (
                    <Image
                      src={croppedImage}
                      alt="Cropped Preview"
                      width={128}
                      height={128}
                      className="w-32 h-32 rounded-full border mb-2"
                      unoptimized
                    />
                  )}
                  <button
                    className="bg-green-600 text-white px-4 py-2 rounded"
                    onClick={handleUpload}
                    disabled={!croppedImage || uploading}
                  >
                    {uploading ? "Uploading..." : "Upload Avatar"}
                  </button>
                </>
              )}
            </div>
          )}

          {error && <div className="text-red-600 mb-2 text-center">{error}</div>}
          {success && <div className="text-green-600 mb-2 text-center">{success}</div>}
        </div>
      </main>
      <AdminFooter />
    </div>
  );
} 