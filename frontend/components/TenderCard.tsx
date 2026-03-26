'use client';

import type { Tender } from '@/lib/api';

function formatCurrency(value: number | null, currency: string): string {
  if (value === null) return '-';
  return new Intl.NumberFormat('pt-PT', { style: 'currency', currency, maximumFractionDigits: 0 }).format(value);
}

function formatDate(date: string | null): string {
  if (!date) return '-';
  return new Date(date).toLocaleDateString('en-GB', {
    day: 'numeric',
    month: 'short',
    year: 'numeric',
  });
}

function deadlineStatus(deadline: string | null): { label: string; class: string } {
  if (!deadline) return { label: 'No deadline', class: 'badge-deadline-normal' };
  const days = Math.ceil((new Date(deadline).getTime() - Date.now()) / (1000 * 60 * 60 * 24));
  if (days < 0) return { label: 'Expired', class: 'badge bg-gray-100 text-gray-500 ring-1 ring-inset ring-gray-200' };
  if (days <= 7) return { label: `${days}d left`, class: 'badge-deadline-urgent' };
  if (days <= 21) return { label: `${days}d left`, class: 'badge-deadline-soon' };
  return { label: formatDate(deadline), class: 'badge-deadline-normal' };
}

function sourceBadgeClass(source: string): string {
  return source === 'ted' ? 'badge-ted' : 'badge-dados';
}

export default function TenderCard({ tender }: { tender: Tender }) {
  const dl = deadlineStatus(tender.deadline);
  const uniqueCpv = Array.from(new Set(tender.cpv_codes || []));

  return (
    <a
      href={`/tenders/${tender.id}`}
      className="card group block p-5 hover:shadow-card-hover hover:ring-brand-200 transition-all duration-200 animate-fade-in"
    >
      <div className="flex items-start justify-between gap-3 mb-3">
        <h3 className="text-[15px] font-semibold text-gray-900 leading-snug line-clamp-2 group-hover:text-brand-700 transition-colors">
          {tender.title}
        </h3>
        <span className={`badge ${sourceBadgeClass(tender.source)} shrink-0 uppercase`}>
          {tender.source}
        </span>
      </div>

      {tender.description && (
        <p className="text-sm text-gray-500 mb-4 line-clamp-2 leading-relaxed">
          {tender.description}
        </p>
      )}

      <div className="flex flex-wrap items-center gap-x-5 gap-y-2 text-[13px]">
        {tender.estimated_value !== null && (
          <div className="flex items-center gap-1.5 text-gray-900 font-semibold">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" className="text-gray-400">
              <line x1="12" y1="1" x2="12" y2="23" />
              <path d="M17 5H9.5a3.5 3.5 0 0 0 0 7h5a3.5 3.5 0 0 1 0 7H6" />
            </svg>
            {formatCurrency(tender.estimated_value, tender.currency)}
          </div>
        )}

        <div className="flex items-center gap-1.5 text-gray-500">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" className="text-gray-400">
            <rect x="3" y="4" width="18" height="18" rx="2" ry="2" />
            <line x1="16" y1="2" x2="16" y2="6" />
            <line x1="8" y1="2" x2="8" y2="6" />
            <line x1="3" y1="10" x2="21" y2="10" />
          </svg>
          <span className={`badge ${dl.class}`}>{dl.label}</span>
        </div>

        {tender.buyer_name && (
          <div className="flex items-center gap-1.5 text-gray-500 truncate max-w-[250px]">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" className="text-gray-400 shrink-0">
              <path d="M3 9l9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z" />
              <polyline points="9 22 9 12 15 12 15 22" />
            </svg>
            <span className="truncate">{tender.buyer_name}</span>
          </div>
        )}

        {tender.buyer_country && (
          <span className="badge bg-gray-50 text-gray-600 ring-1 ring-inset ring-gray-200">
            {tender.buyer_country}
          </span>
        )}
      </div>

      {uniqueCpv.length > 0 && (
        <div className="mt-3 pt-3 border-t border-gray-50 flex flex-wrap gap-1.5">
          {uniqueCpv.slice(0, 4).map((code) => (
            <span
              key={code}
              className="inline-flex items-center px-2 py-0.5 rounded-md text-[11px] font-medium bg-gray-50 text-gray-500 ring-1 ring-inset ring-gray-100"
            >
              {code}
            </span>
          ))}
          {uniqueCpv.length > 4 && (
            <span className="text-[11px] text-gray-400 self-center">+{uniqueCpv.length - 4} more</span>
          )}
        </div>
      )}
    </a>
  );
}
