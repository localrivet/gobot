export const prerender = true;

export async function GET() {
	const robotsTxt = `User-agent: *
Allow: /

Sitemap: https://gobot.dev/sitemap.xml`;

	return new Response(robotsTxt, {
		headers: {
			'Content-Type': 'text/plain'
		}
	});
}
