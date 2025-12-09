// API configuration
export const API_BASE_URL = import.meta.env.VITE_API_URL || '/api/v1';
export const API_REFETCH_INTERVAL = 60000; // 1 minute

// Future: This will be used when we integrate with the backend API
// import { client } from '@/api/client';
//
// client.setConfig({
//   baseUrl: API_BASE_URL,
// });
