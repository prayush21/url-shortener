import { useState } from "react";
import { UrlForm } from "./components/UrlForm";
import { SuccessView } from "./components/SuccessView";
import { createShortUrl } from "./api/client";
import type { UiUrlResponse } from "./api/client";

function App() {
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [successData, setSuccessData] = useState<UiUrlResponse | null>(null);

  const handleSubmit = async (url: string) => {
    setIsLoading(true);
    setError(null);
    try {
      const newUrl = await createShortUrl(url);
      setSuccessData(newUrl);
    } catch (err) {
      setError("Failed to create short URL");
      console.error(err);
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="min-h-screen w-full bg-gray-50 flex justify-center items-center">
      <div className="max-w-4xl mx-auto px-4 py-12">
        <header className="text-center mb-12">
          <h1 className="text-4xl font-bold text-gray-900 mb-4">
            URL Shortener
          </h1>
          <p className="text-lg text-gray-600">
            Shorten your long URLs into memorable links
          </p>
        </header>

        <div className="space-y-8">
          <UrlForm onSubmit={handleSubmit} isLoading={isLoading} />

          {error && (
            <div className="w-full p-4 bg-red-50 border border-red-200 rounded-lg text-red-700 text-center">
              {error}
            </div>
          )}
        </div>

        {successData && (
          <SuccessView
            originalUrl={successData.originalUrl}
            shortUrl={successData.shortUrl}
            onDismiss={() => setSuccessData(null)}
          />
        )}

        <footer className="mt-16 text-center text-gray-500 text-sm">
          <p>
            © {new Date().getFullYear()} URL Shortener. All rights reserved.
          </p>
        </footer>
      </div>
    </div>
  );
}

export default App;
