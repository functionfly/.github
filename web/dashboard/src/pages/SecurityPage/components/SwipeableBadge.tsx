import { useState, useRef, useEffect } from 'react';
import { ChevronLeft, ChevronRight } from 'lucide-react';
import { Button } from '@/components/ui/button';
import type { ComplianceFramework } from '../types';
import { RISK_LEVELS, getStatusRiskLevel } from '../utils/riskColors';

interface SwipeableBadgeProps {
  frameworks: ComplianceFramework[];
}

export function SwipeableBadge({ frameworks }: SwipeableBadgeProps) {
  const [currentIndex, setCurrentIndex] = useState(0);
  const [isDragging, setIsDragging] = useState(false);
  const [startX, setStartX] = useState(0);
  const containerRef = useRef<HTMLDivElement>(null);

  const nextSlide = () => {
    setCurrentIndex((prev) => (prev + 1) % frameworks.length);
  };

  const prevSlide = () => {
    setCurrentIndex((prev) => (prev - 1 + frameworks.length) % frameworks.length);
  };

  const handleTouchStart = (e: React.TouchEvent) => {
    setIsDragging(true);
    setStartX(e.touches[0].clientX);
  };

  const handleTouchMove = (e: React.TouchEvent) => {
    if (!isDragging) return;
    const currentX = e.touches[0].clientX;
    const diff = startX - currentX;

    if (Math.abs(diff) > 50) {
      if (diff > 0) {
        nextSlide();
      } else {
        prevSlide();
      }
      setIsDragging(false);
    }
  };

  const handleTouchEnd = () => {
    setIsDragging(false);
  };

  // Auto-advance every 5 seconds on mobile
  useEffect(() => {
    const interval = setInterval(() => {
      if (window.innerWidth < 768) { // Only on mobile
        nextSlide();
      }
    }, 5000);

    return () => clearInterval(interval);
  }, []);

  const currentFramework = frameworks[currentIndex];
  const riskLevel = getStatusRiskLevel(currentFramework.status);
  const riskColors = RISK_LEVELS[riskLevel];

  return (
    <div className="relative">
      {/* Main badge container */}
      <div
        ref={containerRef}
        className="md:hidden relative overflow-hidden rounded-lg border bg-card"
        onTouchStart={handleTouchStart}
        onTouchMove={handleTouchMove}
        onTouchEnd={handleTouchEnd}
      >
        <div
          className="flex transition-transform duration-300 ease-in-out"
          style={{ transform: `translateX(-${currentIndex * 100}%)` }}
        >
          {frameworks.map((framework, index) => {
            const itemRiskLevel = getStatusRiskLevel(framework.status);
            const itemRiskColors = RISK_LEVELS[itemRiskLevel];

            return (
              <div
                key={framework.name}
                className="w-full flex-shrink-0 p-4"
                style={{ minWidth: '100%' }}
              >
                <div className="flex items-center justify-between mb-2">
                  <h4 className="font-semibold">{framework.name}</h4>
                  <div
                    className="px-2 py-1 rounded-full text-xs font-medium"
                    style={{
                      backgroundColor: itemRiskColors.bgColor,
                      color: itemRiskColors.textColor,
                      border: `1px solid ${itemRiskColors.color}40`
                    }}
                  >
                    {framework.status}
                  </div>
                </div>
                <p className="text-sm text-muted-foreground mb-2">{framework.description}</p>
                <div className="text-xs text-muted-foreground space-y-1">
                  <div>Auditor: {framework.auditor}</div>
                  <div>Last Audit: {framework.lastAudit}</div>
                  <div>Next Audit: {framework.nextAudit}</div>
                </div>
              </div>
            );
          })}
        </div>
      </div>

      {/* Navigation dots */}
      <div className="md:hidden flex justify-center mt-3 space-x-2">
        {frameworks.map((_, index) => (
          <button
            key={index}
            onClick={() => setCurrentIndex(index)}
            className={`w-2 h-2 rounded-full transition-colors ${
              index === currentIndex ? 'bg-primary' : 'bg-muted'
            }`}
            aria-label={`Go to framework ${index + 1}`}
          />
        ))}
      </div>

      {/* Desktop grid view */}
      <div className="hidden md:grid md:grid-cols-2 gap-4">
        {frameworks.map((framework) => {
          const riskLevel = getStatusRiskLevel(framework.status);
          const riskColors = RISK_LEVELS[riskLevel];

          return (
            <div
              key={framework.name}
              className="border rounded-lg p-4"
              style={{ borderColor: riskColors.color + '20' }}
            >
              <div className="flex items-center justify-between mb-2">
                <h4 className="font-semibold">{framework.name}</h4>
                <div
                  className="px-2 py-1 rounded-full text-xs font-medium"
                  style={{
                    backgroundColor: riskColors.bgColor,
                    color: riskColors.textColor,
                    border: `1px solid ${riskColors.color}40`
                  }}
                >
                  {framework.status}
                </div>
              </div>
              <p className="text-sm text-muted-foreground mb-2">{framework.description}</p>
              <div className="text-xs text-muted-foreground space-y-1">
                <div>Auditor: {framework.auditor}</div>
                <div>Last Audit: {framework.lastAudit}</div>
                <div>Next Audit: {framework.nextAudit}</div>
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
}