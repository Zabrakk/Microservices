import os
import sys
import json
import pika


def main():
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


def notify(msg):
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
