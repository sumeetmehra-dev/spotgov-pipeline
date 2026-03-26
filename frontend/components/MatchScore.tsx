'use client';

interface MatchScoreProps {
  score: number;
  vectorScore?: number;
  cpvOverlap?: number;
  showBreakdown?: boolean;
}

function scoreColor(score: number): string {
  if (score >= 0.8) return 'text-green-700 bg-green-100';
  if (score >= 0.5) return 'text-yellow-700 bg-yellow-100';
  return 'text-red-700 bg-red-100';
}

export default function MatchScore({ score, vectorScore, cpvOverlap, showBreakdown }: MatchScoreProps) {
  const pct = (score * 100).toFixed(1);

  return (
    <div className="flex items-center gap-3">
      <span className={`inline-flex items-center px-3 py-1 rounded-full text-sm font-bold ${scoreColor(score)}`}>
        {pct}%
      </span>
      {showBreakdown && (
        <div className="flex gap-2 text-xs text-gray-500">
          {vectorScore !== undefined && <span>Vector: {(vectorScore * 100).toFixed(0)}%</span>}
          {cpvOverlap !== undefined && <span>CPV: {(cpvOverlap * 100).toFixed(0)}%</span>}
        </div>
      )}
    </div>
  );
}
