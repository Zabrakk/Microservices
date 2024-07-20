import os
import requests
from typing import Tuple


def login(request) -> Tuple[str, Tuple[str, int]]:
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
	else:
		return None, (response.text, response.status_code)
