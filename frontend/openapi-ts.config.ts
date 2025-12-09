import { defineConfig } from '@hey-api/openapi-ts';

export default defineConfig({
  input: 'http://localhost:8080/api/v1/openapi.yaml',
  output: {
    path: 'src/api',
    format: 'prettier',
    lint: 'eslint',
  },
  plugins: ['@hey-api/client-fetch', '@tanstack/react-query', '@hey-api/sdk'],
});
