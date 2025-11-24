export interface User {
  id: string;
  email: string;
  startDate: string;
}

export interface AuthData {
  email: string;
  password: string;
}

export interface AuthResponse {
  token: string;
}

