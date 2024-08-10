import os
import jwt # JSON web token
import datetime
import MySQLdb
from typing import Tuple
from flask import Flask, request
from flask_mysqldb import MySQL

server = Flask(__name__)
my_sql = MySQL(server)

# Perform DB connection configuration
server.config['MYSQL_HOST'] 	= os.environ.get('MYSQL_HOST')
server.config['MYSQL_DB'] 		= os.environ.get('MYSQL_DB')
server.config['MYSQL_PORT'] 	= int(os.environ.get('MYSQL_PORT'))
server.config['MYSQL_USER'] 	= os.environ.get('MYSQL_USER')
server.config['MYSQL_PASSWORD'] = os.environ.get('MYSQL_PASSWORD')


@server.route('/login', methods=['POST'])
def login() -> Tuple[str, int]:
	"""
	Checks whether the received POST request's authorization headers include correct credentials
	for a user present in the MySQL DB. If correct credentials are provided, this function returns
	a JWT created based on the credentials. In all other cases an error message along with a status
	code is returned.

	Returns
	- token / error msg, status code
	"""
	# Ensure that the received request includes the authorization header
	auth = request.authorization
	if not auth:
		return "Credentials are missing", 401

	# Create DB cursor
	cursor = my_sql.connection.cursor()
	# Query the DB for the password of the user
	res = cursor.execute(
		"SELECT email, password FROM user WHERE email=%s", (auth.username,)
	)
	# Res is an array
	if res > 0:
		user_row = cursor.fetchone()
		email = user_row[0]
		password = user_row[1]

		if auth.username != email or auth.password != password:
			return 'Credentials were invalid', 401
		else:
			# Check whether the the user is an admin
			is_admin = email == os.getenv('MYSQL_ADMIN_USER') and password == os.getenv('MYSQL_ADMIN_PASSWORD')
			return createJWT(auth.username, os.environ.get('JWT_SECRET'), is_admin), 200
	else:
		return "Credentials were invalid", 401


def createJWT(username: str, jwt_secret: str, is_admin: bool) -> str:
	"""
	Creates a JSON Web Token with an expiration time of 1 day. Used algorithm is HS256.

	Parameters
	- username
	- jwt_secret: JWT secret that should be read from an environment variable
	- is_admin: User is admin. True or False

	Returns
	- str: JSON Web Token
	"""
	return jwt.encode(
		payload={
			'username': username,
			'exp': datetime.datetime.now(datetime.UTC) + datetime.timedelta(days=1),
			'iat': datetime.datetime.now(datetime.UTC),
			'admin': is_admin
		},
		key=jwt_secret,
		algorithm='HS256'
	)


@server.route('/validate', methods=['POST'])
def validate() -> Tuple[str, int]:
	"""
	Checks wheter a valid JSON Web Token is present in the received POST request.

	Returns
	- (str, int): JWT or error msg, status code
	"""
	encoded_jwt = request.headers['Authorization']
	if not encoded_jwt:
		return 'Credentials were invalid', 401

	# Bearer <token>
	encoded_jwt = encoded_jwt.split(' ')[1]
	try:
		decoded = jwt.decode(encoded_jwt, os.environ.get('JWT_SECRET'), algorithms='HS256')
	except:
		return "Not authorized", 403

	return decoded, 200


@server.route('/register', methods=['POST'])
def register() -> Tuple[str, int]:
	"""
	Attempts to register a new user based on the Username and Password included in the
	received POST request's headers. A JWT is returned after successful registrations.
	In all other cases, an error is returned.

	Returns
	- (str, int): JWT or error msg, status code
	"""
	username, password = request.headers['Username'], request.headers['Password']

	# Create DB cursor
	cursor = my_sql.connection.cursor()
	# Attempt to add the new users details to the DB
	try:
		cursor.execute(
			"INSERT INTO user (email, password) VALUES (%s, %s)", (username, password, )
		)
		my_sql.connection.commit()
		return createJWT(username, os.getenv('JWT_SECRET'), False)
	except MySQLdb.IntegrityError as e:
		print(f'Integrity error occured:\n{e}')
		if e.args[0] == 1062:
			# Code 1062 refers to a duplicate entry
			return f'A user has already been registered with the username {username}', 409
		return 'Internal server error', 500
	except Exception as e:
		print(f'Error occured while trying to register user:\n{e}')
		return 'Internal server error', 500


if __name__ == '__main__':
	server.run(host='0.0.0.0', port=5000)
