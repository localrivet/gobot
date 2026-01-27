import { levee } from '$lib/server/levee';
import type { SDKSiteSettings, SDKNavigationMenu } from '@levee/sdk';

export interface LayoutServerData {
	site: SDKSiteSettings | null;
	headerMenu: SDKNavigationMenu | null;
	footerMenu: SDKNavigationMenu | null;
}

export async function load(): Promise<LayoutServerData> {
	if (!levee) {
		return {
			site: null,
			headerMenu: null,
			footerMenu: null
		};
	}

	try {
		const [site, menus] = await Promise.all([
			levee.site.getSiteSettings(),
			levee.site.listNavigationMenus()
		]);

		const headerMenu = menus.menus.find((m: SDKNavigationMenu) => m.location === 'header') ?? null;
		const footerMenu = menus.menus.find((m: SDKNavigationMenu) => m.location === 'footer') ?? null;

		return {
			site,
			headerMenu,
			footerMenu
		};
	} catch (error) {
		console.error('Failed to load site data from Levee:', error);
		return {
			site: null,
			headerMenu: null,
			footerMenu: null
		};
	}
}
