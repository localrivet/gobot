<script lang="ts">
	import { setSEO } from '$lib/utils/seo';

	const seo = setSEO({
		title: 'Privacy Policy - Gobot',
		description: 'Learn how Gobot collects, uses, and protects your personal information.',
		keywords: ['privacy policy', 'data protection', 'gobot privacy'],
		url: '/privacy'
	});

	const lastUpdated = 'December 21, 2025';
</script>

<svelte:head>
	<title>{seo.title}</title>
	<meta name="description" content={seo.description} />
	<link rel="canonical" href={seo.canonical} />
</svelte:head>

<section class="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-20">
	<div class="animate-fade-in">
		<h1 class="font-display text-4xl sm:text-5xl font-black text-base-content leading-tight mb-4">
			Privacy Policy
		</h1>
		<p class="text-base-content/60 mb-12">Last updated: {lastUpdated}</p>

		<div
			class="prose prose-invert prose-headings:font-display prose-h2:text-xl prose-h2:font-bold prose-h3:text-base prose-h3:font-semibold max-w-none prose-headings:text-base-content prose-p:text-base-content/80 prose-li:text-base-content/80"
		>
			<h2>1. Introduction</h2>
			<p>
				Gobot ("we," "our," or "us") is committed to protecting your privacy. This Privacy Policy
				explains how we collect, use, disclose, and safeguard your information when you use our
				brand analysis service.
			</p>

			<h2>2. Information We Collect</h2>
			<h3>Account Information</h3>
			<p>
				When you create an account, we collect your email address and password (stored securely
				using industry-standard encryption).
			</p>
			<h3>Content You Submit</h3>
			<p>
				We collect the marketing copy and URLs you submit for analysis. This content is processed to
				provide our analysis services and may be stored to improve our service.
			</p>
			<h3>Usage Data</h3>
			<p>
				We automatically collect information about how you interact with our service, including
				pages viewed, features used, and analysis history.
			</p>

			<h2>3. How We Use Your Information</h2>
			<p>We use the information we collect to:</p>
			<ul>
				<li>Provide, maintain, and improve our brand analysis services</li>
				<li>Process your analyses and deliver results</li>
				<li>Communicate with you about your account and our services</li>
				<li>Ensure the security and integrity of our platform</li>
				<li>Comply with legal obligations</li>
			</ul>

			<h2>4. Data Sharing</h2>
			<p>We do not sell your personal information. We may share your information with:</p>
			<ul>
				<li>Service providers who assist in operating our platform</li>
				<li>Law enforcement when required by law</li>
				<li>Other parties with your consent</li>
			</ul>

			<h2>5. Data Security</h2>
			<p>
				We implement appropriate technical and organizational measures to protect your personal
				information against unauthorized access, alteration, disclosure, or destruction.
			</p>

			<h2>6. Your Rights</h2>
			<p>You have the right to:</p>
			<ul>
				<li>Access and receive a copy of your personal data</li>
				<li>Correct inaccurate personal data</li>
				<li>Request deletion of your personal data</li>
				<li>Object to processing of your personal data</li>
				<li>Export your data in a portable format</li>
			</ul>

			<h2>7. Cookies</h2>
			<p>
				We use essential cookies to maintain your session and provide our services. We may also use
				analytics cookies to understand how our service is used.
			</p>

			<h2>8. Changes to This Policy</h2>
			<p>
				We may update this Privacy Policy from time to time. We will notify you of any changes by
				posting the new Privacy Policy on this page and updating the "Last updated" date.
			</p>

			<h2>9. Contact Us</h2>
			<p>
				If you have questions about this Privacy Policy, please contact us at privacy@gobot.dev.
			</p>
		</div>
	</div>
</section>
