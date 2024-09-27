package com.example.auth.repositories;

import java.util.Optional;

import com.example.auth.entities.User;
import org.springframework.data.repository.CrudRepository;
import org.springframework.stereotype.Repository;

// Represents the data access layer for the User entity
@Repository
public interface UserRepository extends CrudRepository<User, Integer> {
	Optional<User> findByUsername(String username);
}
