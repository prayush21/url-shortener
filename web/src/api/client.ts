const API_BASE_URL =
  import.meta.env.VITE_API_BASE_URL || "http://localhost:8080";

export interface CreateUrlResponse {
  short_key: string;
  url: string;
}

export interface UiUrlResponse {
  key: string;
  shortUrl: string;
  originalUrl: string;
  createdAt: string;
}

export async function createShortUrl(longUrl: string): Promise<UiUrlResponse> {
  const response = await fetch(`${API_BASE_URL}/api/v1/urls`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({ url: longUrl }),
  });

  if (!response.ok) {
    throw new Error("Failed to create short URL");
  }

  const apiResponse: CreateUrlResponse = await response.json();

  // Transform API response to UI format
  return {
    key: apiResponse.short_key,
    shortUrl: `${API_BASE_URL}/${apiResponse.short_key}`,
    originalUrl: apiResponse.url,
    createdAt: new Date().toISOString(),
  };
}

export async function deleteUrl(key: string): Promise<void> {
  const response = await fetch(`${API_BASE_URL}/api/v1/urls/${key}`, {
    method: "DELETE",
  });

  if (!response.ok && response.status !== 204) {
    throw new Error("Failed to delete URL");
  }
}
