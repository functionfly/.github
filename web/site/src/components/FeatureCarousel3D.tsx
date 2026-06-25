import React, { useEffect, useRef, useState, useCallback } from "react";
import { Chamber, CornerBrace } from "./sc";

interface Feature {
  icon: string;
  title: string;
  desc: string;
  tag: string;
  color: "flame" | "cyan" | "strat" | "taxiway" | "beacon" | "afterburner";
}

interface FeatureCarousel3DProps {
  features: Feature[];
}

const COLOR_MAP: Record<Feature["color"], { bg: string; border: string; glow: string }> = {
  flame: { bg: "rgba(255, 107, 53, 0.1)", border: "rgba(255, 107, 53, 0.3)", glow: "rgba(255, 107, 53, 0.4)" },
  cyan: { bg: "rgba(0, 212, 255, 0.1)", border: "rgba(0, 212, 255, 0.3)", glow: "rgba(0, 212, 255, 0.4)" },
  strat: { bg: "rgba(91, 124, 245, 0.1)", border: "rgba(91, 124, 245, 0.3)", glow: "rgba(91, 124, 245, 0.4)" },
  taxiway: { bg: "rgba(0, 255, 157, 0.1)", border: "rgba(0, 255, 157, 0.3)", glow: "rgba(0, 255, 157, 0.4)" },
  beacon: { bg: "rgba(255, 184, 0, 0.1)", border: "rgba(255, 184, 0, 0.3)", glow: "rgba(255, 184, 0, 0.4)" },
  afterburner: { bg: "rgba(255, 79, 94, 0.1)", border: "rgba(255, 79, 94, 0.3)", glow: "rgba(255, 79, 94, 0.4)" },
};

const SCROLL_SPEED = 5;
const CARD_GAP = 24;

export const FeatureCarousel3D: React.FC<FeatureCarousel3DProps> = ({ features }) => {
  const trackRef = useRef<HTMLDivElement>(null);
  const [cardWidth, setCardWidth] = useState(380);
  const [isMounted, setIsMounted] = useState(false);
  const [isPaused, setIsPaused] = useState(false);
  const [activeIndex, setActiveIndex] = useState(0);
  const [progress, setProgress] = useState(0);
  const animationRef = useRef<number | null>(null);
  const scrollPosRef = useRef(0);
  const lastTimeRef = useRef(0);
  const singleCardWidthRef = useRef(404);
  const velocityRef = useRef(0);
  const lastTouchXRef = useRef(0);
  const touchStartTimeRef = useRef(0);

  // Only run on client
  useEffect(() => {
    setIsMounted(true);
    const updateCardWidth = () => {
      if (typeof window === 'undefined') return;
      let width;
      if (window.innerWidth < 640) {
        width = window.innerWidth - 48;
      } else if (window.innerWidth < 1024) {
        width = 340;
      } else {
        width = 380;
      }
      setCardWidth(width);
      singleCardWidthRef.current = width + CARD_GAP;
    };
    updateCardWidth();
    window.addEventListener("resize", updateCardWidth);
    return () => window.removeEventListener("resize", updateCardWidth);
  }, []);

  const totalWidth = features.length * singleCardWidthRef.current;

  // Snap to nearest card
  const snapToCard = useCallback(() => {
    if (!trackRef.current) return;
    const nearestIndex = Math.round(scrollPosRef.current / singleCardWidthRef.current) % features.length;
    const targetScroll = nearestIndex * singleCardWidthRef.current;
    scrollPosRef.current = targetScroll;
    trackRef.current.scrollTo({ left: targetScroll, behavior: 'smooth' });
    setActiveIndex(nearestIndex);
  }, [features.length]);

  // Animation loop with momentum
  const animate = useCallback((timestamp: number) => {
    if (!trackRef.current) return;
    
    if (!lastTimeRef.current) lastTimeRef.current = timestamp;
    const delta = timestamp - lastTimeRef.current;
    lastTimeRef.current = timestamp;

    if (!isPaused) {
      // Apply momentum
      if (Math.abs(velocityRef.current) > 0.1) {
        scrollPosRef.current += velocityRef.current;
        velocityRef.current *= 0.95; // friction
      } else {
        scrollPosRef.current += (SCROLL_SPEED * delta) / 16;
      }
      
      // Seamless loop
      if (scrollPosRef.current >= totalWidth - cardWidth) {
        scrollPosRef.current = 0;
      } else if (scrollPosRef.current < 0) {
        scrollPosRef.current = totalWidth - cardWidth;
      }
      
      trackRef.current.scrollLeft = scrollPosRef.current;
      
      // Update progress (0 to 1)
      const maxScroll = totalWidth - cardWidth;
      setProgress(maxScroll > 0 ? scrollPosRef.current / maxScroll : 0);
      
      // Update active index for dots
      const idx = Math.floor(scrollPosRef.current / singleCardWidthRef.current) % features.length;
      setActiveIndex(idx);
    }

    animationRef.current = requestAnimationFrame(animate);
  }, [isPaused, cardWidth, totalWidth, features.length]);

  useEffect(() => {
    if (isMounted) {
      animationRef.current = requestAnimationFrame(animate);
    }
    return () => {
      if (animationRef.current) {
        cancelAnimationFrame(animationRef.current);
      }
    };
  }, [isMounted, animate]);

  // Pause on tab visibility change
  useEffect(() => {
    const handleVisibilityChange = () => {
      if (document.hidden) {
        setIsPaused(true);
      } else {
        setIsPaused(false);
        velocityRef.current = 0;
      }
    };
    document.addEventListener('visibilitychange', handleVisibilityChange);
    return () => document.removeEventListener('visibilitychange', handleVisibilityChange);
  }, []);

  // Keyboard shortcuts
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === ' ') {
        e.preventDefault();
        setIsPaused(p => !p);
      }
      const num = parseInt(e.key);
      if (num >= 1 && num <= features.length) {
        e.preventDefault();
        goToIndex(num - 1);
      }
    };
    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [features.length]);

  // Calculate card transforms based on scroll position
  const getCardTransform = (index: number) => {
    const scrollLeft = scrollPosRef.current;
    const cardCenter = index * singleCardWidthRef.current + cardWidth / 2;
    const viewportCenter = scrollLeft + window.innerWidth / 2;
    const offset = (cardCenter - viewportCenter) / singleCardWidthRef.current;
    
    const rotateY = offset * -35;
    const translateZ = Math.abs(offset) * -80;
    const scale = 1 - Math.abs(offset) * 0.08;
    const opacity = Math.max(0.4, 1 - Math.abs(offset) * 0.25);
    const zIndex = 100 - Math.abs(Math.round(offset));

    return { rotateY, translateZ, scale, opacity, zIndex };
  };

  // Content parallax offset
  const getContentParallax = (index: number) => {
    const scrollLeft = scrollPosRef.current;
    const cardCenter = index * singleCardWidthRef.current + cardWidth / 2;
    const viewportCenter = scrollLeft + window.innerWidth / 2;
    const offset = (cardCenter - viewportCenter) / singleCardWidthRef.current;
    return offset * 15; // pixels of parallax offset
  };

  // Handle mouse drag
  const isDragging = useRef(false);
  const startX = useRef(0);
  const scrollStart = useRef(0);

  const handleMouseDown = (e: React.MouseEvent) => {
    if (!trackRef.current) return;
    isDragging.current = true;
    startX.current = e.pageX;
    scrollStart.current = trackRef.current.scrollLeft;
    trackRef.current.style.cursor = "grabbing";
    setIsPaused(true);
    velocityRef.current = 0;
  };

  const handleMouseMove = (e: React.MouseEvent) => {
    if (!isDragging.current || !trackRef.current) return;
    e.preventDefault();
    const x = e.pageX - startX.current;
    const newScroll = scrollStart.current - x;
    velocityRef.current = (scrollStart.current - newScroll) * 0.5;
    trackRef.current.scrollLeft = newScroll;
    scrollPosRef.current = newScroll;
  };

  const handleMouseUp = () => {
    isDragging.current = false;
    if (trackRef.current) {
      trackRef.current.style.cursor = "grab";
    }
    snapToCard();
    setIsPaused(false);
  };

  // Touch handlers with momentum
  const handleTouchStart = (e: React.TouchEvent) => {
    if (!trackRef.current) return;
    lastTouchXRef.current = e.touches[0].pageX;
    touchStartTimeRef.current = Date.now();
    scrollStart.current = trackRef.current.scrollLeft;
    velocityRef.current = 0;
    setIsPaused(true);
  };

  const handleTouchMove = (e: React.TouchEvent) => {
    if (!trackRef.current) return;
    const x = e.touches[0].pageX;
    const deltaX = lastTouchXRef.current - x;
    const newScroll = scrollStart.current + deltaX;
    velocityRef.current = -deltaX * 0.5;
    lastTouchXRef.current = x;
    trackRef.current.scrollLeft = newScroll;
    scrollPosRef.current = newScroll;
  };

  const handleTouchEnd = () => {
    // Calculate swipe velocity
    const touchDuration = Date.now() - touchStartTimeRef.current;
    if (touchDuration < 200 && Math.abs(velocityRef.current) > 2) {
      // Quick swipe - apply momentum
      velocityRef.current *= 2;
    }
    snapToCard();
    setIsPaused(false);
  };

  // Jump to specific card
  const goToIndex = (index: number) => {
    if (!trackRef.current) return;
    const targetScroll = index * singleCardWidthRef.current;
    scrollPosRef.current = targetScroll;
    trackRef.current.scrollTo({ left: targetScroll, behavior: 'smooth' });
    setActiveIndex(index);
  };

  const goNext = () => {
    const nextIndex = (activeIndex + 1) % features.length;
    goToIndex(nextIndex);
  };

  const goPrev = () => {
    const prevIndex = activeIndex <= 0 ? features.length - 1 : activeIndex - 1;
    goToIndex(prevIndex);
  };

  // Show loading state during SSR
  if (!isMounted) {
    return (
      <div className="ff-carousel3d-container">
        <div className="ff-carousel3d-track" style={{ justifyContent: 'center', alignItems: 'center', minHeight: '420px' }}>
          {features.map((feature) => (
            <div key={feature.title} className="ff-carousel3d-card-wrapper" style={{ flex: '0 0 380px' }}>
              <Chamber ribs className="ff-carousel3d-card">
                <div className="ff-carousel3d-icon">{feature.icon}</div>
                <h3 className="ff-carousel3d-title">{feature.title}</h3>
                <p className="ff-carousel3d-desc">{feature.desc}</p>
                <span className="ff-carousel3d-tag">{feature.tag}</span>
              </Chamber>
            </div>
          ))}
        </div>
      </div>
    );
  }

  return (
    <div 
      className="ff-carousel3d-container"
      onMouseEnter={() => setIsPaused(true)}
      onMouseLeave={() => {
        if (!isDragging.current) {
          setIsPaused(false);
          snapToCard();
        }
      }}
    >
      {/* Navigation Arrows */}
      <button className="ff-carousel3d-arrow ff-carousel3d-arrow-left" onClick={goPrev} aria-label="Previous cards">
        <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
          <path d="M15 18l-6-6 6-6" />
        </svg>
      </button>

      <button className="ff-carousel3d-arrow ff-carousel3d-arrow-right" onClick={goNext} aria-label="Next cards">
        <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
          <path d="M9 18l6-6-6-6" />
        </svg>
      </button>

      {/* Gradient Fades */}
      <div className="ff-carousel3d-fade ff-carousel3d-fade-left" />
      <div className="ff-carousel3d-fade ff-carousel3d-fade-right" />

      {/* Carousel Track */}
      <div className="ff-carousel3d-viewport">
        <div
          ref={trackRef}
          className="ff-carousel3d-track"
          onMouseDown={handleMouseDown}
          onMouseMove={handleMouseMove}
          onMouseUp={handleMouseUp}
          onMouseLeave={handleMouseUp}
          onTouchStart={handleTouchStart}
          onTouchMove={handleTouchMove}
          onTouchEnd={handleTouchEnd}
          style={{ cursor: "grab" }}
        >
          {features.map((feature, index) => {
            const colors = COLOR_MAP[feature.color];
            const { rotateY, translateZ, scale, opacity, zIndex } = getCardTransform(index);
            const parallaxOffset = getContentParallax(index);

            return (
              <div
                key={feature.title}
                className="ff-carousel3d-card-wrapper"
                style={{
                  transform: `rotateY(${rotateY}deg) translateZ(${translateZ}px) scale(${scale})`,
                  opacity,
                  zIndex,
                  transition: isPaused ? "none" : "transform 0.05s linear, opacity 0.1s linear",
                }}
              >
                {/* Shadow reflection */}
                <div 
                  className="ff-carousel3d-card-shadow"
                  style={{
                    transform: `rotateY(${rotateY}deg) translateZ(${translateZ - 30}px) scale(${scale * 0.95})`,
                    opacity: opacity * 0.3,
                  }}
                />
                <Chamber ribs className="ff-carousel3d-card">
                  <CornerBrace position="tl" />
                  <CornerBrace position="br" />
                  <div
                    className="ff-carousel3d-card-content"
                    style={{ transform: `translateX(${parallaxOffset}px)` }}
                  >
                    <div
                      className="ff-carousel3d-icon"
                      style={{
                        background: colors.bg,
                        borderColor: colors.border,
                        boxShadow: `0 0 20px ${colors.glow}`,
                      }}
                    >
                      {feature.icon}
                    </div>
                    <h3 className="ff-carousel3d-title">{feature.title}</h3>
                    <p className="ff-carousel3d-desc">{feature.desc}</p>
                    <span
                      className="ff-carousel3d-tag"
                      style={{
                        background: colors.bg,
                        borderColor: colors.border,
                        color: colors.glow.replace("0.4)", "1)"),
                      }}
                    >
                      {feature.tag}
                    </span>
                  </div>
                </Chamber>
              </div>
            );
          })}
        </div>
      </div>

      {/* Progress Bar */}
      <div className="ff-carousel3d-progress">
        <div 
          className="ff-carousel3d-progress-bar"
          style={{ width: `${progress * 100}%` }}
        />
      </div>

      {/* Dots Indicator */}
      <div className="ff-carousel3d-dots">
        {features.map((_, i) => (
          <button
            key={i}
            className={`ff-carousel3d-dot ${i === activeIndex ? "active" : ""}`}
            onClick={() => goToIndex(i)}
            aria-label={`Go to card ${i + 1}`}
          />
        ))}
      </div>

      {/* Pause indicator */}
      {isPaused && (
        <div className="ff-carousel3d-pause-indicator">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="currentColor">
            <rect x="6" y="4" width="4" height="16" />
            <rect x="14" y="4" width="4" height="16" />
          </svg>
        </div>
      )}
    </div>
  );
};

export default FeatureCarousel3D;
