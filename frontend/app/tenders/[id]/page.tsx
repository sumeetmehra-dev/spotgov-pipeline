'use client';

import { useState, useEffect } from 'react';
import { useParams } from 'next/navigation';
import type { Tender } from '@/lib/api';
import { getTender } from '@/lib/api';

function formatCurrency(value: number | null, currency: string): string {
  if (value === null) return 'Not specified';
  return new Intl.NumberFormat('en-EU', { style: 'currency', currency }).format(value);
}

function formatDate(date: string | null): string {
  if (!date) return 'Not specified';
  return new Date(date).toLocaleDateString('en-GB', {
    day: 'numeric',
    month: 'long',
    year: 'numeric',
  });
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
      <div className="flex justify-center py-12">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-brand-600" />
      </div>
    );
  }

  if (!tender) {
    return <p className="text-center py-12 text-gray-500">Tender not found</p>;
  }

  return (
    <div>
      <a href="/" className="text-sm text-brand-600 hover:text-brand-700 mb-4 inline-block">
        &larr; Back to tenders
      </a>

      <div className="bg-white rounded-lg border border-gray-200 p-8">
        <div className="flex justify-between items-start mb-6">
          <h1 className="text-2xl font-bold text-gray-900 flex-1 mr-4">{tender.title}</h1>
          <span className="inline-flex items-center px-3 py-1 rounded-full text-sm font-medium bg-blue-100 text-blue-800">
            {tender.source.toUpperCase()}
          </span>
        </div>

        {tender.description && (
          <div className="mb-6">
            <h2 className="text-sm font-medium text-gray-500 uppercase tracking-wider mb-2">Description</h2>
            <p className="text-gray-700 whitespace-pre-wrap">{tender.description}</p>
          </div>
        )}

        <div className="grid grid-cols-1 md:grid-cols-2 gap-6 mb-6">
          <div>
            <h2 className="text-sm font-medium text-gray-500 uppercase tracking-wider mb-2">Details</h2>
            <dl className="space-y-2">
              <div>
                <dt className="text-sm text-gray-500">Estimated Value</dt>
                <dd className="text-lg font-semibold">{formatCurrency(tender.estimated_value, tender.currency)}</dd>
              </div>
              <div>
                <dt className="text-sm text-gray-500">Procurement Method</dt>
                <dd className="font-medium">{tender.procurement_method || 'Not specified'}</dd>
              </div>
              <div>
                <dt className="text-sm text-gray-500">Source ID</dt>
                <dd className="font-mono text-sm">{tender.source_id}</dd>
              </div>
            </dl>
          </div>

          <div>
            <h2 className="text-sm font-medium text-gray-500 uppercase tracking-wider mb-2">Timeline & Location</h2>
            <dl className="space-y-2">
              <div>
                <dt className="text-sm text-gray-500">Deadline</dt>
                <dd className="font-medium">{formatDate(tender.deadline)}</dd>
              </div>
              <div>
                <dt className="text-sm text-gray-500">Published</dt>
                <dd className="font-medium">{formatDate(tender.publication_date)}</dd>
              </div>
              <div>
                <dt className="text-sm text-gray-500">Buyer</dt>
                <dd className="font-medium">{tender.buyer_name || 'Not specified'}</dd>
              </div>
              <div>
                <dt className="text-sm text-gray-500">Country</dt>
                <dd className="font-medium">{tender.buyer_country || 'Not specified'}</dd>
              </div>
              {tender.place_of_performance && (
                <div>
                  <dt className="text-sm text-gray-500">Place of Performance</dt>
                  <dd className="font-medium">{tender.place_of_performance}</dd>
                </div>
              )}
            </dl>
          </div>
        </div>

        {tender.cpv_codes && tender.cpv_codes.length > 0 && (
          <div>
            <h2 className="text-sm font-medium text-gray-500 uppercase tracking-wider mb-2">CPV Codes</h2>
            <div className="flex flex-wrap gap-2">
              {tender.cpv_codes.map((code) => (
                <span
                  key={code}
                  className="inline-flex items-center px-3 py-1 rounded-full text-sm font-medium bg-gray-100 text-gray-700"
                >
                  {code}
                </span>
              ))}
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
