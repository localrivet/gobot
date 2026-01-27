/**
 * Pricing Page Server Load
 *
 * Reads pricing directly from gobot.yaml at build time.
 * This runs during static site generation, so the YAML is embedded
 * into the built HTML - no API required.
 */

import { readFileSync } from 'fs';
import { join } from 'path';
import yaml from 'js-yaml';

export const prerender = true;

// Types
export interface PricingPrice {
	id: string;
	nickname: string;
	displayName: string;
	unitAmountCents: number;
	currency: string;
	interval: string | undefined;
	intervalCount: number;
	trialDays: number;
	highlighted: boolean;
	displayOrder: number;
	active: boolean;
}

export interface PricingProduct {
	id: string;
	name: string;
	description: string;
	tagline: string;
	features: string[];
	default?: boolean;
	prices: PricingPrice[];
}

export interface PricingPageData {
	products: PricingProduct[];
}

// Read gobot.yaml at build time
function loadPricingFromYaml(): PricingProduct[] {
	// process.cwd() is the app/ directory during build
	const yamlPath = join(process.cwd(), '..', 'etc', 'gobot.yaml');

	try {
		const yamlContent = readFileSync(yamlPath, 'utf8');
		const config = yaml.load(yamlContent) as { Products?: YamlProduct[] };
		const products = config.Products || [];

		return products.map((product) => ({
			id: product.slug,
			name: product.name,
			description: product.description || '',
			tagline: product.tagline || '',
			features: product.features || [],
			default: product.default || false,
			prices: (product.prices || []).map((price) => ({
				id: `${product.slug}-${price.slug}`,
				nickname: price.slug,
				displayName: `${product.name} ${formatInterval(price.interval)}`,
				unitAmountCents: price.amount,
				currency: price.currency || 'usd',
				interval: price.interval,
				intervalCount: 1,
				trialDays: price.trialDays || 0,
				highlighted: product.slug === 'pro' && price.interval === 'month',
				displayOrder: price.interval === 'month' ? 0 : 1,
				active: true
			}))
		}));
	} catch (error) {
		console.error('Failed to load pricing from gobot.yaml:', error);
		return [];
	}
}

function formatInterval(interval?: string): string {
	if (!interval) return '';
	return interval.charAt(0).toUpperCase() + interval.slice(1) + 'ly';
}

// YAML types
interface YamlPrice {
	slug: string;
	amount: number;
	currency?: string;
	interval?: string;
	trialDays?: number;
}

interface YamlProduct {
	slug: string;
	name: string;
	description?: string;
	tagline?: string;
	features?: string[];
	default?: boolean;
	prices?: YamlPrice[];
}

export function load(): PricingPageData {
	return {
		products: loadPricingFromYaml()
	};
}
