import { Levee } from '@levee/sdk';
import { env } from '$env/dynamic/private';

/**
 * Levee client for CMS content (navigation menus, site settings, etc.)
 *
 * IMPORTANT: This is BUILD-TIME only for static sites!
 * - This file is in $lib/server/ so SvelteKit blocks client imports
 * - +layout.server.ts runs during `vite build` to fetch CMS data
 * - The fetched data gets embedded in static HTML
 * - The API key is NEVER sent to the client
 *
 * Set these in .env for local dev or CI/CD build environment:
 * - LEVEE_API_KEY: Your Levee API key
 * - LEVEE_BASE_URL: Optional, defaults to https://api.levee.sh
 */
const apiKey = env.LEVEE_API_KEY;
const baseUrl = env.LEVEE_BASE_URL || 'https://api.levee.sh';

if (!apiKey) {
	console.warn('LEVEE_API_KEY not set - CMS features will be disabled');
}

export const levee = apiKey ? new Levee(apiKey, baseUrl) : null;
