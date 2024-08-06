import os
import json
import requests
from typing import Tuple


def validate_token(request) -> Tuple[bool, Tuple[str, int]]:
	"""
	Sends a POST request to the authorization microservice's /validate route
	and checks if requester's token is valid.

	returns: JWT, error (msg, status code)
	"""
	if not 'Authorization' in request.headers:
		return None, ('Missing credentials', 401)

	token = request.headers['Authorization']
	if not token:
		return None, ('Missing credentials', 401)

	response = requests.post(
		f'http://{os.environ.get("AUTH_SVC_ADDRESS")}/validate',
		headers={'Authorization': token}
	)

	if response.status_code == 200:
		return json.loads(response.text), None
	else:
		return None, (response.text, response.status_code)
