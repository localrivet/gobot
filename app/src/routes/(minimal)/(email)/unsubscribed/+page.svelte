<script lang="ts">
	import { setSEO } from '$lib/utils/seo';

	const seo = setSEO({
		title: 'Unsubscribed - Gobot',
		description: 'You have been successfully unsubscribed from our emails.',
		url: '/unsubscribed'
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
				class="w-16 h-16 mx-auto text-green-500"
				fill="none"
				viewBox="0 0 24 24"
				stroke="currentColor"
			>
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
			</svg>
		</div>
		<h1 class="font-display text-3xl sm:text-4xl font-black text-white leading-tight mb-4">
			You've Been Unsubscribed
		</h1>
		<p class="text-base-400 text-lg mb-8">
			We're sorry to see you go. You will no longer receive emails from us.
		</p>
		<p class="text-base-500 text-sm">
			Changed your mind? You can always resubscribe from your account settings.
		</p>
	</div>
</section>
