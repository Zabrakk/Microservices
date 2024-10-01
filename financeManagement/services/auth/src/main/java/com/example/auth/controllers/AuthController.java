package com.example.auth.controllers;

import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

import com.example.auth.dtos.LoginUserDto;
import com.example.auth.dtos.RegisterUserDto;
import com.example.auth.entities.User;
import com.example.auth.responses.LoginResponse;
import com.example.auth.services.AuthService;
import com.example.auth.services.JwtService;

@RequestMapping("/auth")
@RestController
public class AuthController {
	private final JwtService jwtService;
	private final AuthService authService;

	public AuthController (
		JwtService jwtService,
		AuthService authService
	) {
		this.jwtService = jwtService;
		this.authService = authService;
	}

	@PostMapping("/register")
	public ResponseEntity<User> register(@RequestBody RegisterUserDto dto) {
		User newUser = authService.register(dto);
		return ResponseEntity.ok(newUser);
	}

	@PostMapping("/login")
	public ResponseEntity<LoginResponse> login(@RequestBody LoginUserDto dto) {
		User user = authService.login(dto);
		String token = jwtService.generateToken(user); // Why user ok here?
		LoginResponse resp = new LoginResponse().setToken(token); // TODO: Set expritaion
		return ResponseEntity.ok(resp);
	}
}
