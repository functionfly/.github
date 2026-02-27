import React from 'react';

interface AvatarProps {
  src?: string;
  alt: string;
  fallback?: string;
  className?: string;
  size?: number;
}

export function Avatar({
  src,
  alt,
  fallback,
  className = '',
  size = 40
}: AvatarProps) {
  const [imageError, setImageError] = React.useState(false);

  const handleImageError = () => {
    setImageError(true);
  };

  const getFallbackText = () => {
    if (fallback) return fallback;
    return alt.charAt(0).toUpperCase();
  };

  return (
    <div
      className={`relative inline-flex items-center justify-center overflow-hidden bg-gray-100 rounded-full ${className}`}
      style={{ width: size, height: size }}
    >
      {src && !imageError ? (
        <img
          src={src}
          alt={alt}
          className="w-full h-full object-cover"
          onError={handleImageError}
        />
      ) : (
        <span className="text-sm font-medium text-gray-700">
          {getFallbackText()}
        </span>
      )}
    </div>
  );
}