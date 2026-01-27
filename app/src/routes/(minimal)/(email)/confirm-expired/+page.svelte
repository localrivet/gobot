<script lang="ts">
	import { setSEO } from '$lib/utils/seo';

	const seo = setSEO({
		title: 'Link Expired - Gobot',
		description: 'This confirmation link has expired.',
		url: '/confirm-expired'
	});
</script>

<svelte:head>
	<title>{seo.title}</title>
	<meta name="description" content={seo.description} />
	<link rel="canonical" href={seo.canonical} />
</svelte:head>

<section class="max-w-2xl mx-auto px-4 sm:px-6 lg:px-8 py-20 text-center">
	<div class="animate-fade-in">
		<div class="mb-8">
			<svg
				class="w-16 h-16 mx-auto text-yellow-500"
				fill="none"
				viewBox="0 0 24 24"
				stroke="currentColor"
			>
				<path
					stroke-linecap="round"
					stroke-linejoin="round"
					stroke-width="2"
					d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z"
				/>
			</svg>
		</div>
		<h1 class="font-display text-3xl sm:text-4xl font-black text-white leading-tight mb-4">
			Link Expired
		</h1>
		<p class="text-base-400 text-lg mb-8">
			This confirmation link has expired or is no longer valid.
		</p>
		<p class="text-base-500 text-sm mb-8">
			Please request a new confirmation email from your account or contact support if you need
			assistance.
		</p>
		<a
			href="/"
			class="inline-flex items-center px-6 py-3 bg-base-700 hover:bg-base-600 text-white font-semibold rounded-lg transition-colors"
		>
			Return Home
		</a>
	</div>
</section>
