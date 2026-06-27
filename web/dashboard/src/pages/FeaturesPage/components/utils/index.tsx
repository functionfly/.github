import { useState } from "react";
import { features } from "../../data/features";

// Performance optimization: Lazy load images when they become visible
export const LazyImage = ({ src, alt, className, ...props }: any) => {
  const [isLoaded, setIsLoaded] = useState(false);
  const [isInView, setIsInView] = useState(false);

  return (
    <img
      src={isInView ? src : undefined}
      alt={alt}
      className={`${className} ${isLoaded ? 'opacity-100' : 'opacity-0'} transition-opacity duration-300`}
      loading="lazy"
      onLoad={() => setIsLoaded(true)}
      onError={() => setIsLoaded(true)} // Show placeholder on error
      {...props}
    />
  );
};

// Structured Data Component for Features
export const StructuredData = () => {
  const featuresStructuredData = {
    "@context": "https://schema.org",
    "@type": "WebPage",
    "name": "FunctionFly Features",
    "description": "Discover powerful features for modern developers: multi-provider deployment, intelligent failover, predictive routing, and advanced analytics.",
    "url": "https://functionfly.com/features",
    "mainEntity": {
      "@type": "SoftwareApplication",
      "name": "FunctionFly",
      "description": "Serverless function deployment platform with multi-provider support",
      "offers": [
        {
          "@type": "Offer",
          "name": "Free Tier",
          "description": "Free tier with basic features for getting started",
          "price": "0",
          "priceCurrency": "USD"
        },
        {
          "@type": "Offer",
          "name": "Pro Plan",
          "description": "Professional plan with advanced features",
          "price": "29",
          "priceCurrency": "USD",
          "priceValidUntil": "2028-12-31"
        }
      ],
      "featureList": features.map(feature => ({
        "@type": "PropertyValue",
        "name": feature.title,
        "value": feature.description,
        "category": feature.category
      }))
    },
    "hasPart": features.map(feature => ({
      "@type": "Service",
      "name": feature.title,
      "description": feature.description,
      "category": feature.category,
      "provider": {
        "@type": "Organization",
        "name": "FunctionFly"
      }
    }))
  };

  return (
    <script
      type="application/ld+json"
      dangerouslySetInnerHTML={{
        __html: JSON.stringify(featuresStructuredData)
      }}
    />
  );
};