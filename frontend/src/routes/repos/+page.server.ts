import type { PageServerLoad } from './$types';

export interface Repo {
	id: number;
	name: string;
	url: string;
	local_path: string;
	default_branch: string;
	created_at: string;
}

export const load: PageServerLoad = async ({ fetch, cookies }) => {
	try {
		const res = await fetch('http://localhost:8080/api/v1/repos', {
			headers: {
				Cookie: `tars_session=${cookies.get('tars_session')}`
			}
		});

		if (!res.ok) {
			return { repos: [] as Repo[] };
		}

		const json = await res.json();
		return { repos: (json.data ?? []) as Repo[] };
	} catch {
		return { repos: [] as Repo[] };
	}
};
