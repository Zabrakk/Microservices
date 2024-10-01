package com.example.auth.responses;

public class LoginResponse {
	private String token;
	private long expiresIn;

	public LoginResponse setToken(String token) {
		this.token = token;
		return this;
	}
}
