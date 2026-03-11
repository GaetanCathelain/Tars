export interface User {
	id: string;
	username: string;
	created_at: string;
}

export interface Task {
	id: string;
	title: string;
	status: 'open' | 'running' | 'completed' | 'failed';
	created_by: string;
	created_at: string;
	updated_at: string;
}

export interface Message {
	id: string;
	task_id: string;
	sender_type: 'user' | 'tars' | 'system';
	sender_id?: string;
	content: string;
	created_at: string;
}

export interface WorkerSession {
	id: string;
	task_id: string;
	status: 'running' | 'completed' | 'failed';
	command: string;
	exit_code?: number;
	started_at: string;
	finished_at?: string;
}

export interface AuthResponse {
	token: string;
	user: User;
}
