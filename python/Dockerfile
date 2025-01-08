FROM python:3.12-slim-bookworm


RUN mkdir /usr/src/script
RUN mkdir /usr/src/static
COPY run.sh /usr/src/script
RUN chmod +x /usr/src/script/run.sh
WORKDIR /usr/src/app
# set environment variables
ENV PYTHONDONTWRITEBYTECODE 1
ENV PYTHONUNBUFFERED 1

RUN pip install --upgrade pip
COPY ./requirements.txt .
RUN pip install -r requirements.txt

RUN addgroup --system hn && adduser --system --group hn
RUN chown -R hn:hn .
RUN chown -R hn:hn /usr/src/static

USER hn

COPY src .