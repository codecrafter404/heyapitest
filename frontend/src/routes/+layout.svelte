<script lang="ts">
	import "./layout.css";
	import favicon from "$lib/assets/favicon.svg";
	import { client } from "$lib/client/client.gen";
	import { getApiUrl } from "$lib/config/config";
	import { login } from "$lib/client";

	let { children } = $props();
	client.setConfig({
		baseUrl: getApiUrl(),
	});
	login({ body: { Token: "super-secure-token" } });
</script>

<svelte:head><link rel="icon" href={favicon} /></svelte:head>
{@render children()}
