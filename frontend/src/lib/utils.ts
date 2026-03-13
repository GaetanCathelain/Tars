import { type ClassValue, clsx } from 'clsx';
import { twMerge } from 'tailwind-merge';
import type { Snippet } from 'svelte';

export function cn(...inputs: ClassValue[]) {
	return twMerge(clsx(inputs));
}

// Types used by shadcn-svelte components
export type WithElementRef<T, El extends HTMLElement = HTMLElement> = T & {
	ref?: El | null;
};

type PrimitiveProps = {
	children?: Snippet;
	child?: Snippet<[{ props: Record<string, unknown> }]>;
};

export type WithoutChildren<T> = Omit<T, 'children'>;
export type WithoutChild<T> = Omit<T, 'child'>;
export type WithoutChildrenOrChild<T> = Omit<T, 'children' | 'child'>;
