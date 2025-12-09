import { createFileRoute, Link } from '@tanstack/react-router';
import { useState, useEffect } from 'react';
import { LeaderboardTable } from '@/components/Leaderboard/LeaderboardTable';
import { CombinedPnlChart } from '@/components/Dashboard/CombinedPnlChart';
import { RecentTradesTable } from '@/components/Dashboard/RecentTradesTable';
import { usePersonas, type PersonaSummary } from '@/hooks/usePersonas';
import { FireIcon } from '@heroicons/react/24/solid';

export const Route = createFileRoute('/')({
  component: HomePage,
});

function HeroSection() {
  const { data: personas } = usePersonas();
  const [currentIndex, setCurrentIndex] = useState(0);
  const [isAnimating, setIsAnimating] = useState(false);

  useEffect(() => {
    if (!personas || personas.length <= 1) return;

    const interval = setInterval(() => {
      setIsAnimating(true);
      setTimeout(() => {
        setCurrentIndex(prev => (prev + 1) % personas.length);
        setIsAnimating(false);
      }, 400);
    }, 4000);

    return () => clearInterval(interval);
  }, [personas]);

  const currentPersona = personas?.[currentIndex];
  const displayName = currentPersona?.displayName ?? 'Your';

  return (
    <section className="relative -mx-6 -mt-8 overflow-hidden px-6 py-12 lg:-mx-8 lg:px-8 lg:py-16">
      {/* Background effects */}
      <div className="from-bg-void via-bg-deep pointer-events-none absolute inset-0 bg-linear-to-b to-transparent" />
      <div className="bg-ember-500/8 pointer-events-none absolute -top-40 left-1/2 h-96 w-[800px] -translate-x-1/2 rounded-full blur-[120px]" />

      <div className="relative">
        <div className="mx-auto max-w-4xl">
          {/* Main hero content */}
          <div className="flex flex-col items-center gap-8 lg:flex-row lg:gap-12">
            {/* Avatar section */}
            <div className="relative flex shrink-0 items-center justify-center">
              {/* Rotating ember ring */}
              <div className="animate-spin-slow from-ember-500/40 via-ember-600/20 to-ember-500/40 absolute size-36 rounded-full bg-conic lg:size-44" />
              <div className="from-ember-400/30 via-ember-500/10 absolute size-32 animate-pulse rounded-full bg-radial to-transparent lg:size-40" />

              {/* Avatar stack */}
              <div className="relative size-28 lg:size-36">
                {personas && personas.length > 0 ? (
                  <>
                    {/* Background avatars (stacked) */}
                    {personas.map((persona, index) => {
                      const isActive = index === currentIndex;
                      const offset = (index - currentIndex + personas.length) % personas.length;
                      const scale = offset === 0 ? 1 : 0.7 - offset * 0.1;
                      const zIndex = personas.length - offset;
                      const translateX = offset === 0 ? 0 : offset * 20;
                      const opacity = offset === 0 ? 1 : 0.4 - offset * 0.15;

                      return (
                        <Link
                          key={persona.slug}
                          to="/people/$slug"
                          params={{ slug: persona.slug }}
                          className="absolute inset-0 transition-all duration-500 ease-out"
                          style={{
                            transform: `translateX(${translateX}px) scale(${scale})`,
                            zIndex,
                            opacity: Math.max(0, opacity),
                          }}
                        >
                          <PersonaAvatar persona={persona} isActive={isActive} isAnimating={isAnimating && isActive} />
                        </Link>
                      );
                    })}
                  </>
                ) : (
                  <div className="from-ember-500 to-ember-700 flex size-full items-center justify-center rounded-full bg-linear-to-br">
                    <FireIcon className="size-12 text-white lg:size-16" />
                  </div>
                )}
              </div>
            </div>

            {/* Text content */}
            <div className="text-center lg:text-left">
              {/* Tag line */}
              <div className="mb-4 flex items-center justify-center gap-3 lg:justify-start">
                <div className="to-ember-500/50 h-px w-8 bg-linear-to-r from-transparent" />
                <span className="text-ember-400 text-xs font-semibold tracking-[0.25em] uppercase">
                  Polymarket Tracker
                </span>
                <div className="to-ember-500/50 h-px w-8 bg-linear-to-l from-transparent" />
              </div>

              {/* Main headline */}
              <h1 className="font-display text-text-bright text-4xl font-bold tracking-tight sm:text-5xl lg:text-6xl">
                <span className="block">Watch</span>
                <span className="mt-1 block">
                  {currentPersona ? (
                    <Link
                      to="/people/$slug"
                      params={{ slug: currentPersona.slug }}
                      className="group relative inline-block"
                    >
                      <span
                        className={`from-ember-400 via-ember-500 to-ember-600 inline-block bg-linear-to-r bg-clip-text text-transparent transition-all duration-400 ${
                          isAnimating ? 'translate-y-3 scale-95 opacity-0' : 'translate-y-0 scale-100 opacity-100'
                        }`}
                      >
                        {displayName}&apos;s
                      </span>
                      <span className="from-ember-500/0 via-ember-500/60 to-ember-500/0 absolute -bottom-1 left-0 h-0.5 w-full scale-x-0 rounded-full bg-linear-to-r transition-transform duration-300 group-hover:scale-x-100" />
                    </Link>
                  ) : (
                    <span className="from-ember-400 via-ember-500 to-ember-600 bg-linear-to-r bg-clip-text text-transparent">
                      Your
                    </span>
                  )}{' '}
                  Money{' '}
                  <span className="relative inline-block">
                    <span className="from-ember-400 via-ember-500 to-ember-600 bg-linear-to-r bg-clip-text text-transparent">
                      Burn
                    </span>
                    <span className="from-ember-500/0 via-ember-500/40 to-ember-500/0 absolute -bottom-1 left-0 h-1 w-full rounded-full bg-linear-to-r" />
                  </span>
                </span>
              </h1>

              {/* Subtitle */}
              <p className="text-text-secondary mt-5 max-w-lg text-base/7 lg:text-lg/8">
                Track Polymarket positions and PnL in real-time. See who&apos;s winning, who&apos;s losing, and watch
                the leaderboard shift with every bet.
              </p>

              {/* Persona indicators */}
              {personas && personas.length > 1 && (
                <div className="mt-6 flex items-center justify-center gap-2 lg:justify-start">
                  {personas.map((persona, index) => (
                    <button
                      key={persona.slug}
                      onClick={() => {
                        setIsAnimating(true);
                        setTimeout(() => {
                          setCurrentIndex(index);
                          setIsAnimating(false);
                        }, 400);
                      }}
                      className={`size-2 rounded-full transition-all duration-300 ${
                        index === currentIndex ? 'bg-ember-500 scale-125' : 'bg-text-muted/30 hover:bg-text-muted/50'
                      }`}
                      aria-label={`Show ${persona.displayName}`}
                    />
                  ))}
                </div>
              )}
            </div>
          </div>
        </div>
      </div>
    </section>
  );
}

function PersonaAvatar({
  persona,
  isActive,
  isAnimating,
}: {
  persona: PersonaSummary;
  isActive: boolean;
  isAnimating: boolean;
}) {
  return (
    <div
      className={`relative size-full overflow-hidden rounded-full transition-all duration-400 ${
        isActive ? 'ring-ember-500/60 shadow-ember-500/30 shadow-2xl ring-4' : 'ring-bg-elevated/50 ring-2'
      } ${isAnimating ? 'scale-90' : 'scale-100'}`}
    >
      {persona.image ? (
        <img src={persona.image} alt={persona.displayName} className="size-full object-cover" />
      ) : (
        <div className="from-ember-500 to-ember-700 flex size-full items-center justify-center bg-linear-to-br">
          <span className="font-display text-2xl font-bold text-white lg:text-3xl">
            {persona.displayName.charAt(0)}
          </span>
        </div>
      )}

      {/* Fire overlay effect for active */}
      {isActive && (
        <div className="from-ember-500/0 via-ember-500/10 to-ember-600/30 pointer-events-none absolute inset-0 bg-linear-to-t" />
      )}
    </div>
  );
}

function HomePage() {
  return (
    <div className="space-y-12">
      {/* Hero Section */}
      <HeroSection />

      {/* Leaderboard Section */}
      <section>
        <div className="mb-6 flex items-end justify-between">
          <div>
            <h2 className="font-display text-text-bright text-2xl font-bold">Leaderboard</h2>
            <p className="text-text-muted mt-1 text-sm/6">Rankings by total PnL across all tracked positions</p>
          </div>
        </div>
        <LeaderboardTable />
      </section>

      {/* PnL Chart Section */}
      <section>
        <CombinedPnlChart />
      </section>

      {/* Recent Trades Section */}
      <section>
        <RecentTradesTable />
      </section>
    </div>
  );
}
