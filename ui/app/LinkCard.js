'use client';

import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import {
  faGithub,
  faTwitter,
  faLinkedin,
  faFacebook,
  faInstagram,
  faYoutube,
  faTiktok,
  faTwitch,
  faDiscord,
  faSpotify,
  faSoundcloud,
  faPatreon,
  faMedium,
} from '@fortawesome/free-brands-svg-icons';
import { faLink } from '@fortawesome/free-solid-svg-icons';

function getIconForLink(link) {
  const url = link.url?.toLowerCase() || '';
  if (url.includes('github.com')) return faGithub;
  if (url.includes('twitter.com')) return faTwitter;
  if (url.includes('linkedin.com')) return faLinkedin;
  if (url.includes('facebook.com')) return faFacebook;
  if (url.includes('instagram.com')) return faInstagram;
  if (url.includes('youtube.com')) return faYoutube;
  if (url.includes('tiktok.com')) return faTiktok;
  if (url.includes('twitch.tv')) return faTwitch;
  if (url.includes('discord.gg') || url.includes('discord.com')) return faDiscord;
  if (url.includes('spotify.com')) return faSpotify;
  if (url.includes('soundcloud.com')) return faSoundcloud;
  if (url.includes('patreon.com')) return faPatreon;
  if (url.includes('medium.com')) return faMedium;
  // Add more as needed
  return faLink;
}

export default function LinkCard({ link, onClick, color }) {
  // Use the color prop for background and border
  const style = color
    ? {
        background: `linear-gradient(90deg, ${color} 60%, rgba(255,255,255,0.08))`,
        borderColor: color,
        backdropFilter: 'blur(6px)',
      }
    : { backdropFilter: 'blur(6px)' };

  const icon = getIconForLink(link);

  return (
    <li>
      <a
        href={link.url}
        onClick={() => onClick(link)}
        className="block w-full rounded-xl py-4 px-6 text-lg font-semibold text-white shadow-lg border-2 border-transparent transition-all duration-200 hover:scale-[1.03] hover:shadow-2xl focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-indigo-400"
        style={style}
        tabIndex={0}
        aria-label={link.title}
        target="_blank"
        rel="noopener noreferrer"
      >
        <div className="flex flex-row items-center gap-4">
          <FontAwesomeIcon icon={icon} className="text-2xl min-w-[1.5em]" />
          <div className="flex flex-col items-start">
            <span className="text-lg font-semibold mb-1">{link.title}</span>
            {link.description && (
              <span className="text-sm text-white text-opacity-80 text-left">{link.description}</span>
            )}
          </div>
        </div>
      </a>
    </li>
  );
} 