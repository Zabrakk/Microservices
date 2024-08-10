import os
import requests
from typing import Tuple
from flask import Request


def login(request: Request) -> Tuple[str, Tuple[str, int]]:
	"""
	Sends a POST request to the authorization microservice's /login route
	and checks if the requester is authorized to access provided services.

	returns: Token, error (msg, status code)
	"""
	auth = request.authorization
	if not auth:
		return None, ('Missing credentials', 401)

	basic_auth = (auth.username, auth.password)
	response = requests.post(
		f'http://{os.environ.get("AUTH_SVC_ADDRESS")}/login',
		auth=basic_auth
	)

	if response.status_code == 200:
		return response.text, None
	return None, (response.text, response.status_code)


def register_user(request: Request) -> Tuple[str, Tuple[str, int]]:
	"""
	Sends a POST request to the authorization microservice's /register route
	and which performs the registration of a new user.

	returns: Token, error (msg, status code)
	"""
	if 'Username' not in request.headers:
		return None, ('Username missing from headers', 400)
	if 'Password' not in request.headers:
		return None, ('Password missing from headers', 400)

	response = requests.post(
		f'http://{os.environ.get("AUTH_SVC_ADDRESS")}/register',
		headers=request.headers
	)

	if response.status_code == 200:
		return response.text, None
	return None, (response.text, response.status_code)
