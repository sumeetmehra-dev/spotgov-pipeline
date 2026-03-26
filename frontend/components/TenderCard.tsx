'use client';

import type { Tender } from '@/lib/api';

function formatCurrency(value: number | null, currency: string): string {
  if (value === null) return 'N/A';
  return new Intl.NumberFormat('en-EU', { style: 'currency', currency }).format(value);
}

function formatDate(date: string | null): string {
  if (!date) return 'N/A';
  return new Date(date).toLocaleDateString('en-GB', {
    day: 'numeric',
    month: 'short',
    year: 'numeric',
  });
}

export default function TenderCard({ tender }: { tender: Tender }) {
  return (
    <a
      href={`/tenders/${tender.id}`}
      className="block bg-white rounded-lg border border-gray-200 p-5 hover:border-brand-500 hover:shadow-sm transition-all"
    >
      <div className="flex justify-between items-start mb-3">
        <h3 className="text-lg font-semibold text-gray-900 line-clamp-2 flex-1 mr-4">
          {tender.title}
        </h3>
        <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-blue-100 text-blue-800 whitespace-nowrap">
          {tender.source.toUpperCase()}
        </span>
      </div>

      {tender.description && (
        <p className="text-sm text-gray-600 mb-3 line-clamp-2">{tender.description}</p>
      )}

      <div className="flex flex-wrap gap-4 text-sm text-gray-500">
        <div>
          <span className="font-medium">Value:</span>{' '}
          {formatCurrency(tender.estimated_value, tender.currency)}
        </div>
        <div>
          <span className="font-medium">Deadline:</span> {formatDate(tender.deadline)}
        </div>
        <div>
          <span className="font-medium">Buyer:</span> {tender.buyer_name || 'N/A'}
        </div>
        {tender.buyer_country && (
          <div>
            <span className="font-medium">Country:</span> {tender.buyer_country}
          </div>
        )}
      </div>

      {tender.cpv_codes && tender.cpv_codes.length > 0 && (
        <div className="mt-3 flex flex-wrap gap-1">
          {tender.cpv_codes.slice(0, 5).map((code) => (
            <span
              key={code}
              className="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-gray-100 text-gray-700"
            >
              {code}
            </span>
          ))}
          {tender.cpv_codes.length > 5 && (
            <span className="text-xs text-gray-400">+{tender.cpv_codes.length - 5} more</span>
          )}
        </div>
      )}
    </a>
  );
}
