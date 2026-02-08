<script lang="ts">
	import { getGreeting } from "$lib/client";
	import { client } from "$lib/client/client.gen";
	import { getApiUrl } from "$lib/config/config";
	let res = getGreeting({ path: { user: "Anonymous" } });
</script>

<div>
	<p>Hello world</p>

	<p>URL: {getApiUrl()}</p>
	{#await res}
		<p>Loading...</p>
	{:then x}
		{#if x.error}
			<p>
				Error: {x.error.status}
				{x.error.title}: {x.error.detail}
			</p>
		{:else}
			<p>{x.data?.msg}</p>
		{/if}
		<p>{JSON.stringify(x)}</p>
	{:catch}
		<p>Critical Network Error</p>
	{/await}
</div>
