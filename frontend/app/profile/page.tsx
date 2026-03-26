'use client';

import { useState } from 'react';
import { createCompany, triggerMatching, getMatches, generateEmbeddings } from '@/lib/api';
import type { Company, Match } from '@/lib/api';
import MatchScore from '@/components/MatchScore';

function formatCurrency(value: number | null): string {
  if (value === null) return '-';
  return new Intl.NumberFormat('pt-PT', { style: 'currency', currency: 'EUR', maximumFractionDigits: 0 }).format(value);
}

export default function ProfilePage() {
  const [company, setCompany] = useState<Company | null>(null);
  const [matches, setMatches] = useState<Match[]>([]);
  const [loading, setLoading] = useState(false);
  const [matching, setMatching] = useState(false);
  const [matchStatus, setMatchStatus] = useState<string | null>(null);

  const [form, setForm] = useState({
    name: '',
    description: '',
    cpv_codes: '',
    keywords: '',
    countries: 'PRT',
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
      console.error('Failed to create company profile:', err);
    } finally {
      setLoading(false);
    }
  };

  const handleMatch = async () => {
    if (!company) return;
    setMatching(true);
    setMatchStatus('Generating embeddings...');
    try {
      // Generate any missing embeddings first
      try { await generateEmbeddings(); } catch {}

      setMatchStatus('Running AI matching...');
      await triggerMatching(company.id);

      setMatchStatus('Loading results...');
      const res = await getMatches(company.id, 0, 50);
      setMatches(res.data || []);
      setMatchStatus(null);
    } catch (err) {
      setMatchStatus('Matching failed. Ensure tenders have been ingested first.');
      setTimeout(() => setMatchStatus(null), 5000);
    } finally {
      setMatching(false);
    }
  };

  return (
    <div className="animate-fade-in">
      <div className="mb-8">
        <h1 className="text-2xl font-bold text-gray-900 tracking-tight">Company Profile</h1>
        <p className="text-sm text-gray-500 mt-1">
          Define your company to discover matching procurement opportunities
        </p>
      </div>

      {!company ? (
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
          {/* Form */}
          <form onSubmit={handleSubmit} className="lg:col-span-2 card p-6">
            <h2 className="text-base font-semibold text-gray-900 mb-5">Create your company profile</h2>
            <div className="space-y-5">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1.5">Company Name</label>
                <input
                  type="text"
                  required
                  value={form.name}
                  onChange={(e) => setForm({ ...form, name: e.target.value })}
                  className="input-field"
                  placeholder="e.g. TechLisbon Solutions"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1.5">Description</label>
                <textarea
                  value={form.description}
                  onChange={(e) => setForm({ ...form, description: e.target.value })}
                  rows={3}
                  className="input-field resize-none"
                  placeholder="Describe your company's services, expertise, and target markets..."
                />
              </div>
              <div className="grid grid-cols-1 md:grid-cols-2 gap-5">
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1.5">CPV Codes</label>
                  <input
                    type="text"
                    value={form.cpv_codes}
                    onChange={(e) => setForm({ ...form, cpv_codes: e.target.value })}
                    className="input-field"
                    placeholder="72000000, 72262000"
                  />
                  <p className="text-xs text-gray-400 mt-1">Comma-separated procurement classification codes</p>
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1.5">Keywords</label>
                  <input
                    type="text"
                    value={form.keywords}
                    onChange={(e) => setForm({ ...form, keywords: e.target.value })}
                    className="input-field"
                    placeholder="IT, software, consulting"
                  />
                  <p className="text-xs text-gray-400 mt-1">Terms that describe your capabilities</p>
                </div>
              </div>
              <div className="grid grid-cols-1 md:grid-cols-3 gap-5">
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1.5">Target Countries</label>
                  <input
                    type="text"
                    value={form.countries}
                    onChange={(e) => setForm({ ...form, countries: e.target.value })}
                    className="input-field"
                    placeholder="PRT"
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1.5">Min Value (EUR)</label>
                  <input
                    type="number"
                    value={form.min_value}
                    onChange={(e) => setForm({ ...form, min_value: e.target.value })}
                    className="input-field"
                    placeholder="50000"
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1.5">Max Value (EUR)</label>
                  <input
                    type="number"
                    value={form.max_value}
                    onChange={(e) => setForm({ ...form, max_value: e.target.value })}
                    className="input-field"
                    placeholder="5000000"
                  />
                </div>
              </div>
            </div>
            <button
              type="submit"
              disabled={loading}
              className="btn-primary mt-6 flex items-center gap-2"
            >
              {loading ? (
                <>
                  <div className="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin" />
                  Creating...
                </>
              ) : (
                'Create Profile'
              )}
            </button>
          </form>

          {/* Guide sidebar */}
          <div className="space-y-4">
            <div className="card p-5">
              <h3 className="text-sm font-semibold text-gray-900 mb-2">How it works</h3>
              <ol className="space-y-3 text-sm text-gray-600">
                <li className="flex gap-3">
                  <span className="flex items-center justify-center w-6 h-6 rounded-full bg-brand-50 text-brand-700 text-xs font-bold shrink-0">1</span>
                  <span>Create your company profile with CPV codes and keywords</span>
                </li>
                <li className="flex gap-3">
                  <span className="flex items-center justify-center w-6 h-6 rounded-full bg-brand-50 text-brand-700 text-xs font-bold shrink-0">2</span>
                  <span>AI generates embeddings for semantic understanding</span>
                </li>
                <li className="flex gap-3">
                  <span className="flex items-center justify-center w-6 h-6 rounded-full bg-brand-50 text-brand-700 text-xs font-bold shrink-0">3</span>
                  <span>Composite scoring ranks tenders by relevance</span>
                </li>
              </ol>
            </div>
            <div className="card p-5">
              <h3 className="text-sm font-semibold text-gray-900 mb-2">Scoring formula</h3>
              <div className="space-y-2 text-xs text-gray-500">
                <div className="flex justify-between"><span>Vector similarity</span><span className="font-mono text-gray-700">40%</span></div>
                <div className="flex justify-between"><span>BM25 text match</span><span className="font-mono text-gray-700">30%</span></div>
                <div className="flex justify-between"><span>CPV code overlap</span><span className="font-mono text-gray-700">20%</span></div>
                <div className="flex justify-between"><span>Value range fit</span><span className="font-mono text-gray-700">10%</span></div>
              </div>
            </div>
          </div>
        </div>
      ) : (
        <div className="space-y-6">
          {/* Company card */}
          <div className="card p-6">
            <div className="flex items-start justify-between">
              <div className="flex items-start gap-4">
                <div className="w-12 h-12 rounded-xl bg-brand-50 flex items-center justify-center shrink-0">
                  <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" className="text-brand-600">
                    <path d="M3 9l9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z" /><polyline points="9 22 9 12 15 12 15 22" />
                  </svg>
                </div>
                <div>
                  <h2 className="text-lg font-bold text-gray-900">{company.name}</h2>
                  <p className="text-sm text-gray-500 mt-1 max-w-xl">{company.description}</p>
                  <div className="flex flex-wrap gap-2 mt-3">
                    {company.cpv_codes?.map((code) => (
                      <span key={code} className="badge badge-ted">{code}</span>
                    ))}
                    {company.keywords?.map((kw) => (
                      <span key={kw} className="badge bg-gray-50 text-gray-600 ring-1 ring-inset ring-gray-200">{kw}</span>
                    ))}
                  </div>
                  {(company.min_value || company.max_value) && (
                    <p className="text-xs text-gray-400 mt-2">
                      Value range: {formatCurrency(company.min_value)} &ndash; {formatCurrency(company.max_value)}
                    </p>
                  )}
                </div>
              </div>
              <button
                onClick={handleMatch}
                disabled={matching}
                className="btn-primary flex items-center gap-2 shrink-0"
              >
                {matching ? (
                  <>
                    <div className="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin" />
                    Matching...
                  </>
                ) : (
                  <>
                    <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                      <polygon points="13 2 3 14 12 14 11 22 21 10 12 10 13 2" />
                    </svg>
                    Find Matches
                  </>
                )}
              </button>
            </div>
          </div>

          {/* Match status */}
          {matchStatus && (
            <div className="card p-4 flex items-center gap-3">
              <div className="w-5 h-5 border-2 border-brand-200 border-t-brand-600 rounded-full animate-spin shrink-0" />
              <span className="text-sm text-gray-600">{matchStatus}</span>
            </div>
          )}

          {/* Match results */}
          {matches.length > 0 && (
            <div>
              <div className="flex items-center justify-between mb-4">
                <h2 className="text-base font-semibold text-gray-900">
                  Matched Tenders
                  <span className="ml-2 text-sm font-normal text-gray-500">({matches.length} results)</span>
                </h2>
              </div>
              <div className="space-y-3">
                {matches.map((match, i) => (
                  <div
                    key={match.id}
                    className="card p-5 flex items-center gap-5 animate-slide-up"
                    style={{ animationDelay: `${i * 30}ms` }}
                  >
                    <MatchScore
                      score={match.score}
                      vectorScore={match.vector_score}
                      cpvOverlap={match.cpv_overlap}
                      showBreakdown
                    />
                    <div className="flex-1 min-w-0">
                      <a
                        href={`/tenders/${match.tender_id}`}
                        className="text-sm font-semibold text-gray-900 hover:text-brand-700 transition-colors line-clamp-1"
                      >
                        {match.tender?.title || match.tender_id}
                      </a>
                      <div className="flex items-center gap-3 mt-1 text-xs text-gray-500">
                        {match.tender?.buyer_name && <span>{match.tender.buyer_name}</span>}
                        {match.tender?.estimated_value && (
                          <>
                            <span className="text-gray-300">|</span>
                            <span className="font-semibold text-gray-700">
                              {formatCurrency(match.tender.estimated_value)}
                            </span>
                          </>
                        )}
                      </div>
                    </div>
                    <a
                      href={`/tenders/${match.tender_id}`}
                      className="btn-secondary px-3 py-1.5 text-xs shrink-0"
                    >
                      View
                    </a>
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
