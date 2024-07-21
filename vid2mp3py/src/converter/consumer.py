import os
import sys
import time
import pika
import gridfs
from pymongo import MongoClient

from convert import to_mp3


def main():
	client = MongoClient('host.minikube.internal', 27017)
	db_videos = client.videos
	db_mp3s = client.mp3s
	fs_videos = gridfs.GridFS(db_videos)
	fs_mp3s = gridfs.GridFS(db_mp3s)

	connection = pika.BlockingConnection(
		pika.ConnectionParameters(host='rabbitmq') # Service name will resolve to host IP for RabbitMQ service
	)
	channel = connection.channel()

	def callback(ch, method, properties, body):
		err = to_mp3.start(body, fs_videos, fs_mp3s, ch)
		if err:
			ch.basic_nack(delivery_tag=method.delivery_tag) # Keep message in case of failure
		else:
			ch.basic_ack(delivery_tag=method.delivery_tag)

	channel.basic_consume(
		queue=os.getenv('VIDEO_QUEUE'),
		on_message_callback=callback
	)

	print('Waiting for messages... To exit press CTRL+C')
	channel.start_consuming()


if __name__ == '__main__':
	try:
		main()
	except KeyboardInterrupt:
		print('Interrupted')
		try:
			sys.exit()
		except SystemExit:
			os._exit(0)
