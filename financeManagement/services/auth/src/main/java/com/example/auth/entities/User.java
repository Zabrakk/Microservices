package com.example.auth.entities;

import java.util.Collection;
import java.util.List;

import org.springframework.security.core.GrantedAuthority;
import org.springframework.security.core.userdetails.UserDetails;
import jakarta.persistence.*;

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

	@Override
	public Collection<? extends GrantedAuthority> getAuthorities() {
		// Return an empty list because not implementing roles at this moment
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
