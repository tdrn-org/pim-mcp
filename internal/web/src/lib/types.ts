// Mirror of Go rest.SessionInfo, rest.CredentialInfo, rest.loginRequest structs
export interface CredentialInfo {
	valid: boolean;
}

export interface SessionInfo {
	provider_name: string;
	api_key: string;
	credentials: CredentialInfo;
}

export interface LoginRequest {
	api_key?: string;
}
