package com.example.auth.services;

import java.security.Key;
import java.util.Date;

import io.jsonwebtoken.Claims;
import io.jsonwebtoken.Jwts;
import io.jsonwebtoken.io.Encoders;
import io.jsonwebtoken.security.Keys;

import org.springframework.beans.factory.annotation.Value;
import org.springframework.security.core.userdetails.UserDetails;
import org.springframework.stereotype.Service;

/**
 * Service responsible for the generation of JSON Web Tokens and extracting information from them.
 * Uses a secret key read from application.properties for token signing.
 */
@Service
public class JwtService {
	@Value("${security.jwt.issuer}")
	private String issuer;

	@Value("${security.jwt.secret-key}")
	private String secretKey;

	@Value("${security.jwt.expiration}")
	private long jwtExpiration;

	/**
	 * Generates a JWT token in String format for the given userDetails.
	 * @param userDetails
	 * @return Signed JWT as a string
	 */
	public String generateToken(UserDetails userDetails) {
		return Jwts
			.builder()
			.issuer(issuer)
			.subject(userDetails.getUsername())
			.issuedAt(new Date(System.currentTimeMillis()))
			.expiration(new Date(System.currentTimeMillis() + jwtExpiration))
			.signWith(GetSignInKey())
			.compact();
	}

	/**
	 * Extracts username from JWT.
	 * @param token
	 * @return username.
	 */
	public String extractUsernameFromToken(String token) {
		return extractAlllClaimsFromToken(token).getSubject();
	}

	/**
	 * Extracts all claims from given JWT.
	 * @param token
	 * @return Claims object containing all claims found in the given token.
	 */
	private Claims extractAlllClaimsFromToken(String token) {
		return Jwts
			.parser()
			.build()
			.parseSignedClaims(token)
			.getPayload();
	}

	/**
	 * Converts secretKey into a Key object.
	 * @return HMAC-SHA key derived from the secretKey.
	 */
	private Key GetSignInKey() {
		String encodedKey = Encoders.BASE64.encode(secretKey.getBytes());
		return Keys.hmacShaKeyFor(encodedKey.getBytes());
	}
}
