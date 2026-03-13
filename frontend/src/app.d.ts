// See https://svelte.dev/docs/kit/types#app.d.ts
// for information about these interfaces
declare global {
	namespace App {
		interface Locals {
			user: {
				id: number;
				username: string;
				email: string;
				avatar_url: string;
			} | null;
		}
	}
}

export {};
