package com.example.auth.entities;

import java.util.Collection;
import java.util.List;

import org.springframework.security.core.GrantedAuthority;
import org.springframework.security.core.userdetails.UserDetails;
import jakarta.persistence.*;

/**
 * This class represents a user entity for authentication purposes.
 * User entities are stored to the DB in the table called "users".
 * Implements Spring Security's UserDetails to integrate with
 * Spring's auth systems.
 */
@Table(name = "users")
@Entity
public class User implements UserDetails {
	@Id
	@GeneratedValue(strategy = GenerationType.AUTO)
	@Column(nullable = false)
	private Integer id;

	@Column(unique = true, length = 30, nullable = false)
	private String username;

	@Column(nullable = false)
	private String password;

	/**
	 * Return an empty list because not implementing roles at this point
	 */
	@Override
	public Collection<? extends GrantedAuthority> getAuthorities() {
		return List.of();
	}

	@Override
	public String getUsername() {
		return username;
	}

	@Override
	public String getPassword() {
		return password;
	}

}
