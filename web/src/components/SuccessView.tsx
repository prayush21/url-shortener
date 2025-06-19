import { useState } from "react";

export interface SuccessViewProps {
  originalUrl: string;
  shortUrl: string;
  onDismiss: () => void;
}

export function SuccessView({
  originalUrl,
  shortUrl,
  onDismiss,
}: SuccessViewProps) {
  const [copied, setCopied] = useState(false);

  const copyToClipboard = async () => {
    try {
      await navigator.clipboard.writeText(shortUrl);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    } catch (err) {
      console.error("Failed to copy:", err);
    }
  };

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4">
      <div className="bg-white rounded-lg shadow-xl p-6 max-w-lg w-full">
        <div className="text-center">
          <div className="mb-4">
            <div className="h-12 w-12 bg-green-100 rounded-full flex items-center justify-center mx-auto">
              <svg
                className="h-6 w-6 text-green-600"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth="2"
                  d="M5 13l4 4L19 7"
                />
              </svg>
            </div>
          </div>
          <h3 className="text-lg font-medium text-gray-900 mb-2">
            URL Successfully Shortened!
          </h3>
          <div className="mt-4 space-y-3">
            <div>
              <p className="text-sm text-gray-500 mb-1">Original URL:</p>
              <p className="text-gray-700 break-all">{originalUrl}</p>
            </div>
            <div>
              <p className="text-sm text-gray-500 mb-1">Shortened URL:</p>
              <p className="text-blue-600 break-all font-medium">{shortUrl}</p>
            </div>
          </div>
          <div className="mt-6 space-x-3">
            <button
              onClick={copyToClipboard}
              className="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md shadow-sm text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
            >
              {copied ? "Copied!" : "Copy Short URL"}
            </button>
            <button
              onClick={onDismiss}
              className="inline-flex items-center px-4 py-2 border border-gray-300 text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
            >
              Shorten Another Link
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}
