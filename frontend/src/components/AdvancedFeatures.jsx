import {
  BadgeCheck,
  CalendarClock,
  CheckCircle2,
  ClipboardList,
  Palette,
  Puzzle,
  ShieldCheck,
  SlidersHorizontal,
  Sparkles,
  TimerReset,
  UsersRound,
  X,
} from 'lucide-react';
import { useEffect, useRef, useState } from 'react';

const features = [
  { icon: CalendarClock, title: 'Poll Scheduling', description: 'Schedule polls to start at a specific date and time.' },
  { icon: TimerReset, title: 'Auto-Close Polls', description: 'Automatically close polls after a time or vote limit.' },
  { icon: ShieldCheck, title: 'Anonymous / Authenticated Voting', description: 'Choose how participants can vote.' },
  { icon: SlidersHorizontal, title: 'Multiple Question Types', description: 'Support formats beyond basic single-choice polls.' },
  { icon: Puzzle, title: 'Quiz Mode', description: 'Turn polls into interactive quizzes.' },
  { icon: CheckCircle2, title: 'Correct Answer Support', description: 'Define the correct answer for quiz questions.' },
  { icon: UsersRound, title: 'Team / Workspace Dashboard', description: 'Manage polls collaboratively with teams.' },
  { icon: ShieldCheck, title: 'Role-Based Access', description: 'Set permissions for owners, admins, editors, and viewers.' },
  { icon: ClipboardList, title: 'Poll Templates', description: 'Create new polls from reusable templates.' },
  { icon: Palette, title: 'Custom Branding', description: 'Customize poll colors, logo, and branding.' },
  { icon: BadgeCheck, title: 'Public Results Toggle', description: 'Choose whether participants can see results.' },
];

export const AdvancedFeatures = () => {
  const [selectedFeature, setSelectedFeature] = useState(null);
  const closeButtonRef = useRef(null);
  const previouslyFocusedRef = useRef(null);
  const SelectedIcon = selectedFeature?.icon;

  useEffect(() => {
    if (!selectedFeature) return undefined;

    previouslyFocusedRef.current = document.activeElement;
    closeButtonRef.current?.focus();
    const handleKeyDown = (event) => {
      if (event.key === 'Escape') setSelectedFeature(null);
    };
    document.addEventListener('keydown', handleKeyDown);
    return () => {
      document.removeEventListener('keydown', handleKeyDown);
      previouslyFocusedRef.current?.focus?.();
    };
  }, [selectedFeature]);

  return (
    <section className="advanced-features-section" aria-labelledby="advanced-features-title">
      <div className="advanced-features-header">
        <div>
          <p className="analytics-eyebrow">Roadmap</p>
          <h2 id="advanced-features-title">Advanced Features</h2>
          <p>Powerful features planned for future versions.</p>
        </div>
        <Sparkles size={20} aria-hidden="true" />
      </div>
      <div className="advanced-features-grid">
        {features.map((feature) => {
          const Icon = feature.icon;
          return (
            <button
              type="button"
              className="advanced-feature-card"
              key={feature.title}
              onClick={() => setSelectedFeature(feature)}
            >
              <Icon size={18} aria-hidden="true" />
              <span>
                <strong>{feature.title}</strong>
                <span>{feature.description}</span>
                <small>Coming Soon</small>
              </span>
            </button>
          );
        })}
      </div>

      {selectedFeature && (
        <div className="advanced-feature-dialog-backdrop" role="presentation" onMouseDown={(event) => {
          if (event.target === event.currentTarget) setSelectedFeature(null);
        }}>
          <div
            className="advanced-feature-dialog"
            role="dialog"
            aria-modal="true"
            aria-labelledby="advanced-feature-dialog-title"
            aria-describedby="advanced-feature-dialog-description"
          >
            <button
              ref={closeButtonRef}
              type="button"
              className="advanced-feature-dialog-close"
              aria-label="Close feature details"
              onClick={() => setSelectedFeature(null)}
            >
              <X size={18} />
            </button>
            <SelectedIcon size={22} className="advanced-feature-dialog-icon" aria-hidden="true" />
            <h3 id="advanced-feature-dialog-title">{selectedFeature.title}</h3>
            <p id="advanced-feature-dialog-description">{selectedFeature.description}</p>
            <span className="advanced-feature-badge">Coming Soon</span>
            <p className="advanced-feature-dialog-note">This feature is planned for a future release.</p>
            <button type="button" className="btn btn-primary btn-sm" onClick={() => setSelectedFeature(null)}>
              Close
            </button>
          </div>
        </div>
      )}
    </section>
  );
};
