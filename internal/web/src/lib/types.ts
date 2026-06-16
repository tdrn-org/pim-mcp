// Mirror of Go rest.SessionInfo, rest.CredentialInfo, rest.loginRequest structs
export interface CredentialInfo {
	valid: boolean;
	expiry: string; // ISO 8601 timestamp
}

export interface SessionInfo {
	provider_name: string;
	api_key: string;
	credentials: CredentialInfo;
}

export interface LoginRequest {
	api_key?: string;
}
