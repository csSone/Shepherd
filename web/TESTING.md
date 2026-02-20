# Testing

This directory contains the lightweight Vitest-based testing setup for the Shepherd web frontend (Vite).

What is tested
- Basic unit tests for config loading and API client utilities.
- Simple render helpers to wrap components in necessary providers.

How to run
- Install dependencies: npm install
- Run tests: npm test
- Coverage report: npm run test:coverage

Test structure
- src/lib/configLoader.ts
- src/lib/configLoader.test.ts
- src/lib/api/client.ts
- src/lib/api/client.test.ts
- src/test/setup.ts (global test setup)
- src/test/utils.tsx (render helper)

Notes
- Tests are designed to be minimal, non-intrusive, and not to rely on network calls.
- Coverage is enabled and reports both text and LCOV outputs.
