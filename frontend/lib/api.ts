const API_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

async function fetchAPI<T>(path: string, options?: RequestInit): Promise<T> {
  const res = await fetch(`${API_URL}${path}`, {
    headers: { 'Content-Type': 'application/json' },
    ...options,
  });

  if (!res.ok) {
    const error = await res.json().catch(() => ({ error: res.statusText }));
    throw new Error(error.error || 'API request failed');
  }

  return res.json();
}

export interface Tender {
  id: string;
  source: string;
  source_id: string;
  title: string;
  description: string;
  cpv_codes: string[];
  buyer_name: string;
  buyer_country: string;
  place_of_performance: string;
  procurement_method: string;
  estimated_value: number | null;
  currency: string;
  deadline: string | null;
  publication_date: string | null;
  created_at: string;
}

export interface Company {
  id: string;
  name: string;
  description: string;
  cpv_codes: string[];
  keywords: string[];
  countries: string[];
  min_value: number | null;
  max_value: number | null;
}

export interface Match {
  id: string;
  company_id: string;
  tender_id: string;
  score: number;
  vector_score: number;
  bm25_score: number;
  cpv_overlap: number;
  tender: Tender;
}

export interface PaginatedResponse<T> {
  data: T[];
  page: number;
  page_size: number;
  total: number;
  total_pages: number;
}

export function getTenders(params?: Record<string, string>) {
  const query = params ? '?' + new URLSearchParams(params).toString() : '';
  return fetchAPI<PaginatedResponse<Tender>>(`/api/v1/tenders${query}`);
}

export function getTender(id: string) {
  return fetchAPI<Tender>(`/api/v1/tenders/${id}`);
}

export function searchTenders(q: string, limit = 20) {
  return fetchAPI<{ data: Tender[]; query: string; count: number }>(
    `/api/v1/tenders/search?q=${encodeURIComponent(q)}&limit=${limit}`
  );
}

export function triggerIngestion() {
  return fetchAPI<{ message: string; upserted: number }>('/api/v1/tenders/ingest', {
    method: 'POST',
  });
}

export function createCompany(data: Omit<Company, 'id'>) {
  return fetchAPI<Company>('/api/v1/companies', {
    method: 'POST',
    body: JSON.stringify(data),
  });
}

export function getCompany(id: string) {
  return fetchAPI<Company>(`/api/v1/companies/${id}`);
}

export function updateCompany(id: string, data: Partial<Company>) {
  return fetchAPI<Company>(`/api/v1/companies/${id}`, {
    method: 'PUT',
    body: JSON.stringify(data),
  });
}

export function getMatches(companyId: string, minScore = 0, limit = 20) {
  return fetchAPI<{ data: Match[]; count: number }>(
    `/api/v1/companies/${companyId}/matches?min_score=${minScore}&limit=${limit}`
  );
}

export function triggerMatching(companyId: string) {
  return fetchAPI<{ message: string; matches: number }>('/api/v1/match', {
    method: 'POST',
    body: JSON.stringify({ company_id: companyId }),
  });
}

export function getHealth() {
  return fetchAPI<{ status: string; database: string }>('/api/v1/health');
}

export function getStats() {
  return fetchAPI<{ total_tenders: number; by_source: Record<string, number> }>('/api/v1/stats');
}
