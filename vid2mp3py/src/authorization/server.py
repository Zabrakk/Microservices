import os
import jwt # JSON web token
import datetime
from flask import Flask, request
from flask_mysqldb import MySQL

server = Flask(__name__)
my_sql = MySQL(server)

server.config['MYSQL_HOST'] = 	  os.environ.get('MYSQL_HOST')
server.config['MYSQL_DB'] = 	  os.environ.get('MYSQL_DB')
server.config['MYSQL_PORT'] = 	  os.environ.get('MYSQL_PORT')
server.config['MYSQL_USER'] = 	  os.environ.get('MYSQL_USER')
server.config['MYSQL_PASSWORD'] = os.environ.get('MYSQL_PASSWORD')


@server.route('/login', methods=['POST'])
def login():
	# Ensure that the received request includes the authorization header
	auth = request.authorization
	print(auth)
	if not auth:
		return "Credentials are missing", 401

	# Create DB cursor
	cursor = my_sql.connection.cursor()
	# Query the DB for the password of the user
	res = cursor.execute(
		f'SELECT email, password FROM user WHERE email={auth.username}'
	)
	# Res is an array
	print(res)
	if res > 0:
		user_row = cursor.fetchone()
		email = user_row[0]
		password = user_row[1]

		print(user_row)

		if auth.username != email or auth.password != password:
			return 'Credentials were invalid', 401
		else:
			return createJWT(auth.username, os.environ.get('JWT_SECRET'), True)
	else:
		return "Credentials were invalid", 401


def createJWT(username, jwt_secret, is_admin):
	return jwt.encode(
		{
			'username': username,
			'exp': datetime.datetime.now(datetime.UTC) + datetime.timedelta(days=1),
			'iat': datetime.datetime.now(datetime.UTC),
			'admin': is_admin
		},
		jwt_secret,
		algorithm='HS256'
	)


@server.route('/validate', methods=['POST'])
def validate():
	encoded_jwt = request.headers['Authorization']
	if not encoded_jwt:
		return 'Credentials were invalid', 401

	# Bearer <token>
	encoded_jwt = encoded_jwt.split(' ')[1]
	try:
		decoded = jwt.decode(
			encoded_jwt, os.environ.get('JWT_SECRET'), algorithms='HS256'
		)
	except:
		return "Not authorized", 403

	return decoded, 200


if __name__ == '__main__':
	server.run(host='0.0.0.0', port=5000)
