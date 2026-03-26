'use client';

import { useState } from 'react';
import { createCompany, triggerMatching, getMatches } from '@/lib/api';
import type { Company, Match } from '@/lib/api';
import MatchScore from '@/components/MatchScore';

export default function ProfilePage() {
  const [company, setCompany] = useState<Company | null>(null);
  const [matches, setMatches] = useState<Match[]>([]);
  const [loading, setLoading] = useState(false);
  const [matching, setMatching] = useState(false);

  const [form, setForm] = useState({
    name: '',
    description: '',
    cpv_codes: '',
    keywords: '',
    countries: 'PT',
    min_value: '',
    max_value: '',
  });

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    try {
      const data = {
        name: form.name,
        description: form.description,
        cpv_codes: form.cpv_codes.split(',').map((s) => s.trim()).filter(Boolean),
        keywords: form.keywords.split(',').map((s) => s.trim()).filter(Boolean),
        countries: form.countries.split(',').map((s) => s.trim()).filter(Boolean),
        min_value: form.min_value ? Number(form.min_value) : null,
        max_value: form.max_value ? Number(form.max_value) : null,
      };
      const created = await createCompany(data);
      setCompany(created);
    } catch (err) {
      alert('Failed to create company profile');
    } finally {
      setLoading(false);
    }
  };

  const handleMatch = async () => {
    if (!company) return;
    setMatching(true);
    try {
      await triggerMatching(company.id);
      const res = await getMatches(company.id, 0, 50);
      setMatches(res.data || []);
    } catch (err) {
      alert('Matching failed. Ensure tenders have embeddings (run ingestion + embedding generation first).');
    } finally {
      setMatching(false);
    }
  };

  return (
    <div>
      <h1 className="text-2xl font-bold text-gray-900 mb-6">Company Profile</h1>

      {!company ? (
        <form onSubmit={handleSubmit} className="bg-white rounded-lg border border-gray-200 p-6 max-w-2xl">
          <div className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Company Name *</label>
              <input
                type="text"
                required
                value={form.name}
                onChange={(e) => setForm({ ...form, name: e.target.value })}
                className="w-full rounded-lg border border-gray-300 px-4 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-brand-500"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Description</label>
              <textarea
                value={form.description}
                onChange={(e) => setForm({ ...form, description: e.target.value })}
                rows={3}
                className="w-full rounded-lg border border-gray-300 px-4 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-brand-500"
                placeholder="Describe your company's services and expertise..."
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">CPV Codes (comma-separated)</label>
              <input
                type="text"
                value={form.cpv_codes}
                onChange={(e) => setForm({ ...form, cpv_codes: e.target.value })}
                className="w-full rounded-lg border border-gray-300 px-4 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-brand-500"
                placeholder="45000000, 72000000"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Keywords (comma-separated)</label>
              <input
                type="text"
                value={form.keywords}
                onChange={(e) => setForm({ ...form, keywords: e.target.value })}
                className="w-full rounded-lg border border-gray-300 px-4 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-brand-500"
                placeholder="construction, infrastructure, road building"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Target Countries (comma-separated)</label>
              <input
                type="text"
                value={form.countries}
                onChange={(e) => setForm({ ...form, countries: e.target.value })}
                className="w-full rounded-lg border border-gray-300 px-4 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-brand-500"
              />
            </div>
            <div className="grid grid-cols-2 gap-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Min Contract Value (EUR)</label>
                <input
                  type="number"
                  value={form.min_value}
                  onChange={(e) => setForm({ ...form, min_value: e.target.value })}
                  className="w-full rounded-lg border border-gray-300 px-4 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-brand-500"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Max Contract Value (EUR)</label>
                <input
                  type="number"
                  value={form.max_value}
                  onChange={(e) => setForm({ ...form, max_value: e.target.value })}
                  className="w-full rounded-lg border border-gray-300 px-4 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-brand-500"
                />
              </div>
            </div>
          </div>
          <button
            type="submit"
            disabled={loading}
            className="mt-6 bg-brand-600 text-white px-6 py-2.5 rounded-lg text-sm font-medium hover:bg-brand-700 disabled:opacity-50 transition-colors"
          >
            {loading ? 'Creating...' : 'Create Profile'}
          </button>
        </form>
      ) : (
        <div>
          <div className="bg-white rounded-lg border border-gray-200 p-6 mb-6">
            <div className="flex justify-between items-start">
              <div>
                <h2 className="text-xl font-semibold">{company.name}</h2>
                <p className="text-sm text-gray-600 mt-1">{company.description}</p>
                <div className="flex gap-2 mt-3">
                  {company.cpv_codes?.map((code) => (
                    <span key={code} className="px-2 py-0.5 rounded text-xs bg-gray-100 text-gray-700">
                      {code}
                    </span>
                  ))}
                </div>
              </div>
              <button
                onClick={handleMatch}
                disabled={matching}
                className="bg-brand-600 text-white px-4 py-2 rounded-lg text-sm font-medium hover:bg-brand-700 disabled:opacity-50 transition-colors"
              >
                {matching ? 'Matching...' : 'Find Matches'}
              </button>
            </div>
          </div>

          {matches.length > 0 && (
            <div>
              <h2 className="text-lg font-semibold mb-4">Matched Tenders ({matches.length})</h2>
              <div className="space-y-3">
                {matches.map((match) => (
                  <div key={match.id} className="bg-white rounded-lg border border-gray-200 p-4 flex items-center justify-between">
                    <div className="flex-1 mr-4">
                      <a href={`/tenders/${match.tender_id}`} className="font-medium text-gray-900 hover:text-brand-600">
                        {match.tender?.title || match.tender_id}
                      </a>
                      {match.tender?.buyer_name && (
                        <p className="text-sm text-gray-500 mt-0.5">{match.tender.buyer_name}</p>
                      )}
                    </div>
                    <MatchScore
                      score={match.score}
                      vectorScore={match.vector_score}
                      cpvOverlap={match.cpv_overlap}
                      showBreakdown
                    />
                  </div>
                ))}
              </div>
            </div>
          )}
        </div>
      )}
    </div>
  );
}
