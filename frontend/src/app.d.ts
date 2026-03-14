import type { User } from '$shared/types/models';

// See https://kit.svelte.dev/docs/types#app
declare global {
	namespace App {
		interface Locals {
			user?: User;
		}
		interface PageData {
			user: User | null;
		}
	}
}

export {};
