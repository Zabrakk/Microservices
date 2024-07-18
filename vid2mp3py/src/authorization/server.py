import os
import jwt # Json web token
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
		f'SELECT password FROM user WHERE email={auth.username}'
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
			return "Success", 200
