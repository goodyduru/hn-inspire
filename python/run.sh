#!/bin/sh

cd /usr/src/app

python manage.py makemigrations
python manage.py migrate
python manage.py createsuperuser --noinput
python manage.py collectstatic --noinput
gunicorn base.wsgi:application --bind 0.0.0.0:8000
