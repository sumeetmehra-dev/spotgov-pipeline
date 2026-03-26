'use client';

interface MatchScoreProps {
  score: number;
  vectorScore?: number;
  cpvOverlap?: number;
  showBreakdown?: boolean;
}

function scoreColor(score: number): string {
  if (score >= 0.7) return 'from-emerald-500 to-emerald-600';
  if (score >= 0.5) return 'from-amber-500 to-amber-600';
  return 'from-gray-400 to-gray-500';
}

function scoreRingColor(score: number): string {
  if (score >= 0.7) return 'ring-emerald-200';
  if (score >= 0.5) return 'ring-amber-200';
  return 'ring-gray-200';
}

function scoreLabelColor(score: number): string {
  if (score >= 0.7) return 'text-emerald-700';
  if (score >= 0.5) return 'text-amber-700';
  return 'text-gray-600';
}

export default function MatchScore({ score, vectorScore, cpvOverlap, showBreakdown }: MatchScoreProps) {
  const pct = Math.round(score * 100);

  return (
    <div className="flex items-center gap-4">
      {/* Circular score indicator */}
      <div className="relative flex items-center justify-center w-14 h-14">
        <svg className="w-14 h-14 -rotate-90" viewBox="0 0 56 56">
          <circle cx="28" cy="28" r="24" fill="none" stroke="#f3f4f6" strokeWidth="4" />
          <circle
            cx="28" cy="28" r="24" fill="none"
            className={`stroke-current ${scoreLabelColor(score)}`}
            strokeWidth="4"
            strokeLinecap="round"
            strokeDasharray={`${pct * 1.508} 150.8`}
          />
        </svg>
        <span className={`absolute text-sm font-bold ${scoreLabelColor(score)}`}>{pct}%</span>
      </div>

      {showBreakdown && (
        <div className="flex flex-col gap-1.5">
          {vectorScore !== undefined && (
            <div className="flex items-center gap-2">
              <div className="w-20 h-1.5 rounded-full bg-gray-100 overflow-hidden">
                <div
                  className="h-full rounded-full bg-brand-500"
                  style={{ width: `${Math.round(vectorScore * 100)}%` }}
                />
              </div>
              <span className="text-[11px] text-gray-500 w-28">Semantic {Math.round(vectorScore * 100)}%</span>
            </div>
          )}
          {cpvOverlap !== undefined && (
            <div className="flex items-center gap-2">
              <div className="w-20 h-1.5 rounded-full bg-gray-100 overflow-hidden">
                <div
                  className="h-full rounded-full bg-emerald-500"
                  style={{ width: `${Math.round(cpvOverlap * 100)}%` }}
                />
              </div>
              <span className="text-[11px] text-gray-500 w-28">CPV Match {Math.round(cpvOverlap * 100)}%</span>
            </div>
          )}
        </div>
      )}
    </div>
  );
}
