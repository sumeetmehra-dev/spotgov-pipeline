import type { Metadata } from 'next';
import './globals.css';

export const metadata: Metadata = {
  title: 'SpotGov Pipeline',
  description: 'AI-powered procurement tender matching platform',
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en">
      <body>
        <nav className="bg-white border-b border-gray-200">
          <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
            <div className="flex justify-between h-16 items-center">
              <a href="/" className="text-xl font-bold text-brand-700">
                SpotGov Pipeline
              </a>
              <div className="flex space-x-8">
                <a href="/" className="text-gray-600 hover:text-gray-900 font-medium">
                  Tenders
                </a>
                <a href="/profile" className="text-gray-600 hover:text-gray-900 font-medium">
                  Company Profile
                </a>
              </div>
            </div>
          </div>
        </nav>
        <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">{children}</main>
      </body>
    </html>
  );
}
