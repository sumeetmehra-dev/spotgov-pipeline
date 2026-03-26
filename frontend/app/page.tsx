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

  const handleIngest = async () => {
    setIngesting(true);
    try {
      const res = await triggerIngestion();
      alert(`Ingestion complete: ${res.upserted} tenders upserted`);
      loadTenders(1);
      loadStats();
    } catch (err) {
      alert('Ingestion failed. Check the server logs.');
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
    <div>
      <div className="flex justify-between items-center mb-6">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Procurement Tenders</h1>
          {stats && (
            <p className="text-sm text-gray-500 mt-1">
              {stats.total_tenders} total tenders
              {stats.by_source.ted ? ` | TED: ${stats.by_source.ted}` : ''}
              {stats.by_source.dados ? ` | dados.gov.pt: ${stats.by_source.dados}` : ''}
            </p>
          )}
        </div>
        <button
          onClick={handleIngest}
          disabled={ingesting}
          className="bg-green-600 text-white px-4 py-2 rounded-lg text-sm font-medium hover:bg-green-700 disabled:opacity-50 transition-colors"
        >
          {ingesting ? 'Ingesting...' : 'Run Ingestion'}
        </button>
      </div>

      <div className="mb-6">
        <SearchBar onSearch={handleSearch} />
      </div>

      {searchMode && (
        <div className="mb-4 flex items-center justify-between">
          <p className="text-sm text-gray-600">{total} results found</p>
          <button
            onClick={() => { setPage(1); loadTenders(1); }}
            className="text-sm text-brand-600 hover:text-brand-700"
          >
            Clear search
          </button>
        </div>
      )}

      {loading ? (
        <div className="flex justify-center py-12">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-brand-600" />
        </div>
      ) : tenders.length === 0 ? (
        <div className="text-center py-12 text-gray-500">
          <p className="text-lg">No tenders found</p>
          <p className="text-sm mt-2">Click &quot;Run Ingestion&quot; to fetch tenders from TED and dados.gov.pt</p>
        </div>
      ) : (
        <div className="grid gap-4">
          {tenders.map((tender) => (
            <TenderCard key={tender.id} tender={tender} />
          ))}
        </div>
      )}

      {!searchMode && totalPages > 1 && (
        <div className="flex justify-center gap-2 mt-8">
          <button
            onClick={() => setPage((p) => Math.max(1, p - 1))}
            disabled={page === 1}
            className="px-4 py-2 rounded-lg border border-gray-300 text-sm disabled:opacity-50 hover:bg-gray-50"
          >
            Previous
          </button>
          <span className="px-4 py-2 text-sm text-gray-600">
            Page {page} of {totalPages}
          </span>
          <button
            onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
            disabled={page === totalPages}
            className="px-4 py-2 rounded-lg border border-gray-300 text-sm disabled:opacity-50 hover:bg-gray-50"
          >
            Next
          </button>
        </div>
      )}
    </div>
  );
}
