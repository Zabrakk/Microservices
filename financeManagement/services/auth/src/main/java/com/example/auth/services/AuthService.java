package com.example.auth.services;

import org.springframework.security.authentication.AuthenticationManager;
import org.springframework.security.authentication.UsernamePasswordAuthenticationToken;
import org.springframework.security.crypto.password.PasswordEncoder;
import org.springframework.stereotype.Service;

import com.example.auth.dtos.LoginUserDto;
import com.example.auth.dtos.RegisterUserDto;
import com.example.auth.entities.User;
import com.example.auth.repositories.UserRepository;

@Service
public class AuthService {
	private final UserRepository userRepository;
	private final PasswordEncoder passwordEncoder;
	private final AuthenticationManager authenticationManager;

	public AuthService(
		UserRepository userRepository,
		PasswordEncoder passwordEncoder,
		AuthenticationManager authenticationManager
	) {
		this.userRepository = userRepository;
		this.passwordEncoder = passwordEncoder;
		this.authenticationManager = authenticationManager;
	}

	public User register(RegisterUserDto dto) {
		User user = new User()
			.setUsername(dto.getUsername())
			.setPassword(passwordEncoder.encode(dto.getPassword()));
		return userRepository.save(user);
	}

	public User login(LoginUserDto dto) {
		authenticationManager.authenticate(
			new UsernamePasswordAuthenticationToken(
				dto.getUsername(), dto.getPassword()
			)
		);
		return userRepository.findByUsername(dto.getUsername()).orElseThrow();
	}
}
