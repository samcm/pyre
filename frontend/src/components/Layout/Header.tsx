import { Link } from '@tanstack/react-router';

export function Header() {
  return (
    <header className="border-border-subtle bg-bg-deep relative border-b">
      {/* Subtle ember glow at top */}
      <div className="via-ember-500/30 pointer-events-none absolute inset-x-0 top-0 h-px bg-linear-to-r from-transparent to-transparent" />

      <div className="mx-auto max-w-7xl px-6 py-5 lg:px-8">
        <div className="flex items-center justify-between">
          <Link to="/" className="group flex items-center gap-3">
            {/* Fire icon */}
            <div className="relative">
              <div className="from-ember-500 to-ember-700 group-hover:shadow-ember-500/30 flex size-10 items-center justify-center rounded-lg bg-linear-to-br shadow-lg transition-all duration-300 group-hover:shadow-xl">
                <svg className="size-6 text-white" fill="currentColor" viewBox="0 0 24 24">
                  <path d="M12 23c-3.6 0-6.5-2.9-6.5-6.5 0-1.4.4-2.6 1.2-3.7.7-1 1.6-1.8 2.4-2.6.5-.5.9-1 1.2-1.5.2-.3.3-.7.3-1.1 0-.6-.2-1.1-.5-1.6-.2-.3-.3-.6-.3-1 0-.8.6-1.5 1.4-1.5.4 0 .7.2 1 .5.9 1.1 1.8 2.4 2.4 3.8.5 1.2.7 2.4.7 3.7 0 .9-.1 1.7-.4 2.5-.2.6-.5 1.1-.9 1.6-.4.5-.9.9-1.4 1.3-.5.4-1 .8-1.4 1.2-.3.3-.5.7-.5 1.1 0 .8.6 1.4 1.4 1.4.5 0 1-.3 1.3-.7.2-.3.5-.6.7-.9.3-.4.7-.7 1.2-.7.7 0 1.3.6 1.3 1.3 0 .4-.2.8-.5 1.2-.6.8-1.5 1.4-2.5 1.8-1 .4-2.1.6-3.1.6z" />
                </svg>
              </div>
              {/* Animated ember particles */}
              <div className="bg-ember-400/60 absolute -top-1 left-1/2 size-1.5 -translate-x-1/2 animate-[ember-flicker_2s_ease-in-out_infinite] rounded-full" />
              <div className="bg-ember-300/40 absolute -top-0.5 left-1/3 size-1 animate-[ember-flicker_3s_ease-in-out_infinite_0.5s] rounded-full" />
            </div>

            {/* Logo text */}
            <div className="flex flex-col">
              <span className="font-display text-text-bright text-2xl font-bold tracking-tight">PYRE</span>
              <span className="text-text-muted text-xs font-medium tracking-widest">WATCH IT BURN</span>
            </div>
          </Link>

          {/* Right side - could add navigation or actions here */}
          <div className="flex items-center gap-4">
            <a
              href="https://polymarket.com"
              target="_blank"
              rel="noopener noreferrer"
              className="border-border-subtle bg-bg-card text-text-secondary hover:border-ember-500/30 hover:bg-bg-elevated hover:text-text-primary flex items-center gap-2 rounded-lg border px-4 py-2 text-sm/6 font-medium transition-all duration-200"
            >
              <svg className="size-4" viewBox="0 0 24 24" fill="currentColor">
                <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm-1 17.93c-3.95-.49-7-3.85-7-7.93 0-.62.08-1.21.21-1.79L9 15v1c0 1.1.9 2 2 2v1.93zm6.9-2.54c-.26-.81-1-1.39-1.9-1.39h-1v-3c0-.55-.45-1-1-1H8v-2h2c.55 0 1-.45 1-1V7h2c1.1 0 2-.9 2-2v-.41c2.93 1.19 5 4.06 5 7.41 0 2.08-.8 3.97-2.1 5.39z" />
              </svg>
              Polymarket
            </a>
          </div>
        </div>
      </div>
    </header>
  );
}
