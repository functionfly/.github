interface AnimatedBackgroundGridProps {
  children: React.ReactNode;
}

export function AnimatedBackgroundGrid({ children }: AnimatedBackgroundGridProps) {
  return (
    <section className="relative py-20 lg:py-32 overflow-hidden hero-section-enhanced mesh-gradient-bg">
      <div className="container mx-auto px-4 relative">
        {children}
      </div>
    </section>
  );
}