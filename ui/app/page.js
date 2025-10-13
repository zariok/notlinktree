'use client';

import { useEffect, useState, useCallback } from 'react';
import LinkCard from './LinkCard';

export default function Page() {
  const [links, setLinks] = useState([]);
  const [config, setConfig] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  useEffect(() => {
    const fetchData = async () => {
      try {
        const [configResponse, linksResponse] = await Promise.all([
          fetch('/api/config'),
          fetch('/api/links')
        ]);
        
        if (!configResponse.ok || !linksResponse.ok) {
          throw new Error('Failed to fetch data');
        }
        
        const [configData, linksData] = await Promise.all([
          configResponse.json(),
          linksResponse.json()
        ]);
        
        setConfig(configData.data);
        setLinks(linksData.data.links || []);
        setLoading(false);
      } catch (err) {
        setError('Failed to load data. Please try again later.');
        setLoading(false);
      }
    };
    
    fetchData();
  }, []);

  const handleClick = useCallback(async (link) => {
    try {
      await fetch(`/api/click/${link.id}`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
      });
    } catch (err) {
      // Optionally handle error
    }
  }, []);

  // Dynamic background gradient from config
  const bgStyle = config
    ? {
        background: `linear-gradient(to bottom, ${config.primaryColor || '#6D28D9'}, ${config.secondaryColor || '#3B82F6'})`,
        minHeight: '100vh',
      }
    : {};

  return (
    <div className="min-h-screen flex flex-col items-center justify-center px-4" style={bgStyle}>
      <div className="w-full max-w-md flex flex-col items-center">
        {/* Profile Section */}
        <img
          src="/api/avatar"
          alt="Profile Avatar"
          width={112}
          height={112}
          className="h-28 w-28 rounded-full border-4 border-white shadow-lg mb-4 object-cover"
        />
        <h1 className="text-3xl font-bold text-white mb-2 dark:text-gray-100 text-center">
          {config ? config.username : '...'}
        </h1>
        <p className="text-white text-opacity-80 dark:text-gray-300 mb-8 text-center">
          {config ? config.title : '...'}
        </p>

        {/* Links Section */}
        {loading ? (
          <div className="min-h-[200px] flex items-center justify-center">
            <div className="animate-spin rounded-full h-12 w-12 border-t-2 border-b-2 border-white"></div>
          </div>
        ) : error ? (
          <div className="bg-white bg-opacity-10 p-8 rounded-lg backdrop-blur-sm text-white text-center">
            {error}
          </div>
        ) : links.length === 0 ? (
          <div className="text-center text-white text-opacity-80 dark:text-gray-300">
            No links available yet.
          </div>
        ) : (
          <ul className="w-full flex flex-col gap-4" role="list">
            {links.map((link) => (
              <LinkCard key={link.id} link={link} onClick={handleClick} color={config?.primaryColor} />
            ))}
          </ul>
        )}
      </div>
    </div>
  );
} 