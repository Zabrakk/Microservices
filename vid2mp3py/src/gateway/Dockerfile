FROM python:3.12-slim-bullseye

RUN apt-get update \
	&& apt-get install -y --no-install-recommends --no-install-suggests \
	build-essential \
	&& pip install --cache-dir --upgrade pip

WORKDIR /app
COPY ./requirements.txt /app
RUN pip install --no-cache-dir -r /app/requirements.txt
COPY . /app

EXPOSE 8080

CMD ["python3", "server.py"]
