import { ExclamationTriangleIcon } from '@heroicons/react/24/outline';

interface ErrorStateProps {
  message?: string;
  retry?: () => void;
}

export function ErrorState({ message = 'An error occurred', retry }: ErrorStateProps) {
  return (
    <div className="flex min-h-[400px] flex-col items-center justify-center gap-4 text-center">
      <div className="bg-ember-500/10 flex size-16 items-center justify-center rounded-full">
        <ExclamationTriangleIcon className="text-ember-400 size-8" />
      </div>
      <div>
        <h3 className="font-display text-text-bright text-lg font-semibold">Something went wrong</h3>
        <p className="text-text-secondary mt-1 text-sm/6">{message}</p>
      </div>
      {retry && (
        <button
          onClick={retry}
          className="from-ember-600 to-ember-700 hover:from-ember-500 hover:to-ember-600 rounded-lg bg-linear-to-r px-5 py-2.5 text-sm/6 font-medium text-white shadow-lg transition-all duration-200 hover:shadow-xl"
        >
          Try Again
        </button>
      )}
    </div>
  );
}
