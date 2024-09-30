package com.example.auth.config;

import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.security.authentication.AuthenticationManager;
import org.springframework.security.authentication.AuthenticationProvider;
import org.springframework.security.authentication.dao.DaoAuthenticationProvider;
import org.springframework.security.config.annotation.authentication.configuration.AuthenticationConfiguration;
import org.springframework.security.core.userdetails.UserDetailsService;
import org.springframework.security.core.userdetails.UsernameNotFoundException;
import org.springframework.security.crypto.bcrypt.BCryptPasswordEncoder;

import com.example.auth.repositories.UserRepository;

/**
 * Configuration class defining Beans related to auth, password encoding and user details service.
 */
@Configuration
public class ApplicationConfiguration {
	private final UserRepository userRepository;

	/**
	 * Constructor
	 * @param userRepository UserRepository used to fetch user data for authentication.
	 */
	public ApplicationConfiguration(UserRepository userRepository) {
		this.userRepository = userRepository;
	}

	/**
	 * Bean definition for UserDetailsService, used to load user-specific data during authentication.
	 * @return Lambda function that fetches user by username from the repository.
	 * @throws UsernameNotFoundException
	 */
	@Bean
	UserDetailsService userDetailsService() {
		return username -> userRepository.findByUsername(username)
			.orElseThrow(() -> new UsernameNotFoundException("Username not found!"));
	}

	/**
	 * Bean definition for BCryptPasswordEncoder.
	 * @return New instance of BCryptPasswordEncoder.
	 */
	@Bean
	BCryptPasswordEncoder passwordEncoder() {
		return new BCryptPasswordEncoder();
	}

	/**
	 * Bean definition for AuthenticationManager, responsible for processing auth requests.
	 * @param authConfig AuthenticationConfiguration provided by Spring Security.
	 * @return Instance of AuthenticationManager.
	 * @throws Exception
	 */
	@Bean
	public AuthenticationManager authenticationManager(AuthenticationConfiguration authConfig) throws Exception {
		return authConfig.getAuthenticationManager();
	}

	/**
	 * Bean definition of AuthenticationProvider
	 * @return DaoAuthenticationProvider
	 */
	@Bean
	AuthenticationProvider authenticationProvider() {
		DaoAuthenticationProvider authProvider = new DaoAuthenticationProvider();
		authProvider.setUserDetailsService(userDetailsService());
		authProvider.setPasswordEncoder(passwordEncoder());
		return authProvider;
	}
}
