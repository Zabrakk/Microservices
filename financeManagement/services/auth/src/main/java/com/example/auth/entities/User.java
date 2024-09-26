package com.example.auth.entities;

import jakarta.persistence.*;

@Table(name = "users")
@Entity
public class User {
	@Id
	@GeneratedValue(strategy = GenerationType.AUTO)
	@Column(nullable = false)
	private Integer id;

	@Column(unique = true, length = 30, nullable = false)
	private String username;

	@Column(nullable = false)
	private String password;

	// TODO: GET AND SET?
}
