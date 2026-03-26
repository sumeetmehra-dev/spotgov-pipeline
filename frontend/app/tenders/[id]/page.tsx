'use client';

import { useState, useEffect } from 'react';
import { useParams } from 'next/navigation';
import type { Tender } from '@/lib/api';
import { getTender } from '@/lib/api';

function formatCurrency(value: number | null, currency: string): string {
  if (value === null) return 'Not specified';
  return new Intl.NumberFormat('pt-PT', { style: 'currency', currency, maximumFractionDigits: 0 }).format(value);
}

function formatDate(date: string | null): string {
  if (!date) return 'Not specified';
  return new Date(date).toLocaleDateString('en-GB', {
    day: 'numeric',
    month: 'long',
    year: 'numeric',
  });
}

function deadlineBadge(deadline: string | null): { text: string; className: string } {
  if (!deadline) return { text: 'No deadline', className: 'badge-deadline-normal' };
  const days = Math.ceil((new Date(deadline).getTime() - Date.now()) / (1000 * 60 * 60 * 24));
  if (days < 0) return { text: 'Expired', className: 'badge bg-gray-100 text-gray-500' };
  if (days <= 7) return { text: `${days} days remaining`, className: 'badge-deadline-urgent' };
  if (days <= 21) return { text: `${days} days remaining`, className: 'badge-deadline-soon' };
  return { text: `${days} days remaining`, className: 'badge-deadline-normal' };
}

function InfoRow({ icon, label, children }: { icon: React.ReactNode; label: string; children: React.ReactNode }) {
  return (
    <div className="flex items-start gap-3 py-3">
      <div className="mt-0.5 text-gray-400">{icon}</div>
      <div>
        <dt className="text-xs font-medium text-gray-400 uppercase tracking-wider">{label}</dt>
        <dd className="text-sm font-medium text-gray-900 mt-0.5">{children}</dd>
      </div>
    </div>
  );
}

export default function TenderDetailPage() {
  const params = useParams();
  const [tender, setTender] = useState<Tender | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    if (params.id) {
      getTender(params.id as string)
        .then(setTender)
        .catch(console.error)
        .finally(() => setLoading(false));
    }
  }, [params.id]);

  if (loading) {
    return (
      <div className="flex flex-col items-center justify-center py-20 gap-3">
        <div className="w-8 h-8 border-3 border-gray-200 border-t-brand-600 rounded-full animate-spin" />
        <p className="text-sm text-gray-400">Loading tender details...</p>
      </div>
    );
  }

  if (!tender) {
    return (
      <div className="text-center py-20">
        <p className="text-lg font-semibold text-gray-900">Tender not found</p>
        <a href="/" className="text-sm text-brand-600 hover:text-brand-700 mt-2 inline-block">Back to tenders</a>
      </div>
    );
  }

  const dl = deadlineBadge(tender.deadline);
  const uniqueCpv = Array.from(new Set(tender.cpv_codes || []));

  return (
    <div className="animate-fade-in">
      {/* Back link */}
      <a
        href="/"
        className="inline-flex items-center gap-1.5 text-sm font-medium text-gray-500 hover:text-gray-700 mb-6 transition-colors"
      >
        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
          <polyline points="15 18 9 12 15 6" />
        </svg>
        Back to tenders
      </a>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Main content */}
        <div className="lg:col-span-2 space-y-6">
          {/* Title card */}
          <div className="card p-6">
            <div className="flex items-start justify-between gap-4 mb-4">
              <h1 className="text-xl font-bold text-gray-900 leading-snug">{tender.title}</h1>
              <span className={`badge ${tender.source === 'ted' ? 'badge-ted' : 'badge-dados'} shrink-0 uppercase`}>
                {tender.source}
              </span>
            </div>

            {tender.description && (
              <div className="bg-gray-50 rounded-xl p-4 mt-4">
                <h2 className="text-xs font-semibold text-gray-400 uppercase tracking-wider mb-2">Description</h2>
                <p className="text-sm text-gray-700 leading-relaxed whitespace-pre-wrap">{tender.description}</p>
              </div>
            )}
          </div>

          {/* CPV Codes */}
          {uniqueCpv.length > 0 && (
            <div className="card p-6">
              <h2 className="text-xs font-semibold text-gray-400 uppercase tracking-wider mb-3">CPV Classification Codes</h2>
              <div className="flex flex-wrap gap-2">
                {uniqueCpv.map((code) => (
                  <span
                    key={code}
                    className="inline-flex items-center px-3 py-1.5 rounded-lg text-sm font-medium bg-brand-50 text-brand-700 ring-1 ring-inset ring-brand-100"
                  >
                    {code}
                  </span>
                ))}
              </div>
            </div>
          )}
        </div>

        {/* Sidebar */}
        <div className="space-y-6">
          {/* Key details card */}
          <div className="card p-6 divide-y divide-gray-50">
            <InfoRow
              label="Estimated Value"
              icon={
                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                  <line x1="12" y1="1" x2="12" y2="23" /><path d="M17 5H9.5a3.5 3.5 0 0 0 0 7h5a3.5 3.5 0 0 1 0 7H6" />
                </svg>
              }
            >
              <span className="text-lg font-bold text-gray-900">
                {formatCurrency(tender.estimated_value, tender.currency)}
              </span>
            </InfoRow>

            <InfoRow
              label="Submission Deadline"
              icon={
                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                  <circle cx="12" cy="12" r="10" /><polyline points="12 6 12 12 16 14" />
                </svg>
              }
            >
              <div className="flex flex-col gap-1">
                <span>{formatDate(tender.deadline)}</span>
                <span className={`badge ${dl.className} w-fit`}>{dl.text}</span>
              </div>
            </InfoRow>

            <InfoRow
              label="Published"
              icon={
                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                  <rect x="3" y="4" width="18" height="18" rx="2" ry="2" /><line x1="16" y1="2" x2="16" y2="6" /><line x1="8" y1="2" x2="8" y2="6" /><line x1="3" y1="10" x2="21" y2="10" />
                </svg>
              }
            >
              {formatDate(tender.publication_date)}
            </InfoRow>

            <InfoRow
              label="Buyer"
              icon={
                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                  <path d="M3 9l9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z" /><polyline points="9 22 9 12 15 12 15 22" />
                </svg>
              }
            >
              {tender.buyer_name || 'Not specified'}
            </InfoRow>

            <InfoRow
              label="Country"
              icon={
                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                  <circle cx="12" cy="12" r="10" /><line x1="2" y1="12" x2="22" y2="12" /><path d="M12 2a15.3 15.3 0 0 1 4 10 15.3 15.3 0 0 1-4 10 15.3 15.3 0 0 1-4-10 15.3 15.3 0 0 1 4-10z" />
                </svg>
              }
            >
              {tender.buyer_country || 'Not specified'}
            </InfoRow>

            <InfoRow
              label="Procurement Method"
              icon={
                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                  <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z" /><polyline points="14 2 14 8 20 8" />
                </svg>
              }
            >
              <span className="capitalize">{tender.procurement_method || 'Not specified'}</span>
            </InfoRow>

            <InfoRow
              label="Source ID"
              icon={
                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                  <polyline points="16 18 22 12 16 6" /><polyline points="8 6 2 12 8 18" />
                </svg>
              }
            >
              <span className="font-mono text-xs">{tender.source_id}</span>
            </InfoRow>
          </div>

          {/* View original */}
          {tender.source === 'ted' && (
            <a
              href={`https://ted.europa.eu/en/notice/-/detail/${tender.source_id}`}
              target="_blank"
              rel="noopener noreferrer"
              className="btn-secondary w-full flex items-center justify-center gap-2"
            >
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                <path d="M18 13v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h6" /><polyline points="15 3 21 3 21 9" /><line x1="10" y1="14" x2="21" y2="3" />
              </svg>
              View on TED
            </a>
          )}
        </div>
      </div>
    </div>
  );
}
