package com.example.auth.config;

import java.util.List;

import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.security.authentication.AuthenticationProvider;
import org.springframework.security.config.annotation.web.builders.HttpSecurity;
import org.springframework.security.config.annotation.web.configuration.EnableWebSecurity;
import org.springframework.security.config.annotation.web.configurers.AbstractHttpConfigurer;
import org.springframework.security.config.http.SessionCreationPolicy;
import org.springframework.security.web.SecurityFilterChain;
import org.springframework.security.web.authentication.UsernamePasswordAuthenticationFilter;
import org.springframework.web.cors.CorsConfiguration;
import org.springframework.web.cors.CorsConfigurationSource;
import org.springframework.web.cors.UrlBasedCorsConfigurationSource;

@Configuration
@EnableWebSecurity
public class SecurityConfig {
	private final AuthenticationProvider authProvider;
	private final JwtAuthFilter jwtAuthFilter;

	public SecurityConfig(
		AuthenticationProvider authProvider,
		JwtAuthFilter jwtAuthFilter
	) {
		this.authProvider = authProvider;
		this.jwtAuthFilter = jwtAuthFilter;
	}

	@Bean
	public SecurityFilterChain securityFilterChain(HttpSecurity httpSec) throws Exception {
		httpSec.csrf(AbstractHttpConfigurer::disable);
		httpSec.authorizeHttpRequests((authorize) -> authorize
			.requestMatchers("/auth/**")
			.permitAll()
			.anyRequest()
			.authenticated()
		);
		httpSec.sessionManagement((manager) -> manager
			.sessionCreationPolicy(SessionCreationPolicy.STATELESS)
		);
		httpSec.authenticationProvider(authProvider);
		httpSec.addFilterBefore(jwtAuthFilter, UsernamePasswordAuthenticationFilter.class);
		return httpSec.build();
	}

	@Bean
	CorsConfigurationSource corsConfigurationSource() {
		CorsConfiguration corsConfig = new CorsConfiguration();
		corsConfig.setAllowedOrigins(List.of("http://localhost:8005"));
		corsConfig.setAllowedMethods(List.of("GET", "POST"));
		corsConfig.setAllowedHeaders(List.of("Authorization", "Content-Type"));

		UrlBasedCorsConfigurationSource source = new UrlBasedCorsConfigurationSource();
		source.registerCorsConfiguration("/**", corsConfig);
		return source;
	}
}
