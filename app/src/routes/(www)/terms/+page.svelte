<script lang="ts">
	import { setSEO } from '$lib/utils/seo';

	const seo = setSEO({
		title: 'Terms of Service - Gobot',
		description: 'Read the terms and conditions for using the Gobot brand analysis platform.',
		keywords: ['terms of service', 'terms and conditions', 'gobot terms'],
		url: '/terms'
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
			Terms of Service
		</h1>
		<p class="text-base-content/60 mb-12">Last updated: {lastUpdated}</p>

		<div
			class="prose prose-invert prose-headings:font-display prose-h2:text-xl prose-h2:font-bold max-w-none prose-headings:text-base-content prose-p:text-base-content/80 prose-li:text-base-content/80"
		>
			<h2>1. Acceptance of Terms</h2>
			<p>
				By accessing or using Gobot ("the Service"), you agree to be bound by these Terms of
				Service. If you do not agree to these terms, please do not use our Service.
			</p>

			<h2>2. Description of Service</h2>
			<p>
				Gobot provides AI-powered brand analysis tools that help marketers understand and optimize
				their marketing copy. The Service includes awareness stage detection, value proposition
				extraction, voice and tone analysis, and content transformation features.
			</p>

			<h2>3. User Accounts</h2>
			<p>To use certain features of the Service, you must create an account. You agree to:</p>
			<ul>
				<li>Provide accurate and complete registration information</li>
				<li>Maintain the security of your account credentials</li>
				<li>Notify us immediately of any unauthorized use of your account</li>
				<li>Be responsible for all activities that occur under your account</li>
			</ul>

			<h2>4. Acceptable Use</h2>
			<p>You agree not to:</p>
			<ul>
				<li>Use the Service for any illegal purpose</li>
				<li>Submit content that infringes on intellectual property rights</li>
				<li>Attempt to gain unauthorized access to any part of the Service</li>
				<li>Use the Service to transmit malware or harmful code</li>
				<li>Interfere with or disrupt the Service or servers</li>
				<li>Resell or redistribute the Service without authorization</li>
			</ul>

			<h2>5. Content Ownership</h2>
			<p>
				You retain ownership of any content you submit to the Service. By submitting content, you
				grant us a limited license to process and analyze that content to provide our services. We
				do not claim ownership of your marketing copy or analysis results.
			</p>

			<h2>6. AI-Generated Content</h2>
			<p>
				The Service uses artificial intelligence to analyze and transform content. While we strive
				for accuracy, AI-generated insights and suggestions are provided for informational purposes
				only. You are responsible for reviewing and validating any AI-generated content before use.
			</p>

			<h2>7. Payment Terms</h2>
			<p>
				Paid features are billed according to your selected plan. All fees are non-refundable unless
				otherwise specified. We reserve the right to modify pricing with reasonable notice.
			</p>

			<h2>8. Limitation of Liability</h2>
			<p>
				To the maximum extent permitted by law, Gobot shall not be liable for any indirect,
				incidental, special, consequential, or punitive damages resulting from your use of the
				Service.
			</p>

			<h2>9. Disclaimer of Warranties</h2>
			<p>
				The Service is provided "as is" without warranties of any kind, either express or implied.
				We do not guarantee that the Service will be uninterrupted, secure, or error-free.
			</p>

			<h2>10. Termination</h2>
			<p>
				We may terminate or suspend your account at any time for violation of these terms. You may
				terminate your account at any time by contacting us. Upon termination, your right to use the
				Service will immediately cease.
			</p>

			<h2>11. Changes to Terms</h2>
			<p>
				We reserve the right to modify these terms at any time. We will provide notice of
				significant changes. Your continued use of the Service after changes become effective
				constitutes acceptance of the new terms.
			</p>

			<h2>12. Governing Law</h2>
			<p>
				These Terms shall be governed by and construed in accordance with the laws of the United
				States, without regard to conflict of law provisions.
			</p>

			<h2>13. Contact</h2>
			<p>For questions about these Terms of Service, please contact us at legal@gobot.dev.</p>
		</div>
	</div>
</section>
