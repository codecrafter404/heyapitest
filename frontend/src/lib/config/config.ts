import { PUBLIC_API_URL } from "$env/static/public";

export function getApiUrl(): string {
	if (typeof window != undefined && window.APP_CONFIG != null) {
		let url = window.APP_CONFIG.publicURL;
		if (!url.endsWith("/")) {
			url += "/"
		}
		return url + "api/"
	}
	return PUBLIC_API_URL
}
