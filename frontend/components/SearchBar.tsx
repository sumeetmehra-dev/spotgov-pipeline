'use client';

import { useState } from 'react';

interface SearchBarProps {
  onSearch: (query: string) => void;
  onClear?: () => void;
  placeholder?: string;
}

export default function SearchBar({ onSearch, onClear, placeholder = 'Search tenders by keyword, CPV code, or buyer...' }: SearchBarProps) {
  const [query, setQuery] = useState('');

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (query.trim()) {
      onSearch(query.trim());
    }
  };

  const handleClear = () => {
    setQuery('');
    onClear?.();
  };

  return (
    <form onSubmit={handleSubmit} className="relative">
      <div className="relative flex items-center">
        <div className="absolute left-4 text-gray-400">
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
            <circle cx="11" cy="11" r="8" />
            <path d="m21 21-4.3-4.3" />
          </svg>
        </div>
        <input
          type="text"
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          placeholder={placeholder}
          className="w-full rounded-2xl border-0 ring-1 ring-inset ring-gray-200 bg-white pl-12 pr-32 py-3.5 text-sm text-gray-900 placeholder:text-gray-400 focus:ring-2 focus:ring-brand-500 focus:outline-none shadow-card transition-all"
        />
        <div className="absolute right-2 flex gap-2">
          {query && (
            <button
              type="button"
              onClick={handleClear}
              className="px-3 py-1.5 text-xs text-gray-500 hover:text-gray-700 transition-colors"
            >
              Clear
            </button>
          )}
          <button
            type="submit"
            className="bg-brand-600 text-white px-5 py-2 rounded-xl text-sm font-semibold hover:bg-brand-700 transition-all shadow-sm"
          >
            Search
          </button>
        </div>
      </div>
    </form>
  );
}
