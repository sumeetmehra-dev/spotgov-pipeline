'use client';

import { useState, useEffect, useCallback } from 'react';
import type { Tender, PaginatedResponse } from '@/lib/api';
import { getTenders, searchTenders, triggerIngestion, getStats } from '@/lib/api';
import TenderCard from '@/components/TenderCard';
import SearchBar from '@/components/SearchBar';

export default function DashboardPage() {
  const [tenders, setTenders] = useState<Tender[]>([]);
  const [page, setPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(true);
  const [searchMode, setSearchMode] = useState(false);
  const [ingesting, setIngesting] = useState(false);
  const [stats, setStats] = useState<{ total_tenders: number; by_source: Record<string, number> } | null>(null);

  const loadTenders = useCallback(async (p: number) => {
    setLoading(true);
    try {
      const res = await getTenders({ page: String(p), page_size: '12' });
      setTenders(res.data || []);
      setTotalPages(res.total_pages);
      setTotal(res.total);
      setSearchMode(false);
    } catch (err) {
      console.error('Failed to load tenders:', err);
    } finally {
      setLoading(false);
    }
  }, []);

  const handleSearch = async (query: string) => {
    setLoading(true);
    try {
      const res = await searchTenders(query, 50);
      setTenders(res.data || []);
      setSearchMode(true);
      setTotal(res.count);
    } catch (err) {
      console.error('Search failed:', err);
    } finally {
      setLoading(false);
    }
  };

  const handleClearSearch = () => {
    setPage(1);
    loadTenders(1);
  };

  const handleIngest = async () => {
    setIngesting(true);
    try {
      const res = await triggerIngestion();
      loadTenders(1);
      loadStats();
    } catch (err) {
      console.error('Ingestion failed:', err);
    } finally {
      setIngesting(false);
    }
  };

  const loadStats = async () => {
    try {
      const s = await getStats();
      setStats(s);
    } catch {}
  };

  useEffect(() => {
    loadTenders(page);
    loadStats();
  }, [page, loadTenders]);

  return (
    <div className="animate-fade-in">
      {/* Header */}
      <div className="mb-8">
        <div className="flex justify-between items-start mb-6">
          <div>
            <h1 className="text-2xl font-bold text-gray-900 tracking-tight">Procurement Tenders</h1>
            <p className="text-sm text-gray-500 mt-1">
              AI-powered tender discovery from European procurement sources
            </p>
          </div>
          <button
            onClick={handleIngest}
            disabled={ingesting}
            className="btn-primary flex items-center gap-2"
          >
            {ingesting ? (
              <>
                <div className="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin" />
                Ingesting...
              </>
            ) : (
              <>
                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                  <polyline points="23 4 23 10 17 10" />
                  <path d="M20.49 15a9 9 0 1 1-2.12-9.36L23 10" />
                </svg>
                Run Ingestion
              </>
            )}
          </button>
        </div>

        {/* Stats bar */}
        {stats && (
          <div className="flex gap-3 mb-6">
            <div className="card px-4 py-2.5 flex items-center gap-2">
              <div className="w-2 h-2 rounded-full bg-brand-500" />
              <span className="text-sm font-semibold text-gray-900">{stats.total_tenders.toLocaleString()}</span>
              <span className="text-sm text-gray-500">total</span>
            </div>
            {stats.by_source.ted > 0 && (
              <div className="card px-4 py-2.5 flex items-center gap-2">
                <span className="badge badge-ted">TED</span>
                <span className="text-sm font-semibold text-gray-700">{stats.by_source.ted.toLocaleString()}</span>
              </div>
            )}
            {stats.by_source.dados > 0 && (
              <div className="card px-4 py-2.5 flex items-center gap-2">
                <span className="badge badge-dados">DADOS</span>
                <span className="text-sm font-semibold text-gray-700">{stats.by_source.dados.toLocaleString()}</span>
              </div>
            )}
          </div>
        )}

        {/* Search */}
        <SearchBar onSearch={handleSearch} onClear={handleClearSearch} />
      </div>

      {/* Search mode indicator */}
      {searchMode && (
        <div className="mb-5 flex items-center justify-between">
          <p className="text-sm text-gray-600">
            <span className="font-semibold text-gray-900">{total}</span> results found
          </p>
          <button onClick={handleClearSearch} className="text-sm font-medium text-brand-600 hover:text-brand-700 transition-colors">
            Clear search
          </button>
        </div>
      )}

      {/* Content */}
      {loading ? (
        <div className="flex flex-col items-center justify-center py-20 gap-3">
          <div className="w-8 h-8 border-3 border-gray-200 border-t-brand-600 rounded-full animate-spin" />
          <p className="text-sm text-gray-400">Loading tenders...</p>
        </div>
      ) : tenders.length === 0 ? (
        <div className="text-center py-20">
          <div className="w-16 h-16 mx-auto mb-4 rounded-2xl bg-gray-100 flex items-center justify-center">
            <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="#9ca3af" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round">
              <circle cx="11" cy="11" r="8" />
              <path d="m21 21-4.3-4.3" />
            </svg>
          </div>
          <p className="text-lg font-semibold text-gray-900">No tenders found</p>
          <p className="text-sm text-gray-500 mt-1">Click &quot;Run Ingestion&quot; to fetch tenders from TED and dados.gov.pt</p>
        </div>
      ) : (
        <div className="grid gap-3">
          {tenders.map((tender) => (
            <TenderCard key={tender.id} tender={tender} />
          ))}
        </div>
      )}

      {/* Pagination */}
      {!searchMode && totalPages > 1 && (
        <div className="flex items-center justify-center gap-2 mt-8">
          <button
            onClick={() => setPage((p) => Math.max(1, p - 1))}
            disabled={page === 1}
            className="btn-secondary px-3 py-2"
          >
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
              <polyline points="15 18 9 12 15 6" />
            </svg>
          </button>

          {/* Page numbers */}
          {Array.from({ length: Math.min(5, totalPages) }, (_, i) => {
            let pageNum: number;
            if (totalPages <= 5) {
              pageNum = i + 1;
            } else if (page <= 3) {
              pageNum = i + 1;
            } else if (page >= totalPages - 2) {
              pageNum = totalPages - 4 + i;
            } else {
              pageNum = page - 2 + i;
            }
            return (
              <button
                key={pageNum}
                onClick={() => setPage(pageNum)}
                className={`w-10 h-10 rounded-xl text-sm font-medium transition-all ${
                  page === pageNum
                    ? 'bg-brand-600 text-white shadow-sm'
                    : 'text-gray-600 hover:bg-gray-100'
                }`}
              >
                {pageNum}
              </button>
            );
          })}

          <button
            onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
            disabled={page === totalPages}
            className="btn-secondary px-3 py-2"
          >
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
              <polyline points="9 18 15 12 9 6" />
            </svg>
          </button>
        </div>
      )}
    </div>
  );
}
