import os
import sys
import json
import pika
from typing import Dict


def main():
	"""
	This function consumes messages from RabbitMQ's mp3 queue. When messages are obtained from the queue,
	this function prints out a notification that include's a username and an mp3 file id.
	"""
	# RabbitMQ connection
	connection = pika.BlockingConnection(pika.ConnectionParameters(host='rabbitmq'))
	channel = connection.channel()

	def callback(ch, method, properties, body):
		err = notify(body)
		if err:
			ch.basic_nack(delivery_tag=method.delivery_tag) # Keep message in case of failure
		else:
			ch.basic_ack(delivery_tag=method.delivery_tag)

	channel.basic_consume(
		queue=os.environ.get('MP3_QUEUE'),
		on_message_callback=callback
	)

	print('Waiting for messages... To exit press CTRL+C')
	channel.start_consuming()


def notify(msg: Dict) -> Exception | None:
	"""
	This function prints out a notification that include's a username and an mp3 file id.
	The notification is created based on messages obtained from the RabbitMQ mp3 queue.

	Parameters
	- msg: RabbitMQ message body

	Return
	- e: Any exception that might have occured during notifying
	"""
	try:
		msg = json.loads(msg)
		print(f'Attention user {msg["username"]}! Your mp3 is ready for download.\nfid: {msg["mp3_fid"]}')
	except Exception as e:
		print(f'An error occured while trying to create a notification!\n{e}')
		return e


if __name__ == '__main__':
	try:
		main()
	except KeyboardInterrupt:
		print('Interrupted')
		try:
			sys.exit()
		except SystemExit:
			os._exit(0)
