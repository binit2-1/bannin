import { Star } from 'kubai-icons-react';
import { useEffect, useState } from 'react';
import { Tooltip, TooltipGroup } from './interior/tooltip-group';

// One React component, so both tooltips share the group's delay.
export function HeaderActions({ githubUrl }: { githubUrl: string }) {
  return (
    <TooltipGroup className="flex items-center gap-2">
      <ThemeToggle />

      <Tooltip label="Star on GitHub" side="bottom">
        <a
          href={githubUrl}
          className="group flex items-center gap-1.5 border border-neutral-200 bg-white py-1 pr-3 pl-2 text-sm dark:border-neutral-800 dark:bg-[#111113]"
        >
          <span className="relative size-6 overflow-hidden" aria-hidden="true">
            <span className="absolute inset-0 flex items-center justify-center transition-transform duration-300 ease-in-out group-hover:-translate-y-full group-focus-visible:-translate-y-full">
              <GithubIcon />
            </span>
            <span className="absolute inset-0 flex translate-y-full items-center justify-center transition-transform duration-300 ease-in-out group-hover:translate-y-0 group-focus-visible:translate-y-0">
              <Star size={24} color="#0015FF" variant="soft" trigger="none" />
            </span>
          </span>
          <span>GitHub</span>
        </a>
      </Tooltip>
    </TooltipGroup>
  );
}

function ThemeToggle() {
  const [dark, setDark] = useState(false);

  // The saved theme is applied in Layout.astro before React loads.
  useEffect(() => {
    setDark(document.documentElement.classList.contains('dark'));
  }, []);

  function toggle() {
    const next = document.documentElement.classList.toggle('dark');
    setDark(next);
    try {
      localStorage.setItem('theme', next ? 'dark' : 'light');
    } catch {
      // Storage can be blocked; the theme still changes for this visit.
    }
  }

  return (
    <Tooltip
      side="bottom"
      label={
        <>
          <span className="dark:hidden">Dark theme</span>
          <span className="hidden dark:inline">Light theme</span>
        </>
      }
    >
      <button
        type="button"
        onClick={toggle}
        aria-label="Dark theme"
        aria-pressed={dark}
        className="flex size-8.5 cursor-pointer items-center justify-center"
      >
        <span className="size-4 bg-[#111113] dark:bg-[#FAFAFA]" aria-hidden="true" />
      </button>
    </Tooltip>
  );
}

function GithubIcon() {
  return (
    <svg
      className="size-4"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
    >
      <path d="M15 22v-4a4.8 4.8 0 0 0-1-3.5c3 0 6-2 6-5.5.08-1.25-.27-2.48-1-3.5.28-1.15.28-2.35 0-3.5 0 0-1 0-3 1.5-2.64-.5-5.36-.5-8 0C6 2 5 2 5 2c-.3 1.15-.3 2.35 0 3.5A5.4 5.4 0 0 0 4 9c0 3.5 3 5.5 6 5.5-.39.49-.68 1.05-.85 1.65-.17.6-.22 1.23-.15 1.85v4" />
      <path d="M9 18c-4.51 2-5-2-7-2" />
    </svg>
  );
}
