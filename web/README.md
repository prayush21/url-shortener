# URL Shortener Frontend

A modern React application for shortening URLs with a clean and user-friendly interface.

## Features

- Create short URLs from long ones
- Copy shortened URLs to clipboard
- List and manage your shortened URLs
- Delete URLs you no longer need
- Responsive design with Tailwind CSS
- TypeScript for better type safety

## Setup

1. Install dependencies:

   ```bash
   npm install
   ```

2. Create a `.env` file in the project root with:

   ```
   VITE_API_BASE_URL=http://localhost:8080
   ```

   Adjust the URL according to your backend API location.

3. Start the development server:

   ```bash
   npm run dev
   ```

4. Build for production:
   ```bash
   npm run build
   ```

## Project Structure

- `/src/components` - React components
- `/src/api` - API client and types
- `/src/assets` - Static assets

## Development

The project uses:

- Vite for fast development and building
- React 18 with TypeScript
- Tailwind CSS for styling
- Modern ES6+ features

## Future Improvements

- Add authentication
- Implement URL analytics
- Add categories/tags for URLs
- Support for custom short URLs
- Integration with social sharing
- Space for advertisements
